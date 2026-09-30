package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func ConnectDB() *mongo.Client{
	err:= godotenv.Load(".env")
	if err!=nil {
		log.Println("Unable to load the environment variables file")
	}
	mongourl := os.Getenv("MONGODB_URI")

	if mongourl==""{
		log.Fatal("MONGODB_URL not set!!")
	}

	clientOptions := options.Client().ApplyURI(mongourl)

	client, err:= mongo.Connect(clientOptions)
	if err!=nil{
		log.Fatal("Database couldn't be connected")
	}
	return client

}

var Client *mongo.Client

func OpenCollection(collection string, client *mongo.Client) *mongo.Collection {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Unable to load the environment variables file")
	}
	database := os.Getenv("DATABASE_NAME")

	if database == "" {
		log.Fatal("DB Name not set!!")
	}
	cln := client.Database(database).Collection(collection)
	if cln == nil {
		return nil
	}
	return cln
}