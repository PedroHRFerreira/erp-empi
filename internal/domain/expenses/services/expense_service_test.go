package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	expenserepos "github.com/empi-autocenter/erp-empi/internal/domain/expenses/repositories"
	expenseservices "github.com/empi-autocenter/erp-empi/internal/domain/expenses/services"
	financialservices "github.com/empi-autocenter/erp-empi/internal/domain/financial/services"
	receiptrepos "github.com/empi-autocenter/erp-empi/internal/domain/receipts/repositories"
	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	stockrepos "github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	stockservices "github.com/empi-autocenter/erp-empi/internal/domain/stock/services"
	userrepos "github.com/empi-autocenter/erp-empi/internal/domain/users/repositories"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestExpenseServiceCreatesUpdatesAndArchivesExpenses(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testDB(t)
	expenseService := expenseservices.NewExpenseService(expenserepos.NewExpenseRepository(db))
	start, end := currentDayRange()

	expense, err := expenseService.Create(ctx, expenseservices.ExpenseInput{
		Description: "Conta de luz",
		Category:    "energia",
		AmountCents: 20000,
		SpentAt:     start.Format("2006-01-02"),
		Notes:       "Pagamento mensal",
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := expenseService.Update(ctx, expense.ID, expenseservices.ExpenseInput{
		Description: "Conta de luz ajustada",
		Category:    "energia",
		AmountCents: 21000,
		SpentAt:     start.Format("2006-01-02"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Description != "Conta de luz ajustada" || updated.AmountCents != 21000 {
		t.Fatalf("expected updated expense, got %+v", updated)
	}

	expenses, total, err := expenseService.List(ctx, 10, 0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 || len(expenses) != 1 {
		t.Fatalf("expected one active expense, got total %d and len %d", total, len(expenses))
	}

	if err := expenseService.Archive(ctx, expense.ID); err != nil {
		t.Fatal(err)
	}
	expenses, total, err = expenseService.List(ctx, 10, 0, start, end)
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(expenses) != 0 {
		t.Fatalf("expected archived expense to be hidden, got total %d and len %d", total, len(expenses))
	}
}

func TestExpenseServiceCreatesAndUpdatesReceiptLink(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testDB(t)
	userRepo := userrepos.NewUserRepository(db)
	stockRepo := stockrepos.NewStockRepository(db)
	receiptRepo := receiptrepos.NewReceiptRepository(db)
	userService := userservices.NewUserService(userRepo)
	receiptService := receiptservices.NewReceiptService(receiptRepo, stockRepo, userService)
	expenseService := expenseservices.NewExpenseService(expenserepos.NewExpenseRepository(db))
	admin := createAdmin(t, ctx, userRepo)
	start, _ := currentDayRange()

	receipt, err := receiptService.Create(ctx, admin.ID, receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name:  "Cliente Vinculado",
			Phone: "33999990000",
		},
		VehicleModel:    "Gol",
		VehicleYear:     2020,
		VehiclePlate:    "ABC1D23",
		Services:        "Diagnostico",
		LaborPriceCents: 10000,
	})
	if err != nil {
		t.Fatal(err)
	}

	expense, err := expenseService.Create(ctx, expenseservices.ExpenseInput{
		Description: "Gasolina",
		Category:    "combustível",
		AmountCents: 3000,
		SpentAt:     start.Format("2006-01-02"),
		ReceiptID:   &receipt.ID,
	})
	if err != nil {
		t.Fatal(err)
	}
	if expense.ReceiptID == nil || *expense.ReceiptID != receipt.ID {
		t.Fatalf("expected expense linked to receipt %s, got %+v", receipt.ID, expense.ReceiptID)
	}

	updated, err := expenseService.Update(ctx, expense.ID, expenseservices.ExpenseInput{
		Description: "Gasolina oficina",
		Category:    "combustível",
		AmountCents: 3500,
		SpentAt:     start.Format("2006-01-02"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ReceiptID != nil {
		t.Fatalf("expected receipt link to be cleared, got %+v", updated.ReceiptID)
	}
}

func TestFinancialSummaryUsesPaidReceiptsExpensesCostsAndCardFees(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testDB(t)
	userRepo := userrepos.NewUserRepository(db)
	stockRepo := stockrepos.NewStockRepository(db)
	receiptRepo := receiptrepos.NewReceiptRepository(db)
	userService := userservices.NewUserService(userRepo)
	stockService := stockservices.NewStockService(stockRepo)
	receiptService := receiptservices.NewReceiptService(receiptRepo, stockRepo, userService)
	expenseService := expenseservices.NewExpenseService(expenserepos.NewExpenseRepository(db))
	financialService := financialservices.NewFinancialService(db)
	admin := createAdmin(t, ctx, userRepo)
	start, end := currentDayRange()

	stockItem, err := stockService.Create(ctx, stockservices.StockInput{
		Name:          "Filtro",
		CostCents:     5000,
		MarkupPercent: 0,
		Quantity:      10,
	})
	if err != nil {
		t.Fatal(err)
	}

	paidReceipt, err := receiptService.Create(ctx, admin.ID, receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name:  "Cliente Pago",
			Phone: "33999990000",
		},
		VehicleModel:    "Gol",
		VehicleYear:     2020,
		VehiclePlate:    "ABC1D23",
		Services:        "Servico",
		LaborPriceCents: 10000,
		PaymentMethod:   entities.PaymentMethodCreditCard,
		Installments:    1,
		Items: []receiptservices.ReceiptItemInput{
			{StockItemID: stockItem.ID, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := receiptService.MarkPaid(ctx, paidReceipt.ID); err != nil {
		t.Fatal(err)
	}

	if _, err := receiptService.Create(ctx, admin.ID, receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name:  "Cliente Pendente",
			Phone: "33888880000",
		},
		VehicleModel:    "Uno",
		VehicleYear:     2021,
		VehiclePlate:    "DEF1D23",
		Services:        "Servico pendente",
		LaborPriceCents: 50000,
		Items: []receiptservices.ReceiptItemInput{
			{StockItemID: stockItem.ID, Quantity: 1},
		},
	}); err != nil {
		t.Fatal(err)
	}

	if _, err := expenseService.Create(ctx, expenseservices.ExpenseInput{
		Description: "Gasolina",
		Category:    "combustível",
		AmountCents: 3000,
		SpentAt:     start.Format("2006-01-02"),
		ReceiptID:   &paidReceipt.ID,
	}); err != nil {
		t.Fatal(err)
	}

	summary, err := financialService.Summary(ctx, start, end)
	if err != nil {
		t.Fatal(err)
	}

	if summary.RevenuePaidCents != 22000 {
		t.Fatalf("expected revenue 22000, got %d", summary.RevenuePaidCents)
	}
	if summary.ProductCostCents != 10000 {
		t.Fatalf("expected product cost 10000, got %d", summary.ProductCostCents)
	}
	if summary.CardFeesCents != 2000 {
		t.Fatalf("expected card fees 2000, got %d", summary.CardFeesCents)
	}
	if summary.GrossProfitCents != 12000 {
		t.Fatalf("expected gross profit 12000, got %d", summary.GrossProfitCents)
	}
	if summary.OperationalExpensesCents != 3000 {
		t.Fatalf("expected expenses 3000, got %d", summary.OperationalExpensesCents)
	}
	if summary.OperationalProfitCents != 9000 {
		t.Fatalf("expected operational profit 9000, got %d", summary.OperationalProfitCents)
	}
	if summary.NetProfitCents != 7000 {
		t.Fatalf("expected net profit 7000, got %d", summary.NetProfitCents)
	}
	if summary.HealthStatus != financialservices.HealthStatusGreen {
		t.Fatalf("expected green health status, got %s", summary.HealthStatus)
	}
	if summary.PaidReceiptsCount != 1 || summary.ExpensesCount != 1 {
		t.Fatalf("expected one paid receipt and one expense, got %+v", summary)
	}
	if len(summary.ReceiptCosts) != 1 {
		t.Fatalf("expected one receipt cost summary, got %d", len(summary.ReceiptCosts))
	}
	if summary.ReceiptCosts[0].ReceiptID != paidReceipt.ID {
		t.Fatalf("expected receipt cost for paid receipt, got %+v", summary.ReceiptCosts[0])
	}
	if summary.ReceiptCosts[0].ServiceExpensesCents != 3000 ||
		summary.ReceiptCosts[0].ProductCostCents != 10000 ||
		summary.ReceiptCosts[0].TotalCostCents != 13000 {
		t.Fatalf("unexpected receipt cost summary: %+v", summary.ReceiptCosts[0])
	}
}

func TestExpenseServiceRejectsInvalidExpense(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db := testDB(t)
	expenseService := expenseservices.NewExpenseService(expenserepos.NewExpenseRepository(db))

	_, err := expenseService.Create(ctx, expenseservices.ExpenseInput{
		Description: "Sem valor",
		Category:    "outros",
		AmountCents: 0,
		SpentAt:     "2026-06-18",
	})
	if !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected invalid input for zero amount, got %v", err)
	}

	_, err = expenseService.Create(ctx, expenseservices.ExpenseInput{
		Description: "Sem data",
		Category:    "outros",
		AmountCents: 1000,
	})
	if !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected invalid input for empty date, got %v", err)
	}
}

func testDB(t *testing.T) *gorm.DB {
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

func currentDayRange() (time.Time, time.Time) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.Local)
	return start, start.AddDate(0, 0, 1)
}

func createAdmin(t *testing.T, ctx context.Context, repo *userrepos.UserRepository) *entities.User {
	t.Helper()

	admin := &entities.User{
		Name:              "Admin",
		CPF:               "52998224725",
		Type:              entities.UserTypeAdmin,
		MarkupPercent:     10,
		MachineFeePercent: 10,
	}
	if err := repo.Create(ctx, admin); err != nil {
		t.Fatal(err)
	}
	return admin
}
