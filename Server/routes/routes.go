package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/harshit3011/URL-Shortener/controllers"
	"github.com/harshit3011/URL-Shortener/middleware"
)

func BackendRoutes(router *gin.Engine) {

	router.GET("/", func(ctx *gin.Context) {
		ctx.String(200, "Hi, this is my URL-shortener project")
	})
	router.POST("/register", controllers.RegisterUser())
	router.POST("/login", controllers.LoginUser())
	router.POST("/logout", middleware.AuthMiddleWare(), controllers.LogoutUser())
}
