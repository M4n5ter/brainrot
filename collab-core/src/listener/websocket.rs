use std::sync::Arc;

use actix::prelude::*;
use anyhow::{anyhow, Error, Result};
use bytes::BytesMut;
use dev::Registry;
use fastwebsockets::{
    upgrade::{is_upgrade_request, upgrade},
    CloseCode, FragmentCollectorRead, Frame, Payload, WebSocketError, WebSocketWrite,
};
use http_body_util::Empty;
use hyper::{
    body::{Bytes, Incoming},
    server::conn::http1::Builder,
    service::service_fn,
    upgrade::Upgraded,
    Request, Response,
};
use hyper_util::rt::TokioIo;
use rustc_hash::FxHashMap;
use tokio::{
    io::{AsyncRead, AsyncWrite, ReadHalf, WriteHalf},
    net::TcpListener,
    sync::{mpsc, oneshot, Mutex},
};
use tracing::{error, info, trace};

use crate::{
    config::{ListenerConfig, SETTINGS},
    room::actor::{MessageChan, RoomManagerActor},
    GenericMessage,
};

type WebSocketConnectionAddr =
    Addr<WebsocketConnectionActor<ReadHalf<TokioIo<Upgraded>>, WriteHalf<TokioIo<Upgraded>>>>;

pub struct WebSocketListenerActor {
    connections: FxHashMap<String, WebSocketConnectionAddr>,
    config: ListenerConfig,
}

impl WebSocketListenerActor {
    pub fn new(config: ListenerConfig) -> Self {
        Self {
            connections: Default::default(),
            config,
        }
    }
}

impl Default for WebSocketListenerActor {
    fn default() -> Self {
        Self::new(SETTINGS.listener())
    }
}

impl Actor for WebSocketListenerActor {
    type Context = Context<Self>;

    fn started(&mut self, _: &mut Self::Context) {
        let addr = format!("{}:{}", self.config.host, self.config.port);

        actix::spawn(async move {
            info!("WebSocket listener started on {}", addr);
            match TcpListener::bind(addr.to_owned()).await {
                Ok(listener) => loop {
                    match listener.accept().await {
                        Ok((stream, _)) => {
                            let io = TokioIo::new(stream);
                            let _ = Builder::new()
                                .serve_connection(io, service_fn(handle_connection))
                                .with_upgrades()
                                .await;
                        }
                        Err(err) => error!("TCP accept error: {}", err),
                    }
                },
                Err(err) => error!("TCP bind error: {}", err),
            }
        });
    }
}

impl ArbiterService for WebSocketListenerActor {
    fn service_started(&mut self, ctx: &mut Context<Self>) {
        Registry::set(ctx.address());
    }
}
impl Supervised for WebSocketListenerActor {}

async fn handle_connection(mut req: Request<Incoming>) -> Result<Response<Empty<Bytes>>> {
    if is_upgrade_request(&req) {
        let mut room_id = None;
        let mut connection_id = None;
        if let Some(query) = req.uri().query() {
            for (key, value) in url::form_urlencoded::parse(query.as_bytes()) {
                if key == "room_id" {
                    room_id = Some(value.to_string());
                } else if key == "connection_id" {
                    connection_id = Some(value.to_string());
                }
            }
        } else {
            // unwrap is safe here.
            return Ok(Response::builder().status(400).body(Empty::new()).unwrap());
        }

        if room_id.is_none() || connection_id.is_none() {
            // unwrap is safe here.
            return Ok(Response::builder().status(400).body(Empty::new()).unwrap());
        }

        // we checked that room_id and connection_id are not None
        let room_id = room_id.unwrap();
        let connection_id = connection_id.unwrap();

        let (response, fut) = upgrade(&mut req)?;

        let listener_addr = WebSocketListenerActor::from_registry();
        actix::spawn(async move {
            match fut.await {
                Ok(mut ws) => {
                    ws.set_auto_close(false);
                    let (ws_read, ws_write) = ws.split(tokio::io::split);
                    let addr = WebsocketConnectionActor::new(
                        FragmentCollectorRead::new(ws_read),
                        ws_write,
                        room_id,
                        connection_id.clone(),
                    )
                    .start();
                    listener_addr.do_send(AddWebsocketConnection {
                        id: connection_id,
                        addr,
                    });
                }
                Err(err) => error!("WebSocket upgrade error: {}", err),
            }
        });

        Ok(response)
    } else {
        // unwrap is safe here.
        Ok(Response::builder().status(400).body(Empty::new()).unwrap())
    }
}

impl Handler<AddWebsocketConnection> for WebSocketListenerActor {
    type Result = ();

    fn handle(&mut self, msg: AddWebsocketConnection, _: &mut Self::Context) {
        self.connections.insert(msg.id, msg.addr);
    }
}

