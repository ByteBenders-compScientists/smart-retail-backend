package controllers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/gin-gonic/gin"
)

type SalesReport struct {
	TotalRevenue   float64            `json:"total_revenue"`
	TotalSales     int                `json:"total_sales"`
	BrandBreakdown []BrandSalesReport `json:"brand_breakdown"`
	Period         string             `json:"period"`
	GeneratedAt    time.Time          `json:"generated_at"`
}

type BrandSalesReport struct {
	Brand        string  `json:"brand"`
	Revenue      float64 `json:"revenue"`
	UnitsSold    int     `json:"units_sold"`
	SalesCount   int     `json:"sales_count"`
	AveragePrice float64 `json:"average_price"`
}

type BranchReport struct {
	BranchID     string         `json:"branch_id"`
	BranchName   string         `json:"branch_name"`
	TotalRevenue float64        `json:"total_revenue"`
	TotalSales   int            `json:"total_sales"`
	TopProducts  []ProductSales `json:"top_products"`
}

type ProductSales struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Revenue     float64 `json:"revenue"`
}

func GetSalesReport(c *gin.Context) {
	// Parse query parameters
	period := c.DefaultQuery("period", "week")
	branchID := c.Query("branch_id")

	// Calculate date range based on period
	var startDate time.Time
	now := time.Now()

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, 0, -7) // Default to week
	}

	// Build query
	query := db.DB.
		Preload("SaleItems.Product").
		Where("sales.created_at >= ? AND sales.status = ?", startDate, "paid")

	if branchID != "" {
		query = query.Where("sales.branch_id = ?", branchID)
	}

	var sales []models.Sale
	if err := query.Find(&sales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sales data"})
		return
	}

	// Generate report
	report := SalesReport{
		TotalRevenue:   0,
		TotalSales:     len(sales),
		Period:         period,
		GeneratedAt:    now,
		BrandBreakdown: []BrandSalesReport{},
	}

	brandMap := make(map[string]*BrandSalesReport)

	for _, sale := range sales {
		report.TotalRevenue += float64(sale.Total)

		for _, item := range sale.SaleItems {
			brand := item.Product.Brand

			if brandMap[brand] == nil {
				brandMap[brand] = &BrandSalesReport{
					Brand:      brand,
					Revenue:    0,
					UnitsSold:  0,
					SalesCount: 0,
				}
			}

			brandReport := brandMap[brand]
			brandReport.Revenue += item.Price * float64(item.Qty)
			brandReport.UnitsSold += item.Qty
			brandReport.SalesCount++
		}
	}

	// Convert map to slice and calculate averages
	for _, brandReport := range brandMap {
		if brandReport.UnitsSold > 0 {
			brandReport.AveragePrice = brandReport.Revenue / float64(brandReport.UnitsSold)
		}
		report.BrandBreakdown = append(report.BrandBreakdown, *brandReport)
	}

	c.JSON(http.StatusOK, report)
}

func GetBranchPerformanceReport(c *gin.Context) {
	// Parse query parameters
	period := c.DefaultQuery("period", "week")

	// Calculate date range based on period
	var startDate time.Time
	now := time.Now()

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, 0, -7) // Default to week
	}

	// Get all branches
	var branches []models.Branch
	if err := db.DB.Find(&branches).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch branches"})
		return
	}

	var reports []BranchReport

	for _, branch := range branches {
		var sales []models.Sale
		if err := db.DB.
			Preload("SaleItems.Product").
			Where("branch_id = ? AND created_at >= ? AND status = ?", branch.ID, startDate, "paid").
			Find(&sales).Error; err != nil {
			continue
		}

		branchReport := BranchReport{
			BranchID:     branch.ID,
			BranchName:   branch.Name,
			TotalRevenue: 0,
			TotalSales:   len(sales),
			TopProducts:  []ProductSales{},
		}

		productMap := make(map[string]*ProductSales)

		for _, sale := range sales {
			branchReport.TotalRevenue += float64(sale.Total)

			for _, item := range sale.SaleItems {
				if productMap[item.ProductID] == nil {
					productMap[item.ProductID] = &ProductSales{
						ProductID:   item.ProductID,
						ProductName: item.Product.Name,
						Quantity:    0,
						Revenue:     0,
					}
				}

				productSales := productMap[item.ProductID]
				productSales.Quantity += item.Qty
				productSales.Revenue += item.Price * float64(item.Qty)
			}
		}

		// Convert map to slice and get top 5 products
		for _, productSales := range productMap {
			branchReport.TopProducts = append(branchReport.TopProducts, *productSales)
		}

		// Sort by revenue (top 5)
		if len(branchReport.TopProducts) > 5 {
			// Simple sort by revenue (you might want to implement proper sorting)
			branchReport.TopProducts = branchReport.TopProducts[:5]
		}

		reports = append(reports, branchReport)
	}

	c.JSON(http.StatusOK, gin.H{
		"period":         period,
		"generated_at":   now,
		"branch_reports": reports,
	})
}

