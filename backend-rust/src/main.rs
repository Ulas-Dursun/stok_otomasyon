use warp::Filter;
use dotenv::dotenv;

use db::mongodb_client::{get_counter_collection, get_info_collection, get_product_collection, get_user_collection};
use handlers::info_handler::info_routes;
use handlers::product_handler::product_routes;
use handlers::user_handler::user_routes;
use crate::auth_middleware::handle_rejection;

mod handlers;
mod db;
mod models;
mod services;
mod auth_middleware;

#[tokio::main]
async fn main() {
    dotenv().ok();
    env_logger::init();

    let user_collection = get_user_collection().await;
    let product_collection = get_product_collection().await;
    let info_collection = get_info_collection().await;
    let counter_collection = get_counter_collection().await;

    let routes = user_routes(user_collection, counter_collection.clone())
        .or(product_routes(product_collection, counter_collection.clone()))
        .or(info_routes(info_collection, counter_collection))
        .recover(handle_rejection);

    warp::serve(routes)
        .run(([127, 0, 0, 1], 3030))
        .await;
}

/*mod db;
mod models;
mod services;
mod auth_middleware;

use mongodb::{Client, Collection};
use crate::models::user::User;
use crate::models::product::Product;
use crate::models::info::Info;
use crate::services::user_service::user_operations;
use crate::services::product_service::product_operations;
use crate::services::info_service::info_operations;
use crate::models::counter::Counter;

#[tokio::main]
async fn main() -> mongodb::error::Result<()> {
    let client = Client::with_uri_str("mongodb://localhost:27017").await?;

    // Koleksiyonlar
    let user_db = client.database("user");
    let user_collection: Collection<User> = user_db.collection("user_db");

    let product_db = client.database("product");
    let product_collection: Collection<Product> = product_db.collection("product_db");

    let info_db = client.database("info");
    let info_collection: Collection<Info> = info_db.collection("info_db");

    // Counter koleksiyonu
    let counter_collection = client.database("market_otomasyon").collection("counters");

    println!("Market Otomasyon Sistemine Hoş Geldiniz!");

    loop {
        println!("\nLütfen bir seçenek seçin:");
        println!("1) User işlemleri");
        println!("2) Product işlemleri");
        println!("3) Info işlemleri");
        println!("4) Çıkış");

        let mut choice = String::new();
        std::io::stdin().read_line(&mut choice).expect("Giriş hatası");

        match choice.trim() {
            "1" => user_operations(&user_collection, &counter_collection).await,
            "2" => product_operations(&product_collection, &counter_collection).await,
            "3" => info_operations(&info_collection, &counter_collection).await, // Added counter_collection here
            "4" => {
                println!("Çıkış yapılıyor...");
                break;
            }
            _ => println!("Geçersiz seçenek, lütfen tekrar deneyin."),
        }
    }

    Ok(())
}*/
