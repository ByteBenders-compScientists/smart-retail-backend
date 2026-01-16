// sales model
package models

import "gorm.io/gorm"

type Sale struct {
	gorm.Model
	ID         string     `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BranchID   string     `gorm:"type:uuid;not null"`
	UserID     string     `gorm:"type:uuid;not null"`
	Total      int        `gorm:"not null"`
	Status     string     `gorm:"type:varchar(50);not null"` // pending, paid
	PaymentRef string     `gorm:"type:varchar(100)"`
	SaleItems  []SaleItem `gorm:"foreignKey:SaleID"`
}
