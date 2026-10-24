use serde::{Serialize, Deserialize};

#[derive(Debug, Serialize, Deserialize, Clone)]
pub struct User {
    #[serde(default)]
    pub id: String,
    #[serde(default)]
    pub ad: String,
    #[serde(default)]
    pub soyad: String,
    #[serde(default)]
    pub yas: u32,
    #[serde(default)]
    pub cinsiyet: String,
    #[serde(default)]
    pub kullanici_adi: String,
    #[serde(default)]
    pub sifre: String,
    #[serde(default = "default_role")]
    pub role: String, // "superuser" veya "normal"
}

fn default_role() -> String {
    "normal".to_string()
}

