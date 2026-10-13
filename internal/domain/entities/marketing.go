package entities

import "gorm.io/gorm"

type MarketingLayer struct {
	ID              string  `json:"id"`
	Type            string  `json:"type"`
	X               float64 `json:"x"`
	Y               float64 `json:"y"`
	Width           float64 `json:"width"`
	Height          float64 `json:"height"`
	Text            string  `json:"text,omitempty"`
	FontFamily      string  `json:"fontFamily,omitempty"`
	FontSize        float64 `json:"fontSize,omitempty"`
	Color           string  `json:"color,omitempty"`
	FontWeight      int     `json:"fontWeight,omitempty"`
	Italic          bool    `json:"italic,omitempty"`
	Rotation        float64 `json:"rotation,omitempty"`
	Align           string  `json:"align,omitempty"`
	ImageID         *string `json:"imageId,omitempty"`
	Fit             string  `json:"fit,omitempty"`
	Visible         bool    `json:"visible"`
	BackgroundColor string  `json:"backgroundColor,omitempty"`
}

type MarketingBrand struct {
	ID             uint    `json:"-" gorm:"primaryKey"`
	Name           string  `json:"name" gorm:"size:140"`
	LogoImageID    *string `json:"logoImageId" gorm:"size:36"`
	PrimaryColor   string  `json:"primaryColor" gorm:"size:7"`
	SecondaryColor string  `json:"secondaryColor" gorm:"size:7"`
	Contact        string  `json:"contact" gorm:"size:200"`
}

type MarketingContent struct {
	Layers         []MarketingLayer `json:"layers,omitempty" gorm:"serializer:json;type:jsonb"`
	Kind           string           `json:"kind" gorm:"size:16"`
	Objective      string           `json:"objective" gorm:"size:16"`
	Template       string           `json:"template" gorm:"size:16"`
	Format         string           `json:"format" gorm:"size:16"`
	Title          string           `json:"title" gorm:"size:140"`
	Description    string           `json:"description" gorm:"size:500"`
	PriceCents     *int64           `json:"priceCents"`
	CTA            string           `json:"cta" gorm:"size:100"`
	PrimaryColor   string           `json:"primaryColor" gorm:"size:7"`
	SecondaryColor string           `json:"secondaryColor" gorm:"size:7"`
	Caption        string           `json:"caption" gorm:"size:2200"`
	ImageID        *string          `json:"imageId" gorm:"size:36"`
	ProductID      *string          `json:"productId" gorm:"size:36"`
	BrandName      string           `json:"brandName" gorm:"size:140"`
	BrandContact   string           `json:"brandContact" gorm:"size:200"`
	LogoImageID    *string          `json:"logoImageId" gorm:"size:36"`
}

type MarketingCreation struct {
	ID string `json:"id" gorm:"type:char(36);primaryKey"`
	MarketingContent
	Timestamps
}

func (item *MarketingCreation) BeforeCreate(_ *gorm.DB) error { assignID(&item.ID); return nil }

type MarketingImage struct {
	ID       string `json:"id" gorm:"type:char(36);primaryKey"`
	MimeType string `json:"mimeType" gorm:"size:32"`
	Data     []byte `json:"-"`
	Timestamps
}

func (item *MarketingImage) BeforeCreate(_ *gorm.DB) error { assignID(&item.ID); return nil }
