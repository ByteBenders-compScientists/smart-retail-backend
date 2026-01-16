package controllers

import (
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"

	"github.com/gin-gonic/gin"
)

func CreateSale(c *gin.Context) {
	branchID := c.Param("id")
	userID, _ := c.Get("userID")

	var body struct {
		Items []struct {
			ProductID string `json:"product_id" binding:"required"`
			Qty       int    `json:"qty" binding:"required"`
		} `json:"items" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(body.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one item is required"})
		return
	}

	tx := db.DB.Begin()

	total := 0.0
	sale := models.Sale{
		BranchID: branchID,
		UserID:   userID.(string),
		Status:   "pending",
	}
	if err := tx.Create(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sale"})
		return
	}

	saleItems := []models.SaleItem{}
	for _, item := range body.Items {
		if item.Qty <= 0 {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{"error": "Quantity must be positive"})
			return
		}

		var stock models.Stock
		if err := tx.Where("branch_id = ? AND product_id = ?", branchID, item.ProductID).
			First(&stock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found in branch inventory"})
			return
		}

		if stock.Quantity < item.Qty {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error":      "Insufficient stock",
				"product_id": item.ProductID,
				"available":  stock.Quantity,
				"requested":  item.Qty,
			})
			return
		}

		// Update stock
		stock.Quantity -= item.Qty
		if err := tx.Save(&stock).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update stock"})
			return
		}

		var product models.Product
		if err := tx.First(&product, item.ProductID).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
			return
		}

		line := models.SaleItem{
			SaleID:    sale.ID,
			ProductID: product.ID,
			Qty:       item.Qty,
			Price:     product.Price,
		}
		saleItems = append(saleItems, line)

		total += product.Price * float64(item.Qty)
	}

	// Create sale items
	for _, item := range saleItems {
		if err := tx.Create(&item).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create sale items"})
			return
		}
	}

	// Update sale total
	sale.Total = int(total)
	if err := tx.Save(&sale).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sale total"})
		return
	}

	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete sale"})
		return
	}

	// Load sale with items for response
	if err := db.DB.
		Preload("SaleItems.Product").
		Preload("Branch").
		First(&sale, sale.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load sale details"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Sale created successfully",
		"sale":    sale,
	})
}

func GetSale(c *gin.Context) {
	saleID := c.Param("saleId")

	var sale models.Sale
	if err := db.DB.
		Preload("SaleItems.Product").
		Preload("Branch").
		Preload("User").
		First(&sale, saleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sale not found"})
		return
	}

	c.JSON(http.StatusOK, sale)
}

func GetBranchSales(c *gin.Context) {
	branchID := c.Param("id")

	var sales []models.Sale
	if err := db.DB.
		Preload("SaleItems.Product").
		Preload("User").
		Where("branch_id = ?", branchID).
		Order("created_at DESC").
		Find(&sales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sales"})
		return
	}

	c.JSON(http.StatusOK, sales)
}

func GetAllSales(c *gin.Context) {
	userID, _ := c.Get("userID")
	userRole, _ := c.Get("userRole")

	var sales []models.Sale
	query := db.DB.
		Preload("SaleItems.Product").
		Preload("Branch").
		Preload("User").
		Order("created_at DESC")

	// Filter by user if customer
	if userRole == "customer" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&sales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sales"})
		return
	}

	c.JSON(http.StatusOK, sales)
}

func UpdateSaleStatus(c *gin.Context) {
	saleID := c.Param("saleId")

	var body struct {
		Status string `json:"status" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate status
	validStatuses := map[string]bool{"pending": true, "paid": true, "failed": true, "cancelled": true}
	if !validStatuses[body.Status] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	var sale models.Sale
	if err := db.DB.First(&sale, saleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sale not found"})
		return
	}

	sale.Status = body.Status
	if err := db.DB.Save(&sale).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sale status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sale status updated successfully",
		"sale":    sale,
	})
}

func GetSalesByStatus(c *gin.Context) {
	status := c.Query("status")
	if status == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Status parameter is required"})
		return
	}

	var sales []models.Sale
	if err := db.DB.
		Preload("SaleItems.Product").
		Preload("Branch").
		Preload("User").
		Where("status = ?", status).
		Order("created_at DESC").
		Find(&sales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sales"})
		return
	}

	c.JSON(http.StatusOK, sales)
}
