package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/harshit3011/URL-Shortener/database"
	"github.com/harshit3011/URL-Shortener/models"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

func RegisterUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var registerUser models.RegisterUserDetails

		if err := ctx.ShouldBindJSON(&registerUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User data"})
			return
		}
		validate := validator.New()

		if err := validate.Struct(registerUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}

		hashedPassword, err := HashPassword(registerUser.Password)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to hash the password", "details": err.Error()})
			return
		}
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var userCollection *mongo.Collection = database.OpenCollection("users", database.Client)

		filter := bson.D{{Key: "email", Value: registerUser.Email}}

		var tempUser models.User
		err = userCollection.FindOne(c, filter).Decode(&tempUser)

		if err == nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "User already exists"})
			return
		}
		if err == mongo.ErrNoDocuments {
			var user models.User
			user.ID = bson.NewObjectID().Hex()
			user.Username = registerUser.Username
			user.Email = registerUser.Email
			user.Password = hashedPassword
			user.CreatedAt = time.Now()

			result, err := userCollection.InsertOne(c, user)

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error in saving the user", "details": err.Error()})
			}
			ctx.JSON(http.StatusCreated, result)
			
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error", "details": err.Error()})
			return
		}
	}

}
