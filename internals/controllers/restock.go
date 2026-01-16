package controllers

import (
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

type RestockRequest struct {
	BranchID  string `json:"branch_id" binding:"required"`
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

type BulkRestockRequest struct {
	BranchID string            `json:"branch_id" binding:"required"`
	Items    []BulkRestockItem `json:"items" binding:"required"`
}

type BulkRestockItem struct {
	ProductID string `json:"product_id" binding:"required"`
	Quantity  int    `json:"quantity" binding:"required"`
}

func RestockFromHQ(c *gin.Context) {
	var req RestockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be positive"})
		return
	}

	// Verify branch exists
	var branch models.Branch
	if err := db.DB.First(&branch, req.BranchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	// Verify product exists
	var product models.Product
	if err := db.DB.First(&product, "id = ?", req.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check HQ stock (assuming branch with IsHQ=true is headquarters)
	var hqBranch models.Branch
	if err := db.DB.Where("is_hq = ?", true).First(&hqBranch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Headquarters not found"})
		return
	}

	var hqStock models.Stock
	if err := db.DB.Where("branch_id = ? AND product_id = ?", hqBranch.ID, req.ProductID).
		First(&hqStock).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product not available in HQ"})
		return
	}

	if hqStock.Quantity < req.Quantity {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Insufficient stock in HQ",
			"available": hqStock.Quantity,
			"requested": req.Quantity,
		})
		return
	}

	tx := db.DB.Begin()

	// Deduct from HQ
	hqStock.Quantity -= req.Quantity
	if err := tx.Save(&hqStock).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update HQ stock"})
		return
	}

	// Add to destination branch
	var branchStock models.Stock
	err := tx.Where("branch_id = ? AND product_id = ?", req.BranchID, req.ProductID).
		First(&branchStock).Error

	if err == nil {
		// Update existing stock
		branchStock.Quantity += req.Quantity
		if err := tx.Save(&branchStock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branch stock"})
			return
		}
	} else {
		// Create new stock entry
		branchStock = models.Stock{
			BranchID:  req.BranchID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := tx.Create(&branchStock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create branch stock"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete restock"})
		return
	}

	// Load updated stock for response
	if err := db.DB.
		Preload("Product").
		Preload("Branch").
		First(&branchStock, branchStock.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated stock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":      "Restock completed successfully",
		"stock":        branchStock,
		"hq_remaining": hqStock.Quantity,
	})
}

func BulkRestockFromHQ(c *gin.Context) {
	var req BulkRestockRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(req.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one item is required"})
		return
	}

	// Verify branch exists
	var branch models.Branch
	if err := db.DB.First(&branch, req.BranchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	// Check HQ stock (assuming branch with IsHQ=true is headquarters)
	var hqBranch models.Branch
	if err := db.DB.Where("is_hq = ?", true).First(&hqBranch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Headquarters not found"})
		return
	}

	tx := db.DB.Begin()
	restockResults := []gin.H{}

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "All quantities must be positive"})
			return
		}

		// Verify product exists
		var product models.Product
		if err := tx.First(&product, "id = ?", item.ProductID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		// Check HQ stock
		var hqStock models.Stock
		if err := tx.Where("branch_id = ? AND product_id = ?", hqBranch.ID, item.ProductID).
			First(&hqStock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Product not available in HQ",
				"product_id": item.ProductID,
			})
			return
		}

		if hqStock.Quantity < item.Quantity {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Insufficient stock in HQ",
				"product_id": item.ProductID,
				"available":  hqStock.Quantity,
				"requested":  item.Quantity,
			})
			return
		}

		// Deduct from HQ
		hqStock.Quantity -= item.Quantity
		if err := tx.Save(&hqStock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update HQ stock"})
			return
		}

		// Add to destination branch
		var branchStock models.Stock
		err := tx.Where("branch_id = ? AND product_id = ?", req.BranchID, item.ProductID).
			First(&branchStock).Error

		if err == nil {
			// Update existing stock
			branchStock.Quantity += item.Quantity
			if err := tx.Save(&branchStock).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update branch stock"})
				return
			}
		} else {
			// Create new stock entry
			branchStock = models.Stock{
				BranchID:  req.BranchID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			}
			if err := tx.Create(&branchStock).Error; err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create branch stock"})
				return
			}
		}

		restockResults = append(restockResults, gin.H{
			"product_id":     item.ProductID,
			"product_name":   product.Name,
			"quantity_added": item.Quantity,
			"new_total":      branchStock.Quantity,
		})
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete bulk restock"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Bulk restock completed successfully",
		"branch":  branch.Name,
		"items":   restockResults,
	})
}

func GetHQStock(c *gin.Context) {
	// Find HQ branch
	var hqBranch models.Branch
	if err := db.DB.Where("is_hq = ?", true).First(&hqBranch).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Headquarters not found"})
		return
	}

	var stocks []models.Stock
	if err := db.DB.
		Preload("Product").
		Where("branch_id = ?", hqBranch.ID).
		Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch HQ stock"})
		return
	}

	// Calculate total value and low stock items
	var totalValue float64
	var lowStockItems []models.Stock

	for _, stock := range stocks {
		itemValue := float64(stock.Quantity) * stock.Product.Price
		totalValue += itemValue

		// Consider low stock if less than 50 items in HQ
		if stock.Quantity < 50 {
			lowStockItems = append(lowStockItems, stock)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"branch_name":     hqBranch.Name,
		"stocks":          stocks,
		"total_value":     totalValue,
		"low_stock_items": lowStockItems,
		"total_items":     len(stocks),
	})
}

func GetRestockHistory(c *gin.Context) {
	// This would typically require a separate restock_history table
	// For now, we'll return a placeholder response
	c.JSON(http.StatusOK, gin.H{
		"message": "Restock history feature coming soon",
		"data":    []interface{}{},
	})
}

func GetRestockSuggestions(c *gin.Context) {
	// Get all branches except HQ
	var branches []models.Branch
	if err := db.DB.Where("is_hq = ?", false).Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branches"})
		return
	}

	suggestions := []gin.H{}
	threshold := 10 // Low stock threshold

	for _, branch := range branches {
		var lowStockItems []models.Stock
		if err := db.DB.
			Preload("Product").
			Where("branch_id = ? AND quantity < ?", branch.ID, threshold).
			Find(&lowStockItems).Error; err != nil {
			continue
		}

		if len(lowStockItems) > 0 {
			suggestions = append(suggestions, gin.H{
				"branch_id":       branch.ID,
				"branch_name":     branch.Name,
				"low_stock_count": len(lowStockItems),
				"items":           lowStockItems,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Restock suggestions based on low stock",
		"suggestions": suggestions,
		"threshold":   threshold,
	})
}
