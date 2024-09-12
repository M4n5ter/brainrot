use actix::prelude::*;
use anyhow::Result;
use fastwebsockets::{
    upgrade::{is_upgrade_request, upgrade},
    FragmentCollector, Frame, Payload,
};
use http_body_util::Empty;
use hyper::{
    body::{Bytes, Incoming},
    server::conn::http1::Builder,
    service::service_fn,
    Request, Response,
};
use hyper_util::rt::TokioIo;
use tokio::{
    io::{AsyncRead, AsyncWrite},
    net::TcpListener,
    sync::mpsc,
};
use tracing::error;

use crate::{
    config::ListenerConfig,
    room::actor::{MessageChan, RoomManagerActor},
    GenericMessage,
};

pub struct WebSocketListenerActor {
    addr: String,
}

impl WebSocketListenerActor {
    pub fn new(config: ListenerConfig) -> Self {
        Self {
            addr: format!("{}:{}", config.host, config.port),
        }
    }
}

impl Actor for WebSocketListenerActor {
    type Context = Context<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        let addr = self.addr.clone();

        ctx.spawn(
            async move {
                match TcpListener::bind(addr.to_owned()).await {
                    Ok(listener) => loop {
                        match listener.accept().await {
                            Ok((stream, _)) => {
                                let io = TokioIo::new(stream);

                                actix::spawn(async move {
                                    Builder::new()
                                        .serve_connection(io, service_fn(handle_connection))
                                        .with_upgrades()
                                        .await
                                });
                            }
                            Err(err) => error!("Error: {}", err),
                        }
                    },
                    Err(err) => error!("Error: {}", err),
                }
            }
            .into_actor(self),
        );
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
            return Ok(Response::builder().status(400).body(Empty::new()).unwrap());
        }

        if room_id.is_none() || connection_id.is_none() {
            return Ok(Response::builder().status(400).body(Empty::new()).unwrap());
        }

        // we checked that room_id and connection_id are not None
        let room_id = room_id.unwrap();
        let connection_id = connection_id.unwrap();

        let (response, fut) = upgrade(&mut req)?;

        actix::spawn(async move {
            match fut.await {
                Ok(ws) => {
                    let ws = FragmentCollector::new(ws);
                    handle_ws(ws, room_id, connection_id)
                        .await
                        .unwrap_or_else(|err| {
                            error!("WebSocket handling error: {}", err);
                        })
                }
                Err(err) => error!("WebSocket upgrade error: {}", err),
            }
        });

        Ok(response)
    } else {
        Ok(Response::builder().status(400).body(Empty::new()).unwrap())
    }
}

async fn handle_ws<S>(
    mut ws: FragmentCollector<S>,
    room_id: String,
    connection_id: String,
) -> Result<()>
where
    S: AsyncRead + AsyncWrite + Unpin,
{
    let (msg_tx, ws_rx) = mpsc::channel(1);
    let (ws_tx, mut msg_rx) = mpsc::channel(1);

    RoomManagerActor::from_registry().do_send(MessageChan {
        room_id,
        connection_id,
        tx: ws_tx,
        rx: ws_rx,
    });

    loop {
        tokio::select! {
            Ok(frame) = ws.read_frame() => {
                match frame.opcode {
                    fastwebsockets::OpCode::Close => break,
                    fastwebsockets::OpCode::Binary => {
                        let payload = Bytes::copy_from_slice(&frame.payload);
                        msg_tx.send(GenericMessage::Binary(payload)).await?;
                    }
                    fastwebsockets::OpCode::Text => {
                        let payload = frame.payload.to_vec();
                        msg_tx
                            .send(GenericMessage::Text(String::from_utf8_lossy(&payload).to_string()))
                            .await?;
                    }
                    _ => {}
            }},
            message = msg_rx.recv() => {
                 match message {
                    Some(message) => ws.write_frame(Frame::binary(Payload::Bytes(message.into()))).await?,
                    None => break,
                }
            }
        }
    }
    Ok(())
}
