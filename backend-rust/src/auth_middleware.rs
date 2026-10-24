use warp::Filter;
use jsonwebtoken::{decode, DecodingKey, Validation, Algorithm, TokenData};
use serde::{Deserialize, Serialize};
use warp::reject::{self, Reject};
use warp::{http::StatusCode, Rejection, Reply};
use std::convert::Infallible;
use log::{info, warn, error};
use std::env;
use chrono::{Utc};

// Claims now include role
#[derive(Debug, Serialize, Deserialize)]
struct Claims {
    username: String,
    exp: usize,
    role: String,  // Add role field to JWT claims
}

#[derive(Debug)]
struct Unauthorized;

impl Reject for Unauthorized {}

// JWT authorization middleware
pub fn with_jwt_auth(required_role: &str) -> impl Filter<Extract = (), Error = Rejection> + Clone {
    let required_role = required_role.to_string();
    warp::header::<String>("authorization")
        .and_then(move |token: String| {
            let required_role = required_role.clone();
            async move {
                info!("Received JWT: {}", token);

                match validate_jwt(token.clone()) {
                    Ok(token_data) => {
                        info!("Decoded JWT Claims: {:?}", token_data.claims);
                        let is_super = token_data.claims.role == "superuser" || token_data.claims.role == "super";
                        if token_data.claims.role == required_role || is_super {
                            Ok(())
                        } else {
                            warn!("Unauthorized role: {:?}", token_data.claims.role);
                            Err(reject::custom(Unauthorized))
                        }
                    },
                    Err(e) => {
                        warn!("JWT validation error: {:?}", e);
                        Err(reject::custom(Unauthorized))
                    },
                }
            }
        })
        .untuple_one()
}

fn validate_jwt(token: String) -> Result<TokenData<Claims>, jsonwebtoken::errors::Error> {
    let secret_key = env::var("JWT_SECRET_KEY").unwrap_or_else(|_| {
        error!("JWT_SECRET_KEY environment variable is not set. Using default secret.");
        "your_secret_key".to_string()
    });

    let key = DecodingKey::from_secret(secret_key.as_ref());
    let token = token.trim_start_matches("Bearer ").trim();
    info!("Cleaned JWT token: {}", token);

    let token_data = decode::<Claims>(token, &key, &Validation::new(Algorithm::HS256))?;

    let now = Utc::now().timestamp() as usize;
    if token_data.claims.exp < now {
        warn!("Token has expired: {:?}", token_data.claims);
        return Err(jsonwebtoken::errors::Error::from(jsonwebtoken::errors::ErrorKind::ExpiredSignature));
    }

    Ok(token_data)
}

pub async fn handle_rejection(err: Rejection) -> Result<impl Reply, Infallible> {
    if err.is_not_found() {
        error!("Error: NOT_FOUND");
        Ok(warp::reply::with_status("NOT_FOUND", StatusCode::NOT_FOUND))
    } else if err.find::<Unauthorized>().is_some() {
        error!("Authorization error: Unauthorized access.");
        Ok(warp::reply::with_status("UNAUTHORIZED", StatusCode::UNAUTHORIZED))
    } else {
        error!("Unexpected error: {:?}", err);
        Ok(warp::reply::with_status("INTERNAL_SERVER_ERROR", StatusCode::INTERNAL_SERVER_ERROR))
    }
}