// payments controller
package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/services"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/utils"
	"github.com/gin-gonic/gin"
)

type MpesaInitiateRequest struct {
	OrderID string  `json:"orderId" binding:"required"`
	Phone   string  `json:"phone" binding:"required"`
	Amount  float64 `json:"amount" binding:"required,min=1"`
}

type MpesaCallbackPayload struct {
	Body struct {
		StkCallback struct {
			MerchantRequestID string `json:"MerchantRequestID"`
			CheckoutRequestID string `json:"CheckoutRequestID"`
			ResultCode        int    `json:"ResultCode"`
			ResultDesc        string `json:"ResultDesc"`
			CallbackMetadata  struct {
				Item []struct {
					Name  string      `json:"Name"`
					Value interface{} `json:"Value"`
				} `json:"Item"`
			} `json:"CallbackMetadata"`
		} `json:"stkCallback"`
	} `json:"Body"`
}

func InitiateMpesaPayment(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var req MpesaInitiateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Verify order belongs to user
	var order models.Order
	if err := db.DB.Where("id = ? AND user_id = ?", req.OrderID, userID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	// Check if payment already exists and is completed
	var payment models.Payment
	if err := db.DB.Where("order_id = ?", req.OrderID).First(&payment).Error; err == nil {
		if payment.Status == "completed" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Payment already completed"})
			return
		}
	}

	// Initialize MPESA service
	mpesaService := services.NewMpesaService()

	// Validate MPESA configuration early so we fail fast with a clear message
	if mpesaService.ConsumerKey == "" || mpesaService.ConsumerSecret == "" || mpesaService.Shortcode == "" || mpesaService.Passkey == "" || mpesaService.CallbackURL == "" {
		utils.Logger.Error("MPESA configuration is incomplete")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "MPESA configuration is missing. Please set MPESA_CONSUMER_KEY, MPESA_CONSUMER_SECRET, MPESA_SHORTCODE, MPESA_PASSKEY, and MPESA_CALLBACK_URL"})
		return
	}

	utils.Logger.WithFields(map[string]interface{}{
		"order_id":     req.OrderID,
		"phone":        req.Phone,
		"amount":       req.Amount,
		"callback_url": mpesaService.CallbackURL,
	}).Info("Initiating M-Pesa STK Push")

	// Initiate STK Push using order reference
	response, err := mpesaService.InitiateSTKPush(req.Phone, req.Amount, "ORDER_"+req.OrderID)
	if err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"order_id": req.OrderID,
			"error":    err.Error(),
		}).Error("Failed to initiate MPESA payment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate MPESA payment"})
		return
	}

	// Create or update payment record
	transactionID := "MPESA_" + response.CheckoutRequestID
	checkoutRequestID := response.CheckoutRequestID

	if payment.ID == "" {
		// Create new payment record
		payment = models.Payment{
			OrderID:           req.OrderID,
			Phone:             req.Phone,
			Amount:            req.Amount,
			TransactionID:     &transactionID,
			CheckoutRequestID: &checkoutRequestID,
			Status:            "pending",
		}
		if err := db.DB.Create(&payment).Error; err != nil {
			utils.Logger.WithFields(map[string]interface{}{
				"order_id":            req.OrderID,
				"checkout_request_id": checkoutRequestID,
				"error":               err.Error(),
			}).Error("Failed to create payment record")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment record"})
			return
		}
	} else {
		// Update existing payment record
		if err := db.DB.Model(&payment).Updates(map[string]interface{}{
			"transaction_id":      &transactionID,
			"checkout_request_id": &checkoutRequestID,
			"status":              "pending",
		}).Error; err != nil {
			utils.Logger.WithFields(map[string]interface{}{
				"order_id":            req.OrderID,
				"checkout_request_id": checkoutRequestID,
				"error":               err.Error(),
			}).Error("Failed to update payment record")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update payment record"})
			return
		}
	}

	utils.Logger.WithFields(map[string]interface{}{
		"order_id":            req.OrderID,
		"checkout_request_id": checkoutRequestID,
		"merchant_request_id": response.MerchantRequestID,
	}).Info("Payment initiated successfully, awaiting callback")

	c.JSON(http.StatusOK, gin.H{
		"success":           true,
		"message":           "Payment initiated successfully",
		"transactionId":     transactionID,
		"checkoutRequestId": checkoutRequestID,
		"merchantRequestId": response.MerchantRequestID,
	})
}

