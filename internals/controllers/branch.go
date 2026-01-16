package controllers

import (
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

func CreateBranch(c *gin.Context) {
	var branch models.Branch
	if err := c.ShouldBindJSON(&branch); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if branch.Name == "" || branch.Location == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and location are required"})
		return
	}

	if err := db.DB.Create(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create branch"})
		return
	}

	c.JSON(http.StatusCreated, branch)
}

func GetAllBranches(c *gin.Context) {
	var branches []models.Branch
	if err := db.DB.Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branches"})
		return
	}

	c.JSON(http.StatusOK, branches)
}

func GetBranch(c *gin.Context) {
	branchID := c.Param("id")

	var branch models.Branch
	if err := db.DB.First(&branch, "id = ?", branchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	c.JSON(http.StatusOK, branch)
}

func UpdateBranch(c *gin.Context) {
	branchID := c.Param("id")

	var branch models.Branch
	if err := db.DB.First(&branch, "id = ?", branchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	var updateData models.Branch
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if err := db.DB.Model(&branch).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branch"})
		return
	}

	c.JSON(http.StatusOK, branch)
}

func DeleteBranch(c *gin.Context) {
	branchID := c.Param("id")

	var branch models.Branch
	if err := db.DB.First(&branch, "id = ?", branchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	if err := db.DB.Delete(&branch).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete branch"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Branch deleted successfully"})
}

func GetBranchInventory(c *gin.Context) {
	branchID := c.Param("id")

	var stocks []models.Stock
	if err := db.DB.
		Preload("Product").
		Preload("Branch").
		Where("branch_id = ?", branchID).
		Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch inventory"})
		return
	}

	// Calculate total value and low stock items
	var totalValue float64
	var lowStockItems []models.Stock

	for _, stock := range stocks {
		itemValue := float64(stock.Quantity) * stock.Product.Price
		totalValue += itemValue

		// Consider low stock if less than 10 items
		if stock.Quantity < 10 {
			lowStockItems = append(lowStockItems, stock)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"branch_id":       branchID,
		"stocks":          stocks,
		"total_value":     totalValue,
		"low_stock_items": lowStockItems,
		"total_items":     len(stocks),
	})
}

func AddStockToBranch(c *gin.Context) {
	branchID := c.Param("id")

	var body struct {
		ProductID string `json:"product_id" binding:"required"`
		Quantity  int    `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if body.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be positive"})
		return
	}

	// Check if product exists
	var product models.Product
	if err := db.DB.First(&product, "id = ?", body.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check if stock already exists for this branch and product
	var existingStock models.Stock
	err := db.DB.Where("branch_id = ? AND product_id = ?", branchID, body.ProductID).First(&existingStock).Error

	if err == nil {
		// Update existing stock
		existingStock.Quantity += body.Quantity
		if err := db.DB.Save(&existingStock).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
			return
		}
		c.JSON(http.StatusOK, existingStock)
		return
	}

	// Create new stock entry
	stock := models.Stock{
		BranchID:  branchID,
		ProductID: body.ProductID,
		Quantity:  body.Quantity,
	}

	if err := db.DB.Create(&stock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create stock"})
		return
	}

	c.JSON(http.StatusCreated, stock)
}

func UpdateStock(c *gin.Context) {
	branchID := c.Param("id")
	stockID := c.Param("stockId")

	var body struct {
		Quantity int `json:"quantity" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if body.Quantity < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity cannot be negative"})
		return
	}

	var stock models.Stock
	if err := db.DB.Where("id = ? AND branch_id = ?", stockID, branchID).First(&stock).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock not found"})
		return
	}

	stock.Quantity = body.Quantity
	if err := db.DB.Save(&stock).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
		return
	}

	c.JSON(http.StatusOK, stock)
}
