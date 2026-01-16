package controllers

import (
	"net/http"
	"strconv"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/services"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/utils"
	"github.com/gin-gonic/gin"
)

func InitiateMpesa(c *gin.Context) {
	var body struct {
		SaleID uint   `json:"sale_id" binding:"required"`
		Phone  string `json:"phone" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get sale details
	var sale models.Sale
	if err := db.DB.First(&sale, body.SaleID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Sale not found"})
		return
	}

	if sale.Status == "paid" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Sale already paid"})
		return
	}

	// Initialize MPESA service
	mpesaService := services.NewMpesaService()

	// Initiate STK Push
	response, err := mpesaService.InitiateSTKPush(body.Phone, float64(sale.Total), "SALE_"+strconv.Itoa(int(body.SaleID)))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to initiate MPESA payment"})
		return
	}

	// Update sale with payment reference
	paymentRef := "MPESA_" + response.CheckoutRequestID
	if err := db.DB.Model(&models.Sale{}).
		Where("id = ?", body.SaleID).
		Update("payment_ref", paymentRef).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sale"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":             "MPESA payment initiated successfully",
		"checkout_request_id": response.CheckoutRequestID,
		"merchant_request_id": response.MerchantRequestID,
		"payment_ref":         paymentRef,
	})
}

func MpesaWebhook(c *gin.Context) {
	var body struct {
		Body struct {
			StkCallback struct {
				MerchantRequestID string `json:"MerchantRequestID"`
				CheckoutRequestID string `json:"CheckoutRequestID"`
				ResultCode        int    `json:"ResultCode"`
				ResultDesc        string `json:"ResultDesc"`
				CallbackMetadata  struct {
					Item []struct {
						Name  string `json:"Name"`
						Value string `json:"Value"`
					} `json:"Item"`
				} `json:"CallbackMetadata"`
			} `json:"stkCallback"`
		} `json:"Body"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook format"})
		return
	}

	callback := body.Body.StkCallback
	paymentRef := "MPESA_" + callback.CheckoutRequestID

	if callback.ResultCode == 0 { // Success
		// Extract amount from metadata
		var amount float64
		var mpesaReceipt string
		for _, item := range callback.CallbackMetadata.Item {
			if item.Name == "Amount" {
				amount, _ = strconv.ParseFloat(item.Value, 64)
			}
			if item.Name == "MpesaReceiptNumber" {
				mpesaReceipt = item.Value
			}
		}

		// Update sale status to paid
		if err := db.DB.Model(&models.Sale{}).
			Where("payment_ref = ?", paymentRef).
			Updates(map[string]interface{}{
				"status":      "paid",
				"payment_ref": mpesaReceipt,
			}).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sale status"})
			return
		}

		utils.Logger.WithFields(map[string]interface{}{
			"checkout_request_id": callback.CheckoutRequestID,
			"merchant_request_id": callback.MerchantRequestID,
			"amount":              amount,
			"mpesa_receipt":       mpesaReceipt,
		}).Info("Payment completed successfully")
	} else {
		// Payment failed
		utils.Logger.WithFields(map[string]interface{}{
			"checkout_request_id": callback.CheckoutRequestID,
			"merchant_request_id": callback.MerchantRequestID,
			"result_code":         callback.ResultCode,
			"result_desc":         callback.ResultDesc,
		}).Error("Payment failed")

		// Update sale status to failed
		db.DB.Model(&models.Sale{}).
			Where("payment_ref = ?", paymentRef).
			Update("status", "failed")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Webhook processed successfully"})
}
