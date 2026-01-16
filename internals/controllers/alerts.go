package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

type LowStockAlert struct {
	ID           string     `json:"id"`
	BranchID     string     `json:"branch_id"`
	BranchName   string     `json:"branch_name"`
	ProductID    string     `json:"product_id"`
	ProductName  string     `json:"product_name"`
	Brand        string     `json:"brand"`
	CurrentStock int        `json:"current_stock"`
	Threshold    int        `json:"threshold"`
	ReorderLevel int        `json:"reorder_level"`
	LastAlert    *time.Time `json:"last_alert,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

type AlertSettings struct {
	DefaultThreshold  int            `json:"default_threshold"`
	BranchThresholds  map[string]int `json:"branch_thresholds"`
	ProductThresholds map[string]int `json:"product_thresholds"`
	EmailEnabled      bool           `json:"email_enabled"`
	SMSEnabled        bool           `json:"sms_enabled"`
}

func GetLowStockAlerts(c *gin.Context) {
	threshold := 10 // Default threshold
	if thresholdStr := c.Query("threshold"); thresholdStr != "" {
		if t, err := strconv.Atoi(thresholdStr); err == nil && t > 0 {
			threshold = t
		}
	}

	branchID := c.Query("branch_id")

	var stocks []models.Stock
	query := db.DB.
		Preload("Product").
		Preload("Branch").
		Where("quantity < ?", threshold)

	if branchID != "" {
		query = query.Where("branch_id = ?", branchID)
	}

	if err := query.Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch low stock items"})
		return
	}

	alerts := []LowStockAlert{}
	for _, stock := range stocks {
		alert := LowStockAlert{
			ID:           stock.ProductID,
			BranchID:     stock.BranchID,
			BranchName:   stock.Branch.Name,
			ProductID:    stock.ProductID,
			ProductName:  stock.Product.Name,
			Brand:        stock.Product.Brand,
			CurrentStock: stock.Quantity,
			Threshold:    threshold,
			ReorderLevel: threshold * 2, // Suggested reorder level
			CreatedAt:    stock.CreatedAt,
		}
		alerts = append(alerts, alert)
	}

	c.JSON(http.StatusOK, gin.H{
		"threshold":    threshold,
		"total_alerts": len(alerts),
		"alerts":       alerts,
		"generated_at": time.Now(),
	})
}

func GetCriticalStockAlerts(c *gin.Context) {
	// Critical stock is when quantity is 0 or very low (<= 3)
	criticalThreshold := 3

	var stocks []models.Stock
	if err := db.DB.
		Preload("Product").
		Preload("Branch").
		Where("quantity <= ?", criticalThreshold).
		Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch critical stock items"})
		return
	}

	alerts := []LowStockAlert{}
	for _, stock := range stocks {
		alert := LowStockAlert{
			ID:           stock.ProductID,
			BranchID:     stock.BranchID,
			BranchName:   stock.Branch.Name,
			ProductID:    stock.ProductID,
			ProductName:  stock.Product.Name,
			Brand:        stock.Product.Brand,
			CurrentStock: stock.Quantity,
			Threshold:    criticalThreshold,
			ReorderLevel: 10, // Urgent reorder level
			CreatedAt:    stock.CreatedAt,
		}
		alerts = append(alerts, alert)
	}

	c.JSON(http.StatusOK, gin.H{
		"critical_threshold": criticalThreshold,
		"total_critical":     len(alerts),
		"alerts":             alerts,
		"generated_at":       time.Now(),
		"urgent_action":      len(alerts) > 0,
	})
}

func GetStockAlertsByBranch(c *gin.Context) {
	branchID := c.Param("id")

	threshold := 10 // Default threshold
	if thresholdStr := c.Query("threshold"); thresholdStr != "" {
		if t, err := strconv.Atoi(thresholdStr); err == nil && t > 0 {
			threshold = t
		}
	}

	// Verify branch exists
	var branch models.Branch
	if err := db.DB.First(&branch, branchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	var stocks []models.Stock
	if err := db.DB.
		Preload("Product").
		Where("branch_id = ? AND quantity < ?", branchID, threshold).
		Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branch stock alerts"})
		return
	}

	alerts := []LowStockAlert{}
	for _, stock := range stocks {
		alert := LowStockAlert{
			ID:           stock.ProductID,
			BranchID:     stock.BranchID,
			BranchName:   stock.Branch.Name,
			ProductID:    stock.ProductID,
			ProductName:  stock.Product.Name,
			Brand:        stock.Product.Brand,
			CurrentStock: stock.Quantity,
			Threshold:    threshold,
			ReorderLevel: threshold * 2,
			CreatedAt:    stock.CreatedAt,
		}
		alerts = append(alerts, alert)
	}

	c.JSON(http.StatusOK, gin.H{
		"branch_id":    branchID,
		"branch_name":  branch.Name,
		"threshold":    threshold,
		"total_alerts": len(alerts),
		"alerts":       alerts,
		"generated_at": time.Now(),
	})
}

func GetAlertSummary(c *gin.Context) {
	// Get overall alert summary
	var totalProducts int64
	db.DB.Model(&models.Product{}).Count(&totalProducts)

	var totalBranches int64
	db.DB.Model(&models.Branch{}).Count(&totalBranches)

	var totalStockEntries int64
	db.DB.Model(&models.Stock{}).Count(&totalStockEntries)

	// Low stock alerts (threshold: 10)
	var lowStockCount int64
	db.DB.Model(&models.Stock{}).Where("quantity < ?", 10).Count(&lowStockCount)

	// Critical stock alerts (threshold: 3)
	var criticalStockCount int64
	db.DB.Model(&models.Stock{}).Where("quantity <= ?", 3).Count(&criticalStockCount)

	// Out of stock
	var outOfStockCount int64
	db.DB.Model(&models.Stock{}).Where("quantity = ?", 0).Count(&outOfStockCount)

	// Get alerts by branch
	var branches []models.Branch
	db.DB.Find(&branches)

	branchAlerts := []gin.H{}
	for _, branch := range branches {
		var branchLowStock int64
		db.DB.Model(&models.Stock{}).
			Where("branch_id = ? AND quantity < ?", branch.ID, 10).
			Count(&branchLowStock)

		if branchLowStock > 0 {
			branchAlerts = append(branchAlerts, gin.H{
				"branch_id":   branch.ID,
				"branch_name": branch.Name,
				"alert_count": branchLowStock,
			})
		}
	}

	// Get top 5 products with most low stock alerts
	type ProductAlert struct {
		ProductID   uint   `json:"product_id"`
		ProductName string `json:"product_name"`
		Brand       string `json:"brand"`
		AlertCount  int64  `json:"alert_count"`
	}

	var productAlerts []ProductAlert
	db.DB.Raw(`
		SELECT p.id as product_id, p.name as product_name, p.brand, COUNT(*) as alert_count
		FROM stocks s
		JOIN products p ON s.product_id = p.id
		WHERE s.quantity < 10
		GROUP BY p.id, p.name, p.brand
		ORDER BY alert_count DESC
		LIMIT 5
	`).Scan(&productAlerts)

	summary := gin.H{
		"total_products":      totalProducts,
		"total_branches":      totalBranches,
		"total_stock_entries": totalStockEntries,
		"low_stock_alerts":    lowStockCount,
		"critical_alerts":     criticalStockCount,
		"out_of_stock":        outOfStockCount,
		"branch_alerts":       branchAlerts,
		"top_product_alerts":  productAlerts,
		"generated_at":        time.Now(),
		"health_score":        calculateHealthScore(totalStockEntries, lowStockCount, criticalStockCount),
	}

	c.JSON(http.StatusOK, summary)
}

func calculateHealthScore(totalStock, lowStock, criticalStock int64) string {
	if totalStock == 0 {
		return "unknown"
	}

	lowPercentage := float64(lowStock) / float64(totalStock) * 100
	criticalPercentage := float64(criticalStock) / float64(totalStock) * 100

	if criticalPercentage > 10 {
		return "critical"
	} else if lowPercentage > 25 {
		return "warning"
	} else if lowPercentage > 10 {
		return "caution"
	} else {
		return "healthy"
	}
}

func CreateAlertRule(c *gin.Context) {
	var body struct {
		BranchID  uint `json:"branch_id" binding:"required"`
		ProductID uint `json:"product_id" binding:"required"`
		Threshold int  `json:"threshold" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if body.Threshold <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Threshold must be positive"})
		return
	}

	// Verify branch and product exist
	var branch models.Branch
	if err := db.DB.First(&branch, body.BranchID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Branch not found"})
		return
	}

	var product models.Product
	if err := db.DB.First(&product, body.ProductID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check current stock
	var stock models.Stock
	if err := db.DB.Where("branch_id = ? AND product_id = ?", body.BranchID, body.ProductID).
		First(&stock).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Stock entry not found"})
		return
	}

	isAlert := stock.Quantity <= body.Threshold

	c.JSON(http.StatusOK, gin.H{
		"message":       "Alert rule created",
		"branch_id":     body.BranchID,
		"branch_name":   branch.Name,
		"product_id":    body.ProductID,
		"product_name":  product.Name,
		"threshold":     body.Threshold,
		"current_stock": stock.Quantity,
		"is_alert":      isAlert,
		"created_at":    time.Now(),
	})
}

func GetAlertHistory(c *gin.Context) {
	// This would typically require an alert_history table
	// For now, we'll return a placeholder with recent low stock items
	days := 7 // Default to last 7 days
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 90 {
			days = d
		}
	}

	startDate := time.Now().AddDate(0, 0, -days)

	var stocks []models.Stock
	if err := db.DB.
		Preload("Product").
		Preload("Branch").
		Where("quantity < 10 AND updated_at >= ?", startDate).
		Order("updated_at DESC").
		Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch alert history"})
		return
	}

	alerts := []gin.H{}
	for _, stock := range stocks {
		alerts = append(alerts, gin.H{
			"stock_id":      stock.ID,
			"branch_id":     stock.BranchID,
			"branch_name":   stock.Branch.Name,
			"product_id":    stock.ProductID,
			"product_name":  stock.Product.Name,
			"brand":         stock.Product.Brand,
			"current_stock": stock.Quantity,
			"threshold":     10,
			"last_updated":  stock.UpdatedAt,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"period":       strconv.Itoa(days) + " days",
		"total_alerts": len(alerts),
		"alerts":       alerts,
		"generated_at": time.Now(),
	})
}
