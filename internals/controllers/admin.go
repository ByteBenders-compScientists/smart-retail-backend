// admin controller
package controllers

import (
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

type RestockRequest struct {
	BranchID  string `json:"branchId" binding:"required"`
	ProductID string `json:"productId" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required,min=1"`
}

func RestockBranch(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req RestockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	tx := db.DB.Begin()

	// Get current inventory
	var inventory models.BranchInventory
	if err := tx.Where("branch_id = ? AND product_id = ?", req.BranchID, req.ProductID).First(&inventory).Error; err != nil {
		// Create new inventory record if doesn't exist
		inventory = models.BranchInventory{
			BranchID:  req.BranchID,
			ProductID: req.ProductID,
			Quantity:  0,
		}
		if err := tx.Create(&inventory).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create inventory record"})
			return
		}
	}

	previousQuantity := inventory.Quantity
	newQuantity := previousQuantity + req.Quantity

	// Update inventory
	if err := tx.Model(&inventory).Update("quantity", newQuantity).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update inventory"})
		return
	}

	// Create restock log
	restockLog := models.RestockLog{
		BranchID:         req.BranchID,
		ProductID:        req.ProductID,
		QuantityAdded:    req.Quantity,
		PreviousQuantity: previousQuantity,
		NewQuantity:      newQuantity,
		RestockedBy:      userID.(string),
	}

	if err := tx.Create(&restockLog).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create restock log"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusOK, gin.H{
		"message": "Branch restocked successfully",
		"updatedInventory": gin.H{
			"branchId":    req.BranchID,
			"productId":   req.ProductID,
			"previousQty": previousQuantity,
			"addedQty":    req.Quantity,
			"newQty":      newQuantity,
		},
	})
}

func GetSalesReports(c *gin.Context) {
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")
	branchID := c.Query("branchId")
	productID := c.Query("productId")

	query := db.DB.Model(&models.Order{}).
		Preload("Branch").
		Preload("OrderItems.Product").
		Where("payment_status = ?", "completed")

	// Apply filters
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}
	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	var orders []models.Order
	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sales data"})
		return
	}

	// Calculate sales by brand
	salesByBrand := map[string]gin.H{
		"Coke":   {"units": 0, "revenue": 0.0},
		"Fanta":  {"units": 0, "revenue": 0.0},
		"Sprite": {"units": 0, "revenue": 0.0},
	}

	// Calculate sales by branch
	salesByBranch := map[string]float64{}

	grandTotal := 0.0

	for _, order := range orders {
		grandTotal += order.TotalAmount

		// Sales by branch
		salesByBranch[order.Branch.Name] += order.TotalAmount

		// Sales by brand
		for _, item := range order.OrderItems {
			if brandData, exists := salesByBrand[item.ProductBrand]; exists {
				brandData["units"] = brandData["units"].(int) + item.Quantity
				brandData["revenue"] = brandData["revenue"].(float64) + item.Subtotal
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"salesByBrand":  salesByBrand,
		"salesByBranch": salesByBranch,
		"grandTotal":    grandTotal,
		"filters": gin.H{
			"startDate": startDate,
			"endDate":   endDate,
			"branchId":  branchID,
			"productId": productID,
		},
	})
}

func GetBranchReport(c *gin.Context) {
	branchID := c.Param("branchId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	var branch models.Branch
	if err := db.DB.First(&branch, "id = ?", branchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	query := db.DB.Model(&models.Order{}).
		Preload("OrderItems.Product").
		Where("branch_id = ? AND payment_status = ?", branchID, "completed")

	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	var orders []models.Order
	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branch sales"})
		return
	}

	// Calculate metrics
	totalRevenue := 0.0
	productSales := map[string]int{}

	for _, order := range orders {
		totalRevenue += order.TotalAmount
		for _, item := range order.OrderItems {
			productSales[item.ProductBrand] += item.Quantity
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"branch": branch,
		"branchSales": gin.H{
			"totalRevenue": totalRevenue,
			"totalOrders":  len(orders),
			"topProducts":  productSales,
		},
	})
}

func GetInventory(c *gin.Context) {
	branchID := c.Query("branchId")

	query := db.DB.Model(&models.BranchInventory{}).
		Preload("Branch").
		Preload("Product")

	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	var inventories []models.BranchInventory
	if err := query.Find(&inventories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory"})
		return
	}

	// Add low stock alerts
	lowStockThreshold := 20
	lowStockItems := []gin.H{}

	for _, inventory := range inventories {
		if inventory.Quantity <= lowStockThreshold {
			lowStockItems = append(lowStockItems, gin.H{
				"branchId":     inventory.BranchID,
				"branchName":   inventory.Branch.Name,
				"productId":    inventory.ProductID,
				"productName":  inventory.Product.Name,
				"currentStock": inventory.Quantity,
				"threshold":    lowStockThreshold,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"inventory":         inventories,
		"lowStockAlerts":    lowStockItems,
		"lowStockThreshold": lowStockThreshold,
	})
}

func GetRestockLogs(c *gin.Context) {
	branchID := c.Query("branchId")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	query := db.DB.Model(&models.RestockLog{}).
		Preload("Branch").
		Preload("Product").
		Preload("RestockedByUser")

	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	var logs []models.RestockLog
	if err := query.Order("created_at DESC").Find(&logs).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch restock logs"})
		return
	}

	c.JSON(http.StatusOK, logs)
}
