package models

import (
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type Order struct {
	ID          string    `bson:"id" json:"id"`
	Username    string    `bson:"username" json:"username"`
	ProductID   string    `bson:"product_id" json:"product_id"`
	ProductName string    `bson:"product_name" json:"product_name"`
	Quantity    uint      `bson:"quantity" json:"quantity"`
	CreatedAt   time.Time `bson:"created_at" json:"created_at"`
	Status      string    `bson:"status" json:"status"` // "Tamamlandı"
}

var OrderCollection *mongo.Collection
