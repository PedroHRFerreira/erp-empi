package entities

import (
	"time"

	"gorm.io/gorm"
)

type ReceiptStatus string
type PaymentMethod string

const (
	ReceiptStatusPending   ReceiptStatus = "pending"
	ReceiptStatusPaid      ReceiptStatus = "paid"
	ReceiptStatusCancelled ReceiptStatus = "cancelled"
)

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodDebitCard  PaymentMethod = "debit_card"
	PaymentMethodPix        PaymentMethod = "pix"
	PaymentMethodCash       PaymentMethod = "cash"
)

type Receipt struct {
	ID                 string        `json:"id" gorm:"type:char(36);primaryKey"`
	UserID             string        `json:"userId" gorm:"type:char(36);not null;index"`
	User               User          `json:"user" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	VehicleModel       string        `json:"vehicleModel" gorm:"size:140;not null"`
	VehicleYear        int           `json:"vehicleYear" gorm:"not null"`
	VehiclePlate       string        `json:"vehiclePlate" gorm:"size:12;not null;index"`
	Services           string        `json:"services" gorm:"size:700;not null"`
	LaborPriceCents    int64         `json:"laborPriceCents" gorm:"not null;default:0"`
	DiscountCents      int64         `json:"discountCents" gorm:"not null;default:0"`
	ProductsTotalCents int64         `json:"productsTotalCents" gorm:"not null;default:0"`
	SubtotalCents      int64         `json:"subtotalCents" gorm:"not null;default:0"`
	CardFeePercent     float64       `json:"cardFeePercent" gorm:"not null;default:0"`
	CardFeeCents       int64         `json:"cardFeeCents" gorm:"not null;default:0"`
	PaymentMethod      PaymentMethod `json:"paymentMethod" gorm:"size:30;not null;default:cash"`
	Installments       int           `json:"installments" gorm:"not null;default:1"`
	PriceCents         int64         `json:"priceCents" gorm:"not null"`
	Status             ReceiptStatus `json:"status" gorm:"size:20;not null;default:pending;index"`
	Notes              string        `json:"notes" gorm:"size:700"`
	PaidAt             *time.Time    `json:"paidAt"`
	Items              []ReceiptItem `json:"items" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Expenses           []Expense     `json:"expenses,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Timestamps
}

type ReceiptItem struct {
	ID              string    `json:"id" gorm:"type:char(36);primaryKey"`
	ReceiptID       string    `json:"receiptId" gorm:"type:char(36);not null;index"`
	StockItemID     string    `json:"stockItemId" gorm:"type:char(36);not null;index"`
	StockItem       StockItem `json:"stockItem" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Quantity        int       `json:"quantity" gorm:"not null"`
	UnitCostCents   int64     `json:"unitCostCents" gorm:"not null"`
	UnitResaleCents int64     `json:"unitResaleCents" gorm:"not null"`
	MarkupPercent   float64   `json:"markupPercent" gorm:"not null"`
	Timestamps
}

func (receipt *Receipt) BeforeCreate(_ *gorm.DB) error {
	assignID(&receipt.ID)
	if receipt.Status == "" {
		receipt.Status = ReceiptStatusPending
	}
	if receipt.PaymentMethod == "" {
		receipt.PaymentMethod = PaymentMethodCash
	}
	if receipt.Installments == 0 {
		receipt.Installments = 1
	}
	return nil
}

func (item *ReceiptItem) BeforeCreate(_ *gorm.DB) error {
	assignID(&item.ID)
	return nil
}
