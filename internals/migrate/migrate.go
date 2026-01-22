// automigrate runs database migrations automatically on application startup.
package main

import (
	"fmt"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/initialisers"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
)

func init() {
	initialisers.LoadEnv()
	db.ConnectDB()
}

func main() {
	// Enable UUID extension for PostgreSQL
	if err := db.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error; err != nil {
		fmt.Println("Warning: Could not create uuid-ossp extension:", err)
		fmt.Println("Trying pgcrypto extension as fallback...")
		// Try pgcrypto as fallback (gen_random_uuid)
		if err := db.DB.Exec("CREATE EXTENSION IF NOT EXISTS \"pgcrypto\"").Error; err != nil {
			fmt.Println("Error: Could not create pgcrypto extension:", err)
		}
	}

	db.DB.AutoMigrate(
		&models.User{},
		&models.Branch{},
		&models.Product{},
		&models.BranchInventory{},
		&models.Order{},
		&models.OrderItem{},
		&models.Payment{},
		&models.RestockLog{},
	)

	fmt.Println("Database migration completed")
	
	// Seed initial data
	seedData()
}
