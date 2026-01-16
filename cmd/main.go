package main

import (
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/api"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/initialisers"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/utils"
)

func init() {
	initialisers.LoadEnv()
	utils.InitLogger()
	db.ConnectDB()
}

func main() {
	r := api.SetUpRoutes()

	utils.Logger.Info("Smart Retail server starting on http://localhost:8080")
	r.Run(":8080")
}
