package services

import (
	"context"
	"errors"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"gorm.io/gorm"
)

type MetricsService struct {
	db *gorm.DB
}

type Summary struct {
	ClientsTotal             int64           `json:"clientsTotal"`
	ReceiptsTotal            int64           `json:"receiptsTotal"`
	ReceiptsPaid             int64           `json:"receiptsPaid"`
	ReceiptsPending          int64           `json:"receiptsPending"`
	ReceiptsCancelled        int64           `json:"receiptsCancelled"`
	RevenuePaidCents         int64           `json:"revenuePaidCents"`
	RevenuePendingCents      int64           `json:"revenuePendingCents"`
	AverageTicketPaidCents   int64           `json:"averageTicketPaidCents"`
	StockItemsTotal          int64           `json:"stockItemsTotal"`
	StockUnitsAvailableTotal int64           `json:"stockUnitsAvailableTotal"`
	StockUnitsUsedTotal      int64           `json:"stockUnitsUsedTotal"`
	LastReceipt              *ReceiptMetric  `json:"lastReceipt"`
	LastStockItem            *StockMetric    `json:"lastStockItem"`
	TopProducts              []TopProduct    `json:"topProducts"`
	LowStockProducts         []StockMetric   `json:"lowStockProducts"`
	RecentClients            []ClientMetric  `json:"recentClients"`
	PendingReceipts          []ReceiptMetric `json:"pendingReceipts"`
	PaidReceipts             []ReceiptMetric `json:"paidReceipts"`
}

type TopProduct struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	UsedQuantity int    `json:"usedQuantity"`
}

type StockMetric struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	Quantity     int    `json:"quantity"`
	UsedQuantity int    `json:"usedQuantity"`
	CreatedAt    string `json:"createdAt"`
}

type ClientMetric struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	CPF           string `json:"cpf"`
	ReceiptsCount int64  `json:"receiptsCount"`
	LastReceiptAt string `json:"lastReceiptAt"`
}

type ReceiptMetric struct {
	ID         string `json:"id"`
	ClientName string `json:"clientName"`
	PriceCents int64  `json:"priceCents"`
	Status     string `json:"status"`
	CreatedAt  string `json:"createdAt"`
}

func NewMetricsService(db *gorm.DB) *MetricsService {
	return &MetricsService{db: db}
}

func (service *MetricsService) Summary(ctx context.Context) (*Summary, error) {
	summary := &Summary{
		TopProducts:      []TopProduct{},
		LowStockProducts: []StockMetric{},
		RecentClients:    []ClientMetric{},
		PendingReceipts:  []ReceiptMetric{},
		PaidReceipts:     []ReceiptMetric{},
	}
	if err := service.db.WithContext(ctx).Model(&entities.User{}).Where("type = ?", entities.UserTypeClient).Count(&summary.ClientsTotal).Error; err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Model(&entities.Receipt{}).Count(&summary.ReceiptsTotal).Error; err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Model(&entities.Receipt{}).Where("status = ?", entities.ReceiptStatusPaid).Count(&summary.ReceiptsPaid).Error; err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Model(&entities.Receipt{}).Where("status = ?", entities.ReceiptStatusPending).Count(&summary.ReceiptsPending).Error; err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Model(&entities.Receipt{}).Where("status = ?", entities.ReceiptStatusCancelled).Count(&summary.ReceiptsCancelled).Error; err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Model(&entities.Receipt{}).Where("status = ?", entities.ReceiptStatusPaid).Select("COALESCE(SUM(price_cents), 0)").Scan(&summary.RevenuePaidCents).Error; err != nil {
		return nil, err
	}
	if err := service.db.WithContext(ctx).Model(&entities.Receipt{}).Where("status = ?", entities.ReceiptStatusPending).Select("COALESCE(SUM(price_cents), 0)").Scan(&summary.RevenuePendingCents).Error; err != nil {
		return nil, err
	}
	if summary.ReceiptsPaid > 0 {
		summary.AverageTicketPaidCents = summary.RevenuePaidCents / summary.ReceiptsPaid
	}
	if err := service.loadStockTotals(ctx, summary); err != nil {
		return nil, err
	}
	if err := service.loadLatestReceipt(ctx, summary); err != nil {
		return nil, err
	}
	if err := service.loadLatestStockItem(ctx, summary); err != nil {
		return nil, err
	}
	if err := service.loadTopProducts(ctx, summary); err != nil {
		return nil, err
	}
	if err := service.loadLowStockProducts(ctx, summary); err != nil {
		return nil, err
	}
	if err := service.loadClientMetrics(ctx, summary); err != nil {
		return nil, err
	}
	if err := service.loadReceiptMetrics(ctx, summary, entities.ReceiptStatusPending, &summary.PendingReceipts); err != nil {
		return nil, err
	}
	if err := service.loadReceiptMetrics(ctx, summary, entities.ReceiptStatusPaid, &summary.PaidReceipts); err != nil {
		return nil, err
	}
	return summary, nil
}

