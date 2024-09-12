use actix::{Actor, Arbiter, System};
use anyhow::{Context, Result};
use collab_core::{
    config::Settings, listener::websocket::WebSocketListenerActor, room::actor::RoomManagerActor,
    storage::actor::S3Actor,
};
use tracing::info;

fn main() -> Result<()> {
    let settings = Settings::new()?;

    let subscriber = tracing_subscriber::fmt::Subscriber::builder()
        .with_env_filter(tracing_subscriber::EnvFilter::from_default_env())
        .with_max_level(settings.logger.get_level()?)
        .finish();
    tracing::subscriber::set_global_default(subscriber)?;

    let system_rt = tokio::runtime::Builder::new_multi_thread()
        .worker_threads(1)
        .enable_all()
        .build()
        .context("Failed to build system runtime")?;

    let listener_rt = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .context("Failed to build listener runtime")?;

    System::with_tokio_rt(|| system_rt).block_on(async move {
        S3Actor::new(settings.s3).start();
        RoomManagerActor::start_default();
        Arbiter::with_tokio_rt(|| listener_rt).spawn_fn(|| {
            WebSocketListenerActor::new(settings.listener).start();
        });

        tokio::signal::ctrl_c()
            .await
            .expect("Failed to listen for Ctrl-C");
        info!("Shutting down");
    });
    Ok(())
}
