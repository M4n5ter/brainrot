use actix::{Arbiter, ArbiterService, System, SystemService};
use anyhow::{Context, Result};
use collab_core::{
    config::SETTINGS, listener::websocket::WebSocketListenerActor, room::actor::RoomManagerActor,
    storage::s3::S3Actor,
};
use tracing::info;

fn main() -> Result<()> {
    let subscriber = tracing_subscriber::fmt::Subscriber::builder()
        .with_env_filter(
            tracing_subscriber::EnvFilter::from_default_env()
                .add_directive("loro=warn".parse()?)
                .add_directive(SETTINGS.logger().get_level()?.into()),
        )
        .finish();
    tracing::subscriber::set_global_default(subscriber)?;

    let listener_rt = match SETTINGS.listener().threads {
        0 => {
            info!("Listener uses the number of cores available to the system.");
            tokio::runtime::Builder::new_multi_thread()
                .enable_all()
                .build()
                .context("Failed to build listener runtime")?
        }
        1 => {
            info!("Listener running in single-threaded mode");
            tokio::runtime::Builder::new_current_thread()
                .thread_name("listener")
                .enable_all()
                .build()
                .context("Failed to build listener runtime")?
        }
        n => {
            info!("Listener running in multi-threaded mode with {} threads", n);
            tokio::runtime::Builder::new_multi_thread()
                .worker_threads(n)
                .enable_all()
                .build()
                .context("Failed to build listener runtime")?
        }
    };

    System::new().block_on(async move {
        S3Actor::start_service(&Arbiter::current());
        RoomManagerActor::start_service(&Arbiter::current());
        Arbiter::with_tokio_rt(|| listener_rt).spawn_fn(|| {
            WebSocketListenerActor::start_service();
        });

        tokio::signal::ctrl_c()
            .await
            .expect("Failed to listen for Ctrl-C");
        info!("Shutting down");
    });

    Ok(())
}
