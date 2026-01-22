// cors middleware
package middlewares

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:3000,http://localhost:3001,http://localhost:5173"
	}

	allowedOrigins := strings.Split(corsOrigins, ",")

	config := cors.Config{
		AllowOrigins: allowedOrigins,

		AllowMethods: []string{
			"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS",
		},

		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
		},

		AllowCredentials: true,

		MaxAge: 12 * time.Hour,
	}

	return cors.New(config)
}
