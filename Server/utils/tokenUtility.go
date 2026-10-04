package utils

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/harshit3011/URL-Shortener/database"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

var ACCESS_SECRET []byte = []byte(os.Getenv("SECRET_ACCESS_KEY"))
var REFRESH_SECRET []byte = []byte(os.Getenv("SECRET_REFRESH_KEY"))

type AccessClaims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func GenerateTokenPair(userId, email string) (*TokenPair, error) {
	accessClaims := AccessClaims{
		UserID: userId,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	refreshClaims := RefreshClaims{
		UserID: userId,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString(ACCESS_SECRET)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString(REFRESH_SECRET)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func ValidateAccessToken(tokenString string) (*AccessClaims, error) {
	claims := &AccessClaims{}

	token, err := jwt.ParseWithClaims(
		tokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {

			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}

			return ACCESS_SECRET, nil
		},
	)
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid access token")
	}

	return claims, nil
}

func RefreshAccessToken(refreshTokenString *string) (*string, error) {
	claims := &RefreshClaims{}

	token, err := jwt.ParseWithClaims(
		*refreshTokenString,
		claims,
		func(token *jwt.Token) (interface{}, error) {
			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("unexpected signing method")
			}
			return REFRESH_SECRET, nil
		},
	)

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	accessClaims := AccessClaims{
		UserID: claims.UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	newAccessToken, err := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		accessClaims,
	).SignedString(ACCESS_SECRET)

	if err != nil {
		return nil, err
	}

	return &newAccessToken, nil
}

func GetAccessToken(ctx *gin.Context) (string, error) {
	authorization := ctx.GetHeader("Authorization")
	if authorization == "" {
		return "", errors.New("authorization header is required")
	}

	parts := strings.Fields(authorization)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", errors.New("authorization header must use the Bearer scheme")
	}

	return parts[1], nil

}

func UpdateAllTokens(userId, accessToken, refreshToken string, client *mongo.Client) (err error) {
	c, cancel := context.WithTimeout(context.Background(),3*time.Second)
	defer cancel()

	updatedAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updatedData := bson.D{
		{Key: "$set", Value: bson.M{
			"access_token": accessToken,
			"refresh_token": refreshToken,
			"updated_at": updatedAt,
		}},
	}
	var userCollection *mongo.Collection = database.OpenCollection("users",client)

	result, err := userCollection.UpdateOne(c, bson.M{"_id": userId}, updatedData)
	if err != nil {
		return err
	}
	if result.MatchedCount == 0 {
		return errors.New("user not found while updating tokens")
	}
	return nil
}
