package controllers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshit3011/URL-Shortener/database"
	"github.com/harshit3011/URL-Shortener/models"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func BenchmarkRedirect(b *testing.B) {

	gin.SetMode(gin.ReleaseMode)

	database.ConnectRedis()
	defer database.RedisClient.Close()

	router := gin.New()
	router.GET("/redirect/:shortcode", RedirectUrl())

	err := database.RedisClient.Set(
		b.Context(),
		"url:000001",
		"https://example.com",
		0,
	).Err()

	if err != nil {
		b.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/redirect/000001", nil)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {

		recorder := httptest.NewRecorder()

		router.ServeHTTP(recorder, req)

		if recorder.Code != 302 {
			b.Fatalf("expected status 302, got %d", recorder.Code)
		}
	}
}

func TestRedirect(t *testing.T) {
	gin.SetMode(gin.ReleaseMode)

	database.ConnectRedis()
	defer database.RedisClient.Close()

	router := gin.New()
	router.GET("/redirect/:shortcode", RedirectUrl())

	err := database.RedisClient.Set(
		context.Background(),
		"url:000001",
		"https://example.com",
		0,
	).Err()

	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/redirect/000001", nil)
	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, recorder.Code)
	}
}

func TestRedirectCacheMiss(t *testing.T) {

	err := godotenv.Load("../.env")

	if err != nil {
		t.Fatal(err)
	}

	gin.SetMode(gin.ReleaseMode)

	client := database.ConnectDB()

	if client == nil {
		t.Fatal("Mongo client is nil")
	}

	database.Client = client

	defer client.Disconnect(context.Background())

	database.ConnectRedis()

	defer database.RedisClient.Close()

	router := gin.New()

	router.GET("/redirect/:shortcode", RedirectUrl())

	ctx := context.Background()

	originalURL := "https://example.com"

	shortcode := bson.NewObjectID().Hex()

	urlID := bson.NewObjectID().Hex()

	urlCollection := database.OpenCollection("urls", database.Client)

	url := &models.URL{
		ID:           urlID,
		OriginalURL:  originalURL,
		ShortenedURL: shortcode,
		CreatedAt:    time.Now(),
	}

	_, err = urlCollection.InsertOne(ctx, url)

	if err != nil {
		t.Fatal(err)
	}

	defer urlCollection.DeleteOne(ctx, bson.D{{Key: "_id", Value: urlID}})
	defer database.RedisClient.Del(ctx, "url:"+shortcode)

	req := httptest.NewRequest("GET", "/redirect/"+shortcode, nil)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusFound {
		t.Fatalf("expected status %d, got %d", http.StatusFound, recorder.Code)
	}
}
