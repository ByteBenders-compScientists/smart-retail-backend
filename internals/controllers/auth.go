// auth controller
package controllers

import (
	"net/http"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/utils"
	"github.com/gin-gonic/gin"
)

func Register(c *gin.Context) {
	var user models.User
	c.BindJSON(&user)

	user.Password = utils.HashPassword(user.Password)
	user.Role = "customer"

	db.DB.Create(&user)

	c.JSON(http.StatusOK, gin.H{"message": "registered"})
}

func Login(c *gin.Context) {
	var body struct {
		Email    string
		Password string
	}
	c.BindJSON(&body)

	var user models.User
	db.DB.Where("email = ?", body.Email).First(&user)

	if !utils.CheckPassword(body.Password, user.Password) {
		c.JSON(401, gin.H{"error": "invalid credentials"})
		return
	}

	token := utils.GenerateJWT(user.ID, user.Role)
	c.JSON(200, gin.H{"token": token})
}
