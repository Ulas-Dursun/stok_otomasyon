use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
pub struct Counter {
    pub _id: String,
    pub seq: i32,
}
