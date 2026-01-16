package controllers

import (
	"net/http"
	"time"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/utils"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

type OfflineSyncRequest struct {
	ClientID string               `json:"client_id" binding:"required"`
	Sales    []OfflineSaleRequest `json:"sales" binding:"required"`
}

type OfflineSaleRequest struct {
	ClientTxnID string              `json:"client_txn_id" binding:"required"`
	BranchID    string              `json:"branch_id" binding:"required"`
	Items       []OfflineSaleItem   `json:"items" binding:"required"`
	Total       int                 `json:"total" binding:"required"`
	CreatedAt   time.Time           `json:"created_at"`
	PaymentInfo *OfflinePaymentInfo `json:"payment_info,omitempty"`
}

type OfflineSaleItem struct {
	ProductID string  `json:"product_id" binding:"required"`
	Qty       int     `json:"qty" binding:"required"`
	Price     float64 `json:"price" binding:"required"`
}

type OfflinePaymentInfo struct {
	Method    string `json:"method"` // cash, mpesa, card
	Reference string `json:"reference,omitempty"`
	Phone     string `json:"phone,omitempty"`
	Status    string `json:"status"` // pending, completed, failed
}

type SyncResult struct {
	ClientTxnID string `json:"client_txn_id"`
	Status      string `json:"status"` // synced, duplicate, failed, insufficient_stock
	ServerID    string `json:"server_id,omitempty"`
	Error       string `json:"error,omitempty"`
}

func SyncOffline(c *gin.Context) {
	var req OfflineSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if len(req.Sales) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No sales to sync"})
		return
	}

	logger := utils.Logger.WithFields(logrus.Fields{
		"client_id":   req.ClientID,
		"sales_count": len(req.Sales),
	})

	results := []SyncResult{}
	successCount := 0
	duplicateCount := 0
	failedCount := 0

	for _, offlineSale := range req.Sales {
		result := SyncResult{
			ClientTxnID: offlineSale.ClientTxnID,
		}

		// Check for duplicates using client transaction ID
		var existingSale models.Sale
		if err := db.DB.Where("payment_ref = ?", offlineSale.ClientTxnID).First(&existingSale).Error; err == nil {
			result.Status = "duplicate"
			result.ServerID = existingSale.ID
			duplicateCount++
			results = append(results, result)
			continue
		}

		// Validate sale data
		if offlineSale.BranchID == "" || len(offlineSale.Items) == 0 {
			result.Status = "failed"
			result.Error = "Invalid sale data"
			failedCount++
			results = append(results, result)
			continue
		}

		// Process sale in transaction
		tx := db.DB.Begin()

		// Create sale record
		sale := models.Sale{
			BranchID:   offlineSale.BranchID,
			UserID:     "", // Will be set from authenticated user
			Total:      offlineSale.Total,
			Status:     "pending",
			PaymentRef: offlineSale.ClientTxnID,
		}

		// Set user ID from authenticated context
		if userID, exists := c.Get("userID"); exists {
			if userIDStr, ok := userID.(string); ok {
				sale.UserID = userIDStr
			}
		}

		if err := tx.Create(&sale).Error; err != nil {
			tx.Rollback()
			result.Status = "failed"
			result.Error = "Failed to create sale"
			failedCount++
			results = append(results, result)
			continue
		}

		// Process sale items and update stock
		calculatedTotal := 0.0
		for _, item := range offlineSale.Items {
			// Check stock availability
			var stock models.Stock
			if err := tx.Where("branch_id = ? AND product_id = ?", offlineSale.BranchID, item.ProductID).
				First(&stock).Error; err != nil {
				tx.Rollback()
				result.Status = "insufficient_stock"
				result.Error = "Product not found in inventory"
				failedCount++
				results = append(results, result)
				continue
			}

			if stock.Quantity < item.Qty {
				tx.Rollback()
				result.Status = "insufficient_stock"
				result.Error = "Insufficient stock for product"
				failedCount++
				results = append(results, result)
				continue
			}

			// Update stock
			stock.Quantity -= item.Qty
			if err := tx.Save(&stock).Error; err != nil {
				tx.Rollback()
				result.Status = "failed"
				result.Error = "Failed to update stock"
				failedCount++
				results = append(results, result)
				continue
			}

			// Create sale item
			saleItem := models.SaleItem{
				SaleID: sale.ID,
				Price:  item.Price,
			}

			if err := tx.Create(&saleItem).Error; err != nil {
				tx.Rollback()
				result.Status = "failed"
				result.Error = "Failed to create sale item"
				failedCount++
				results = append(results, result)
				continue
			}

			calculatedTotal += item.Price * float64(item.Qty)
		}

		// Verify total matches calculated total
		if calculatedTotal != float64(offlineSale.Total) {
			tx.Rollback()
			result.Status = "failed"
			result.Error = "Total amount mismatch"
			failedCount++
			results = append(results, result)
			continue
		}

		// Handle payment info if provided
		if offlineSale.PaymentInfo != nil {
			switch offlineSale.PaymentInfo.Status {
			case "completed":
				sale.Status = "paid"
				sale.PaymentRef = offlineSale.PaymentInfo.Reference
			case "failed":
				sale.Status = "failed"
			default:
				sale.Status = "pending"
			}
		}

		// Update sale with final status
		if err := tx.Save(&sale).Error; err != nil {
			tx.Rollback()
			result.Status = "failed"
			result.Error = "Failed to update sale status"
			failedCount++
			results = append(results, result)
			continue
		}

		if err := tx.Commit().Error; err != nil {
			result.Status = "failed"
			result.Error = "Transaction failed"
			failedCount++
			results = append(results, result)
			continue
		}

		result.Status = "synced"
		result.ServerID = sale.ID
		successCount++
		results = append(results, result)
	}

	logger.WithFields(logrus.Fields{
		"success_count":   successCount,
		"duplicate_count": duplicateCount,
		"failed_count":    failedCount,
	}).Info("Offline sync completed")

	c.JSON(http.StatusOK, gin.H{
		"message": "Sync completed",
		"summary": gin.H{
			"total":     len(req.Sales),
			"synced":    successCount,
			"duplicate": duplicateCount,
			"failed":    failedCount,
		},
		"results": results,
	})
}

