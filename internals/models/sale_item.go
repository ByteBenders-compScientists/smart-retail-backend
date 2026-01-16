package models

import "gorm.io/gorm"

type SaleItem struct {
	gorm.Model
	SaleID    uint    `gorm:"not null"`
	ProductID uint    `gorm:"not null"`
	Qty       int     `gorm:"not null"`
	Price     float64 `gorm:"not null"`
}
