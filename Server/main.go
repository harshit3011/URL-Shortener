package main

import (
	"context"
	"log"

	// "github.com/gin-gonic/gin"
	"github.com/harshit3011/URL-Shortener/database"
)

func main() {

	// router := gin.Default()

	client := database.ConnectDB()

	err := client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal("MongoDB couldn't be connected!!")
	}
	log.Println("MongoDB connected successfully!")

	// fmt.Println("This is my url shortener project")

}
