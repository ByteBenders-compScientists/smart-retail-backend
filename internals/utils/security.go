// authentication related utility functions
package utils

import (
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func HashPassword(p string) string {
	hash, _ := bcrypt.GenerateFromPassword([]byte(p), 10)
	return string(hash)
}

func CheckPassword(p, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(p)) == nil
}

func GenerateJWT(userID uint, role string) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"role":    role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	})

	t, _ := token.SignedString([]byte(os.Getenv("JWT_SECRET")))
	return t
}
