package controllers

import (
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

func GetBranchStock(c *gin.Context) {
	branchID := c.Param("id")

	var stock []models.Stock
	db.DB.
		Preload("Product").
		Where("branch_id = ?", branchID).
		Find(&stock)

	c.JSON(http.StatusOK, stock)
}
