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
	PaymentMethodLegacy     PaymentMethod = "legacy"
	PaymentMethodMixed      PaymentMethod = "mixed"
)

type Receipt struct {
	ID                 string           `json:"id" gorm:"type:char(36);primaryKey"`
	UserID             *string          `json:"userId,omitempty" gorm:"type:char(36);index"`
	User               *User            `json:"user,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	VehicleModel       string           `json:"vehicleModel" gorm:"size:140"`
	VehicleYear        int              `json:"vehicleYear"`
	VehiclePlate       string           `json:"vehiclePlate" gorm:"size:12;index"`
	Services           string           `json:"services" gorm:"size:700;not null"`
	LaborPriceCents    int64            `json:"laborPriceCents" gorm:"not null;default:0"`
	DiscountCents      int64            `json:"discountCents" gorm:"not null;default:0"`
	ProductsTotalCents int64            `json:"productsTotalCents" gorm:"not null;default:0"`
	SubtotalCents      int64            `json:"subtotalCents" gorm:"not null;default:0"`
	CardFeePercent     float64          `json:"cardFeePercent" gorm:"not null;default:0"`
	CardFeeCents       int64            `json:"cardFeeCents" gorm:"not null;default:0"`
	PaymentMethod      PaymentMethod    `json:"paymentMethod" gorm:"size:30;not null;default:cash"`
	Installments       int              `json:"installments" gorm:"not null;default:1"`
	PriceCents         int64            `json:"priceCents" gorm:"not null"`
	Status             ReceiptStatus    `json:"status" gorm:"size:20;not null;default:pending;index"`
	Notes              string           `json:"notes" gorm:"size:700"`
	PaidAt             *time.Time       `json:"paidAt"`
	Payments           []ReceiptPayment `json:"payments" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Items              []ReceiptItem    `json:"items" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Expenses           []Expense        `json:"expenses,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
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

// ReceiptPayment stores the base allocation separately from the card surcharge.
type ReceiptPayment struct {
	ID              string        `json:"id" gorm:"type:char(36);primaryKey"`
	ReceiptID       string        `json:"receiptId" gorm:"type:char(36);not null;uniqueIndex:idx_receipt_payment_position"`
	Position        int           `json:"position" gorm:"not null;uniqueIndex:idx_receipt_payment_position"`
	PaymentMethod   PaymentMethod `json:"paymentMethod" gorm:"size:30;not null"`
	BaseAmountCents int64         `json:"baseAmountCents" gorm:"not null"`
	CardFeePercent  float64       `json:"cardFeePercent" gorm:"not null;default:0"`
	CardFeeCents    int64         `json:"cardFeeCents" gorm:"not null;default:0"`
	AmountCents     int64         `json:"amountCents" gorm:"not null"`
	Installments    int           `json:"installments" gorm:"not null;default:1"`
	Timestamps
}

func (payment *ReceiptPayment) BeforeCreate(_ *gorm.DB) error {
	assignID(&payment.ID)
	return nil
}

// EffectivePayments preserves historical amounts without recalculating old fees.
func (receipt *Receipt) EffectivePayments() []ReceiptPayment {
	if len(receipt.Payments) > 0 {
		return receipt.Payments
	}
	method := receipt.PaymentMethod
	if method == "" {
		method = PaymentMethodCash
	}
	installments := receipt.Installments
	if installments < 1 {
		installments = 1
	}
	return []ReceiptPayment{{ReceiptID: receipt.ID, PaymentMethod: method, BaseAmountCents: receipt.PriceCents - receipt.CardFeeCents, CardFeePercent: receipt.CardFeePercent, CardFeeCents: receipt.CardFeeCents, AmountCents: receipt.PriceCents, Installments: installments}}
}
