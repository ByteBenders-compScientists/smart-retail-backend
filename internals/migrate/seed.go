// seed initial data for the application
package main

import (
	"fmt"
	"log"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
)

func seedData() {
	fmt.Println("Seeding initial data...")

	seedBranches()
	seedProducts()
	seedBranchInventory()
	seedAdminUser()

	fmt.Println("Seeding completed!")
}

func seedBranches() {
	branches := []models.Branch{
		{
			ID:            "branch-nairobi",
			Name:          "Nairobi",
			IsHeadquarter: true,
			Address:       "Nairobi CBD, Kenya",
			Phone:         "+254 700 000 001",
			Status:        "active",
		},
		{
			ID:            "branch-kisumu",
			Name:          "Kisumu",
			IsHeadquarter: false,
			Address:       "Kisumu City, Kenya",
			Phone:         "+254 700 000 002",
			Status:        "active",
		},
		{
			ID:            "branch-mombasa",
			Name:          "Mombasa",
			IsHeadquarter: false,
			Address:       "Mombasa Road, Kenya",
			Phone:         "+254 700 000 003",
			Status:        "active",
		},
		{
			ID:            "branch-nakuru",
			Name:          "Nakuru",
			IsHeadquarter: false,
			Address:       "Nakuru Town, Kenya",
			Phone:         "+254 700 000 004",
			Status:        "active",
		},
		{
			ID:            "branch-eldoret",
			Name:          "Eldoret",
			IsHeadquarter: false,
			Address:       "Eldoret Town, Kenya",
			Phone:         "+254 700 000 005",
			Status:        "active",
		},
	}

	for _, branch := range branches {
		var existingBranch models.Branch
		if err := db.DB.Where("id = ?", branch.ID).First(&existingBranch).Error; err != nil {
			if err := db.DB.Create(&branch).Error; err != nil {
				log.Printf("Failed to create branch %s: %v", branch.Name, err)
			} else {
				fmt.Printf("Created branch: %s\n", branch.Name)
			}
		} else {
			fmt.Printf("Branch already exists: %s\n", branch.Name)
		}
	}
}

