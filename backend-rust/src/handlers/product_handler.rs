use warp::Filter;
use mongodb::Collection;
use crate::models::product::Product;
use crate::models::counter::Counter;
use std::convert::Infallible;
use crate::services::counter_service::get_next_sequence_value;
use mongodb::bson::{doc};
use futures::stream::TryStreamExt;
use crate::auth_middleware::with_jwt_auth;
use serde::Serialize;

#[derive(Serialize)]
pub struct OrderResponse {
    pub success: bool,
    pub message: String,
    pub product_name: Option<String>,
    pub remaining_stock: Option<u32>,
}

#[derive(Serialize)]
pub struct RestockResponse {
    pub success: bool,
    pub message: String,
    pub new_stock: Option<u32>,
}

pub fn product_routes(
    product_collection: Collection<Product>,
    counter_collection: Collection<Counter>
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path("add_product")
        .and(warp::post())
        .and(with_jwt_auth("superuser"))  // SuperUser can add products
        .and(warp::body::json())
        .and(with_db(product_collection.clone()))
        .and(with_counter(counter_collection.clone()))
        .and_then(handle_add_product)
        .or(warp::path("delete_product")
            .and(warp::delete())
            .and(with_jwt_auth("superuser"))  // SuperUser can delete products
            .and(warp::path::param())
            .and(with_db(product_collection.clone()))
            .and_then(handle_delete_product)
        )
        .or(warp::path("list_products")
            .and(warp::get())
            .and(with_jwt_auth("normal"))  // Both SuperUser and NormalUser can list products
            .and(with_db(product_collection.clone()))
            .and_then(handle_list_products)
        )
        .or(warp::path("update_product")
            .and(warp::put())
            .and(with_jwt_auth("superuser"))  // SuperUser can update products
            .and(warp::path::param())
            .and(warp::body::json())
            .and(with_db(product_collection.clone()))
            .and_then(handle_update_product)
        )
        .or(warp::path("order_product")
            .and(warp::post())
            .and(with_jwt_auth("normal"))  // Both normal and superuser can place order
            .and(warp::path::param())
            .and(with_db(product_collection.clone()))
            .and_then(handle_order_product)
        )
        .or(warp::path("restock_product")
            .and(warp::post())
            .and(with_jwt_auth("superuser"))  // SuperUser can quick restock
            .and(warp::path::param())
            .and(with_db(product_collection))
            .and_then(handle_restock_product)
        )
}

fn with_db(db: Collection<Product>) -> impl Filter<Extract = (Collection<Product>,), Error = Infallible> + Clone {
    warp::any().map(move || db.clone())
}

fn with_counter(counter: Collection<Counter>) -> impl Filter<Extract = (Collection<Counter>,), Error = Infallible> + Clone {
    warp::any().map(move || counter.clone())
}

async fn handle_add_product(
    mut new_product: Product,
    db: Collection<Product>,
    counter: Collection<Counter>
) -> Result<impl warp::Reply, Infallible> {
    let product_id = get_next_sequence_value(&counter, "product").await;
    new_product.id = product_id.to_string();

    match db.insert_one(new_product, None).await {
        Ok(_) => Ok(warp::reply::json(&"Product added successfully")),
        Err(e) => {
            eprintln!("Failed to insert product: {:?}", e);
            Ok(warp::reply::json(&"Failed to add product"))
        }
    }
}

async fn handle_delete_product(
    product_id: String,
    db: Collection<Product>
) -> Result<impl warp::Reply, Infallible> {
    let filter = doc! { "id": product_id };

    match db.delete_one(filter, None).await {
        Ok(delete_result) => {
            if delete_result.deleted_count > 0 {
                Ok(warp::reply::json(&"Product deleted successfully"))
            } else {
                Ok(warp::reply::json(&"Product not found"))
            }
        }
        Err(e) => {
            eprintln!("Failed to delete product: {:?}", e);
            Ok(warp::reply::json(&"Failed to delete product"))
        }
    }
}

