// user model
package models

import (
	"time"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	ID        string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name      string    `gorm:"type:varchar(100);not null"`
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null"`
	Phone     string    `gorm:"type:varchar(20);not null"`
	Password  string    `gorm:"type:varchar(255);not null"`
	Role      string    `gorm:"type:varchar(50);not null;check:role IN ('customer', 'admin')"` // customer or admin
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
}
