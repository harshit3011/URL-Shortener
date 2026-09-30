package main

import (
	"context"
	"log"
	"os"

	// "github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin"
	"github.com/harshit3011/URL-Shortener/controllers"
	"github.com/harshit3011/URL-Shortener/database"
	"github.com/joho/godotenv"
)

func main() {

	router := gin.Default()

	router.GET("/hello", func(ctx *gin.Context) {
		ctx.String(200,"Hello, this is my URL-shortener project")
	})

	router.POST("/registerUser",controllers.RegisterUser())
	err := godotenv.Load(".env")
	if err!=nil{
		log.Fatal("Unable to find .env file")
	}
	client := database.ConnectDB()

	database.Client = client

	port := os.Getenv("PORT")
	if port==""{
		port = "8081"
	}
	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal("MongoDB couldn't be connected!!")
	}
	log.Println("MongoDB connected successfully!")

	defer func(){
		if err:=client.Disconnect(context.Background()); err!=nil{
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	err= router.Run(":"+port); if err != nil {
		log.Fatal("Server couldn't be connected!!")
	}
}
