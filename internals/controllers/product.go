package controllers

import (
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

func CreateProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if product.Name == "" || product.Brand == "" || product.Price <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name, brand, and positive price are required"})
		return
	}

	if err := db.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create product"})
		return
	}

	c.JSON(http.StatusCreated, product)
}

func GetAllProducts(c *gin.Context) {
	var products []models.Product
	if err := db.DB.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func GetProduct(c *gin.Context) {
	productID := c.Param("id")

	var product models.Product
	if err := db.DB.First(&product, "id = ?", productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func UpdateProduct(c *gin.Context) {
	productID := c.Param("id")

	var product models.Product
	if err := db.DB.First(&product, "id = ?", productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	var updateData models.Product
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Validate price if provided
	if updateData.Price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Price cannot be negative"})
		return
	}

	if err := db.DB.Model(&product).Updates(updateData).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product"})
		return
	}

	c.JSON(http.StatusOK, product)
}

func DeleteProduct(c *gin.Context) {
	productID := c.Param("id")

	var product models.Product
	if err := db.DB.First(&product, "id = ?", productID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Check if product has stock entries
	var stockCount int64
	db.DB.Model(&models.Stock{}).Where("product_id = ?", productID).Count(&stockCount)
	if stockCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete product with existing stock entries"})
		return
	}

	if err := db.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete product"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}

func GetProductsByBrand(c *gin.Context) {
	brand := c.Query("brand")
	if brand == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Brand parameter is required"})
		return
	}

	var products []models.Product
	if err := db.DB.Where("brand ILIKE ?", "%"+brand+"%").Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch products by brand"})
		return
	}

	c.JSON(http.StatusOK, products)
}

func GetProductStockAcrossBranches(c *gin.Context) {
	productID := c.Param("id")

	var stocks []models.Stock
	if err := db.DB.
		Preload("Branch").
		Preload("Product").
		Where("product_id = ?", productID).
		Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch product stock"})
		return
	}

	// Calculate total stock across all branches
	totalStock := 0
	for _, stock := range stocks {
		totalStock += stock.Quantity
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id":    productID,
		"total_stock":   totalStock,
		"branch_stocks": stocks,
		"branches":      len(stocks),
	})
}
