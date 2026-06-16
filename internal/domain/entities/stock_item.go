package entities

import "gorm.io/gorm"

type StockItem struct {
	ID               string  `json:"id" gorm:"type:char(36);primaryKey"`
	Name             string  `json:"name" gorm:"size:140;not null;index"`
	Description      string  `json:"description" gorm:"size:500"`
	CostCents        int64   `json:"costCents" gorm:"not null"`
	MarkupPercent    float64 `json:"markupPercent" gorm:"not null;default:10"`
	ResalePriceCents int64   `json:"resalePriceCents" gorm:"not null"`
	Quantity         int     `json:"quantity" gorm:"not null;default:0"`
	UsedQuantity     int     `json:"usedQuantity" gorm:"not null;default:0"`
	Active           bool    `json:"active" gorm:"not null;default:true;index"`
	Timestamps
}

func (item *StockItem) BeforeCreate(_ *gorm.DB) error {
	assignID(&item.ID)
	return nil
}
