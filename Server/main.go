package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshit3011/URL-Shortener/database"
	"github.com/harshit3011/URL-Shortener/kafka"
	"github.com/harshit3011/URL-Shortener/middleware"
	"github.com/harshit3011/URL-Shortener/routes"
	"github.com/joho/godotenv"
)

func main() {

	router := gin.Default()

	router.Use(middleware.Observability())
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}
	client := database.ConnectDB()

	database.Client = client

	port := os.Getenv("PORT")
	if port == "" {
		port = "8081"
	}
	err = client.Ping(context.Background(), nil)

	if err != nil {
		log.Fatal("MongoDB couldn't be connected!!")
	}
	log.Println("MongoDB connected successfully!")

	database.CreateIndexes()
	database.ConnectRedis()

	kafka.InitProducer()
	kafka.InitConsumer()
	consumerCtx, stopConsumer := context.WithCancel(context.Background())
	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		kafka.ConsumeClicks(consumerCtx)
	}()

	defer func() {
		if err := database.RedisClient.Close(); err != nil {
			log.Printf("Failed to disconnect from Redis: %v", err)
		}
	}()

	defer func() {
		if err := client.Disconnect(context.Background()); err != nil {
			log.Fatalf("Failed to disconnect from MongoDB: %v", err)
		}
	}()

	routes.BackendRoutes(router)

	server := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Server couldn't be connected!!")
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit

	log.Println("Shutting down server...")
	stopConsumer()
	if err := kafka.Reader.Close(); err != nil {
		log.Printf("Failed to close Kafka consumer: %v", err)
	}
	<-consumerDone

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}
}
