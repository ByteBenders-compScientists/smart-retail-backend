// stock model
package models

import "gorm.io/gorm"

type Stock struct {
	gorm.Model
	BranchID  uint `gorm:"not null"`
	ProductID uint `gorm:"not null"`
	Quantity  int  `gorm:"not null"`
}