func MpesaCallback(c *gin.Context) {
	// Log incoming callback
	utils.Logger.Info("M-Pesa callback received")

	var callback MpesaCallbackPayload
	if err := c.ShouldBindJSON(&callback); err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"error": err.Error(),
		}).Error("Failed to parse M-Pesa callback payload")
		// ALWAYS return 200 OK to M-Pesa to prevent retries
		c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
		return
	}

	stkCallback := callback.Body.StkCallback
	checkoutRequestID := stkCallback.CheckoutRequestID
	resultCode := stkCallback.ResultCode

	utils.Logger.WithFields(map[string]interface{}{
		"checkout_request_id": checkoutRequestID,
		"merchant_request_id": stkCallback.MerchantRequestID,
		"result_code":         resultCode,
		"result_desc":         stkCallback.ResultDesc,
	}).Info("Processing M-Pesa callback")

	// Find payment by checkout request ID
	var payment models.Payment
	if err := db.DB.Where("checkout_request_id = ?", checkoutRequestID).First(&payment).Error; err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"checkout_request_id": checkoutRequestID,
			"error":               err.Error(),
		}).Error("Payment not found for CheckoutRequestID")
		// ALWAYS return 200 OK to M-Pesa
		c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
		return
	}

	tx := db.DB.Begin()

	// Update payment status
	paymentStatus := "failed"
	mpesaReceipt := ""

	if resultCode == 0 { // Success
		paymentStatus = "completed"

		// Extract amount and receipt from metadata - safe type conversion
		for _, item := range stkCallback.CallbackMetadata.Item {
			if item.Name == "MpesaReceiptNumber" {
				// M-Pesa can send Value as string or number, handle both
				switch v := item.Value.(type) {
				case string:
					mpesaReceipt = v
				case float64:
					mpesaReceipt = fmt.Sprintf("%.0f", v)
				default:
					mpesaReceipt = fmt.Sprintf("%v", v)
				}
				break
			}
		}

		utils.Logger.WithFields(map[string]interface{}{
			"checkout_request_id": checkoutRequestID,
			"mpesa_receipt":       mpesaReceipt,
			"order_id":            payment.OrderID,
		}).Info("Payment successful, extracted receipt number")
	}

	// Store M-Pesa response
	responseJSON, _ := json.Marshal(callback)
	responseStr := string(responseJSON)

	if err := tx.Model(&payment).Updates(map[string]interface{}{
		"status":         paymentStatus,
		"mpesa_response": &responseStr,
	}).Error; err != nil {
		tx.Rollback()
		utils.Logger.WithFields(map[string]interface{}{
			"checkout_request_id": checkoutRequestID,
			"order_id":            payment.OrderID,
			"error":               err.Error(),
		}).Error("Failed to update payment status")
		// ALWAYS return 200 OK to M-Pesa
		c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
		return
	}

	// Update order status if payment completed
	if resultCode == 0 {
		// Update transaction ID with MPesa receipt
		if err := tx.Model(&payment).Update("transaction_id", mpesaReceipt).Error; err != nil {
			tx.Rollback()
			utils.Logger.WithFields(map[string]interface{}{
				"checkout_request_id": checkoutRequestID,
				"order_id":            payment.OrderID,
				"error":               err.Error(),
			}).Error("Failed to update transaction ID")
			// ALWAYS return 200 OK to M-Pesa
			c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
			return
		}

		if err := tx.Model(&models.Order{}).Where("id = ?", payment.OrderID).Updates(map[string]interface{}{
			"payment_status":       "completed",
			"order_status":         "completed",
			"mpesa_transaction_id": &mpesaReceipt,
		}).Error; err != nil {
			tx.Rollback()
			utils.Logger.WithFields(map[string]interface{}{
				"checkout_request_id": checkoutRequestID,
				"order_id":            payment.OrderID,
				"error":               err.Error(),
			}).Error("Failed to update order status")
			// ALWAYS return 200 OK to M-Pesa
			c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
			return
		}

		utils.Logger.WithFields(map[string]interface{}{
			"checkout_request_id": checkoutRequestID,
			"merchant_request_id": stkCallback.MerchantRequestID,
			"mpesa_receipt":       mpesaReceipt,
			"order_id":            payment.OrderID,
		}).Info("Payment completed successfully")
	} else {
		utils.Logger.WithFields(map[string]interface{}{
			"checkout_request_id": checkoutRequestID,
			"merchant_request_id": stkCallback.MerchantRequestID,
			"result_code":         resultCode,
			"result_desc":         stkCallback.ResultDesc,
			"order_id":            payment.OrderID,
		}).Error("Payment failed")

		// Update order status to failed
		if err := tx.Model(&models.Order{}).Where("id = ?", payment.OrderID).Update("order_status", "cancelled").Error; err != nil {
			tx.Rollback()
			utils.Logger.WithFields(map[string]interface{}{
				"checkout_request_id": checkoutRequestID,
				"order_id":            payment.OrderID,
				"error":               err.Error(),
			}).Error("Failed to update order status to cancelled")
			// ALWAYS return 200 OK to M-Pesa
			c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
			return
		}
	}

	if err := tx.Commit().Error; err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"checkout_request_id": checkoutRequestID,
			"order_id":            payment.OrderID,
			"error":               err.Error(),
		}).Error("Failed to commit transaction")
		// ALWAYS return 200 OK to M-Pesa
		c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
		return
	}

	utils.Logger.WithFields(map[string]interface{}{
		"checkout_request_id": checkoutRequestID,
		"order_id":            payment.OrderID,
		"status":              paymentStatus,
	}).Info("M-Pesa callback processed successfully")

	c.JSON(http.StatusOK, gin.H{"ResultCode": 0, "ResultDesc": "Accepted"})
}

func GetPaymentStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	orderID := c.Param("orderId")

	// Single joined lookup to verify ownership and fetch payment fields with minimal columns
	var result struct {
		Status        string  `json:"status"`
		TransactionID *string `json:"transactionId"`
	}

	if err := db.DB.Table("payments").
		Select("payments.status, payments.transaction_id").
		Joins("JOIN orders ON orders.id = payments.order_id").
		Where("payments.order_id = ? AND orders.user_id = ? AND orders.deleted_at IS NULL", orderID, userID).
		Take(&result).Error; err != nil {
		utils.Logger.WithFields(map[string]interface{}{
			"order_id": orderID,
			"user_id":  userID,
		}).Error("Payment not found for status check")
		c.JSON(http.StatusNotFound, gin.H{"error": "Payment not found"})
		return
	}

	utils.Logger.WithFields(map[string]interface{}{
		"order_id": orderID,
		"status":   result.Status,
	}).Info("Payment status checked")

	response := gin.H{
		"status": result.Status,
	}

	if result.TransactionID != nil {
		response["transactionId"] = *result.TransactionID
	}

	c.JSON(http.StatusOK, response)
}
