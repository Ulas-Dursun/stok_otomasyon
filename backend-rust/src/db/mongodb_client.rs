use mongodb::{Client, options::ClientOptions, Collection};
use crate::models::{user::User, product::Product, info::Info, counter::Counter};

pub async fn get_user_collection() -> Collection<User> {
    let client_uri = "mongodb://localhost:27017";
    let client_options = ClientOptions::parse(client_uri).await.unwrap();
    let client = Client::with_options(client_options).unwrap();

    let db = client.database("user");
    db.collection::<User>("user_db")
}

pub async fn get_product_collection() -> Collection<Product> {
    let client_uri = "mongodb://localhost:27017";
    let client_options = ClientOptions::parse(client_uri).await.unwrap();
    let client = Client::with_options(client_options).unwrap();

    let db = client.database("product");
    db.collection::<Product>("product_db")
}

pub async fn get_info_collection() -> Collection<Info> {
    let client_uri = "mongodb://localhost:27017";
    let client_options = ClientOptions::parse(client_uri).await.unwrap();
    let client = Client::with_options(client_options).unwrap();

    let db = client.database("info");
    db.collection::<Info>("info_db")
}

pub async fn get_counter_collection() -> Collection<Counter> {
    let client_uri = "mongodb://localhost:27017";
    let client_options = ClientOptions::parse(client_uri).await.unwrap();
    let client = Client::with_options(client_options).unwrap();

    let db = client.database("market_otomasyon");
    db.collection::<Counter>("counters")
}
