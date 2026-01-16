// sales model
package models

import "gorm.io/gorm"

type Sale struct {
	gorm.Model
	BranchID   uint   `gorm:"not null"`
	UserID     uint   `gorm:"not null"`
	Total      int    `gorm:"not null"`
	Status     string `gorm:"type:varchar(50);not null"` // pending, paid
	PaymentRef string `gorm:"type:varchar(100)"`
}
