use actix::prelude::*;
use bytes::Bytes;
use dev::SystemRegistry;
use loro::LoroDoc;
use rustc_hash::FxHashMap;
use tokio::sync::mpsc::{Receiver, Sender};
use tracing::{error, trace};

use crate::{
    storage::s3::{ReadFile, S3Actor, WriteFile},
    storage::StorageErrorKind,
    GenericMessage,
};

/// RoomManagerActor 是一个系统级别的Actor，用于管理所有的房间
///
/// ```rust
/// // 在某处初始化 RoomManagerActor
/// RoomManagerActor::start_default();
///
/// // 在需要使用 RoomManagerActor 的地方获取其地址
/// let room_manager = RoomManagerActor::from_registry();
/// ```
pub struct RoomManagerActor {
    rooms: FxHashMap<RoomID, Addr<RoomActor>>,
}
type RoomID = String;

impl RoomManagerActor {
    pub fn new() -> Self {
        RoomManagerActor {
            rooms: FxHashMap::default(),
        }
    }
}

impl Default for RoomManagerActor {
    fn default() -> Self {
        Self::new()
    }
}

impl Actor for RoomManagerActor {
    type Context = Context<Self>;
}

impl SystemService for RoomManagerActor {
    fn service_started(&mut self, ctx: &mut Context<Self>) {
        SystemRegistry::set(ctx.address())
    }
}
impl Supervised for RoomManagerActor {}

impl Handler<MessageChan> for RoomManagerActor {
    type Result = ();
    fn handle(&mut self, mut mc: MessageChan, ctx: &mut Self::Context) {
        let connection = ConnectionActor::new(mc.tx).start();
        ctx.spawn(
            async move {
                let room_manager = RoomManagerActor::from_registry();
                let room = {
                    match room_manager
                        .send(JoinRoom {
                            room_id: mc.room_id.clone(),
                            connection_id: mc.connection_id.clone(),
                            connection: connection.clone(),
                        })
                        .await
                    {
                        Ok(room) => room,
                        Err(e) => {
                            error!("Failed to join room: {:?}", e);
                            return;
                        }
                    }
                };

                loop {
                    let Some(msg) = mc.rx.recv().await else {
                        room_manager.do_send(LeaveRoom {
                            room_id: mc.room_id,
                            connection_id: mc.connection_id,
                        });
                        break;
                    };
                    match msg {
                        GenericMessage::ConnectionClosed => {
                            connection
                                .send(SendMessage(GenericMessage::ConnectionClosed))
                                .await
                                .unwrap();
                            trace!(
                                "room {}'s connection {} is closed",
                                mc.room_id,
                                mc.connection_id
                            );
                            room.do_send(RemoveConnection {
                                id: mc.connection_id.clone(),
                            });
                            break;
                        }
                        _ => {
                            room.do_send(BroadcastMessage {
                                message: msg,
                                sender_id: mc.connection_id.clone(),
                            });
                        }
                    }
                }
            }
            .into_actor(self),
        );
    }
}

#[derive(Message)]
#[rtype(result = "()")]
pub struct MessageChan {
    /// `rx` receives messages from the client
    pub rx: Receiver<GenericMessage>,
    /// `tx` sends messages to the client
    pub tx: Sender<GenericMessage>,
    pub room_id: RoomID,
    pub connection_id: ConnectionID,
}

impl Handler<RemoveRoom> for RoomManagerActor {
    type Result = ();

