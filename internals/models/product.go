// product model
package models

import "gorm.io/gorm"

type Product struct {
	gorm.Model
	ID    string  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name  string  `gorm:"type:varchar(100);not null"`
	Brand string  `gorm:"type:varchar(100);not null"`
	Price float64 `gorm:"not null"` // cents
}
