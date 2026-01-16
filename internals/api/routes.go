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

	// MPESA webhook (no authentication required)
	api.POST("/mpesa/webhook", controllers.MpesaWebhook)

	// protected routes
	protected := api.Group("/")
	protected.Use(middlewares.JWTAuthMiddleware())
	{
		// user profile
		protected.GET("/me", func(c *gin.Context) {
			userID, _ := c.Get("userID")
			userRole, _ := c.Get("userRole")
			c.JSON(200, gin.H{
				"userID":  userID,
				"role":    userRole,
				"message": "Access granted to protected route",
			})
		})

		// admin only routes
		admin := protected.Group("/admin")
		admin.Use(middlewares.AdminAuthMiddleware())
		{
			// Branch management
			admin.POST("/branches", controllers.CreateBranch)
			admin.GET("/branches", controllers.GetAllBranches)
			admin.GET("/branches/:id", controllers.GetBranch)
			admin.PUT("/branches/:id", controllers.UpdateBranch)
			admin.DELETE("/branches/:id", controllers.DeleteBranch)
			admin.GET("/branches/:id/inventory", controllers.GetBranchInventory)
			admin.POST("/branches/:id/stock", controllers.AddStockToBranch)
			admin.PUT("/branches/:id/stock/:stockId", controllers.UpdateStock)
			admin.GET("/branches/:id/low-stock", controllers.GetLowStockAlerts)
			
			// Product management
			admin.POST("/products", controllers.CreateProduct)
			admin.GET("/products", controllers.GetAllProducts)
			admin.GET("/products/:id", controllers.GetProduct)
			admin.PUT("/products/:id", controllers.UpdateProduct)
			admin.DELETE("/products/:id", controllers.DeleteProduct)
			admin.GET("/products/brand", controllers.GetProductsByBrand)
			admin.GET("/products/:id/stock", controllers.GetProductStockAcrossBranches)
			
			// Sales and stock
			admin.GET("/branches/:id/stocks", controllers.GetBranchStock)
			admin.POST("/branches/:id/sales", controllers.CreateSale)
			admin.GET("/sales", controllers.GetAllSales)
			admin.GET("/sales/:saleId", controllers.GetSale)
			admin.PUT("/sales/:saleId/status", controllers.UpdateSaleStatus)
			admin.GET("/sales/status", controllers.GetSalesByStatus)
			admin.GET("/branches/:id/sales", controllers.GetBranchSales)
			
			// Restocking from HQ
			admin.POST("/restock", controllers.RestockFromHQ)
			admin.POST("/restock/bulk", controllers.BulkRestockFromHQ)
			admin.GET("/restock/hq-stock", controllers.GetHQStock)
			admin.GET("/restock/history", controllers.GetRestockHistory)
			admin.GET("/restock/suggestions", controllers.GetRestockSuggestions)
			
			// Reports
			admin.GET("/reports/sales", controllers.GetSalesReport)
			admin.GET("/reports/branches", controllers.GetBranchPerformanceReport)
			admin.GET("/reports/low-stock", controllers.GetLowStockReport)
			admin.GET("/reports/revenue", controllers.GetRevenueSummary)
			admin.GET("/reports/trends", controllers.GetDailySalesTrend)
			
			// Sync management
			admin.GET("/sync/pending", controllers.GetPendingSync)
			admin.PUT("/sync/:saleId/resolve", controllers.ResolveSyncConflict)
			
			// Alerts and notifications
			admin.GET("/alerts/low-stock", controllers.GetLowStockAlerts)
			admin.GET("/alerts/critical", controllers.GetCriticalStockAlerts)
			admin.GET("/alerts/summary", controllers.GetAlertSummary)
			admin.GET("/alerts/history", controllers.GetAlertHistory)
			admin.POST("/alerts/rules", controllers.CreateAlertRule)
		}

		// customer routes (admin and customer can access)
		customer := protected.Group("/customer")
		customer.Use(middlewares.CustomerAuthMiddleware())
		{
			// Branch viewing
			customer.GET("/branches", controllers.GetAllBranches)
			customer.GET("/branches/:id", controllers.GetBranch)
			customer.GET("/branches/:id/stocks", controllers.GetBranchStock)
			customer.GET("/branches/:id/alerts", controllers.GetStockAlertsByBranch)
			
			// Product viewing
			customer.GET("/products", controllers.GetAllProducts)
			customer.GET("/products/:id", controllers.GetProduct)
			customer.GET("/products/brand", controllers.GetProductsByBrand)
			
			// Sales and payments
			customer.POST("/branches/:id/sales", controllers.CreateSale)
			customer.GET("/sales", controllers.GetAllSales)
			customer.GET("/sales/:saleId", controllers.GetSale)
			customer.POST("/mpesa/initiate", controllers.InitiateMpesa)
			customer.POST("/sync", controllers.SyncOffline)
			customer.GET("/sync/status", controllers.GetSyncStatus)
		}
	}

	return r
}
