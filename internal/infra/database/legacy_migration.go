package database

import (
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"gorm.io/gorm"
)

const legacyProductionMigrationID = "2026-08-23-legacy-production-baseline"

type dataMigration struct {
	ID        string    `gorm:"size:100;primaryKey"`
	AppliedAt time.Time `gorm:"not null"`
}

func (dataMigration) TableName() string { return "data_migrations" }

func requireLegacyMigrationApplied(db *gorm.DB) error {
	var pendingExpenses int64
	if err := db.Model(&entities.Expense{}).
		Where("archived_at IS NULL AND NOT EXISTS (?)", db.Model(&entities.PayableInstallment{}).Select("1").Where("payable_installments.expense_id = expenses.id")).
		Count(&pendingExpenses).Error; err != nil {
		return err
	}
	var pendingStock int64
	if err := db.Model(&entities.StockItem{}).
		Where("NOT EXISTS (?)", db.Model(&entities.StockPurchaseItem{}).Select("1").Where("stock_purchase_items.stock_item_id = stock_items.id")).
		Count(&pendingStock).Error; err != nil {
		return err
	}
	if pendingExpenses+pendingStock > 0 {
		return fmt.Errorf("legacy data migration is pending for %d expenses and %d stock items; run cmd/migrate before starting the API", pendingExpenses, pendingStock)
	}
	return nil
}

// migrateLegacyProductionData converts records created before cash, purchases
// and payables existed. AutoMigrate calls it inside the same PostgreSQL
// transaction as the schema changes, so any failed invariant rolls everything
// back and prevents the API from starting on a partial migration.
func migrateLegacyProductionData(tx *gorm.DB) error {
	var applied dataMigration
	err := tx.First(&applied, "id = ?", legacyProductionMigrationID).Error
	alreadyApplied := err == nil
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if err := validateLegacyInputs(tx); err != nil {
		return err
	}
	if err := migrateLegacyExpenses(tx); err != nil {
		return err
	}
	if err := migrateLegacyStock(tx); err != nil {
		return err
	}
	if err := validateLegacyOutputs(tx); err != nil {
		return err
	}
	if alreadyApplied {
		return nil
	}
	return tx.Create(&dataMigration{ID: legacyProductionMigrationID, AppliedAt: time.Now()}).Error
}

func validateLegacyInputs(tx *gorm.DB) error {
	var invalidExpenses int64
	if err := tx.Model(&entities.Expense{}).
		Where("archived_at IS NULL AND amount_cents <= 0").
		Count(&invalidExpenses).Error; err != nil {
		return err
	}
	if invalidExpenses > 0 {
		return fmt.Errorf("legacy migration blocked: %d active expenses have non-positive values", invalidExpenses)
	}

	var invalidStock int64
	if err := tx.Model(&entities.StockItem{}).
		Where("cost_cents < 0 OR quantity < 0 OR used_quantity < 0").
		Count(&invalidStock).Error; err != nil {
		return err
	}
	if invalidStock > 0 {
		return fmt.Errorf("legacy migration blocked: %d stock items have negative values", invalidStock)
	}
	return nil
}

func migrateLegacyExpenses(tx *gorm.DB) error {
	var expenses []entities.Expense
	if err := tx.Where("archived_at IS NULL").Find(&expenses).Error; err != nil {
		return err
	}
	for index := range expenses {
		expense := &expenses[index]
		var count int64
		if err := tx.Model(&entities.PayableInstallment{}).Where("expense_id = ?", expense.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		paidAt := expense.SpentAt
		row := entities.PayableInstallment{
			ExpenseID: &expense.ID, Number: 1, AmountCents: expense.AmountCents,
			DueDate: expense.SpentAt, Status: entities.PayablePaid,
			PlannedMethod: entities.PayableMethodLegacy,
			PaymentMethod: entities.PaymentMethodLegacy, PaidAt: &paidAt,
		}
		if err := tx.Create(&row).Error; err != nil {
			return err
		}
	}
	return nil
}

func migrateLegacyStock(tx *gorm.DB) error {
	var stocks []entities.StockItem
	if err := tx.Find(&stocks).Error; err != nil {
		return err
	}
	for index := range stocks {
		stock := &stocks[index]
		var count int64
		if err := tx.Model(&entities.StockPurchaseItem{}).Where("stock_item_id = ?", stock.ID).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			continue
		}
		originalQuantity := stock.Quantity + stock.UsedQuantity
		if stock.CostCents > 0 && int64(originalQuantity) > math.MaxInt64/stock.CostCents {
			return fmt.Errorf("legacy migration blocked: stock item %s total overflows", stock.ID)
		}
		total := stock.CostCents * int64(originalQuantity)
		purchasedAt := stock.CreatedAt
		if purchasedAt.IsZero() {
			purchasedAt = stock.UpdatedAt
		}
		if purchasedAt.IsZero() {
			purchasedAt = time.Now()
		}
		purchase := entities.StockPurchase{
			StockItemID: stock.ID, SupplierName: "Saldo inicial legado — " + stock.Name,
			TotalCents: total, Status: entities.StockPurchaseConfirmed, PurchasedAt: purchasedAt,
			Items: []entities.StockPurchaseItem{{
				StockItemID: stock.ID, Quantity: originalQuantity, UnitCostCents: stock.CostCents,
				SubtotalCents: total, HasStockSnapshot: false,
			}},
		}
		if err := tx.Create(&purchase).Error; err != nil {
			return err
		}
		paidAt := purchasedAt
		purchaseID := purchase.ID
		installment := entities.PayableInstallment{
			StockPurchaseID: &purchaseID, Number: 1, AmountCents: total,
			DueDate: purchasedAt, Status: entities.PayablePaid,
			PlannedMethod: entities.PayableMethodLegacy,
			PaymentMethod: entities.PaymentMethodLegacy, PaidAt: &paidAt,
		}
		if err := tx.Create(&installment).Error; err != nil {
			return err
		}
	}
	return nil
}

func validateLegacyOutputs(tx *gorm.DB) error {
	var expensesWithoutInstallment int64
	if err := tx.Model(&entities.Expense{}).
		Where("archived_at IS NULL AND NOT EXISTS (?)", tx.Model(&entities.PayableInstallment{}).Select("1").Where("payable_installments.expense_id = expenses.id")).
		Count(&expensesWithoutInstallment).Error; err != nil {
		return err
	}
	if expensesWithoutInstallment > 0 {
		return fmt.Errorf("legacy migration incomplete: %d active expenses have no installment", expensesWithoutInstallment)
	}

	var stocksWithoutHistory int64
	if err := tx.Model(&entities.StockItem{}).
		Where("NOT EXISTS (?)", tx.Model(&entities.StockPurchaseItem{}).Select("1").Where("stock_purchase_items.stock_item_id = stock_items.id")).
		Count(&stocksWithoutHistory).Error; err != nil {
		return err
	}
	if stocksWithoutHistory > 0 {
		return fmt.Errorf("legacy migration incomplete: %d stock items have no purchase history", stocksWithoutHistory)
	}

	var legacyCashEntries int64
	if err := tx.Model(&entities.CashEntry{}).Where("payment_method = ?", entities.PaymentMethodLegacy).Count(&legacyCashEntries).Error; err != nil {
		return err
	}
	if legacyCashEntries > 0 {
		return fmt.Errorf("legacy migration invalid: %d legacy cash entries were created", legacyCashEntries)
	}
	return nil
}
