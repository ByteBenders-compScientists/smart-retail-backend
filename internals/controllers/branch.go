package controllers

import (
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

func CreateBranch(c *gin.Context) {
	var body struct {
		ID            string `json:"id" binding:"required"`
		Name          string `json:"name" binding:"required,oneof=Nairobi Kisumu Mombasa Nakuru Eldoret"`
		IsHeadquarter bool   `json:"isHeadquarter"`
		Address       string `json:"address" binding:"required"`
		Phone         string `json:"phone" binding:"required"`
		Status        string `json:"status"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	branch := models.Branch{
		ID:            body.ID,
		Name:          body.Name,
		IsHeadquarter: body.IsHeadquarter,
		Address:       body.Address,
		Phone:         body.Phone,
		Status:        "active",
	}

	if body.Status != "" {
		branch.Status = body.Status
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
