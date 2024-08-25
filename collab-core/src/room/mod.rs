use std::sync::Arc;

use rustc_hash::FxHashMap;
use tokio::sync::{mpsc, Mutex};

pub struct RoomManager {
    rooms: FxHashMap<RoomID, Arc<Mutex<Room>>>,
}

type RoomID = String;

impl RoomManager {
    pub fn new() -> Self {
        RoomManager {
            rooms: FxHashMap::default(),
        }
    }

    pub fn get_or_create_room(&mut self, room_id: impl Into<String>) -> Arc<Mutex<Room>> {
        let room_id = room_id.into();
        Arc::clone(self.rooms.entry(room_id.to_owned()).or_insert_with(|| {
            Arc::new(Mutex::new(Room {
                _id: room_id,
                connections: FxHashMap::default(),
            }))
        }))
    }
}

impl Default for RoomManager {
    fn default() -> Self {
        Self::new()
    }
}

pub struct Room {
    _id: String,
    connections: FxHashMap<ConnectionID, Connection>,
}
type ConnectionID = String;

struct Connection {
    sender: mpsc::Sender<Vec<u8>>,
}

impl Room {
    pub async fn broardcast(&mut self, message: Vec<u8>, sender_id: &str) {
        let payload = message;

        for (id, conn) in self.connections.iter_mut() {
            if id != sender_id {
                let payload = payload.to_owned();
                if let Err(err) = conn.sender.send(payload).await {
                    eprintln!("Error: {}", err);
                }
            }
        }
    }

    pub async fn add_connection(&mut self, id: String, sender: mpsc::Sender<Vec<u8>>) {
        self.connections
            .insert(id.to_owned(), Connection { sender });
    }

    pub async fn remove_connection(&mut self, id: String) {
        self.connections.remove(&id);
    }
}