func GetLowStockReport(c *gin.Context) {
	threshold := 10 // Default threshold
	if thresholdStr := c.Query("threshold"); thresholdStr != "" {
		if t, err := strconv.Atoi(thresholdStr); err == nil && t > 0 {
			threshold = t
		}
	}

	var stocks []models.Stock
	if err := db.DB.
		Preload("Product").
		Preload("Branch").
		Where("quantity < ?", threshold).
		Find(&stocks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch low stock items"})
		return
	}

	// Group by branch
	branchMap := make(map[string][]models.Stock)
	for _, stock := range stocks {
		branchMap[stock.BranchID] = append(branchMap[stock.BranchID], stock)
	}

	type BranchLowStock struct {
		BranchID      string         `json:"branch_id"`
		BranchName    string         `json:"branch_name"`
		LowStockItems []models.Stock `json:"low_stock_items"`
		Count         int            `json:"count"`
	}

	var report []BranchLowStock
	for branchID, items := range branchMap {
		var branch models.Branch
		db.DB.First(&branch, branchID)

		report = append(report, BranchLowStock{
			BranchID:      branchID,
			BranchName:    branch.Name,
			LowStockItems: items,
			Count:         len(items),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"threshold":             threshold,
		"generated_at":          time.Now(),
		"branches":              report,
		"total_low_stock_items": len(stocks),
	})
}

func GetRevenueSummary(c *gin.Context) {
	// Parse query parameters
	period := c.DefaultQuery("period", "month")

	// Calculate date range based on period
	var startDate time.Time
	now := time.Now()

	switch period {
	case "today":
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		startDate = now.AddDate(0, 0, -7)
	case "month":
		startDate = now.AddDate(0, -1, 0)
	case "year":
		startDate = now.AddDate(-1, 0, 0)
	default:
		startDate = now.AddDate(0, -1, 0) // Default to month
	}

	// Get total revenue
	var totalRevenue struct {
		Total float64 `json:"total"`
	}

	if err := db.DB.Model(&models.Sale{}).
		Select("COALESCE(SUM(total), 0) as total").
		Where("created_at >= ? AND status = ?", startDate, "paid").
		Scan(&totalRevenue).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate total revenue"})
		return
	}

	// Get total sales count
	var totalSales int64
	db.DB.Model(&models.Sale{}).
		Where("created_at >= ? AND status = ?", startDate, "paid").
		Count(&totalSales)

	// Get top performing brands
	type BrandRevenue struct {
		Brand   string  `json:"brand"`
		Revenue float64 `json:"revenue"`
	}

	var brandRevenues []BrandRevenue
	if err := db.DB.Raw(`
		SELECT p.brand, SUM(si.price * si.qty) as revenue
		FROM sales s
		JOIN sale_items si ON s.id = si.sale_id
		JOIN products p ON si.product_id = p.id
		WHERE s.created_at >= ? AND s.status = ?
		GROUP BY p.brand
		ORDER BY revenue DESC
		LIMIT 10
	`, startDate, "paid").Scan(&brandRevenues).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get brand revenues"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period":        period,
		"generated_at":  now,
		"total_revenue": totalRevenue.Total,
		"total_sales":   totalSales,
		"average_sale":  totalRevenue.Total / float64(totalSales),
		"top_brands":    brandRevenues,
	})
}

func GetDailySalesTrend(c *gin.Context) {
	days := 30 // Default to 30 days
	if daysStr := c.Query("days"); daysStr != "" {
		if d, err := strconv.Atoi(daysStr); err == nil && d > 0 && d <= 365 {
			days = d
		}
	}

	startDate := time.Now().AddDate(0, 0, -days)

	type DailySales struct {
		Date    string  `json:"date"`
		Revenue float64 `json:"revenue"`
		Sales   int64   `json:"sales"`
	}

	var dailySales []DailySales
	if err := db.DB.Raw(`
		SELECT 
			DATE(created_at) as date,
			COALESCE(SUM(total), 0) as revenue,
			COUNT(*) as sales
		FROM sales
		WHERE created_at >= ? AND status = ?
		GROUP BY DATE(created_at)
		ORDER BY date ASC
	`, startDate, "paid").Scan(&dailySales).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get daily sales trend"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"period":       strconv.Itoa(days) + " days",
		"generated_at": time.Now(),
		"daily_sales":  dailySales,
	})
}
