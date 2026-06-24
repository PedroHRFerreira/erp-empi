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
	if err := dropLegacyUserCPFUniqueIndex(db); err != nil {
		return err
	}

	return db.AutoMigrate(
		new(entities.User),
		new(entities.StockItem),
		new(entities.Receipt),
		new(entities.ReceiptItem),
		new(entities.Expense),
	)
}

func dropLegacyUserCPFUniqueIndex(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	if !db.Migrator().HasTable(new(entities.User)) {
		return nil
	}
	if err := db.Exec("ALTER TABLE IF EXISTS users DROP CONSTRAINT IF EXISTS users_cpf_key").Error; err != nil {
		return err
	}
	if err := db.Exec("DROP INDEX IF EXISTS idx_users_identity").Error; err != nil {
		return err
	}
	if err := db.Exec("ALTER TABLE IF EXISTS users ALTER COLUMN cpf DROP NOT NULL").Error; err != nil {
		return err
	}
	return db.Exec("CREATE INDEX IF NOT EXISTS idx_users_phone ON users(phone)").Error
}
