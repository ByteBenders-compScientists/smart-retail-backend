// branch inventory model
package models

import (
	"time"
	"gorm.io/gorm"
)

type BranchInventory struct {
	gorm.Model
	ID            string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BranchID      string    `gorm:"type:varchar(50);not null"`
	Branch        Branch    `gorm:"foreignKey:BranchID"`
	ProductID     string    `gorm:"type:uuid;not null"`
	Product       Product   `gorm:"foreignKey:ProductID"`
	Quantity      int       `gorm:"not null;default:0"`
	LastRestocked time.Time `gorm:"autoCreateTime"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}

// Add unique constraint for branch-product combination
func (BranchInventory) TableName() string {
	return "branch_inventories"
}
