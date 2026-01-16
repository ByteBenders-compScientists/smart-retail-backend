package api

import (
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/controllers"
	"github.com/gin-gonic/gin"
)

func RegisterRoutes(r *gin.Engine) {
	api := r.Group("/api/v1")

	// health endpoint
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	// auth routes
	api.POST("/register", controllers.Register)
	api.POST("/login", controllers.Login)
}
