// branch model
package models

import "gorm.io/gorm"

type Branch struct {
	gorm.Model
	ID       string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name     string `gorm:"type:varchar(100);not null"`
	Location string `gorm:"type:varchar(255);not null"`
	IsHQ     bool   `gorm:"default:false"`
}
