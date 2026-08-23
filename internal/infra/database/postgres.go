package database

import (
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	goalservices "github.com/empi-autocenter/erp-empi/internal/domain/goals/services"
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
	if err := requireLegacyMigrationApplied(db); err != nil {
		return nil, err
	}
	return &Client{DB: db}, nil
}

func AutoMigrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		return migrateSchema(tx)
	})
}

// Migrate is the explicit maintenance-window entry point. It applies schema
// and versioned data migrations atomically and is intentionally not invoked by
// normal API startup.
func Migrate(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := migrateSchema(tx); err != nil {
			return err
		}
		return migrateLegacyProductionData(tx)
	})
}

func migrateSchema(tx *gorm.DB) error {
	if err := dropLegacyUserCPFUniqueIndex(tx); err != nil {
		return err
	}
	if err := makePayableOriginNullable(tx); err != nil {
		return err
	}
	if err := tx.AutoMigrate(
		new(entities.User),
		new(entities.StockItem),
		new(entities.Receipt),
		new(entities.ReceiptItem),
		new(entities.Expense),
		new(entities.CashSession),
		new(entities.CashEntry),
		new(entities.StockPurchase),
		new(entities.StockPurchaseItem),
		new(entities.PayableInstallment),
		new(goalservices.MonthlyGoal),
		new(dataMigration),
	); err != nil {
		return err
	}
	if err := backfillLegacyPurchaseItems(tx); err != nil {
		return err
	}
	return nil
}

func makePayableOriginNullable(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" || !db.Migrator().HasTable(new(entities.PayableInstallment)) {
		return nil
	}
	return db.Exec("ALTER TABLE payable_installments ALTER COLUMN stock_purchase_id DROP NOT NULL").Error
}

func backfillLegacyPurchaseItems(db *gorm.DB) error {
	var purchases []entities.StockPurchase
	if err := db.Where("stock_item_id <> ''").Find(&purchases).Error; err != nil {
		return err
	}
	for _, purchase := range purchases {
		var count int64
		if err := db.Model(&entities.StockPurchaseItem{}).Where("stock_purchase_id = ?", purchase.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		item := entities.StockPurchaseItem{StockPurchaseID: purchase.ID, StockItemID: purchase.StockItemID, Quantity: 1, UnitCostCents: purchase.TotalCents, SubtotalCents: purchase.TotalCents}
		if err := db.Create(&item).Error; err != nil {
			return err
		}
	}
	return nil
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
