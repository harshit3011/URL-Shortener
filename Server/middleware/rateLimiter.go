package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/harshit3011/URL-Shortener/database"
)

func RateLimiter() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		id, exists := ctx.Get("user_id")

		if !exists {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unathorized User"})
			return
		}
		user_id := id.(string)

		key := "rate:user:" + user_id

		rateCtx, rateCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer rateCancel()

		count, err := database.RedisClient.Incr(rateCtx, key).Result()

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Rate Limiter error", "details": err.Error()})
			return
		}
		if count == 1 {
			_, err := database.RedisClient.Expire(rateCtx, key, time.Minute).Result()
			if err != nil {
				ctx.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Couldn't set rate limit expiry",
					"details": err.Error(),
				})
				return
			}
		}

		if count > 10 {
			ctx.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate Limit Reached"})
			return
		}

		ctx.Next()
	}
}
