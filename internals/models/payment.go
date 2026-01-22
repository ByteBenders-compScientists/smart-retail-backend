// payment model
package models

import (
	"time"
	"gorm.io/gorm"
)

type Payment struct {
	gorm.Model
	ID               string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrderID          string    `gorm:"type:uuid;not null;uniqueIndex"`
	Order            Order     `gorm:"foreignKey:OrderID"`
	Phone            string    `gorm:"type:varchar(20);not null"`
	Amount           float64   `gorm:"not null"`
	TransactionID    *string   `gorm:"type:varchar(255)"`
	CheckoutRequestID *string  `gorm:"type:varchar(255)"`
	Status           string    `gorm:"type:varchar(20);not null;default:'pending';check:status IN ('pending', 'completed', 'failed')"`
	MpesaResponse    *string   `gorm:"type:text"` // JSON response from M-Pesa
	CreatedAt        time.Time `gorm:"autoCreateTime"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime"`
}
