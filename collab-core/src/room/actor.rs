use actix::prelude::*;
use anyhow::Result;
use rustc_hash::FxHashMap;
use tokio::sync::mpsc::{Receiver, Sender};

pub async fn message_handler(
    mut msg_rx: Receiver<GenericMessage>,
    msg_tx: Sender<GenericMessage>,
    room_id: RoomID,
    connection_id: ConnectionID,
) -> Result<()> {
    let room_manager = RoomManagerActor::from_registry();

    // 创建Connection Actor
    let connection = ConnectionActor::new(msg_tx).start();

    // 加入房间
    let room = room_manager
        .send(JoinRoom {
            room_id: room_id.clone(),
            connection_id: connection_id.clone(),
            connection: connection.clone(),
        })
        .await?;

    // 处理消息
    loop {
        tokio::select! {
            // 接收到消息后广播到房间内的其他连接
            Some(msg) = msg_rx.recv() => {
                room.do_send(BroadcastMessage {
                    message: msg,
                    sender_id: connection_id.clone(),
                });
            }
            else => {
                break;
            }
        }
    }

    // 离开房间
    room_manager.do_send(LeaveRoom {
        room_id,
        connection_id,
    });

    Ok(())
}

/// RoomManagerActor 是一个系统级别的Actor，用于管理所有的房间
///
/// ```rust
/// // 在某处初始化 RoomManagerActor
/// RoomManager::start_default();
///
/// // 在需要使用 RoomManagerActor 的地方获取其地址
/// let room_manager = RoomManager::from_registry();
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

// 使 RoomManager 成为系统服务
impl SystemService for RoomManagerActor {}
impl Supervised for RoomManagerActor {}

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
}
type ConnectionID = String;

impl RoomActor {
    pub fn new() -> Self {
        RoomActor {
            connections: FxHashMap::default(),
        }
    }
}

impl Actor for RoomActor {
    type Context = Context<Self>;
}

impl Handler<BroadcastMessage> for RoomActor {
    type Result = ();

    fn handle(&mut self, msg: BroadcastMessage, _: &mut Self::Context) {
        for (id, conn) in self.connections.iter_mut() {
            if id != &msg.sender_id {
                conn.do_send(SendMessage(msg.message.clone()));
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

impl Handler<StopRoom> for RoomActor {
    type Result = ();

    fn handle(&mut self, _: StopRoom, ctx: &mut Self::Context) {
        // TODO: 有需要的话需要进行资源清理
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
    type Result = ();

    fn handle(&mut self, msg: SendMessage, ctx: &mut Self::Context) {
        let tx = self.tx.clone();
        let fut = async move {
            if let Err(e) = tx.send(msg.0).await {
                eprintln!("Failed to send message: {:?}", e)
            }
        };
        ctx.spawn(fut.into_actor(self));
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

#[derive(Clone, Debug)]
pub enum GenericMessage {
    Text(String),
    Binary(Vec<u8>),
}

impl From<String> for GenericMessage {
    fn from(s: String) -> Self {
        GenericMessage::Text(s)
    }
}

impl From<Vec<u8>> for GenericMessage {
    fn from(v: Vec<u8>) -> Self {
        GenericMessage::Binary(v)
    }
}

impl From<GenericMessage> for Vec<u8> {
    fn from(msg: GenericMessage) -> Self {
        match msg {
            GenericMessage::Text(s) => s.into_bytes(),
            GenericMessage::Binary(v) => v,
        }
    }
}

impl From<GenericMessage> for String {
    fn from(msg: GenericMessage) -> Self {
        match msg {
            GenericMessage::Text(s) => s,
            GenericMessage::Binary(v) => String::from_utf8_lossy(&v).to_string(),
        }
    }
}
