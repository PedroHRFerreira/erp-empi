package database

import (
	"context"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	financialservices "github.com/empi-autocenter/erp-empi/internal/domain/financial/services"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestLegacyProductionMigrationPreservesDataAndIsIdempotent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:legacy-production?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	createLegacySchema(t, db)

	createdAt := "2026-07-02T12:00:00Z"
	archivedAt := "2026-07-03T12:00:00Z"
	if err := db.Exec(`INSERT INTO expenses (id, description, category, amount_cents, spent_at, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`, "expense-active", "Energia", "energia", 2500, createdAt, "", createdAt, createdAt).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO expenses (id, description, category, amount_cents, spent_at, notes, archived_at, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`, "expense-archived", "Antigo", "outros", 900, archivedAt, "", archivedAt, archivedAt, archivedAt).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO stock_items (id, name, description, cost_cents, markup_percent, resale_price_cents, quantity, used_quantity, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, "stock-legacy", "Filtro", "", 1000, 50, 1500, 5, 3, true, createdAt, createdAt).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`INSERT INTO stock_items (id, name, description, cost_cents, markup_percent, resale_price_cents, quantity, used_quantity, active, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, "stock-zero", "Item zerado", "", 500, 50, 750, 0, 0, false, createdAt, createdAt).Error; err != nil {
		t.Fatal(err)
	}

	if err := AutoMigrate(db); err != nil {
		t.Fatalf("schema migration failed: %v", err)
	}
	if err := requireLegacyMigrationApplied(db); err == nil {
		t.Fatal("API startup should be blocked while legacy data migration is pending")
	}
	if err := Migrate(db); err != nil {
		t.Fatalf("first migration failed: %v", err)
	}
	if err := requireLegacyMigrationApplied(db); err != nil {
		t.Fatalf("API startup remained blocked after migration: %v", err)
	}
	assertLegacyMigration(t, db)

	if err := Migrate(db); err != nil {
		t.Fatalf("second migration failed: %v", err)
	}
	assertLegacyMigration(t, db)

	start, _ := time.Parse(time.RFC3339, "2026-07-01T00:00:00Z")
	end, _ := time.Parse(time.RFC3339, "2026-08-01T00:00:00Z")
	summary, err := financialservices.NewFinancialService(db).Summary(context.Background(), start, end)
	if err != nil {
		t.Fatal(err)
	}
	if summary.OperationalExpensesCents != 2500 || summary.ExpensesCount != 1 {
		t.Fatalf("legacy operational totals changed: %+v", summary)
	}
	if summary.StockExpensesCents != 8000 || summary.StockPaymentsCount != 2 {
		t.Fatalf("legacy stock totals missing: %+v", summary)
	}
	rows, total, err := financialservices.NewFinancialService(db).RealizedExpenses(context.Background(), financialservices.RealizedExpenseAll, 20, 0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(rows) != 3 {
		t.Fatalf("legacy realized rows missing: total=%d rows=%+v", total, rows)
	}
	for _, row := range rows {
		if row.PaymentMethod != entities.PaymentMethodLegacy {
			t.Fatalf("legacy payment method missing: %+v", row)
		}
	}
}

func createLegacySchema(t *testing.T, db *gorm.DB) {
	t.Helper()
	statements := []string{
		`CREATE TABLE expenses (id character(36) PRIMARY KEY, receipt_id character(36), description varchar(180) NOT NULL, category varchar(80) NOT NULL, amount_cents bigint NOT NULL, spent_at datetime NOT NULL, notes varchar(700), archived_at datetime, created_at datetime, updated_at datetime)`,
		`CREATE TABLE stock_items (id character(36) PRIMARY KEY, name varchar(140) NOT NULL, description varchar(700), cost_cents bigint NOT NULL, markup_percent numeric NOT NULL, resale_price_cents bigint NOT NULL, quantity bigint NOT NULL, used_quantity bigint NOT NULL, active boolean NOT NULL, created_at datetime, updated_at datetime)`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatal(err)
		}
	}
}

func assertLegacyMigration(t *testing.T, db *gorm.DB) {
	t.Helper()
	var activeInstallments []entities.PayableInstallment
	if err := db.Where("expense_id = ?", "expense-active").Find(&activeInstallments).Error; err != nil {
		t.Fatal(err)
	}
	if len(activeInstallments) != 1 || activeInstallments[0].Status != entities.PayablePaid || activeInstallments[0].PaymentMethod != entities.PaymentMethodLegacy || activeInstallments[0].CashEntryID != nil {
		t.Fatalf("unexpected active expense migration: %+v", activeInstallments)
	}
	var archivedCount int64
	if err := db.Model(&entities.PayableInstallment{}).Where("expense_id = ?", "expense-archived").Count(&archivedCount).Error; err != nil {
		t.Fatal(err)
	}
	if archivedCount != 0 {
		t.Fatalf("archived expense received %d installments", archivedCount)
	}

	var purchases []entities.StockPurchase
	if err := db.Preload("Items").Preload("Installments").Find(&purchases).Error; err != nil {
		t.Fatal(err)
	}
	if len(purchases) != 2 {
		t.Fatalf("unexpected stock migration: %+v", purchases)
	}
	var total int64
	for _, purchase := range purchases {
		total += purchase.TotalCents
		if len(purchase.Items) != 1 || len(purchase.Installments) != 1 || purchase.Installments[0].PaymentMethod != entities.PaymentMethodLegacy || purchase.Installments[0].CashEntryID != nil {
			t.Fatalf("unexpected stock migration: %+v", purchases)
		}
	}
	if total != 8000 {
		t.Fatalf("legacy purchase total = %d", total)
	}
	var stock entities.StockItem
	if err := db.First(&stock, "id = ?", "stock-legacy").Error; err != nil {
		t.Fatal(err)
	}
	if stock.Quantity != 5 || stock.UsedQuantity != 3 || stock.CostCents != 1000 || stock.ResalePriceCents != 1500 {
		t.Fatalf("stock balance changed: %+v", stock)
	}
	var cashCount int64
	if err := db.Model(&entities.CashEntry{}).Count(&cashCount).Error; err != nil {
		t.Fatal(err)
	}
	if cashCount != 0 {
		t.Fatalf("legacy migration created %d cash entries", cashCount)
	}
	var migrationCount int64
	if err := db.Model(&dataMigration{}).Where("id = ?", legacyProductionMigrationID).Count(&migrationCount).Error; err != nil {
		t.Fatal(err)
	}
	if migrationCount != 1 {
		t.Fatalf("migration marker count = %d", migrationCount)
	}
}
