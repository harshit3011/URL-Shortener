package database

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func CreateIndexes() {
	indexModel := mongo.IndexModel{
		Keys: bson.D{
			{Key: "shortened_url", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	}
	urlCollection := OpenCollection("urls", Client)
	_, err := urlCollection.Indexes().CreateOne(context.Background(), indexModel)

	if err != nil {
		log.Fatal("Failed to create indices", err)
	}
}
