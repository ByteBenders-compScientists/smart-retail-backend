// auth controller
package controllers

import (
	"net/http"
	"os"
	"time"

	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/db"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/models"
	"github.com/ByteBenders-compScientists/smart-retail-backend/internals/utils"
	"github.com/gin-gonic/gin"
)

const defaultAuthCookieName = "auth_token"

func setAuthCookie(c *gin.Context, token string) {
	cookieName := os.Getenv("JWT_COOKIE_NAME")
	if cookieName == "" {
		cookieName = defaultAuthCookieName
	}
	secure := os.Getenv("COOKIE_SECURE") == "true"
	maxAge := int((24 * time.Hour).Seconds())

	// Lax avoids CSRF on GET while allowing same-site POST in typical SPA flows
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(cookieName, token, maxAge, "/", "", secure, true)
}

func Register(c *gin.Context) {
	var body struct {
		Name            string `json:"name" binding:"required"`
		Email           string `json:"email" binding:"required,email"`
		Phone           string `json:"phone" binding:"required"`
		Password        string `json:"password" binding:"required"`
		ConfirmPassword string `json:"confirmPassword" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if body.Password != body.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
		return
	}

	var existingUser models.User
	if err := db.DB.Where("email = ? OR phone = ?", body.Email, body.Phone).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email or phone already exists"})
		return
	}

	user := models.User{
		Name:     body.Name,
		Email:    body.Email,
		Phone:    body.Phone,
		Password: utils.HashPassword(body.Password),
		Role:     "customer",
	}

	if err := db.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	token := utils.GenerateJWT(user.ID, user.Role)
	setAuthCookie(c, token)

	c.JSON(http.StatusCreated, gin.H{
		"message": "User registered successfully",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"phone": user.Phone,
			"role":  user.Role,
		},
		"token": token,
	})
}

func Login(c *gin.Context) {
	var body struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	var user models.User
	if err := db.DB.Where("email = ?", body.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if !utils.CheckPassword(body.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token := utils.GenerateJWT(user.ID, user.Role)
	setAuthCookie(c, token)

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"phone": user.Phone,
			"role":  user.Role,
		},
		"token": token,
	})
}

func Logout(c *gin.Context) {
	cookieName := os.Getenv("JWT_COOKIE_NAME")
	if cookieName == "" {
		cookieName = defaultAuthCookieName
	}
	
	c.SetCookie(cookieName, "", -1, "/", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

func GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := db.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
		"phone": user.Phone,
		"role":  user.Role,
		"createdAt": user.CreatedAt,
	})
}
