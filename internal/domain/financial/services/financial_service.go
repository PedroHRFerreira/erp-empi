package services

import (
	"context"
	"sort"
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

type ReceiptCostSummary struct {
	ReceiptID            string `json:"receiptId"`
	ClientName           string `json:"clientName"`
	VehicleModel         string `json:"vehicleModel"`
	VehiclePlate         string `json:"vehiclePlate"`
	ServiceExpensesCents int64  `json:"serviceExpensesCents"`
	ProductCostCents     int64  `json:"productCostCents"`
	TotalCostCents       int64  `json:"totalCostCents"`
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
	ReceiptCosts             []ReceiptCostSummary     `json:"receiptCosts"`
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
	receiptCosts, err := service.loadReceiptCosts(ctx, start, end)
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
		ReceiptCosts:             receiptCosts,
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

func (service *FinancialService) loadReceiptCosts(ctx context.Context, start time.Time, end time.Time) ([]ReceiptCostSummary, error) {
	type expenseCostRow struct {
		ReceiptID            string
		ServiceExpensesCents int64
	}
	type productCostRow struct {
		ReceiptID        string
		ProductCostCents int64
	}

	var expenseRows []expenseCostRow
	err := service.db.WithContext(ctx).
		Table("expenses").
		Select("expenses.receipt_id, COALESCE(SUM(expenses.amount_cents), 0) AS service_expenses_cents").
		Joins("JOIN receipts ON receipts.id = expenses.receipt_id").
		Where("expenses.archived_at IS NULL").
		Where("expenses.receipt_id IS NOT NULL").
		Where("receipts.status <> ?", entities.ReceiptStatusCancelled).
		Where("expenses.spent_at >= ? AND expenses.spent_at < ?", start, end).
		Group("expenses.receipt_id").
		Scan(&expenseRows).
		Error
	if err != nil {
		return nil, err
	}

	var productRows []productCostRow
	err = service.db.WithContext(ctx).
		Table("receipt_items").
		Select("receipts.id AS receipt_id, COALESCE(SUM(receipt_items.unit_cost_cents * receipt_items.quantity), 0) AS product_cost_cents").
		Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipts.status = ?", entities.ReceiptStatusPaid).
		Where("COALESCE(receipts.paid_at, receipts.updated_at) >= ? AND COALESCE(receipts.paid_at, receipts.updated_at) < ?", start, end).
		Group("receipts.id").
		Scan(&productRows).
		Error
	if err != nil {
		return nil, err
	}

	costsByReceiptID := map[string]*ReceiptCostSummary{}
	receiptIDs := make([]string, 0, len(expenseRows)+len(productRows))

	ensureCost := func(receiptID string) *ReceiptCostSummary {
		if cost, exists := costsByReceiptID[receiptID]; exists {
			return cost
		}
		cost := &ReceiptCostSummary{ReceiptID: receiptID}
		costsByReceiptID[receiptID] = cost
		receiptIDs = append(receiptIDs, receiptID)
		return cost
	}

	for _, row := range expenseRows {
		ensureCost(row.ReceiptID).ServiceExpensesCents = row.ServiceExpensesCents
	}
	for _, row := range productRows {
		ensureCost(row.ReceiptID).ProductCostCents = row.ProductCostCents
	}
	if len(receiptIDs) == 0 {
		return []ReceiptCostSummary{}, nil
	}

	var receipts []entities.Receipt
	err = service.db.WithContext(ctx).
		Preload("User").
		Where("id IN ?", receiptIDs).
		Find(&receipts).
		Error
	if err != nil {
		return nil, err
	}

	for _, receipt := range receipts {
		cost := costsByReceiptID[receipt.ID]
		if cost == nil {
			continue
		}
		cost.ClientName = receiptClientName(receipt)
		cost.VehicleModel = receiptVehicleModel(receipt)
		cost.VehiclePlate = receiptVehiclePlate(receipt)
		cost.TotalCostCents = cost.ServiceExpensesCents + cost.ProductCostCents
	}

	receiptCosts := make([]ReceiptCostSummary, 0, len(costsByReceiptID))
	for _, cost := range costsByReceiptID {
		if cost.TotalCostCents > 0 {
			receiptCosts = append(receiptCosts, *cost)
		}
	}
	sort.Slice(receiptCosts, func(i int, j int) bool {
		if receiptCosts[i].TotalCostCents == receiptCosts[j].TotalCostCents {
			return receiptCosts[i].ClientName < receiptCosts[j].ClientName
		}
		return receiptCosts[i].TotalCostCents > receiptCosts[j].TotalCostCents
	})
	if len(receiptCosts) > 5 {
		receiptCosts = receiptCosts[:5]
	}
	return receiptCosts, nil
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

func receiptClientName(receipt entities.Receipt) string {
	if receipt.User != nil && receipt.User.Name != "" {
		return receipt.User.Name
	}
	return "Recibo rápido"
}

func receiptVehicleModel(receipt entities.Receipt) string {
	if receipt.VehicleModel != "" {
		return receipt.VehicleModel
	}
	return "Sem veículo"
}

func receiptVehiclePlate(receipt entities.Receipt) string {
	if receipt.VehiclePlate != "" {
		return receipt.VehiclePlate
	}
	return "-"
}
