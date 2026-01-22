// product model
package models

import (
	"time"
	"gorm.io/gorm"
)

type Product struct {
	gorm.Model
	ID            string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name          string    `gorm:"type:varchar(100);not null"`
	Brand         string    `gorm:"type:varchar(50);not null;check:brand IN ('Coke', 'Fanta', 'Sprite')"`
	Description   string    `gorm:"type:text"`
	Price         float64   `gorm:"not null"` // Current price in KSh
	OriginalPrice float64   `gorm:"not null"` // Original price in KSh
	Image         string    `gorm:"type:varchar(255)"`
	Rating        float64   `gorm:"default:0"`
	Reviews       int       `gorm:"default:0"`
	Category      string    `gorm:"type:varchar(50);not null;default:'Soft Drinks'"`
	Volume        string    `gorm:"type:varchar(20);default:'500ml'"`
	Unit          string    `gorm:"type:varchar(20);default:'single'"`
	Tags          string    `gorm:"type:json;default:'[]'"` // JSON array of tags
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}
