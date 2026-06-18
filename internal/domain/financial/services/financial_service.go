package services

import (
	"context"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"gorm.io/gorm"
)

type HealthStatus string

const (
	HealthStatusRed    HealthStatus = "red"
	HealthStatusYellow HealthStatus = "yellow"
	HealthStatusGreen  HealthStatus = "green"
)

type FinancialService struct {
	db *gorm.DB
}

type ExpenseCategorySummary struct {
	Category    string `json:"category"`
	AmountCents int64  `json:"amountCents"`
	Count       int64  `json:"count"`
}

type Summary struct {
	StartDate                string                   `json:"startDate"`
	EndDate                  string                   `json:"endDate"`
	PaidReceiptsCount        int64                    `json:"paidReceiptsCount"`
	ExpensesCount            int64                    `json:"expensesCount"`
	RevenuePaidCents         int64                    `json:"revenuePaidCents"`
	ProductCostCents         int64                    `json:"productCostCents"`
	CardFeesCents            int64                    `json:"cardFeesCents"`
	GrossProfitCents         int64                    `json:"grossProfitCents"`
	OperationalExpensesCents int64                    `json:"operationalExpensesCents"`
	OperationalProfitCents   int64                    `json:"operationalProfitCents"`
	NetProfitCents           int64                    `json:"netProfitCents"`
	NetMarginPercent         float64                  `json:"netMarginPercent"`
	HealthStatus             HealthStatus             `json:"healthStatus"`
	ExpensesByCategory       []ExpenseCategorySummary `json:"expensesByCategory"`
}

type receiptTotals struct {
	PaidReceiptsCount int64
	RevenuePaidCents  int64
	CardFeesCents     int64
}

type expenseTotals struct {
	ExpensesCount            int64
	OperationalExpensesCents int64
}

func NewFinancialService(db *gorm.DB) *FinancialService {
	return &FinancialService{db: db}
}

func (service *FinancialService) Summary(ctx context.Context, start time.Time, end time.Time) (*Summary, error) {
	receipts, err := service.loadReceiptTotals(ctx, start, end)
	if err != nil {
		return nil, err
	}
	productCostCents, err := service.loadProductCost(ctx, start, end)
	if err != nil {
		return nil, err
	}
	expenses, err := service.loadExpenseTotals(ctx, start, end)
	if err != nil {
		return nil, err
	}
	categoryTotals, err := service.loadExpensesByCategory(ctx, start, end)
	if err != nil {
		return nil, err
	}

	grossProfitCents := receipts.RevenuePaidCents - productCostCents
	operationalProfitCents := grossProfitCents - expenses.OperationalExpensesCents
	netProfitCents := operationalProfitCents - receipts.CardFeesCents
	netMarginPercent := 0.0
	if receipts.RevenuePaidCents > 0 {
		netMarginPercent = (float64(netProfitCents) / float64(receipts.RevenuePaidCents)) * 100
	}

	return &Summary{
		StartDate:                start.Format("2006-01-02"),
		EndDate:                  end.AddDate(0, 0, -1).Format("2006-01-02"),
		PaidReceiptsCount:        receipts.PaidReceiptsCount,
		ExpensesCount:            expenses.ExpensesCount,
		RevenuePaidCents:         receipts.RevenuePaidCents,
		ProductCostCents:         productCostCents,
		CardFeesCents:            receipts.CardFeesCents,
		GrossProfitCents:         grossProfitCents,
		OperationalExpensesCents: expenses.OperationalExpensesCents,
		OperationalProfitCents:   operationalProfitCents,
		NetProfitCents:           netProfitCents,
		NetMarginPercent:         netMarginPercent,
		HealthStatus:             healthStatus(netProfitCents, netMarginPercent),
		ExpensesByCategory:       categoryTotals,
	}, nil
}

func (service *FinancialService) loadReceiptTotals(ctx context.Context, start time.Time, end time.Time) (*receiptTotals, error) {
	var totals receiptTotals
	err := service.db.WithContext(ctx).
		Model(&entities.Receipt{}).
		Where("status = ?", entities.ReceiptStatusPaid).
		Where("COALESCE(paid_at, updated_at) >= ? AND COALESCE(paid_at, updated_at) < ?", start, end).
		Select(`
			COUNT(*) AS paid_receipts_count,
			COALESCE(SUM(price_cents), 0) AS revenue_paid_cents,
			COALESCE(SUM(card_fee_cents), 0) AS card_fees_cents
		`).
		Scan(&totals).
		Error
	return &totals, err
}

func (service *FinancialService) loadProductCost(ctx context.Context, start time.Time, end time.Time) (int64, error) {
	var productCostCents int64
	err := service.db.WithContext(ctx).
		Table("receipt_items").
		Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipts.status = ?", entities.ReceiptStatusPaid).
		Where("COALESCE(receipts.paid_at, receipts.updated_at) >= ? AND COALESCE(receipts.paid_at, receipts.updated_at) < ?", start, end).
		Select("COALESCE(SUM(receipt_items.unit_cost_cents * receipt_items.quantity), 0)").
		Scan(&productCostCents).
		Error
	return productCostCents, err
}

func (service *FinancialService) loadExpenseTotals(ctx context.Context, start time.Time, end time.Time) (*expenseTotals, error) {
	var totals expenseTotals
	err := service.db.WithContext(ctx).
		Model(&entities.Expense{}).
		Where("archived_at IS NULL").
		Where("spent_at >= ? AND spent_at < ?", start, end).
		Select(`
			COUNT(*) AS expenses_count,
			COALESCE(SUM(amount_cents), 0) AS operational_expenses_cents
		`).
		Scan(&totals).
		Error
	return &totals, err
}

func (service *FinancialService) loadExpensesByCategory(ctx context.Context, start time.Time, end time.Time) ([]ExpenseCategorySummary, error) {
	var totals []ExpenseCategorySummary
	err := service.db.WithContext(ctx).
		Model(&entities.Expense{}).
		Where("archived_at IS NULL").
		Where("spent_at >= ? AND spent_at < ?", start, end).
		Select("category, COALESCE(SUM(amount_cents), 0) AS amount_cents, COUNT(*) AS count").
		Group("category").
		Order("amount_cents desc, category asc").
		Scan(&totals).
		Error
	if totals == nil {
		totals = []ExpenseCategorySummary{}
	}
	return totals, err
}

func healthStatus(netProfitCents int64, netMarginPercent float64) HealthStatus {
	if netProfitCents < 0 {
		return HealthStatusRed
	}
	if netMarginPercent < 15 {
		return HealthStatusYellow
	}
	return HealthStatusGreen
}
