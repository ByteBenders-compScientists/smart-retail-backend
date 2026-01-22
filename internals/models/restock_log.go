// restock log model
package models

import (
	"time"
	"gorm.io/gorm"
)

type RestockLog struct {
	gorm.Model
	ID               string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BranchID         string    `gorm:"type:varchar(50);not null"`
	Branch           Branch    `gorm:"foreignKey:BranchID"`
	ProductID        string    `gorm:"type:uuid;not null"`
	Product          Product  `gorm:"foreignKey:ProductID"`
	QuantityAdded    int       `gorm:"not null"`
	PreviousQuantity int       `gorm:"not null"`
	NewQuantity      int       `gorm:"not null"`
	RestockedBy      string    `gorm:"type:uuid;not null"`
	RestockedByUser  User      `gorm:"foreignKey:RestockedBy"`
	CreatedAt        time.Time `gorm:"autoCreateTime"`
}
