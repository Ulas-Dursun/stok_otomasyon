use serde::{Serialize, Deserialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct Info {
    #[serde(default)]
    pub id: String,
    #[serde(default)]
    pub adres: String,
    #[serde(default)]
    pub telefon: String,
    #[serde(default)]
    pub calisan_sayisi: u32,
}
