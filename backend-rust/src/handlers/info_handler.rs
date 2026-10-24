use warp::Filter;
use mongodb::Collection;
use crate::models::info::Info;
use crate::models::counter::Counter;
use crate::services::counter_service::get_next_sequence_value;
use mongodb::bson::doc;
use futures::stream::TryStreamExt;
use std::convert::Infallible;
use crate::auth_middleware::with_jwt_auth;  // JWT doğrulama filtresi

pub fn info_routes(
    info_collection: Collection<Info>,
    counter_collection: Collection<Counter>
) -> impl Filter<Extract = impl warp::Reply, Error = warp::Rejection> + Clone {
    warp::path("add_info")
        .and(warp::post())
        .and(with_jwt_auth("superuser"))  // JWT doğrulama
        .and(warp::body::json())
        .and(with_db(info_collection.clone()))
        .and(with_counter(counter_collection.clone()))
        .and_then(handle_add_info)
        .or(warp::path("delete_info")
            .and(warp::delete())
            .and(with_jwt_auth("superuser"))  // JWT doğrulama
            .and(warp::path::param())
            .and(with_db(info_collection.clone()))
            .and_then(handle_delete_info)
        )
        .or(warp::path("list_info")
            .and(warp::get())
            .and(with_jwt_auth("normal"))
            .and(with_db(info_collection.clone()))
            .and_then(handle_list_info)
        )
        .or(warp::path("update_info")
            .and(warp::put())
            .and(with_jwt_auth("superuser"))  // JWT doğrulama
            .and(warp::path::param())
            .and(warp::body::json())
            .and(with_db(info_collection))
            .and_then(handle_update_info)
        )
}

fn with_db(db: Collection<Info>) -> impl Filter<Extract = (Collection<Info>,), Error = Infallible> + Clone {
    warp::any().map(move || db.clone())
}

fn with_counter(counter: Collection<Counter>) -> impl Filter<Extract = (Collection<Counter>,), Error = Infallible> + Clone {
    warp::any().map(move || counter.clone())
}

async fn handle_add_info(
    mut new_info: Info,
    db: Collection<Info>,
    counter: Collection<Counter>
) -> Result<impl warp::Reply, Infallible> {
    let info_id = get_next_sequence_value(&counter, "info").await;
    new_info.id = info_id.to_string();  // ID'yi String formatına çeviriyoruz.

    match db.insert_one(new_info, None).await {
        Ok(_) => Ok(warp::reply::json(&"Info added successfully")),
        Err(e) => {
            eprintln!("Failed to insert info: {:?}", e);
            Ok(warp::reply::json(&"Failed to add info"))
        }
    }
}

async fn handle_delete_info(
    info_id: String,
    db: Collection<Info>
) -> Result<impl warp::Reply, Infallible> {
    let filter = doc! { "id": info_id };

    match db.delete_one(filter, None).await {
        Ok(delete_result) => {
            if delete_result.deleted_count > 0 {
                Ok(warp::reply::json(&"Info deleted successfully"))
            } else {
                Ok(warp::reply::json(&"Info not found"))
            }
        }
        Err(e) => {
            eprintln!("Failed to delete info: {:?}", e);
            Ok(warp::reply::json(&"Failed to delete info"))
        }
    }
}

async fn handle_list_info(
    db: Collection<Info>
) -> Result<impl warp::Reply, Infallible> {
    let mut info_list = Vec::new();
    match db.find(None, None).await {
        Ok(mut cursor) => {
            while let Ok(Some(info)) = cursor.try_next().await {
                info_list.push(info);
            }
        }
        Err(e) => eprintln!("Failed to execute find info: {:?}", e),
    }

    Ok(warp::reply::json(&info_list))
}

async fn handle_update_info(
    info_id: String,
    updated_info: Info,
    db: Collection<Info>
) -> Result<impl warp::Reply, Infallible> {
    let filter = doc! { "id": info_id };

    let update = doc! {
        "$set": {
            "adres": updated_info.adres,
            "telefon": updated_info.telefon,
            "calisan_sayisi": updated_info.calisan_sayisi,
        }
    };

    match db.update_one(filter, update, None).await {
        Ok(update_result) => {
            if update_result.matched_count > 0 {
                Ok(warp::reply::json(&"Info updated successfully"))
            } else {
                Ok(warp::reply::json(&"Info not found"))
            }
        }
        Err(e) => {
            eprintln!("Failed to update info: {:?}", e);
            Ok(warp::reply::json(&"Failed to update info"))
        }
    }
}