#[derive(Message)]
#[rtype(result = "()")]
pub struct AddWebsocketConnection {
    id: String,
    addr: WebSocketConnectionAddr,
}

impl Handler<RemoveWebsocketConnection> for WebSocketListenerActor {
    type Result = ();

    fn handle(&mut self, msg: RemoveWebsocketConnection, _: &mut Self::Context) {
        self.connections.remove(&msg.0);
    }
}

#[derive(Message)]
#[rtype(result = "()")]
pub struct RemoveWebsocketConnection(String);

struct WebsocketConnectionActor<R, W>
where
    R: AsyncRead + Unpin,
    W: AsyncWrite + Unpin,
{
    read_half: Arc<Mutex<FragmentCollectorRead<R>>>,
    write_half: Arc<Mutex<WebSocketWrite<W>>>,
    room_id: String,
    connection_id: String,
}

impl<R, W> WebsocketConnectionActor<R, W>
where
    R: AsyncRead + Unpin,
    W: AsyncWrite + Unpin,
{
    pub fn new(
        read_half: FragmentCollectorRead<R>,
        write_half: WebSocketWrite<W>,
        room_id: String,
        connection_id: String,
    ) -> Self {
        Self {
            read_half: Arc::new(Mutex::new(read_half)),
            write_half: Arc::new(Mutex::new(write_half)),
            room_id,
            connection_id,
        }
    }
}

impl<R, W> Actor for WebsocketConnectionActor<R, W>
where
    R: AsyncRead + Unpin + Send + 'static,
    W: AsyncWrite + Unpin + Send + 'static,
{
    type Context = Context<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        // `ws_tx` and `ws_rx` represent the channels for room actors to communicate with the websocket world.
        // `msg_tx` and `msg_rx` represent the channels for the websocket world to communicate with the room actors.
        // 即 `ws_xx` 是 room actor 用来和 websocket actor 通信的 channel，`msg_xx` 是 websocket actor 用来和 room actor 通信的 channel。
        let (msg_tx, ws_rx) = mpsc::channel(1);
        let (ws_tx, mut msg_rx) = mpsc::channel(1);

        RoomManagerActor::from_registry().do_send(MessageChan {
            room_id: self.room_id.to_owned(),
            connection_id: self.connection_id.to_owned(),
            tx: ws_tx,
            rx: ws_rx,
        });

        let addr1 = ctx.address();
        let addr2 = ctx.address();
        let addr3 = ctx.address();

        let (close_tx1, mut close_rx) = mpsc::channel(1);
        let close_tx2 = close_tx1.clone();
        ctx.spawn(
            async move {
                let addr = addr1.clone();
                while let Ok(read_frame_result) = addr.send(ReadFrame).await {
                    let msg = match read_frame_result {
                        Ok(msg) => msg,
                        Err(e) => {
                            // https://github.com/denoland/fastwebsockets/issues/64#issuecomment-2024122393
                            // ReadFrame 的消息处理函数中 read_frame() 不是 cancel safe 的。
                            // 当客户端浏览器刷新时，ctx 会因为 CloseConnection 的消息处理函数最终被 stop，此处的错误为 Unexpected EOF，暂时来看不需要处理。
                            match e.downcast::<WebSocketError>() {
                                Ok(WebSocketError::UnexpectedEOF) => break,
                                Ok(e) => {
                                    error!("Failed to read frame: {:?}", e);
                                    break;
                                }
                                Err(e) => {
                                    error!("Failed to read frame: {:?}", e);
                                    break;
                                }
                            };
                        }
                    };

                    if let Err(e) = msg_tx.send(msg).await {
                        error!("Failed to send message to room actors: {:?}", e);
                    }
                }
                let _ = close_tx1.send(()).await;
            }
            .into_actor(self),
        );
        ctx.spawn(
            async move {
                let addr = addr2.clone();
                while let Some(message) = msg_rx.recv().await {
                    match message {
                        GenericMessage::Binary(_) => {
                            addr.do_send(WriteFrame {
                                payload: BytesMut::from(message),
                                opcode: fastwebsockets::OpCode::Binary,
                            });
                        }
                        GenericMessage::Text(_) => {
                            addr.do_send(WriteFrame {
                                payload: BytesMut::from(message),
                                opcode: fastwebsockets::OpCode::Text,
                            });
                        }
                        GenericMessage::ConnectionClosed => {
                            break;
                        }
                    }
                }
                let _ = close_tx2.send(()).await;
            }
            .into_actor(self),
        );

        ctx.spawn(
            async move {
                let _ = close_rx.recv().await;
                addr3.do_send(CloseConnection);
            }
            .into_actor(self),
        );
    }

    fn stopped(&mut self, _: &mut Self::Context) {
        WebSocketListenerActor::from_registry()
            .do_send(RemoveWebsocketConnection(self.connection_id.clone()));
    }
}

