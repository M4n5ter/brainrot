use actix::{Arbiter, ArbiterService, SystemService};
use anyhow::{Context, Result};
use collab_core::{
    config::SETTINGS, listener::websocket::WebSocketListenerActor, room::actor::RoomManagerActor,
    storage::s3::S3Actor,
};
use tracing::info;

#[actix::main]
async fn main() -> Result<()> {
    let subscriber = tracing_subscriber::fmt::Subscriber::builder()
        .with_env_filter(
            tracing_subscriber::EnvFilter::from_default_env()
                .add_directive("loro=info".parse()?)
                .add_directive(SETTINGS.logger().get_level()?.into()),
        )
        .finish();
    tracing::subscriber::set_global_default(subscriber)?;

    let listener_rt = tokio::runtime::Builder::new_multi_thread()
        .worker_threads(1) // Maybe we need more threads in the future. Put `tokio::spawn` to use them.
        .enable_all()
        .build()
        .context("Failed to build listener runtime")?;

    let listener_arbiter = Arbiter::with_tokio_rt(|| listener_rt);

    S3Actor::start_service(&Arbiter::current());
    RoomManagerActor::start_service(&Arbiter::current());
    listener_arbiter.spawn_fn(|| {
        WebSocketListenerActor::start_service();
    });

    tokio::signal::ctrl_c()
        .await
        .expect("Failed to listen for Ctrl-C");
    info!("Shutting down");

    Ok(())
}
