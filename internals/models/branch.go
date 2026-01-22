// branch model
package models

import (
	"time"
	"gorm.io/gorm"
)

type Branch struct {
	gorm.Model
	ID           string    `gorm:"type:varchar(50);primaryKey"` // e.g., 'branch-nairobi'
	Name         string    `gorm:"type:varchar(50);not null;check:name IN ('Nairobi', 'Kisumu', 'Mombasa', 'Nakuru', 'Eldoret')"`
	IsHeadquarter bool      `gorm:"default:false"`
	Address      string    `gorm:"type:varchar(255);not null"`
	Phone        string    `gorm:"type:varchar(20);not null"`
	Status       string    `gorm:"type:varchar(20);not null;default:'active';check:status IN ('active', 'inactive')"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
}
