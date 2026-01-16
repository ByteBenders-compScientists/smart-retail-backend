package models

import "gorm.io/gorm"

type SaleItem struct {
	gorm.Model
	ID        string  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	SaleID    string  `gorm:"type:uuid;not null"`
	ProductID string  `gorm:"type:uuid;not null"`
	Product   Product `gorm:"foreignKey:ProductID"`
	Qty       int     `gorm:"not null"`
	Price     float64 `gorm:"not null"`
}