func seedProducts() {
	products := []models.Product{
		{
			Name:          "Coca-Cola Original 500ml",
			Brand:         "Coke",
			Price:         60.00,
			OriginalPrice: 65.00,
			Description:   "Classic Coca-Cola taste. Available in single bottles or crates.",
			Image:         "https://i.postimg.cc/y6SN9pt5/coke.png",
			Rating:        4.8,
			Reviews:       1250,
			Category:      "Soft Drinks",
			Volume:        "500ml",
			Unit:          "single",
			Tags:          `["popular", "classic", "carbonated", "original"]`,
		},
		{
			Name:          "Coca-Cola Original 500ml Crate",
			Brand:         "Coke",
			Price:         1400.00,
			OriginalPrice: 1560.00,
			Description:   "Coca-Cola 500ml crate of 24 bottles. Perfect for parties and events.",
			Image:         "https://i.postimg.cc/VLQkFBcd/cokes.png",
			Rating:        4.8,
			Reviews:       850,
			Category:      "Soft Drinks",
			Volume:        "500ml x 24",
			Unit:          "crate",
			Tags:          `["popular", "classic", "carbonated", "bulk", "crate"]`,
		},
		{
			Name:          "Coca-Cola Original 1 Litre",
			Brand:         "Coke",
			Price:         110.00,
			OriginalPrice: 120.00,
			Description:   "Coca-Cola in a larger 1 litre bottle. Great for sharing.",
			Image:         "https://i.postimg.cc/J4VzQcWQ/litre.webp",
			Rating:        4.7,
			Reviews:       680,
			Category:      "Soft Drinks",
			Volume:        "1L",
			Unit:          "single",
			Tags:          `["popular", "classic", "carbonated", "large", "sharing"]`,
		},
		{
			Name:          "Fanta Orange 500ml",
			Brand:         "Fanta",
			Price:         60.00,
			OriginalPrice: 65.00,
			Description:   "Bursting with orange flavor. Refreshing anytime.",
			Image:         "https://i.postimg.cc/fRMWGzyz/orangee.png",
			Rating:        4.6,
			Reviews:       980,
			Category:      "Soft Drinks",
			Volume:        "500ml",
			Unit:          "single",
			Tags:          `["citrus", "fruity", "orange", "refreshing"]`,
		},
		{
			Name:          "Fanta Orange 500ml Crate",
			Brand:         "Fanta",
			Price:         1400.00,
			OriginalPrice: 1560.00,
			Description:   "Fanta Orange 500ml crate of 24 bottles. Bulk savings!",
			Image:         "https://i.postimg.cc/bNcwRHjG/fantas.png",
			Rating:        4.6,
			Reviews:       520,
			Category:      "Soft Drinks",
			Volume:        "500ml x 24",
			Unit:          "crate",
			Tags:          `["citrus", "fruity", "orange", "bulk", "crate"]`,
		},
		{
			Name:          "Fanta Orange 2 Litre",
			Brand:         "Fanta",
			Price:         180.00,
			OriginalPrice: 195.00,
			Description:   "Fanta Orange in 2 litre bottle. Maximum refreshment.",
			Image:         "https://i.postimg.cc/CxwM3h19/fant.png",
			Rating:        4.5,
			Reviews:       420,
			Category:      "Soft Drinks",
			Volume:        "2L",
			Unit:          "single",
			Tags:          `["citrus", "fruity", "orange", "large", "sharing"]`,
		},
		{
			Name:          "Sprite Lemon-Lime 500ml",
			Brand:         "Sprite",
			Price:         60.00,
			OriginalPrice: 65.00,
			Description:   "Crisp, clean lemon-lime flavor. Caffeine-free.",
			Image:         "https://i.postimg.cc/wj9xCqvr/sp.png",
			Rating:        4.7,
			Reviews:       1120,
			Category:      "Soft Drinks",
			Volume:        "500ml",
			Unit:          "single",
			Tags:          `["lemon", "lime", "crisp", "caffeine-free"]`,
		},
		{
			Name:          "Sprite Lemon-Lime 500ml Crate",
			Brand:         "Sprite",
			Price:         1400.00,
			OriginalPrice: 1560.00,
			Description:   "Sprite 500ml crate of 24 bottles. Stock up and save.",
			Image:         "https://i.postimg.cc/N0Ms0dR7/spritecrate.png",
			Rating:        4.7,
			Reviews:       640,
			Category:      "Soft Drinks",
			Volume:        "500ml x 24",
			Unit:          "crate",
			Tags:          `["lemon", "lime", "crisp", "caffeine-free", "bulk", "crate"]`,
		},
		{
			Name:          "Coca-Cola Zero Sugar 500ml",
			Brand:         "Coke",
			Price:         65.00,
			OriginalPrice: 70.00,
			Description:   "All Coca-Cola taste, zero sugar. Zero calories.",
			Image:         "https://i.postimg.cc/R0FS0gwP/zero.png",
			Rating:        4.5,
			Reviews:       760,
			Category:      "Diet Drinks",
			Volume:        "500ml",
			Unit:          "single",
			Tags:          `["zero-sugar", "diet", "caffeine-free", "calorie-free"]`,
		},
		{
			Name:          "Sprite Zero Sugar 1 Litre",
			Brand:         "Sprite",
			Price:         115.00,
			OriginalPrice: 125.00,
			Description:   "Great Sprite taste with zero sugar and zero calories.",
			Image:         "https://i.postimg.cc/Vk4s1x0Z/spritezero.png",
			Rating:        4.3,
			Reviews:       420,
			Category:      "Diet Drinks",
			Volume:        "1L",
			Unit:          "single",
			Tags:          `["zero-sugar", "diet", "caffeine-free", "calorie-free", "large"]`,
		},
		{
			Name:          "Fanta Pineapple 500ml",
			Brand:         "Fanta",
			Price:         60.00,
			OriginalPrice: 65.00,
			Description:   "Tropical pineapple flavor. Sweet and refreshing.",
			Image:         "https://i.postimg.cc/CLGLrBF4/pine.webp",
			Rating:        4.4,
			Reviews:       540,
			Category:      "Soft Drinks",
			Volume:        "500ml",
			Unit:          "single",
			Tags:          `["tropical", "pineapple", "fruity", "exotic"]`,
		},
		{
			Name:          "Coca-Cola Vanilla 500ml",
			Brand:         "Coke",
			Price:         70.00,
			OriginalPrice: 75.00,
			Description:   "Classic Coca-Cola with smooth vanilla twist. Limited edition.",
			Image:         "https://i.postimg.cc/mgVZRv1m/vani.png",
			Rating:        4.9,
			Reviews:       890,
			Category:      "Special Editions",
			Volume:        "500ml",
			Unit:          "single",
			Tags:          `["vanilla", "limited-edition", "special", "flavored"]`,
		},
	}

	for _, product := range products {
		var existingProduct models.Product
		if err := db.DB.Where("name = ?", product.Name).First(&existingProduct).Error; err != nil {
			if err := db.DB.Create(&product).Error; err != nil {
				log.Printf("Failed to create product %s: %v", product.Name, err)
			} else {
				fmt.Printf("Created product: %s\n", product.Name)
			}
		} else {
			fmt.Printf("Product already exists: %s\n", existingProduct.Name)
		}
	}
}

func seedBranchInventory() {
	var branches []models.Branch
	if err := db.DB.Find(&branches).Error; err != nil {
		log.Printf("Failed to fetch branches: %v", err)
		return
	}

	var products []models.Product
	if err := db.DB.Find(&products).Error; err != nil {
		log.Printf("Failed to fetch products: %v", err)
		return
	}

	for _, branch := range branches {
		for _, product := range products {
			var inventory models.BranchInventory
			if err := db.DB.Where("branch_id = ? AND product_id = ?", branch.ID, product.ID).First(&inventory).Error; err != nil {
				// Set initial stock levels
				initialStock := 100
				if branch.Name != "Nairobi" {
					initialStock = 50 // Other branches start with less stock
				}

				inventory = models.BranchInventory{
					BranchID:  branch.ID,
					ProductID: product.ID,
					Quantity:  initialStock,
				}

				if err := db.DB.Create(&inventory).Error; err != nil {
					log.Printf("Failed to create inventory for branch %s, product %s: %v", branch.Name, product.Name, err)
				} else {
					fmt.Printf("Created inventory: %s - %s (Qty: %d)\n", branch.Name, product.Name, initialStock)
				}
			}
		}
	}
}

func seedAdminUser() {
	var existingUser models.User
	if err := db.DB.Where("email = ?", "admin@drinx.com").First(&existingUser).Error; err != nil {
		admin := models.User{
			Name:     "System Administrator",
			Email:    "admin@drinx.com",
			Phone:    "+254 700 000 000",
			Password: "$2a$10$92IXUNpkjO0rOQ5byMi.Ye4oKoEa3Ro9llC/.og/at2.uheWG/igi", // password: "password"
			Role:     "admin",
		}

		if err := db.DB.Create(&admin).Error; err != nil {
			log.Printf("Failed to create admin user: %v", err)
		} else {
			fmt.Printf("Created admin user: %s (password: password)\n", admin.Email)
		}
	} else {
		fmt.Printf("Admin user already exists: %s\n", existingUser.Email)
	}
}
