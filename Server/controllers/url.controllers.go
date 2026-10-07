package controllers

import (
	"context"
	"log"
	"math/big"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshit3011/URL-Shortener/database"
	"github.com/harshit3011/URL-Shortener/models"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	"golang.org/x/sync/singleflight"
)

var urlGroup singleflight.Group

func GetUrls() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user_id, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
			return
		}
		user_id = user_id.(string)

		var userCollection *mongo.Collection = database.OpenCollection("users", database.Client)
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		userFilter := bson.D{{Key: "_id", Value: user_id}}

		var user models.User
		err := userCollection.FindOne(c, userFilter).Decode(&user)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
			return
		}
		var urls []models.URL
		urls = user.URLS

		if len(urls) == 0 {
			ctx.JSON(http.StatusOK, gin.H{"Result": "This user doesn't have any URLs yet"})
			return
		}
		ctx.JSON(http.StatusOK, urls)
	}
}

func ShortenUrl() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user_id, exists := ctx.Get("user_id")
		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user"})
			return
		}
		user_id = user_id.(string)
		var userCollection *mongo.Collection = database.OpenCollection("users", database.Client)
		c, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		userFilter := bson.D{{Key: "_id", Value: user_id}}

		var user models.User
		err := userCollection.FindOne(c, userFilter).Decode(&user)
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "User not found", "details": err.Error()})
			return
		}

		counterCollection := database.OpenCollection("counters", database.Client)

		counterFilter := bson.D{
			{Key: "_id", Value: "url"},
		}

		update := bson.D{
			{Key: "$inc", Value: bson.D{
				{Key: "sequence", Value: 1},
			}},
		}

		opts := options.FindOneAndUpdate().
			SetUpsert(true).
			SetReturnDocument(options.After)

		var counter models.Counter

		err = counterCollection.FindOneAndUpdate(
			c,
			counterFilter,
			update,
			opts,
		).Decode(&counter)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Counter couldn't be created/updated", "details": err.Error()})
			return
		}
		num := big.NewInt(counter.Sequence)
		shortUrl := num.Text(62)

		for len(shortUrl) < 6 {
			shortUrl = "0" + shortUrl
		}

		var longUrl struct {
			URL string `json:"url"`
		}

		err = ctx.ShouldBindJSON(&longUrl)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid url payload", "detail": err.Error()})
			return
		}
		url := &models.URL{}

		url.ID = bson.NewObjectID().Hex()
		url.OriginalURL = longUrl.URL
		url.ShortenedURL = shortUrl
		url.BelongsTo = user.ID
		url.CreatedAt = time.Now()

		_, err = userCollection.UpdateOne(c,
			bson.D{
				{Key: "_id", Value: user.ID},
				{Key: "urls", Value: nil},
			},
			bson.D{
				{Key: "$set", Value: bson.D{
					{Key: "urls", Value: bson.A{}},
				}},
			},
		)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Couldn't initialize the user's URL list", "details": err.Error()})
			return
		}

		urlCollection := database.OpenCollection("urls", database.Client)
		_, err = urlCollection.InsertOne(c, url)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error in saving the URL", "details": err.Error()})
			return
		}

		updateResult, err := userCollection.UpdateOne(c, userFilter, bson.D{
			{Key: "$push", Value: bson.D{
				{Key: "urls", Value: url},
			}},
		})
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "URL was created but couldn't be added to the user's URL list", "details": err.Error()})
			return
		}
		if updateResult.MatchedCount == 0 {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "URL was created but the user was not found"})
			return
		}

		ctx.JSON(http.StatusCreated, url)
	}
}

func RedirectUrl() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		shortcode := ctx.Param("shortcode")

		key := "url:" + shortcode
		redisGetCtx, redisGetCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer redisGetCancel()

		cachedResult, err := database.RedisClient.Get(redisGetCtx, key).Result()

		if err == nil {
			log.Printf("Redis cache HIT: shortcode=%s", shortcode)
			ctx.Redirect(http.StatusFound, cachedResult)
			return
		} else if err == redis.Nil {
			log.Printf("Redis cache MISS: shortcode=%s", shortcode)
			_, err, _ := urlGroup.Do(shortcode, func() (interface{}, error) {
				mongoCtx, mongoCancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer mongoCancel()
				urlCollection := database.OpenCollection("urls", database.Client)

				filter := bson.D{{
					Key: "shortened_url", Value: shortcode,
				}}

				var url models.URL
				err = urlCollection.FindOne(mongoCtx, filter).Decode(&url)

				if err != nil {
					ctx.JSON(http.StatusNotFound, gin.H{"error": "URL not found", "details": err.Error()})
					return nil, err
				}

				redisSetCtx, redisSetCancel := context.WithTimeout(context.Background(), 3*time.Second)
				defer redisSetCancel()
				_, err := database.RedisClient.Set(redisSetCtx, key, url.OriginalURL, time.Hour).Result()
				if err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Key couldn't be set in redis", "details": err.Error()})
					return nil, err
				}

				ctx.Redirect(http.StatusFound, url.OriginalURL)
				return url.OriginalURL, nil
			})
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "SingleFlight error", "details": err.Error()})
				return
			}
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Redis internal error", "details": err.Error()})
			return
		}

	}
}
