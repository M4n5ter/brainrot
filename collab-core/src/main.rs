use std::sync::Arc;

use anyhow::Result;
use collab_core::room::{Room, RoomManager};
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
    sync::{mpsc, Mutex},
};

#[tokio::main]
async fn main() -> Result<()> {
    let room_manager = Arc::new(Mutex::new(RoomManager::new()));

    let listner = TcpListener::bind("0.0.0.0:8999").await?;
    loop {
        let (stream, _) = listner.accept().await?;
        let io = TokioIo::new(stream);
        let room_manager = Arc::clone(&room_manager);
        tokio::spawn(async move {
            if let Err(err) = Builder::new()
                .serve_connection(
                    io,
                    service_fn(move |req| server_upgrade(req, Arc::clone(&room_manager))),
                )
                // https://github.com/hyperium/hyper/issues/1752
                .with_upgrades()
                .await
            {
                eprintln!("Error: {}", err);
            }
        });
    }
}

async fn server_upgrade(
    mut req: Request<Incoming>,
    room_manager: Arc<Mutex<RoomManager>>,
) -> Result<Response<Empty<Bytes>>> {
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

    tokio::spawn(async move {
        let ws = fut.await;
        match ws {
            Ok(ws) => {
                if let Err(err) = handle(
                    room_manager,
                    FragmentCollector::new(ws),
                    room_id,
                    connection_id,
                )
                .await
                {
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
    room_manager: Arc<Mutex<RoomManager>>,
    mut ws: FragmentCollector<S>,
    room_id: String,
    connection_id: String,
) -> Result<()>
where
    S: AsyncRead + AsyncWrite + Unpin,
{
    let (tx, mut rx) = mpsc::channel(1);

    // 获取房间并添加连接
    let room = {
        let mut manager = room_manager.lock().await;
        manager.get_or_create_room(room_id)
    };
    {
        room.lock()
            .await
            .add_connection(connection_id.to_owned(), tx)
            .await
    }

    let room_clone: Arc<Mutex<Room>> = Arc::clone(&room);
    let connection_id_clone = connection_id.to_owned();
    loop {
        tokio::select! {
            message = rx.recv() => {
                if let Some(message) = message {
                    ws.write_frame(Frame::binary(Payload::Owned(message))).await?;
                } else {
                    break
                }
            }
            frame = ws.read_frame() => {
                if let Ok(frame) = frame{
                    match frame.opcode {
                        fastwebsockets::OpCode::Close => break,
                        fastwebsockets::OpCode::Binary => {
                            let mut room = room_clone.lock().await;
                            room.broardcast(frame.payload.to_owned(), &connection_id_clone).await;
                        }
                        _ => {}
                    }
                }else {
                    break
                }
            }
        }
    }
    room.lock().await.remove_connection(connection_id).await;
    Ok(())
}
