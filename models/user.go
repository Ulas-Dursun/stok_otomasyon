package models

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type User struct {
	ID           string `bson:"id" json:"id"`
	Ad           string `bson:"ad" json:"ad"`
	Soyad        string `bson:"soyad" json:"soyad"`
	Yas          uint   `bson:"yas" json:"yas"`
	Cinsiyet     string `bson:"cinsiyet" json:"cinsiyet"`
	KullaniciAdi string `bson:"kullanici_adi" json:"kullanici_adi"`
	Sifre        string `bson:"sifre" json:"sifre"`
	Role         string `bson:"role" json:"role"` // "superuser" veya "normal"
}

type Counter struct {
	ID  string `bson:"_id"`
	Seq int32  `bson:"seq"`
}

var Client *mongo.Client
var UserCollection *mongo.Collection
var CounterCollection *mongo.Collection

// ConnectDatabase fonksiyonu, veritabanı bağlantısını sağlar.
func ConnectDatabase() error {
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("MongoDB sunucusuna bağlantı başarısız: %v", err)
	}

	fmt.Println("MongoDB bağlantısı başarılı.")

	UserCollection = client.Database("user").Collection("user_db")
	CounterCollection = client.Database("market_otomasyon").Collection("counters")
	OrderCollection = client.Database("market_otomasyon").Collection("orders")
	Client = client
	return nil
}

// GetNextSequenceValue Rust API ile uyumlu sıralı ID üretir.
func GetNextSequenceValue(sequenceName string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"_id": sequenceName}
	update := bson.M{"$inc": bson.M{"seq": 1}}
	opts := options.FindOneAndUpdate().SetUpsert(true).SetReturnDocument(options.After)

	var counter Counter
	err := CounterCollection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&counter)
	if err != nil {
		return "1"
	}
	return strconv.Itoa(int(counter.Seq))
}
