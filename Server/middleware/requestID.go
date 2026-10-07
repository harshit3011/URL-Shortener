package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Observability() gin.HandlerFunc{
	return func(ctx *gin.Context) {
		request_id := uuid.New().String()
		ctx.Set("request_id", request_id)
		ctx.Header("X-Request-ID", request_id)

		start := time.Now()
		ctx.Next()
		duration:= time.Since(start)
		log.Printf(
            "request_id=%s method=%s path=%s status=%d duration=%s",
            request_id,
            ctx.Request.Method,
            ctx.Request.URL.Path,
            ctx.Writer.Status(),
            duration,
        )
	}
}