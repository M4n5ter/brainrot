use actix::prelude::*;
use bytes::Bytes;
use dev::SystemRegistry;
use loro::LoroDoc;
use rustc_hash::FxHashMap;
use tokio::sync::mpsc::{Receiver, Sender};
use tracing::error;

use crate::{storage::s3::S3Actor, GenericMessage};

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
                            connection,
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
                    tokio::select! {
                        Some(msg) = mc.rx.recv() => {
                            room.do_send(BroadcastMessage {
                                message: msg,
                                sender_id: mc.connection_id.clone(),
                            });
                        }
                        else => {
                            room_manager.do_send(LeaveRoom {
                                room_id: mc.room_id.clone(),
                                connection_id: mc.connection_id.clone(),
                            });
                            break;
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
            .or_insert_with(|| RoomActor::new().start())
            .clone();
        room.do_send(AddConnection {
            id: msg.connection_id.clone(),
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
    connections: FxHashMap<ConnectionID, Addr<ConnectionActor>>,
    doc: LoroDoc,
}
type ConnectionID = String;

impl RoomActor {
    pub fn new() -> Self {
        RoomActor {
            connections: FxHashMap::default(),
            doc: LoroDoc::new(),
        }
    }
}

impl Actor for RoomActor {
    type Context = Context<Self>;
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
        msg.connection.do_send(SendMessage(GenericMessage::from(
            self.doc.export_snapshot(),
        )));
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
        self.connections.remove(&msg.id);
        if self.connections.is_empty() {
            ctx.notify(StopRoom);
        }
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct RemoveConnection {
    id: ConnectionID,
}

impl Handler<SyncDoc> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: SyncDoc, _: &mut Self::Context) {
        if let Err(e) = self.doc.import(&msg.0) {
            error!("Failed to import doc: {:?}", e);
        };
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct SyncDoc(Bytes);

impl Handler<GetSnapshot> for RoomActor {
    type Result = BytesMessage;

    fn handle(&mut self, _: GetSnapshot, _: &mut Self::Context) -> Self::Result {
        BytesMessage(self.doc.export_snapshot().into())
    }
}

#[derive(Message)]
#[rtype(result = "BytesMessage")]
struct GetSnapshot;

#[derive(MessageResponse)]
struct BytesMessage(Bytes);

impl Handler<StopRoom> for RoomActor {
    type Result = ();

    fn handle(&mut self, _: StopRoom, ctx: &mut Self::Context) {
        let s3 = S3Actor::from_registry();
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
        // TODO: 有需要的话需要进行资源清理
        ctx.stop();
    }
}

#[derive(Message)]
#[rtype(result = "()")]
struct StopConnection;