    fn handle(&mut self, msg: RemoveRoom, _: &mut Self::Context) {
        self.rooms.remove(&msg.0);
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct RemoveRoom(RoomID);

impl Handler<JoinRoom> for RoomManagerActor {
    type Result = Addr<RoomActor>;

    fn handle(&mut self, msg: JoinRoom, _: &mut Self::Context) -> Self::Result {
        let room = self
            .rooms
            .entry(msg.room_id.clone())
            .or_insert_with(|| RoomActor::new(msg.room_id, msg.connection_id.clone()).start())
            .clone();

        room.do_send(AddConnection {
            id: msg.connection_id,
            connection: msg.connection,
        });
        room
    }
}

#[derive(Message)]
#[rtype(result = "Addr<RoomActor>")]
struct JoinRoom {
    room_id: RoomID,
    connection_id: ConnectionID,
    connection: Addr<ConnectionActor>,
}

impl Handler<LeaveRoom> for RoomManagerActor {
    type Result = ();

    fn handle(&mut self, msg: LeaveRoom, _: &mut Self::Context) {
        if let Some(room) = self.rooms.get_mut(&msg.room_id) {
            room.do_send(RemoveConnection {
                id: msg.connection_id,
            });
        }
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct LeaveRoom {
    room_id: RoomID,
    connection_id: ConnectionID,
}

struct RoomActor {
    id: RoomID,
    owner: ConnectionID,
    connections: FxHashMap<ConnectionID, Addr<ConnectionActor>>,
    doc: Option<LoroDoc>,
    doc_path: String,
}
type ConnectionID = String;

impl RoomActor {
    pub fn new(id: String, owner: ConnectionID) -> Self {
        RoomActor {
            id: id.clone(),
            owner: owner.clone(),
            connections: FxHashMap::default(),
            doc: None,
            doc_path: format!("{}/collab_room/{}/doc", owner, id),
        }
    }
}

impl Actor for RoomActor {
    type Context = Context<Self>;

    fn started(&mut self, ctx: &mut Self::Context) {
        ctx.notify(SyncFromRemote);
    }

    fn stopped(&mut self, _: &mut Self::Context) {
        trace!("{}'s room {} is stopped", self.owner, self.id);
    }
}

impl Handler<BroadcastMessage> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: BroadcastMessage, ctx: &mut Self::Context) {
        let sender_id = msg.sender_id;
        let msg: Bytes = msg.message.into();
        ctx.notify(SyncDoc(msg.clone()));
        for (id, conn) in self.connections.iter_mut() {
            if id != &sender_id {
                conn.do_send(SendMessage(msg.clone().into()));
            }
        }
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct BroadcastMessage {
    message: GenericMessage,
    sender_id: String,
}

impl Handler<AddConnection> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: AddConnection, _: &mut Self::Context) {
        trace!("Adding connection {} to room {}", msg.id, self.id);
        self.connections.insert(msg.id, msg.connection);
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct AddConnection {
    id: ConnectionID,
    connection: Addr<ConnectionActor>,
}

/// 从房间中移除连接，如果房间为空则停止房间
impl Handler<RemoveConnection> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: RemoveConnection, ctx: &mut Self::Context) {
        trace!("Removing connection {} from room {}", msg.id, self.id);

        let addr = ctx.address();
        if self.connections.len() == 1 {
            ctx.spawn(
                async move {
                    RoomManagerActor::from_registry().do_send(RemoveRoom(msg.id));
                    addr.do_send(ClearConnections);

                    if let Err(e) = addr.send(SyncToRemote).await {
                        error!("Failed to sync to remote: {:?}", e);
                    }

                    if let Err(e) = addr.send(StopRoom).await {
                        error!("Failed to stop room: {:?}", e);
                    };
                }
                .into_actor(self),
            );
        } else {
            self.connections.remove(&msg.id);
        }
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct RemoveConnection {
    id: ConnectionID,
}

impl Handler<ClearConnections> for RoomActor {
    type Result = ();

    fn handle(&mut self, _: ClearConnections, _: &mut Self::Context) {
        self.connections.clear();
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct ClearConnections;

impl Handler<SyncDoc> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: SyncDoc, _: &mut Self::Context) {
        if self.doc.is_none() {
            self.doc = Some(LoroDoc::new());
        }

        // unwrap is safe here because we just checked if self.doc is None
        if let Err(e) = self.doc.as_mut().unwrap().import(&msg.0) {
            error!("Failed to import doc: {:?}", e);
        }
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct SyncDoc(Bytes);

impl Handler<SyncFromRemote> for RoomActor {
    type Result = ResponseActFuture<Self, ()>;

    fn handle(&mut self, _: SyncFromRemote, ctx: &mut Self::Context) -> Self::Result {
        if self.doc.is_some() {
            return Box::pin(async {}.into_actor(self));
        }

        trace!("Syncing {}'s room {} from remote", self.owner, self.id);

        let addr = ctx.address();
        let s3 = S3Actor::from_registry();
        let doc_path = self.doc_path.clone();
        Box::pin(
            async move {
                match s3.send(ReadFile { path: doc_path }).await {
                    Ok(Ok(data)) => {
                        if let Err(e) = addr.send(SetDoc(data.clone())).await {
                            error!("Failed to set doc: {:?}", e);
                        }

                        addr.do_send(BroadcastMessage {
                            message: GenericMessage::Binary(data),
                            sender_id: "".to_string(),
                        });
                    }
                    Ok(Err(e)) => {
                        if e.kind() == StorageErrorKind::NotFound {
                            return;
                        }
                        error!("Failed to read file because of opendal Error: {:?}", e);
                    }
                    Err(e) => {
                        error!("Failed to read file because of MailboxError: {:?}", e);
                    }
                }
            }
            .into_actor(self),
        )
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct SyncFromRemote;

impl Handler<SyncToRemote> for RoomActor {
    type Result = ResponseActFuture<Self, ()>;

    fn handle(&mut self, _: SyncToRemote, ctx: &mut Self::Context) -> Self::Result {
        trace!("Syncing {}'s room {} to remote", self.owner, self.id);

        let owner = self.owner.clone();
        let room_id = self.id.clone();
        let addr = ctx.address();
        let s3 = S3Actor::from_registry();
        let doc_path = self.doc_path.clone();
        Box::pin(
            async move {
                trace!("Getting {}'s{} room snapshot", owner, room_id);
                let doc = match addr.send(GetSnapshot).await {
                    Ok(Ok(doc)) => doc.0,
                    Ok(Err(e)) => {
                        error!("Failed to get snapshot: {:?}", e);
                        return;
                    }
                    Err(e) => {
                        error!("Failed to get snapshot: {:?}", e);
                        return;
                    }
                };

                trace!("Writing {}'s{} room snapshot to S3", owner, room_id);
                if let Err(e) = s3
                    .send(WriteFile {
                        path: doc_path,
                        data: doc,
                    })
                    .await
                {
                    error!("Failed to write file: {:?}", e);
                };
            }
            .into_actor(self),
        )
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct SyncToRemote;

impl Handler<SetDoc> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: SetDoc, _: &mut Self::Context) {
        if self.doc.is_none() {
            self.doc = Some(LoroDoc::new());
        }

        // unwrap is safe here because we just checked if self.doc is None
        if let Err(e) = self.doc.as_mut().unwrap().import(&msg.0) {
            error!("Failed to import snapshot: {:?}", e);
        }
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct SetDoc(Bytes);

impl Handler<GetSnapshot> for RoomActor {
    type Result = Result<BytesMessage, String>;

    fn handle(&mut self, _: GetSnapshot, _: &mut Self::Context) -> Self::Result {
        if let Some(doc) = &self.doc {
            Ok(BytesMessage(doc.export_snapshot().into()))
        } else {
            Err("Doc is not initialized".to_string())
        }
    }
}

#[derive(Message)]
#[rtype(result = "Result<BytesMessage, String>")]
struct GetSnapshot;

#[derive(MessageResponse)]
struct BytesMessage(Bytes);

impl Handler<StopRoom> for RoomActor {
    type Result = ();

    fn handle(&mut self, _: StopRoom, ctx: &mut Self::Context) -> Self::Result {
        ctx.stop();
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct StopRoom;

impl Handler<CheckRoomEmpty> for RoomActor {
    type Result = bool;

    fn handle(&mut self, _: CheckRoomEmpty, _: &mut Self::Context) -> Self::Result {
        self.connections.is_empty()
    }
}

#[derive(Message)]
#[rtype(result = "bool")]
struct CheckRoomEmpty;

struct ConnectionActor {
    tx: Sender<GenericMessage>,
}

impl ConnectionActor {
    pub fn new(tx: Sender<GenericMessage>) -> Self {
        ConnectionActor { tx }
    }
}

impl Actor for ConnectionActor {
    type Context = Context<Self>;
}

impl Handler<SendMessage> for ConnectionActor {
    type Result = ResponseActFuture<Self, ()>;

    fn handle(&mut self, msg: SendMessage, _: &mut Self::Context) -> Self::Result {
        let tx = self.tx.clone();
        Box::pin(
            async move {
                if let Err(e) = tx.send(msg.0).await {
                    error!("Failed to send message: {:?}", e)
                }
            }
            .into_actor(self),
        )
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct SendMessage(GenericMessage);

impl Handler<StopConnection> for ConnectionActor {
    type Result = ();

    fn handle(&mut self, _: StopConnection, ctx: &mut Self::Context) {
        ctx.stop();
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct StopConnection;
