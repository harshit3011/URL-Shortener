package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/harshit3011/URL-Shortener/database"
	"github.com/harshit3011/URL-Shortener/models"
	"github.com/harshit3011/URL-Shortener/utils"
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
			user.URLS = []models.URL{}

			result, err := userCollection.InsertOne(c, user)

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error in saving the user", "details": err.Error()})
				return
			}
			ctx.JSON(http.StatusCreated, result)

		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Internal error", "details": err.Error()})
			return
		}
	}

}

func LoginUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var loginDetails models.LoginUserDetails

		err := ctx.ShouldBindJSON(&loginDetails)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user data", "details": err.Error()})
			return
		}
		validate := validator.New()
		if err := validate.Struct(loginDetails); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Validation failed", "details": err.Error()})
			return
		}
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		var userCollection *mongo.Collection = database.OpenCollection("users", database.Client)

		var foundUser models.User

		if loginDetails.Username == "" {
			filter := bson.D{{Key: "email", Value: loginDetails.Email}}
			err := userCollection.FindOne(c, filter).Decode(&foundUser)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
				return
			}
			err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(loginDetails.Password))
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
				return
			}

			tokenPair, err := utils.GenerateTokenPair(foundUser.ID, foundUser.Email)

			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
				return
			}
			err = utils.UpdateAllTokens(foundUser.ID, tokenPair.AccessToken, tokenPair.RefreshToken, database.Client)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tokens"})
				return
			}

			http.SetCookie(ctx.Writer, &http.Cookie{
				Name:  "access_token",
				Value: tokenPair.AccessToken,
				Path:  "/",
				// Domain:   "localhost",
				MaxAge:   86400,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteNoneMode,
			})
			http.SetCookie(ctx.Writer, &http.Cookie{
				Name:  "refresh_token",
				Value: tokenPair.RefreshToken,
				Path:  "/",
				// Domain:   "localhost",
				MaxAge:   604800,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteNoneMode,
			})
			ctx.JSON(http.StatusCreated, gin.H{
				"id":           foundUser.ID,
				"username":     foundUser.Username,
				"email":        foundUser.Email,
				"access_token": tokenPair.AccessToken,
			})
		} else {
			filter := bson.D{{Key: "username", Value: loginDetails.Username}}
			err := userCollection.FindOne(c, filter).Decode(&foundUser)
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
				return
			}
			err = bcrypt.CompareHashAndPassword([]byte(foundUser.Password), []byte(loginDetails.Password))
			if err != nil {
				ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
				return
			}
			tokenPair, err := utils.GenerateTokenPair(foundUser.ID, foundUser.Email)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate tokens"})
				return
			}

			err = utils.UpdateAllTokens(foundUser.ID, tokenPair.AccessToken, tokenPair.RefreshToken, database.Client)
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tokens"})
				return
			}

			http.SetCookie(ctx.Writer, &http.Cookie{
				Name:     "access_token",
				Value:    tokenPair.AccessToken,
				Path:     "/",
				MaxAge:   86400,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteNoneMode,
			})

			http.SetCookie(ctx.Writer, &http.Cookie{
				Name:     "refresh_token",
				Value:    tokenPair.RefreshToken,
				Path:     "/",
				MaxAge:   604800,
				Secure:   true,
				HttpOnly: true,
				SameSite: http.SameSiteNoneMode,
			})

			ctx.JSON(http.StatusCreated, gin.H{
				"id":           foundUser.ID,
				"username":     foundUser.Username,
				"email":        foundUser.Email,
				"access_token": tokenPair.AccessToken,
			})
		}
	}
}

func LogoutUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User is not logged in"})
			return
		}
		userID, ok := id.(string)
		if !ok {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user session"})
			return
		}
		err := utils.UpdateAllTokens(userID, "", "", database.Client)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tokens"})
			return
		}

		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:     "access_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})

		http.SetCookie(ctx.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			Secure:   true,
			HttpOnly: true,
			SameSite: http.SameSiteNoneMode,
		})
		ctx.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
	}
}