func GetSyncStatus(c *gin.Context) {
	clientID := c.Query("client_id")
	if clientID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "client_id parameter is required"})
		return
	}

	// Get all sales synced by this client
	var sales []models.Sale
	if err := db.DB.
		Preload("SaleItems.Product").
		Preload("Branch").
		Where("payment_ref LIKE ?", clientID+"%").
		Order("created_at DESC").
		Find(&sales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sync status"})
		return
	}

	// Group by status
	statusCounts := make(map[string]int)
	for _, sale := range sales {
		statusCounts[sale.Status]++
	}

	c.JSON(http.StatusOK, gin.H{
		"client_id":     clientID,
		"total_synced":  len(sales),
		"status_counts": statusCounts,
		"recent_sales":  sales,
	})
}

func GetPendingSync(c *gin.Context) {
	// Get all pending sales that might need payment confirmation
	var sales []models.Sale
	if err := db.DB.
		Preload("SaleItems.Product").
		Preload("Branch").
		Preload("User").
		Where("status = ?", "pending").
		Order("created_at DESC").
		Find(&sales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pending sales"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"pending_count": len(sales),
		"pending_sales": sales,
	})
}

func ResolveSyncConflict(c *gin.Context) {
	saleID := c.Param("saleId")

	var body struct {
		Action string `json:"action" binding:"required"` // approve, reject, delete
		Reason string `json:"reason,omitempty"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var sale models.Sale
	if err := db.DB.First(&sale, "id = ?", saleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sale not found"})
		return
	}

	switch body.Action {
	case "approve":
		sale.Status = "paid"
	case "reject":
		sale.Status = "failed"
	case "delete":
		// Need to restore stock before deleting
		tx := db.DB.Begin()

		// Get sale items and restore stock
		var saleItems []models.SaleItem
		if err := tx.Where("sale_id = ?", saleID).Find(&saleItems).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get sale items"})
			return
		}

		for _, item := range saleItems {
			var stock models.Stock
			if err := tx.Where("branch_id = ? AND product_id = ?", sale.BranchID, item.ProductID).
				First(&stock).Error; err == nil {
				stock.Quantity += item.Qty
				tx.Save(&stock)
			}
		}

		// Delete sale items and sale
		if err := tx.Where("sale_id = ?", saleID).Delete(&models.SaleItem{}).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sale items"})
			return
		}

		if err := tx.Delete(&sale).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete sale"})
			return
		}

		if err := tx.Commit().Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete deletion"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Sale deleted and stock restored"})
		return
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}

	if err := db.DB.Save(&sale).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sale"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sale status updated",
		"sale":    sale,
	})
}
