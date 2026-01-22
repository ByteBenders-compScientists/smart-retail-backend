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

	// auth routes (no authentication required)
	auth.POST("/register", controllers.Register)
	auth.POST("/login", controllers.Login)

	// auth routes requiring authentication
	authProtected := auth.Group("/")
	authProtected.Use(middlewares.JWTAuthMiddleware())
	{
		authProtected.GET("/me", controllers.GetCurrentUser)
		authProtected.POST("/logout", controllers.Logout)
	}

	// M-Pesa webhook (no authentication required)
	api.POST("/payments/mpesa/callback", controllers.MpesaCallback)

	// protected routes
	protected := api.Group("/")
	protected.Use(middlewares.JWTAuthMiddleware())
	{
		// Product routes (accessible by all authenticated users)
		protected.GET("/products", controllers.GetAllProducts)
		protected.GET("/products/:id", controllers.GetProduct)
		protected.GET("/products/branch/:branchId", controllers.GetBranchInventory)

		// Branch routes (accessible by all authenticated users)
		protected.GET("/branches", controllers.GetAllBranches)
		protected.GET("/branches/:id", controllers.GetBranch)

		// Order routes (customer accessible)
		protected.POST("/orders", controllers.CreateOrder)
		protected.GET("/orders", controllers.GetUserOrders)
		protected.GET("/orders/:id", controllers.GetOrderById)

		// Payment routes (customer accessible)
		protected.POST("/payments/mpesa/initiate", controllers.InitiateMpesaPayment)
		protected.GET("/payments/:orderId/status", controllers.GetPaymentStatus)

		// admin only routes
		admin := protected.Group("/admin")
		admin.Use(middlewares.AdminAuthMiddleware())
		{
			// Branch management
			admin.POST("/branches", controllers.CreateBranch)
			admin.PUT("/branches/:id", controllers.UpdateBranch)
			admin.DELETE("/branches/:id", controllers.DeleteBranch)

			// Product management
			admin.POST("/products", controllers.CreateProduct)
			admin.PUT("/products/:id", controllers.UpdateProduct)
			admin.DELETE("/products/:id", controllers.DeleteProduct)
			admin.GET("/products/brand", controllers.GetProductsByBrand)
			admin.GET("/products/:id/stock", controllers.GetProductStockAcrossBranches)

			// Restocking
			admin.POST("/restock", controllers.RestockBranch)
			admin.GET("/inventory", controllers.GetInventory)
			admin.GET("/restock-logs", controllers.GetRestockLogs)

			// Reports
			admin.GET("/reports/sales", controllers.GetSalesReports)
			admin.GET("/reports/branch/:branchId", controllers.GetBranchReport)
		}
	}

	return r
}
