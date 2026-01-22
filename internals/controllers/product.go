package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

func CreateProduct(c *gin.Context) {
	var body struct {
		Name          string   `json:"name" binding:"required"`
		Brand         string   `json:"brand" binding:"required,oneof=Coke Fanta Sprite"`
		Description   string   `json:"description"`
		Price         float64  `json:"price" binding:"required,min=0"`
		OriginalPrice float64  `json:"originalPrice" binding:"required,min=0"`
		Image         string   `json:"image"`
		Rating        float64  `json:"rating"`
		Reviews       int      `json:"reviews"`
		Category      string   `json:"category"`
		Volume        string   `json:"volume"`
		Unit          string   `json:"unit"`
		Tags          []string `json:"tags"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Convert tags to JSON
	tagsJSON := "[]"
	if len(body.Tags) > 0 {
		tagsBytes, err := json.Marshal(body.Tags)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tags format"})
			return
		}
		tagsJSON = string(tagsBytes)
	}

	product := models.Product{
		Name:          body.Name,
		Brand:         body.Brand,
		Description:   body.Description,
		Price:         body.Price,
		OriginalPrice: body.OriginalPrice,
		Image:         body.Image,
		Rating:        body.Rating,
		Reviews:       body.Reviews,
		Category:      "Soft Drinks",
		Volume:        "500ml",
		Unit:          "single",
		Tags:          tagsJSON,
	}

	if body.Category != "" {
		product.Category = body.Category
	}
	if body.Volume != "" {
		product.Volume = body.Volume
	}
	if body.Unit != "" {
		product.Unit = body.Unit
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

	// Check if product has inventory entries
	var inventoryCount int64
	db.DB.Model(&models.BranchInventory{}).Where("product_id = ?", productID).Count(&inventoryCount)
	if inventoryCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete product with existing inventory entries"})
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

	var stocks []models.BranchInventory
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

func GetBranchInventory(c *gin.Context) {
	branchID := c.Param("branchId")

	var inventories []models.BranchInventory
	if err := db.DB.
		Preload("Product").
		Where("branch_id = ?", branchID).
		Find(&inventories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branch inventory"})
		return
	}

	// Convert to ProductWithStock format matching frontend
	products := []gin.H{}
	for _, inventory := range inventories {
		// Parse tags JSON to array
		var tags []string
		if inventory.Product.Tags != "" && inventory.Product.Tags != "[]" {
			if err := json.Unmarshal([]byte(inventory.Product.Tags), &tags); err != nil {
				tags = []string{} // Default to empty array on error
			}
		}

		products = append(products, gin.H{
			"id":            inventory.Product.ID,
			"name":          inventory.Product.Name,
			"brand":         inventory.Product.Brand,
			"description":   inventory.Product.Description,
			"price":         inventory.Product.Price,
			"originalPrice": inventory.Product.OriginalPrice,
			"image":         inventory.Product.Image,
			"rating":        inventory.Product.Rating,
			"reviews":       inventory.Product.Reviews,
			"stock":         inventory.Quantity,
			"category":      inventory.Product.Category,
			"volume":        inventory.Product.Volume,
			"unit":          inventory.Product.Unit,
			"tags":          tags,
			"available":     inventory.Quantity > 0,
		})
	}

	c.JSON(http.StatusOK, products)
}
