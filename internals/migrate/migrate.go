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
	db.DB.AutoMigrate(
		&models.User{},
		&models.Branch{},
		&models.Product{},
		&models.Stock{},
		&models.Sale{},
		&models.SaleItem{},
	)

	fmt.Println("Database migration completed")
}
