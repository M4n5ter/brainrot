use actix::Actor;
use anyhow::Result;
use collab_core::room::actor::{self, RoomManagerActor};
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

#[actix::main]
async fn main() -> Result<()> {
    RoomManagerActor::start_default();

    let listner = TcpListener::bind("0.0.0.0:8999").await?;
    loop {
        let (stream, _) = listner.accept().await?;
        let io = TokioIo::new(stream);
        actix::spawn(async move {
            if let Err(err) = Builder::new()
                .serve_connection(io, service_fn(server_upgrade))
                // https://github.com/hyperium/hyper/issues/1752
                .with_upgrades()
                .await
            {
                eprintln!("Error: {}", err);
            }
        });
    }
}

async fn server_upgrade(mut req: Request<Incoming>) -> Result<Response<Empty<Bytes>>> {
    if !is_upgrade_request(&req) {
        // unwrap is safe here because we know the assume "if any previously configured argument failed to parse or get converted to the internal representation." is false
        return Ok(Response::builder().status(400).body(Empty::new()).unwrap());
    }

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
        let ws = fut.await;
        match ws {
            Ok(ws) => {
                if let Err(err) = handle(FragmentCollector::new(ws), room_id, connection_id).await {
                    eprintln!("Error: {}", err);
                }
            }
            Err(err) => {
                eprintln!("Upgrade failed: {}", err);
            }
        }
    });

    Ok(response)
}

async fn handle<S>(
    mut ws: FragmentCollector<S>,
    room_id: String,
    connection_id: String,
) -> Result<()>
where
    S: AsyncRead + AsyncWrite + Unpin,
{
    let (msg_tx, ws_rx) = mpsc::channel(1);
    let (ws_tx, mut msg_rx) = mpsc::channel(1);
    actix::spawn(actor::message_handler(
        ws_rx,
        ws_tx.clone(),
        room_id.clone(),
        connection_id.clone(),
    ));

    loop {
        tokio::select! {
            message = msg_rx.recv() => {
                if let Some(message) = message {
                    ws.write_frame(Frame::binary(Payload::Owned(message.into()))).await?;
                } else {
                    break
                }
            }
            frame = ws.read_frame() => {
                if let Ok(frame) = frame{
                    match frame.opcode {
                        fastwebsockets::OpCode::Close => break,
                        fastwebsockets::OpCode::Binary => {
                            msg_tx.send(actor::GenericMessage::Binary(frame.payload.to_vec())).await?;
                        }
                        fastwebsockets::OpCode::Text => {
                            msg_tx.send(actor::GenericMessage::Text(String::from_utf8_lossy(&frame.payload).to_string())).await?;
                        }
                        _ => {}
                    }
                }else {
                    break
                }
            }
        }
    }
    Ok(())
}
