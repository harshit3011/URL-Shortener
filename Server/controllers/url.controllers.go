package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshit3011/URL-Shortener/database"
	"github.com/harshit3011/URL-Shortener/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func GetUrls() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		user_id, exists := ctx.Get("user_id")
		user_id = user_id.(string)
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
			return
		}

		var userCollection *mongo.Collection = database.OpenCollection("users", database.Client)
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		filter := bson.D{{Key: "_id", Value: user_id}}

		var user models.User
		err:= userCollection.FindOne(c, filter).Decode(&user)
		if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
				return
		}
		var urls []models.URL
		urls=user.URLS

		if len(urls)==0{
			ctx.JSON(http.StatusFound,gin.H{"Result":"This user doesn't have any URLs yet"})
			return
		}
		ctx.JSON(http.StatusFound, urls)
	}
}

