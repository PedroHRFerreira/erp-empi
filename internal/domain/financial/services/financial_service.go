package services

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
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

type RealizedExpenseOrigin string

const (
	RealizedExpenseAll         RealizedExpenseOrigin = "all"
	RealizedExpenseOperational RealizedExpenseOrigin = "operational"
	RealizedExpenseStock       RealizedExpenseOrigin = "stock"
)

// RealizedExpense is a read model. Operational expenses remain owned by the
// expenses domain; stock payments are immutable ledger movements.
type RealizedExpense struct {
	ID                string                 `json:"id"`
	Origin            RealizedExpenseOrigin  `json:"origin"`
	Description       string                 `json:"description"`
	Category          string                 `json:"category"`
	AmountCents       int64                  `json:"amountCents"`
	OccurredAt        time.Time              `json:"occurredAt"`
	PaymentMethod     entities.PaymentMethod `json:"paymentMethod,omitempty"`
	SupplierName      string                 `json:"supplierName,omitempty"`
	InstallmentNumber int                    `json:"installmentNumber,omitempty"`
	StockPurchaseID   string                 `json:"stockPurchaseId,omitempty"`
	ExpenseID         string                 `json:"expenseId,omitempty"`
	InstallmentID     string                 `json:"installmentId,omitempty"`
	Editable          bool                   `json:"editable"`
}

type Summary struct {
	StartDate                  string                   `json:"startDate"`
	EndDate                    string                   `json:"endDate"`
	PaidReceiptsCount          int64                    `json:"paidReceiptsCount"`
	ExpensesCount              int64                    `json:"expensesCount"`
	RevenuePaidCents           int64                    `json:"revenuePaidCents"`
	ProductCostCents           int64                    `json:"productCostCents"`
	CardFeesCents              int64                    `json:"cardFeesCents"`
	GrossProfitCents           int64                    `json:"grossProfitCents"`
	OperationalExpensesCents   int64                    `json:"operationalExpensesCents"`
	StockExpensesCents         int64                    `json:"stockExpensesCents"`
	TotalRealizedExpensesCents int64                    `json:"totalRealizedExpensesCents"`
	StockPaymentsCount         int64                    `json:"stockPaymentsCount"`
	OperationalProfitCents     int64                    `json:"operationalProfitCents"`
	NetProfitCents             int64                    `json:"netProfitCents"`
	NetMarginPercent           float64                  `json:"netMarginPercent"`
	HealthStatus               HealthStatus             `json:"healthStatus"`
	ExpensesByCategory         []ExpenseCategorySummary `json:"expensesByCategory"`
	ReceiptCosts               []ReceiptCostSummary     `json:"receiptCosts"`
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
	stockExpensesCents, stockPaymentsCount, err := service.loadStockExpenseTotals(ctx, start, end)
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
		StartDate:                  start.Format("2006-01-02"),
		EndDate:                    end.AddDate(0, 0, -1).Format("2006-01-02"),
		PaidReceiptsCount:          receipts.PaidReceiptsCount,
		ExpensesCount:              expenses.ExpensesCount,
		RevenuePaidCents:           receipts.RevenuePaidCents,
		ProductCostCents:           productCostCents,
		CardFeesCents:              receipts.CardFeesCents,
		GrossProfitCents:           grossProfitCents,
		OperationalExpensesCents:   expenses.OperationalExpensesCents,
		StockExpensesCents:         stockExpensesCents,
		TotalRealizedExpensesCents: expenses.OperationalExpensesCents + stockExpensesCents,
		StockPaymentsCount:         stockPaymentsCount,
		OperationalProfitCents:     operationalProfitCents,
		NetProfitCents:             netProfitCents,
		NetMarginPercent:           netMarginPercent,
		HealthStatus:               healthStatus(netProfitCents, netMarginPercent),
		ExpensesByCategory:         categoryTotals,
		ReceiptCosts:               receiptCosts,
	}, nil
}

func (service *FinancialService) RealizedExpenses(ctx context.Context, origin RealizedExpenseOrigin, limit int, offset int, start time.Time, end time.Time) ([]RealizedExpense, int64, error) {
	origin = RealizedExpenseOrigin(strings.ToLower(strings.TrimSpace(string(origin))))
	if origin == "" {
		origin = RealizedExpenseAll
	}
	if origin != RealizedExpenseAll && origin != RealizedExpenseOperational && origin != RealizedExpenseStock {
		return nil, 0, apperrors.ErrInvalidInput
	}

	rows := make([]RealizedExpense, 0)
	if origin != RealizedExpenseStock {
		operational, err := service.loadOperationalExpenses(ctx, start, end)
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, operational...)
	}
	if origin != RealizedExpenseOperational {
		stock, err := service.loadStockExpenses(ctx, start, end)
		if err != nil {
			return nil, 0, err
		}
		rows = append(rows, stock...)
	}

	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].OccurredAt.Equal(rows[j].OccurredAt) {
			return rows[i].ID > rows[j].ID
		}
		return rows[i].OccurredAt.After(rows[j].OccurredAt)
	})
	total := int64(len(rows))
	if offset >= len(rows) {
		return []RealizedExpense{}, total, nil
	}
	last := offset + limit
	if last > len(rows) {
		last = len(rows)
	}
	return rows[offset:last], total, nil
}

