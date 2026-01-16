package api

import (
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/controllers"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/middlewares"
	"github.com/gin-gonic/gin"
)

func SetUpRoutes() *gin.Engine {
	r := gin.Default()

	// cors middleware
	r.Use(middlewares.CORSMiddleware())
	
	// logging middleware
	r.Use(middlewares.RequestLogger())

	api := r.Group("/api/v1")

	// health endpoint
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status": "ok",
		})
	})

	auth := api.Group("/auth")

	// auth routes
	auth.POST("/register", controllers.Register)
	auth.POST("/login", controllers.Login)

	// protected routes
	protected := api.Group("/protected")
	protected.Use(middlewares.JWTAuthMiddleware())
	{
		// user profile
		protected.GET("/profile", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			userRole, _ := c.Get("userRole")
			c.JSON(200, gin.H{
				"userID": userID,
				"role":   userRole,
				"message": "Access granted to protected route",
			})
		})

		// admin only routes
		admin := protected.Group("/admin")
		admin.Use(middlewares.AdminAuthMiddleware())
		{
			admin.GET("/dashboard", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Admin dashboard access"})
			})
		}

		// customer routes (admin and customer can access)
		customer := protected.Group("/customer")
		customer.Use(middlewares.CustomerAuthMiddleware())
		{
			customer.GET("/products", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Customer products access"})
			})
		}
	}

	return r
}
