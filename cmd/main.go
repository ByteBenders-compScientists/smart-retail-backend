package main

import (
	"log"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/api"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/initialisers"
	"github.com/gin-gonic/gin"
)

func init() {
	initialisers.LoadEnv()
	db.ConnectDB()
}

func main() {
	r := gin.Default()

	api.RegisterRoutes(r)

	log.Println("Smart Retail running on http://localhost:8080")
	r.Run(":8080")
}
