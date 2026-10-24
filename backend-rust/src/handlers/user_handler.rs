use warp::Filter;
use mongodb::Collection;
use crate::models::user::User;
use crate::models::counter::Counter;
use std::convert::Infallible;
use crate::services::counter_service::get_next_sequence_value;
use mongodb::bson::{doc};
use futures::stream::TryStreamExt;
use sha2::{Sha256, Digest};
use crate::auth_middleware::with_jwt_auth;

pub fn user_routes(
    user_collection: Collection<User>,
    counter_collection: Collection<Counter>
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path("add_user")
        .and(warp::post())
        .and(with_jwt_auth("superuser"))
        .and(warp::body::json())
        .and(with_db(user_collection.clone()))
        .and(with_counter(counter_collection.clone()))
        .and_then(handle_add_user)
        .or(warp::path("delete_user")
            .and(warp::delete())
            .and(with_jwt_auth("superuser"))
            .and(warp::path::param())
            .and(with_db(user_collection.clone()))
            .and_then(handle_delete_user)
        )
        .or(warp::path("list_users")
            .and(warp::get())
            .and(with_jwt_auth("superuser"))
            .and(with_db(user_collection.clone()))
            .and_then(handle_list_users)
        )
        .or(warp::path("update_user")
            .and(warp::put())
            .and(with_jwt_auth("superuser"))
            .and(warp::path::param())
            .and(warp::body::json())
            .and(with_db(user_collection))
            .and_then(handle_update_user)
        )
}

fn with_db(db: Collection<User>) -> impl Filter<Extract = (Collection<User>,), Error = Infallible> + Clone {
    warp::any().map(move || db.clone())
}

fn with_counter(counter: Collection<Counter>) -> impl Filter<Extract = (Collection<Counter>,), Error = Infallible> + Clone {
    warp::any().map(move || counter.clone())
}

async fn handle_add_user(
    mut new_user: User,
    db: Collection<User>,
    counter: Collection<Counter>
) -> Result<impl warp::Reply, Infallible> {
    let user_id = get_next_sequence_value(&counter, "user").await;
    new_user.id = user_id.to_string();

    if new_user.role.is_empty() {
        new_user.role = "normal".to_string();
    } else if new_user.role == "super" {
        new_user.role = "superuser".to_string();
    }

    // Şifre zaten SHA-256 hex (64 karakter) değilse hashle
    if new_user.sifre.len() != 64 || !new_user.sifre.chars().all(|c| c.is_ascii_hexdigit()) {
        let mut hasher = Sha256::new();
        hasher.update(new_user.sifre.as_bytes());
        let result = hasher.finalize();
        new_user.sifre = format!("{:x}", result);
    }

    match db.insert_one(new_user, None).await {
        Ok(_) => Ok(warp::reply::json(&"User added successfully")),
        Err(e) => {
            eprintln!("Failed to insert user: {:?}", e);
            Ok(warp::reply::json(&"Failed to add user"))
        }
    }
}

async fn handle_delete_user(
    user_id: String,
    db: Collection<User>
) -> Result<impl warp::Reply, Infallible> {
    let filter = doc! { "id": user_id };

    match db.delete_one(filter, None).await {
        Ok(delete_result) => {
            if delete_result.deleted_count > 0 {
                Ok(warp::reply::json(&"User deleted successfully"))
            } else {
                Ok(warp::reply::json(&"User not found"))
            }
        }
        Err(e) => {
            eprintln!("Failed to delete user: {:?}", e);
            Ok(warp::reply::json(&"Failed to delete user"))
        }
    }
}

async fn handle_list_users(
    db: Collection<User>
) -> Result<impl warp::Reply, Infallible> {
    let mut users = Vec::new();
    match db.find(None, None).await {
        Ok(mut cursor) => {
            while let Ok(Some(user)) = cursor.try_next().await {
                users.push(user);
            }
        }
        Err(e) => eprintln!("Failed to execute find users: {:?}", e),
    }

    Ok(warp::reply::json(&users))
}

async fn handle_update_user(
    user_id: String,
    mut updated_user: User,
    db: Collection<User>
) -> Result<impl warp::Reply, Infallible> {
    let filter = doc! { "id": user_id };

    let mut set_doc = doc! {
        "ad": updated_user.ad,
        "soyad": updated_user.soyad,
        "yas": updated_user.yas,
        "cinsiyet": updated_user.cinsiyet,
        "kullanici_adi": updated_user.kullanici_adi,
    };

    if !updated_user.sifre.is_empty() {
        if updated_user.sifre.len() != 64 || !updated_user.sifre.chars().all(|c| c.is_ascii_hexdigit()) {
            let mut hasher = Sha256::new();
            hasher.update(updated_user.sifre.as_bytes());
            let result = hasher.finalize();
            updated_user.sifre = format!("{:x}", result);
        }
        set_doc.insert("sifre", updated_user.sifre);
    }

    if !updated_user.role.is_empty() {
        let role = if updated_user.role == "super" {
            "superuser".to_string()
        } else {
            updated_user.role
        };
        set_doc.insert("role", role);
    }

    let update = doc! { "$set": set_doc };

    match db.update_one(filter, update, None).await {
        Ok(update_result) => {
            if update_result.matched_count > 0 {
                Ok(warp::reply::json(&"User updated successfully"))
            } else {
                Ok(warp::reply::json(&"User not found"))
            }
        }
        Err(e) => {
            eprintln!("Failed to update user: {:?}", e);
            Ok(warp::reply::json(&"Failed to update user"))
        }
    }
}