func (service *MetricsService) loadStockTotals(ctx context.Context, summary *Summary) error {
	type totals struct {
		ItemsTotal          int64
		UnitsAvailableTotal int64
		UnitsUsedTotal      int64
	}
	var row totals
	if err := service.db.WithContext(ctx).
		Model(&entities.StockItem{}).
		Where("active = ?", true).
		Select("COUNT(*) as items_total, COALESCE(SUM(quantity), 0) as units_available_total, COALESCE(SUM(used_quantity), 0) as units_used_total").
		Scan(&row).Error; err != nil {
		return err
	}
	summary.StockItemsTotal = row.ItemsTotal
	summary.StockUnitsAvailableTotal = row.UnitsAvailableTotal
	summary.StockUnitsUsedTotal = row.UnitsUsedTotal
	return nil
}

func (service *MetricsService) loadLatestReceipt(ctx context.Context, summary *Summary) error {
	var receipt entities.Receipt
	err := service.db.WithContext(ctx).Preload("User").Order("created_at desc").First(&receipt).Error
	if err == nil {
		summary.LastReceipt = &ReceiptMetric{
			ID:         receipt.ID,
			ClientName: receipt.User.Name,
			PriceCents: receipt.PriceCents,
			Status:     string(receipt.Status),
			CreatedAt:  receipt.CreatedAt.Format(time.RFC3339),
		}
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (service *MetricsService) loadLatestStockItem(ctx context.Context, summary *Summary) error {
	var product entities.StockItem
	err := service.db.WithContext(ctx).Where("active = ?", true).Order("created_at desc").First(&product).Error
	if err == nil {
		summary.LastStockItem = &StockMetric{
			ID:           product.ID,
			Name:         product.Name,
			Quantity:     product.Quantity,
			UsedQuantity: product.UsedQuantity,
			CreatedAt:    product.CreatedAt.Format(time.RFC3339),
		}
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil
	}
	return err
}

func (service *MetricsService) loadTopProducts(ctx context.Context, summary *Summary) error {
	var products []entities.StockItem
	if err := service.db.WithContext(ctx).Where("active = ?", true).Order("used_quantity desc").Limit(5).Find(&products).Error; err != nil {
		return err
	}
	for _, product := range products {
		summary.TopProducts = append(summary.TopProducts, TopProduct{ID: product.ID, Name: product.Name, UsedQuantity: product.UsedQuantity})
	}
	return nil
}

func (service *MetricsService) loadLowStockProducts(ctx context.Context, summary *Summary) error {
	var products []entities.StockItem
	if err := service.db.WithContext(ctx).Where("active = ? AND quantity <= ?", true, 3).Order("quantity asc, name asc").Limit(5).Find(&products).Error; err != nil {
		return err
	}
	for _, product := range products {
		summary.LowStockProducts = append(summary.LowStockProducts, StockMetric{
			ID:           product.ID,
			Name:         product.Name,
			Quantity:     product.Quantity,
			UsedQuantity: product.UsedQuantity,
			CreatedAt:    product.CreatedAt.Format(time.RFC3339),
		})
	}
	return nil
}

func (service *MetricsService) loadClientMetrics(ctx context.Context, summary *Summary) error {
	type row struct {
		ID            string
		Name          string
		CPF           string
		ReceiptsCount int64
		LastReceiptAt *time.Time
	}
	var rows []row
	err := service.db.WithContext(ctx).
		Table("users").
		Select("users.id, users.name, users.cpf, COUNT(receipts.id) as receipts_count, MAX(receipts.created_at) as last_receipt_at").
		Joins("LEFT JOIN receipts ON receipts.user_id = users.id").
		Where("users.type = ?", entities.UserTypeClient).
		Group("users.id, users.name, users.cpf").
		Order("last_receipt_at desc").
		Limit(5).
		Scan(&rows).Error
	if err != nil {
		return err
	}
	for _, row := range rows {
		lastReceiptAt := ""
		if row.LastReceiptAt != nil {
			lastReceiptAt = row.LastReceiptAt.Format(time.RFC3339)
		}
		summary.RecentClients = append(summary.RecentClients, ClientMetric{
			ID:            row.ID,
			Name:          row.Name,
			CPF:           row.CPF,
			ReceiptsCount: row.ReceiptsCount,
			LastReceiptAt: lastReceiptAt,
		})
	}
	return nil
}

func (service *MetricsService) loadReceiptMetrics(ctx context.Context, _ *Summary, status entities.ReceiptStatus, target *[]ReceiptMetric) error {
	var receipts []entities.Receipt
	if err := service.db.WithContext(ctx).Preload("User").Where("status = ?", status).Order("created_at desc").Limit(5).Find(&receipts).Error; err != nil {
		return err
	}
	for _, receipt := range receipts {
		*target = append(*target, ReceiptMetric{
			ID:         receipt.ID,
			ClientName: receipt.User.Name,
			PriceCents: receipt.PriceCents,
			Status:     string(receipt.Status),
			CreatedAt:  receipt.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		})
	}
	return nil
}
