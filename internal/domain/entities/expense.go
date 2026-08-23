package entities

import (
	"time"

	"gorm.io/gorm"
)

type Expense struct {
	ID            string               `json:"id" gorm:"type:char(36);primaryKey"`
	ReceiptID     *string              `json:"receiptId" gorm:"type:char(36);index"`
	Receipt       *Receipt             `json:"receipt,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Description   string               `json:"description" gorm:"size:180;not null"`
	Category      string               `json:"category" gorm:"size:80;not null;index"`
	AmountCents   int64                `json:"amountCents" gorm:"not null"`
	SpentAt       time.Time            `json:"spentAt" gorm:"not null;index"`
	Notes         string               `json:"notes" gorm:"size:700"`
	AffectsProfit bool                 `json:"affectsProfit" gorm:"not null;default:true"`
	ArchivedAt    *time.Time           `json:"archivedAt" gorm:"index"`
	Installments  []PayableInstallment `json:"installments,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Timestamps
}

func (expense *Expense) BeforeCreate(_ *gorm.DB) error {
	assignID(&expense.ID)
	return nil
}
