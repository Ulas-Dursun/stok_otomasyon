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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_product_serialization() {
        let product = Product {
            id: "1".to_string(),
            adi: "Test Ürün".to_string(),
            miktar: 10,
            stok_durumu: true,
            satis_durumu: true,
        };

        let json = serde_json::to_string(&product).expect("Failed to serialize product");
        assert!(json.contains("Test Ürün"));
        assert!(json.contains("\"miktar\":10"));

        let deserialized: Product = serde_json::from_str(&json).expect("Failed to deserialize product");
        assert_eq!(deserialized.id, "1");
        assert_eq!(deserialized.miktar, 10);
        assert!(deserialized.stok_durumu);
    }

    #[test]
    fn test_product_stock_depletion() {
        let mut product = Product {
            id: "2".to_string(),
            adi: "Tükenen Ürün".to_string(),
            miktar: 1,
            stok_durumu: true,
            satis_durumu: true,
        };

        product.miktar -= 1;
        product.stok_durumu = product.miktar > 0;

        assert_eq!(product.miktar, 0);
        assert!(!product.stok_durumu);
    }
}
