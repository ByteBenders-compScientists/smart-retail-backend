// stock model
package models

import "gorm.io/gorm"

type Stock struct {
	gorm.Model
	ID        string `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	BranchID  string `gorm:"type:uuid;not null"`
	Branch    Branch
	ProductID string `gorm:"type:uuid;not null"`
	Product   Product
	Quantity  int `gorm:"not null"`
}
