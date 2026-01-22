// orders controller
package controllers

import (
	"net/http"
	"time"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

func CreateOrder(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var body struct {
		BranchID   string `json:"branchId" binding:"required"`
		Items      []struct {
			ProductID    string  `json:"productId" binding:"required"`
			ProductBrand string  `json:"productBrand" binding:"required"`
			Quantity     int     `json:"quantity" binding:"required,min=1"`
			Price        float64 `json:"price" binding:"required,min=0"`
			Subtotal     float64 `json:"subtotal" binding:"required,min=0"`
		} `json:"items" binding:"required,min=1"`
		TotalAmount float64 `json:"totalAmount" binding:"required,min=0"`
		Phone       string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	tx := db.DB.Begin()

	// Create order
	order := models.Order{
		UserID:        userID.(string),
		BranchID:      body.BranchID,
		TotalAmount:   body.TotalAmount,
		PaymentStatus: "pending",
		PaymentMethod: "mpesa",
		OrderStatus:   "processing",
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	// Create order items
	orderItems := []models.OrderItem{}
	for _, item := range body.Items {
		orderItem := models.OrderItem{
			OrderID:      order.ID,
			ProductID:    item.ProductID,
			ProductBrand: item.ProductBrand,
			Quantity:     item.Quantity,
			Price:        item.Price,
			Subtotal:     item.Subtotal,
		}
		orderItems = append(orderItems, orderItem)
	}

	if err := tx.Create(&orderItems).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order items"})
		return
	}

	// Create payment record
	payment := models.Payment{
		OrderID: order.ID,
		Phone:   body.Phone,
		Amount:  body.TotalAmount,
		Status:  "pending",
	}

	if err := tx.Create(&payment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment record"})
		return
	}

	tx.Commit()

	c.JSON(http.StatusCreated, gin.H{
		"order": order,
		"paymentUrl": "/api/payments/mpesa/initiate",
	})
}

func GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var orders []models.Order
	if err := db.DB.Where("user_id = ?", userID).
		Preload("Branch").
		Preload("OrderItems.Product").
		Order("created_at DESC").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

func GetOrderById(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orderID := c.Param("id")
	
	var order models.Order
	if err := db.DB.Where("id = ? AND user_id = ?", orderID, userID).
		Preload("Branch").
		Preload("OrderItems.Product").
		First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func UpdateOrderStatus(c *gin.Context) {
	orderID := c.Param("id")
	
	var body struct {
		Status string `json:"status" binding:"required,oneof=processing completed cancelled"`
	}
	
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid status"})
		return
	}

	var order models.Order
	if err := db.DB.Where("id = ?", orderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	updates := map[string]interface{}{
		"order_status": body.Status,
	}
	
	if body.Status == "completed" {
		now := time.Now()
		updates["completed_at"] = &now
		updates["payment_status"] = "completed"
	}

	if err := db.DB.Model(&order).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order updated successfully"})
}
