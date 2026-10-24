use serde::{Serialize, Deserialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Product {
    #[serde(default)]
    pub id: String,
    #[serde(default)]
    pub adi: String,
    #[serde(default)]
    pub miktar: u32,
    #[serde(default)]
    pub stok_durumu: bool,
    #[serde(default)]
    pub satis_durumu: bool,
}
