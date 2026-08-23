package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	financialservices "github.com/empi-autocenter/erp-empi/internal/domain/financial/services"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestRealizedExpensesCombinesOnlyPaidOutflowsWithoutDuplication(t *testing.T) {
	t.Parallel()
	db := financialTestDB(t)
	ctx := context.Background()
	service := financialservices.NewFinancialService(db)
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	operational := entities.Expense{Description: "Energia", Category: "energia", AmountCents: 12000, SpentAt: start.AddDate(0, 0, 2), AffectsProfit: true}
	if err := db.Create(&operational).Error; err != nil {
		t.Fatal(err)
	}
	paidAt := start.AddDate(0, 0, 2)
	installment := entities.PayableInstallment{ExpenseID: &operational.ID, Number: 1, AmountCents: 12000, DueDate: paidAt, Status: entities.PayablePaid, PlannedMethod: entities.PayableMethodPix, PaymentMethod: entities.PaymentMethodPix, PaidAt: &paidAt}
	if err := db.Create(&installment).Error; err != nil {
		t.Fatal(err)
	}
	operationalEntry := entities.CashEntry{Direction: entities.CashEntryOut, Kind: entities.CashEntryExpense, PaymentMethod: entities.PaymentMethodPix, AmountCents: -12000, Description: operational.Description, ReferenceType: "payable_installment", ReferenceID: installment.ID, OccurredAt: paidAt}
	if err := db.Create(&operationalEntry).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&installment).Update("cash_entry_id", operationalEntry.ID).Error; err != nil {
		t.Fatal(err)
	}

	stockItem := entities.StockItem{Name: "Filtro", CostCents: 10000, ResalePriceCents: 11000, MarkupPercent: 10, Active: true}
	if err := db.Create(&stockItem).Error; err != nil {
		t.Fatal(err)
	}
	cash := cashservices.NewCashService(db)
	purchase, err := cash.CreatePurchase(ctx, cashservices.PurchaseInput{SupplierName: "Fornecedor Alfa", Items: []cashservices.PurchaseItemInput{{StockItemID: stockItem.ID, Quantity: 3, UnitCostCents: 10000}}, Installments: []cashservices.PurchaseInstallmentInput{{AmountCents: 10000, DueDate: "2026-08-06", PlannedMethod: entities.PayableMethodDebitCard}, {AmountCents: 20000, DueDate: "2026-09-06", PlannedMethod: entities.PayableMethodBoleto}}})
	if err != nil {
		t.Fatal(err)
	}
	pendingRows, pendingTotal, err := service.RealizedExpenses(ctx, financialservices.RealizedExpenseStock, 10, 0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if pendingTotal != 0 || len(pendingRows) != 0 {
		t.Fatalf("pending installments must not be realized: total=%d rows=%+v", pendingTotal, pendingRows)
	}
	paid, err := cash.PayInstallment(ctx, purchase.Installments[0].ID, entities.PaymentMethodDebitCard)
	if err != nil {
		t.Fatal(err)
	}

	rows, total, err := service.RealizedExpenses(ctx, financialservices.RealizedExpenseAll, 10, 0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(rows) != 2 {
		t.Fatalf("expected two unique realized outflows, total=%d rows=%d", total, len(rows))
	}
	if rows[0].Origin != financialservices.RealizedExpenseStock || rows[0].AmountCents != 10000 || rows[0].SupplierName != purchase.SupplierName || rows[0].Editable {
		t.Fatalf("unexpected stock row: %+v", rows[0])
	}
	if rows[1].Origin != financialservices.RealizedExpenseOperational || rows[1].PaymentMethod != entities.PaymentMethodPix || rows[1].Editable {
		t.Fatalf("unexpected operational row: %+v", rows[1])
	}

	stock, stockTotal, err := service.RealizedExpenses(ctx, financialservices.RealizedExpenseStock, 10, 0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if stockTotal != 1 || len(stock) != 1 || stock[0].InstallmentID != paid.ID {
		t.Fatalf("unexpected stock filter: total=%d rows=%+v", stockTotal, stock)
	}
	operationalRows, operationalTotal, err := service.RealizedExpenses(ctx, financialservices.RealizedExpenseOperational, 10, 0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if operationalTotal != 1 || len(operationalRows) != 1 || operationalRows[0].ID != operationalEntry.ID || operationalRows[0].ExpenseID != operational.ID {
		t.Fatalf("unexpected operational filter: total=%d rows=%+v", operationalTotal, operationalRows)
	}
	page, allTotal, err := service.RealizedExpenses(ctx, financialservices.RealizedExpenseAll, 1, 1, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if allTotal != 2 || len(page) != 1 || page[0].Origin != financialservices.RealizedExpenseOperational {
		t.Fatalf("unexpected page: total=%d rows=%+v", allTotal, page)
	}

	summary, err := service.Summary(ctx, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if summary.OperationalExpensesCents != 12000 || summary.StockExpensesCents != 10000 || summary.TotalRealizedExpensesCents != 22000 || summary.StockPaymentsCount != 1 {
		t.Fatalf("unexpected realized totals: %+v", summary)
	}
}

func TestRealizedExpensesRejectsUnknownOriginAndRespectsPeriod(t *testing.T) {
	t.Parallel()
	db := financialTestDB(t)
	service := financialservices.NewFinancialService(db)
	start := time.Date(2026, 8, 1, 0, 0, 0, 0, time.Local)
	expense := entities.Expense{Description: "Fora do período", Category: "outros", AmountCents: 500, SpentAt: start.AddDate(0, -1, 0), AffectsProfit: true}
	if err := db.Create(&expense).Error; err != nil {
		t.Fatal(err)
	}
	rows, total, err := service.RealizedExpenses(context.Background(), financialservices.RealizedExpenseOperational, 10, 0, start, start.AddDate(0, 1, 0))
	if err != nil || total != 0 || len(rows) != 0 {
		t.Fatalf("expected empty period, total=%d rows=%+v err=%v", total, rows, err)
	}
	_, _, err = service.RealizedExpenses(context.Background(), "future", 10, 0, start, start.AddDate(0, 1, 0))
	if !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func financialTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	return db
}
