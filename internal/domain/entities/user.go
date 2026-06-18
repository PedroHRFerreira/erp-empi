package entities

import (
	"time"

	"gorm.io/gorm"
)

type UserType string

const (
	UserTypeAdmin  UserType = "admin"
	UserTypeClient UserType = "client"
)

type User struct {
	ID                    string     `json:"id" gorm:"type:char(36);primaryKey"`
	Name                  string     `json:"name" gorm:"size:140;not null;index"`
	CPF                   string     `json:"cpf" gorm:"size:11;index"`
	PasswordHash          *string    `json:"-" gorm:"size:255"`
	Type                  UserType   `json:"type" gorm:"size:20;not null;index"`
	Email                 string     `json:"email" gorm:"size:180"`
	Phone                 string     `json:"phone" gorm:"size:20;index"`
	MarkupPercent         float64    `json:"markupPercent" gorm:"not null;default:10"`
	MachineFeePercent     float64    `json:"machineFeePercent" gorm:"not null;default:0"`
	InstallmentFeePercent float64    `json:"installmentFeePercent" gorm:"not null;default:0"`
	Address               string     `json:"address" gorm:"size:255"`
	Notes                 string     `json:"notes" gorm:"size:500"`
	ArchivedAt            *time.Time `json:"archivedAt,omitempty" gorm:"index"`
	Timestamps
}

func (user *User) BeforeCreate(_ *gorm.DB) error {
	assignID(&user.ID)
	return nil
}
