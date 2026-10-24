use mongodb::bson::doc;
use mongodb::Collection;
use crate::models::counter::Counter;

pub async fn get_next_sequence_value(counter_collection: &Collection<Counter>, sequence_name: &str) -> i32 {
    let filter = doc! { "_id": sequence_name };
    let update = doc! { "$inc": { "seq": 1 } };
    let options = mongodb::options::FindOneAndUpdateOptions::builder()
        .upsert(true) // Eğer belge yoksa oluşturur
        .return_document(mongodb::options::ReturnDocument::After)
        .build();

    match counter_collection.find_one_and_update(filter, update, options).await {
        Ok(Some(updated_counter)) => updated_counter.seq,
        Ok(None) => {
            eprintln!("Counter not found for sequence name: {}", sequence_name);
            // Burada uygun bir varsayılan değer döndürebilirsiniz veya hata fırlatabilirsiniz
            0 // Varsayılan bir değer döndürmek yerine uygun bir hata işleme yapabilirsiniz
        },
        Err(e) => {
            eprintln!("Failed to get next sequence value: {}", e);
            // Burada uygun bir varsayılan değer döndürebilirsiniz veya hata fırlatabilirsiniz
            0 // Varsayılan bir değer döndürmek yerine uygun bir hata işleme yapabilirsiniz
        }
    }
}

#[allow(dead_code)]
pub async fn initialize_counter(counter_collection: &Collection<Counter>, sequence_name: &str) -> Result<(), mongodb::error::Error> {
    let filter = doc! { "_id": sequence_name };
    let update = doc! { "$setOnInsert": { "seq": 0 } };
    let options = mongodb::options::FindOneAndUpdateOptions::builder()
        .upsert(true) // Eğer belge yoksa oluşturur
        .return_document(mongodb::options::ReturnDocument::After)
        .build();

    counter_collection.find_one_and_update(filter, update, options).await?;
    Ok(())
}