func (service *FinancialService) loadOperationalExpenses(ctx context.Context, start, end time.Time) ([]RealizedExpense, error) {
	type row struct {
		ID, ExpenseID, Description, Category string
		AmountCents                          int64
		OccurredAt                           time.Time
		PaymentMethod                        entities.PaymentMethod
		InstallmentNumber                    int
		InstallmentID                        string
	}
	var source []row
	err := service.db.WithContext(ctx).Table("payable_installments").
		Select("COALESCE(payable_installments.cash_entry_id, payable_installments.id) AS id, expenses.id AS expense_id, expenses.description, expenses.category, payable_installments.amount_cents, payable_installments.paid_at AS occurred_at, payable_installments.payment_method, payable_installments.number AS installment_number, payable_installments.id AS installment_id").
		Joins("JOIN expenses ON expenses.id = payable_installments.expense_id").
		Where("payable_installments.status = ?", entities.PayablePaid).
		Where("payable_installments.paid_at >= ? AND payable_installments.paid_at < ?", start, end).Scan(&source).Error
	if err != nil {
		return nil, err
	}
	rows := make([]RealizedExpense, 0, len(source))
	for _, item := range source {
		rows = append(rows, RealizedExpense{ID: item.ID, Origin: RealizedExpenseOperational, Description: item.Description, Category: item.Category, AmountCents: item.AmountCents, OccurredAt: item.OccurredAt, PaymentMethod: item.PaymentMethod, InstallmentNumber: item.InstallmentNumber, ExpenseID: item.ExpenseID, InstallmentID: item.InstallmentID, Editable: false})
	}
	return rows, nil
}

func (service *FinancialService) loadStockExpenses(ctx context.Context, start, end time.Time) ([]RealizedExpense, error) {
	type stockExpenseRow struct {
		ID                string
		Description       string
		AmountCents       int64
		OccurredAt        time.Time
		PaymentMethod     entities.PaymentMethod
		SupplierName      string
		InstallmentNumber int
		StockPurchaseID   string
		InstallmentID     string
	}
	var source []stockExpenseRow
	err := service.db.WithContext(ctx).Table("payable_installments").
		Select("COALESCE(payable_installments.cash_entry_id, payable_installments.id) AS id, 'Compra de estoque: ' || stock_purchases.supplier_name AS description, payable_installments.amount_cents, payable_installments.paid_at AS occurred_at, payable_installments.payment_method, stock_purchases.supplier_name, payable_installments.number AS installment_number, payable_installments.stock_purchase_id, payable_installments.id AS installment_id").
		Joins("JOIN stock_purchases ON stock_purchases.id = payable_installments.stock_purchase_id").
		Where("payable_installments.status = ?", entities.PayablePaid).
		Where("payable_installments.paid_at >= ? AND payable_installments.paid_at < ?", start, end).
		Scan(&source).Error
	if err != nil {
		return nil, err
	}
	rows := make([]RealizedExpense, 0, len(source))
	for _, row := range source {
		rows = append(rows, RealizedExpense{ID: row.ID, Origin: RealizedExpenseStock, Description: row.Description, Category: "Estoque", AmountCents: row.AmountCents, OccurredAt: row.OccurredAt, PaymentMethod: row.PaymentMethod, SupplierName: row.SupplierName, InstallmentNumber: row.InstallmentNumber, StockPurchaseID: row.StockPurchaseID, InstallmentID: row.InstallmentID, Editable: false})
	}
	return rows, nil
}

func (service *FinancialService) loadStockExpenseTotals(ctx context.Context, start, end time.Time) (int64, int64, error) {
	type totals struct {
		AmountCents int64
		Count       int64
	}
	var result totals
	err := service.db.WithContext(ctx).Table("payable_installments").
		Select("COALESCE(SUM(payable_installments.amount_cents), 0) AS amount_cents, COUNT(*) AS count").
		Where("payable_installments.stock_purchase_id IS NOT NULL").
		Where("payable_installments.status = ?", entities.PayablePaid).
		Where("payable_installments.paid_at >= ? AND payable_installments.paid_at < ?", start, end).
		Scan(&result).Error
	return result.AmountCents, result.Count, err
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
	err := service.db.WithContext(ctx).Table("payable_installments").
		Joins("JOIN expenses ON expenses.id = payable_installments.expense_id").
		Where("expenses.affects_profit = ?", true).
		Where("payable_installments.status = ?", entities.PayablePaid).
		Where("payable_installments.paid_at >= ? AND payable_installments.paid_at < ?", start, end).
		Select(`
			COUNT(*) AS expenses_count,
			COALESCE(SUM(payable_installments.amount_cents), 0) AS operational_expenses_cents
		`).
		Scan(&totals).
		Error
	return &totals, err
}

func (service *FinancialService) loadExpensesByCategory(ctx context.Context, start time.Time, end time.Time) ([]ExpenseCategorySummary, error) {
	var totals []ExpenseCategorySummary
	err := service.db.WithContext(ctx).Table("payable_installments").
		Joins("JOIN expenses ON expenses.id = payable_installments.expense_id").
		Where("expenses.affects_profit = ?", true).
		Where("payable_installments.status = ?", entities.PayablePaid).
		Where("payable_installments.paid_at >= ? AND payable_installments.paid_at < ?", start, end).
		Select("expenses.category, COALESCE(SUM(payable_installments.amount_cents), 0) AS amount_cents, COUNT(*) AS count").
		Group("expenses.category").
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
