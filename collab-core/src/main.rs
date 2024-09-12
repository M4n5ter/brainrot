use actix::{Actor, Arbiter, System};
use anyhow::Result;
use collab_core::{
    listener::websocket::WebSocketListenerActor, room::actor::RoomManagerActor,
    storage::actor::S3Actor,
};

fn main() -> Result<()> {
    let system_rt = tokio::runtime::Builder::new_multi_thread()
        .worker_threads(1)
        .enable_all()
        .build()?;

    let listener_rt = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()?;

    System::with_tokio_rt(|| system_rt).block_on(async move {
        S3Actor::start_default();
        RoomManagerActor::start_default();
        Arbiter::with_tokio_rt(|| listener_rt).spawn_fn(|| {
            WebSocketListenerActor::new("0.0.0.0:8999").start();
        });

        tokio::signal::ctrl_c().await.unwrap();
        println!("\nShutting down...");
    });
    Ok(())
}
