use std::{sync::Arc, time::Duration};

use actix::prelude::*;
use bytes::Bytes;
use opendal::{
    layers::LoggingLayer, raw::PresignedRequest, services::S3, Entry, Metadata, Operator,
};
use tracing::error;

use crate::config::S3Config;

#[derive(Debug)]
pub struct S3Actor {
    operator: Arc<Operator>,
}

impl S3Actor {
    pub fn new(config: S3Config) -> Self {
        let builder = S3::default()
            .endpoint(&config.endpoint)
            .access_key_id(&config.access_key)
            .secret_access_key(&config.secret_key)
            .region(&config.region)
            .bucket(&config.bucket);
        let op = {
            match Operator::new(builder) {
                Ok(opb) => opb.layer(LoggingLayer::default()).finish(),
                Err(e) => {
                    panic!("Failed to create S3 operator: {}", e);
                }
            }
        };
        S3Actor {
            operator: Arc::new(op),
        }
    }
}

impl Default for S3Actor {
    // This should not be used. Use `S3Actor::new()` instead.
    fn default() -> Self {
        panic!("S3Actor::default() should not be used. Use `S3Actor::new()` instead.");
    }
}

impl Actor for S3Actor {
    type Context = Context<Self>;
}

impl SystemService for S3Actor {}
impl Supervised for S3Actor {}

impl Handler<CreateDir> for S3Actor {
    type Result = ResponseActFuture<Self, Result<(), String>>;

    fn handle(&mut self, mut msg: CreateDir, _: &mut Self::Context) -> Self::Result {
        if !msg.path.ends_with('/') {
            msg.path.push('/');
        };

        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move {
                operator
                    .create_dir(&msg.path)
                    .await
                    .map_err(|e| e.to_string())
            }
            .into_actor(self),
        )
    }
}

#[derive(Message)]
#[rtype(result = "Result<(), String>")]
pub struct CreateDir {
    pub path: String,
}

impl Handler<ReadFile> for S3Actor {
    type Result = ResponseActFuture<Self, Result<Bytes, String>>;

    fn handle(&mut self, msg: ReadFile, _: &mut Self::Context) -> Self::Result {
        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move { operator.read(&msg.path).await.map_err(|e| e.to_string()) }
                .into_actor(self)
                .map(|res, _, _| match res {
                    Ok(data) => Ok(data.to_bytes()),
                    Err(e) => {
                        error!("Failed to read file: {}", e);
                        Err(e)
                    }
                }),
        )
    }
}

#[derive(Message)]
#[rtype(result = "Result<Bytes, String>")]
pub struct ReadFile {
    pub path: String,
}

impl Handler<WriteFile> for S3Actor {
    type Result = ResponseActFuture<Self, Result<(), String>>;

    fn handle(&mut self, msg: WriteFile, _: &mut Self::Context) -> Self::Result {
        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move {
                operator
                    .write(&msg.path, msg.data)
                    .await
                    .map_err(|e| e.to_string())
            }
            .into_actor(self),
        )
    }
}

#[derive(Message)]
#[rtype(result = "Result<(), String>")]
pub struct WriteFile {
    pub path: String,
    pub data: Bytes,
}

impl Handler<Delete> for S3Actor {
    type Result = ResponseActFuture<Self, Result<(), String>>;

    #[tracing::instrument]
    fn handle(&mut self, msg: Delete, _: &mut Self::Context) -> Self::Result {
        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move { operator.delete(&msg.path).await.map_err(|e| e.to_string()) }
                .into_actor(self),
        )
    }
}

#[derive(Message, Debug)]
#[rtype(result = "Result<(), String>")]
pub struct Delete {
    pub path: String,
}

impl Handler<Copy> for S3Actor {
    type Result = ResponseActFuture<Self, Result<(), String>>;

    fn handle(&mut self, msg: Copy, _: &mut Self::Context) -> Self::Result {
        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move {
                operator
                    .copy(&msg.src, &msg.dst)
                    .await
                    .map_err(|e| e.to_string())
            }
            .into_actor(self),
        )
    }
}

#[derive(Message)]
#[rtype(result = "Result<(), String>")]
pub struct Copy {
    pub src: String,
    pub dst: String,
}

impl Handler<List> for S3Actor {
    type Result = ResponseActFuture<Self, Result<Vec<Entry>, String>>;

    #[tracing::instrument]
    fn handle(&mut self, msg: List, _: &mut Self::Context) -> Self::Result {
        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move { operator.list(&msg.path).await.map_err(|e| e.to_string()) }
                .into_actor(self),
        )
    }
}

#[derive(Message, Debug)]
#[rtype(result = "Result<Vec<Entry>, String>")]
pub struct List {
    pub path: String,
}

impl Handler<Stat> for S3Actor {
    type Result = ResponseActFuture<Self, Result<Metadata, String>>;

    fn handle(&mut self, msg: Stat, _: &mut Self::Context) -> Self::Result {
        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move { operator.stat(&msg.path).await.map_err(|e| e.to_string()) }
                .into_actor(self),
        )
    }
}

#[derive(Message)]
#[rtype(result = "Result<Metadata, String>")]
pub struct Stat {
    pub path: String,
}

impl Handler<PresignRead> for S3Actor {
    type Result = ResponseActFuture<Self, Result<PresignedRequest, String>>;

    #[tracing::instrument]
    fn handle(&mut self, msg: PresignRead, _: &mut Self::Context) -> Self::Result {
        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move {
                operator
                    .presign_read(&msg.path, msg.expires)
                    .await
                    .map_err(|e| e.to_string())
            }
            .into_actor(self),
        )
    }
}

#[derive(Message, Debug)]
#[rtype(result = "Result<PresignedRequest, String>")]
pub struct PresignRead {
    pub path: String,
    pub expires: Duration,
}

impl Handler<PresignWrite> for S3Actor {
    type Result = ResponseActFuture<Self, Result<PresignedRequest, String>>;

    #[tracing::instrument]
    fn handle(&mut self, msg: PresignWrite, _: &mut Self::Context) -> Self::Result {
        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move {
                operator
                    .presign_write(&msg.path, msg.expires)
                    .await
                    .map_err(|e| e.to_string())
            }
            .into_actor(self),
        )
    }
}

#[derive(Message, Debug)]
#[rtype(result = "Result<PresignedRequest, String>")]
pub struct PresignWrite {
    pub path: String,
    pub expires: Duration,
}

impl Handler<PresignStat> for S3Actor {
    type Result = ResponseActFuture<Self, Result<PresignedRequest, String>>;

    #[tracing::instrument]
    fn handle(&mut self, msg: PresignStat, _: &mut Self::Context) -> Self::Result {
        let operator = Arc::clone(&self.operator);
        Box::pin(
            async move {
                operator
                    .presign_stat(&msg.path, msg.expires)
                    .await
                    .map_err(|e| e.to_string())
            }
            .into_actor(self),
        )
    }
}

#[derive(Message, Debug)]
#[rtype(result = "Result<PresignedRequest, String>")]
pub struct PresignStat {
    pub path: String,
    pub expires: Duration,
}