async fn handle_list_products(
    db: Collection<Product>
) -> Result<impl warp::Reply, Infallible> {
    let mut products = Vec::new();
    match db.find(None, None).await {
        Ok(mut cursor) => {
            while let Ok(Some(product)) = cursor.try_next().await {
                products.push(product);
            }
        }
        Err(e) => eprintln!("Failed to execute find products: {:?}", e),
    }

    Ok(warp::reply::json(&products))
}

async fn handle_update_product(
    product_id: String,
    updated_product: Product,
    db: Collection<Product>
) -> Result<impl warp::Reply, Infallible> {
    let filter = doc! { "id": product_id };
    let update = doc! {

         "$set": {
                    "adi": updated_product.adi,
                    "miktar": updated_product.miktar,
                    "stok_durumu": updated_product.stok_durumu,
                    "satis_durumu": updated_product.satis_durumu,
                }
            };
    match db.update_one(filter, update, None).await {
        Ok(update_result) => {
            if update_result.matched_count > 0 {
                Ok(warp::reply::json(&"Product updated successfully"))
            } else {
                Ok(warp::reply::json(&"Product not found"))
            }
        }
        Err(e) => {
            eprintln!("Failed to update product: {:?}", e);
            Ok(warp::reply::json(&"Failed to update product"))
        }
    }
}

async fn handle_order_product(
    product_id: String,
    db: Collection<Product>
) -> Result<impl warp::Reply, Infallible> {
    let filter = doc! { "id": &product_id };
    match db.find_one(filter.clone(), None).await {
        Ok(Some(product)) => {
            if product.miktar == 0 || !product.satis_durumu {
                return Ok(warp::reply::json(&OrderResponse {
                    success: false,
                    message: "Ürün stokta yok veya satışa kapalı".to_string(),
                    product_name: Some(product.adi),
                    remaining_stock: Some(0),
                }));
            }
            let new_miktar = product.miktar - 1;
            let new_stok = new_miktar > 0;
            let update = doc! {
                "$set": {
                    "miktar": new_miktar,
                    "stok_durumu": new_stok,
                }
            };
            let _ = db.update_one(filter, update, None).await;
            Ok(warp::reply::json(&OrderResponse {
                success: true,
                message: "Sipariş başarıyla oluşturuldu".to_string(),
                product_name: Some(product.adi),
                remaining_stock: Some(new_miktar),
            }))
        }
        Ok(None) => Ok(warp::reply::json(&OrderResponse {
            success: false,
            message: "Ürün bulunamadı".to_string(),
            product_name: None,
            remaining_stock: None,
        })),
        Err(e) => {
            eprintln!("Failed to order product: {:?}", e);
            Ok(warp::reply::json(&OrderResponse {
                success: false,
                message: "Veritabanı hatası".to_string(),
                product_name: None,
                remaining_stock: None,
            }))
        }
    }
}

async fn handle_restock_product(
    product_id: String,
    db: Collection<Product>
) -> Result<impl warp::Reply, Infallible> {
    let filter = doc! { "id": &product_id };
    match db.find_one(filter.clone(), None).await {
        Ok(Some(product)) => {
            let new_miktar = product.miktar + 10;
            let update = doc! {
                "$set": {
                    "miktar": new_miktar,
                    "stok_durumu": true,
                }
            };
            let _ = db.update_one(filter, update, None).await;
            Ok(warp::reply::json(&RestockResponse {
                success: true,
                message: "Stok başarıyla +10 adet artırıldı".to_string(),
                new_stock: Some(new_miktar),
            }))
        }
        Ok(None) => Ok(warp::reply::json(&RestockResponse {
            success: false,
            message: "Ürün bulunamadı".to_string(),
            new_stock: None,
        })),
        Err(e) => {
            eprintln!("Failed to restock product: {:?}", e);
            Ok(warp::reply::json(&RestockResponse {
                success: false,
                message: "Stok artırılamadı".to_string(),
                new_stock: None,
            }))
        }
    }
}

