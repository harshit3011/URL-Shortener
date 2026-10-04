package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/harshit3011/URL-Shortener/database"
	"github.com/harshit3011/URL-Shortener/utils"
)

func AuthMiddleWare() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		token, err := utils.GetAccessToken(ctx)

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized user", "details": err.Error()})
			ctx.Abort()
			return
		}
		if token == "" {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Token not provided", "details": err.Error()})
			ctx.Abort()
			return
		}

		claims, err := utils.ValidateAccessToken(token)

		if err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			ctx.Abort()
			return
		}

		active, err := utils.IsAccessTokenActive(claims.UserID, token, database.Client)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify login session"})
			ctx.Abort()
			return
		}
		if !active {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User is not logged in"})
			ctx.Abort()
			return
		}

		ctx.Set("user_id", claims.UserID)
		ctx.Next()
	}
}
