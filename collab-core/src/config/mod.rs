use std::sync::LazyLock;

use anyhow::anyhow;
use config::{Environment, File};
use serde::Deserialize;

pub static SETTINGS: LazyLock<Settings> =
    LazyLock::new(|| Settings::new().expect("Failed to load settings"));

#[derive(Debug, Clone, Deserialize)]
pub struct Settings {
    logger: LoggerConfig,
    listener: ListenerConfig,
    s3: S3Config,
}

impl Settings {
    fn new() -> Result<Self, config::ConfigError> {
        let settings = config::Config::builder()
            .add_source(File::with_name("collab-core.toml").required(false))
            .add_source(File::with_name("config/collab-core.toml").required(false))
            .add_source(File::with_name("etc/collab-core.toml").required(false))
            .add_source(File::with_name("/etc/collab-core/collab-core.toml").required(false))
            .add_source(File::with_name("/etc/collab-core/config.toml").required(false))
            .add_source(File::with_name("~/.collab-core.toml").required(false))
            .add_source(File::with_name("~/.config/collab-core.toml").required(false))
            .add_source(
                Environment::with_prefix("COLLAB")
                    .try_parsing(true)
                    .separator("_")
                    .list_separator(" "),
            )
            .build()?;
        settings.try_deserialize()
    }

    pub fn logger(&self) -> LoggerConfig {
        self.logger.clone()
    }

    pub fn listener(&self) -> ListenerConfig {
        self.listener.clone()
    }

    pub fn s3(&self) -> S3Config {
        self.s3.clone()
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct LoggerConfig {
    pub level: String,
}

impl LoggerConfig {
    pub fn get_level(&self) -> Result<tracing::Level, anyhow::Error> {
        match self.level.to_lowercase().as_str() {
            "trace" => Ok(tracing::Level::TRACE),
            "debug" => Ok(tracing::Level::DEBUG),
            "info" => Ok(tracing::Level::INFO),
            "warn" => Ok(tracing::Level::WARN),
            "error" => Ok(tracing::Level::ERROR),
            _ => Err(anyhow!("Invalid log level: {}", self.level)),
        }
    }
}

#[derive(Debug, Clone, Deserialize)]
pub struct ListenerConfig {
    pub host: String,
    pub port: u16,
}

#[derive(Debug, Clone, Deserialize)]
pub struct S3Config {
    pub endpoint: String,
    pub region: String,
    pub bucket: String,
    pub access_key: String,
    pub secret_key: String,
}
