// order model
package models

import (
	"time"
	"gorm.io/gorm"
)

type Order struct {
	gorm.Model
	ID                   string    `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	UserID               string    `gorm:"type:uuid;not null"`
	User                 User      `gorm:"foreignKey:UserID"`
	BranchID             string    `gorm:"type:varchar(50);not null"`
	Branch               Branch    `gorm:"foreignKey:BranchID"`
	TotalAmount          float64   `gorm:"not null"`
	PaymentStatus        string    `gorm:"type:varchar(20);not null;default:'pending';check:payment_status IN ('pending', 'completed', 'failed')"`
	PaymentMethod        string    `gorm:"type:varchar(20);not null;default:'mpesa';check:payment_method IN ('mpesa')"`
	MpesaTransactionID   *string   `gorm:"type:varchar(255)"`
	OrderStatus          string    `gorm:"type:varchar(20);not null;default:'processing';check:order_status IN ('processing', 'completed', 'cancelled')"`
	CreatedAt            time.Time `gorm:"autoCreateTime"`
	UpdatedAt            time.Time `gorm:"autoUpdateTime"`
	CompletedAt          *time.Time
	OrderItems           []OrderItem `gorm:"foreignKey:OrderID"`
}

type OrderItem struct {
	gorm.Model
	ID          string  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	OrderID     string  `gorm:"type:uuid;not null"`
	Order       Order   `gorm:"foreignKey:OrderID"`
	ProductID   string  `gorm:"type:uuid;not null"`
	Product     Product `gorm:"foreignKey:ProductID"`
	ProductBrand string `gorm:"type:varchar(50);not null;check:product_brand IN ('Coke', 'Fanta', 'Sprite')"`
	Quantity    int     `gorm:"not null"`
	Price       float64 `gorm:"not null"`
	Subtotal    float64 `gorm:"not null"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}
