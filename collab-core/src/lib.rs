use std::ops::Deref;

use bytes::{Bytes, BytesMut};

pub mod config;
pub mod listener;
pub mod room;
pub mod storage;

#[derive(Clone, Debug)]
pub enum GenericMessage {
    Text(String),
    Binary(Bytes),
    ConnectionClosed,
}

impl From<String> for GenericMessage {
    fn from(s: String) -> Self {
        GenericMessage::Text(s)
    }
}

impl From<Vec<u8>> for GenericMessage {
    fn from(v: Vec<u8>) -> Self {
        GenericMessage::Binary(Bytes::from(v))
    }
}

impl From<Bytes> for GenericMessage {
    fn from(b: Bytes) -> Self {
        GenericMessage::Binary(b)
    }
}

impl From<GenericMessage> for String {
    fn from(msg: GenericMessage) -> Self {
        match msg {
            GenericMessage::Text(s) => s,
            GenericMessage::Binary(v) => String::from_utf8_lossy(&v).to_string(),
            GenericMessage::ConnectionClosed => "Connection closed".to_string(),
        }
    }
}

impl From<GenericMessage> for Bytes {
    fn from(msg: GenericMessage) -> Self {
        match msg {
            GenericMessage::Text(s) => Bytes::from(s),
            GenericMessage::Binary(v) => v,
            GenericMessage::ConnectionClosed => Bytes::from("Connection closed"),
        }
    }
}

impl From<GenericMessage> for BytesMut {
    fn from(value: GenericMessage) -> Self {
        Bytes::from(value).into()
    }
}

impl Deref for GenericMessage {
    type Target = [u8];

    fn deref(&self) -> &Self::Target {
        match self {
            GenericMessage::Text(s) => s.as_bytes(),
            GenericMessage::Binary(v) => v.as_ref(),
            GenericMessage::ConnectionClosed => b"Connection closed",
        }
    }
}
