package database

import (
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Client struct {
	DB *gorm.DB
}

func NewPostgresClient(dsn string) (*Client, error) {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	if err := AutoMigrate(db); err != nil {
		return nil, err
	}
	return &Client{DB: db}, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		new(entities.User),
		new(entities.StockItem),
		new(entities.Receipt),
		new(entities.ReceiptItem),
	)
}