impl<R, W> Handler<ReadFrame> for WebsocketConnectionActor<R, W>
where
    R: AsyncRead + Unpin + Send + 'static,
    W: AsyncWrite + Unpin + Send + 'static,
{
    type Result = ResponseActFuture<Self, Result<GenericMessage, Error>>;

    fn handle(&mut self, _: ReadFrame, _: &mut Self::Context) -> Self::Result {
        let read_half = Arc::clone(&self.read_half);
        let (tx, rx) = oneshot::channel();

        tokio::spawn(async move {
            let msg = match read_half
                .lock()
                .await
                .read_frame(&mut |frame| async move {
                    if frame.opcode == fastwebsockets::OpCode::Close {
                        return Err(fastwebsockets::WebSocketError::ConnectionClosed);
                    }
                    Ok(())
                })
                .await
            {
                Ok(frame) => match frame.opcode {
                    fastwebsockets::OpCode::Binary => Ok(GenericMessage::Binary(
                        Bytes::copy_from_slice(&frame.payload),
                    )),
                    fastwebsockets::OpCode::Text => Ok(GenericMessage::Text(
                        String::from_utf8_lossy(&frame.payload).to_string(),
                    )),
                    fastwebsockets::OpCode::Close => {
                        trace!("Received close frame");
                        Ok(GenericMessage::ConnectionClosed)
                    }
                    _ => Err(anyhow!("Unsupported opcode")),
                },
                Err(err) => Err(err.into()),
            };
            let _ = tx.send(msg);
        });

        Box::pin(
            async move {
                match rx.await {
                    Ok(msg) => msg,
                    Err(e) => Err(e.into()),
                }
            }
            .into_actor(self),
        )
    }
}

#[derive(Message)]
#[rtype(result = "Result<GenericMessage, Error>")]
struct ReadFrame;

impl<R, W> Handler<WriteFrame> for WebsocketConnectionActor<R, W>
where
    R: AsyncRead + Unpin + Send + 'static,
    W: AsyncWrite + Unpin + Send + 'static,
{
    type Result = ();

    fn handle(&mut self, msg: WriteFrame, _: &mut Self::Context) -> Self::Result {
        let write_half = Arc::clone(&self.write_half);
        tokio::spawn(async move {
            let mut write_half = write_half.lock().await;
            match msg.opcode {
                fastwebsockets::OpCode::Binary => {
                    if let Err(e) = write_half
                        .write_frame(Frame::binary(Payload::Bytes(msg.payload)))
                        .await
                    {
                        error!("Failed to write frame: {:?}", e);
                    }
                }
                fastwebsockets::OpCode::Text => {
                    if let Err(e) = write_half
                        .write_frame(Frame::text(Payload::Bytes(msg.payload)))
                        .await
                    {
                        error!("Failed to write frame: {:?}", e);
                    }
                }
                fastwebsockets::OpCode::Close => {
                    if let Err(e) = write_half
                        .write_frame(Frame::close(CloseCode::Normal.into(), &msg.payload))
                        .await
                    {
                        error!("Failed to write frame: {:?}", e);
                    }
                }
                _ => {
                    error!("Write frame: Unsupported opcode");
                }
            }
        });
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct WriteFrame {
    payload: BytesMut,
    opcode: fastwebsockets::OpCode,
}

impl<R, W> Handler<CloseConnection> for WebsocketConnectionActor<R, W>
where
    R: AsyncRead + Unpin + Send + 'static,
    W: AsyncWrite + Unpin + Send + 'static,
{
    type Result = ResponseActFuture<Self, ()>;

    fn handle(&mut self, _: CloseConnection, ctx: &mut Self::Context) -> Self::Result {
        let addr = ctx.address();
        Box::pin(
            async move {
                let reason = "server closed this connection";

                if let Err(e) = addr
                    .send(WriteFrame {
                        payload: reason.into(),
                        opcode: fastwebsockets::OpCode::Close,
                    })
                    .await
                {
                    error!("Failed to send close frame: {:?}", e);
                };
                addr.do_send(StopWebsocketConnection);
            }
            .into_actor(self),
        )
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct CloseConnection;

impl<R, W> Handler<StopWebsocketConnection> for WebsocketConnectionActor<R, W>
where
    R: AsyncRead + Unpin + Send + 'static,
    W: AsyncWrite + Unpin + Send + 'static,
{
    type Result = ();

    fn handle(&mut self, _: StopWebsocketConnection, ctx: &mut Self::Context) {
        ctx.stop();
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct StopWebsocketConnection;
