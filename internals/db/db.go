// database connection (postgres)
package db

import (
	"log"
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func ConnectDB() {
	var err error
	DB, err = gorm.Open(postgres.Open(os.Getenv("DB_URI")), &gorm.Config{})

	if err != nil {
		panic("Failed to connect to database!")
	}

	log.Println("Database connection established")
}
