use bytes::{Bytes, BytesMut};

pub mod actor;

impl MessageWithOwner<&str> for Bytes {
    fn set_owner(&self, owner: &str) -> Bytes {
        let mut buffer = BytesMut::with_capacity(self.len() + owner.len());
        buffer.extend_from_slice(owner.as_bytes());
        buffer.extend_from_slice(b":");
        buffer.extend_from_slice(self);
        buffer.freeze()
    }
}

trait MessageWithOwner<T> {
    fn set_owner(&self, owner: T) -> Self;
}
