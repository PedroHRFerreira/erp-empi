package services

import (
	"context"
	"sort"
	"strings"

	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
)

type SupplierSummary struct {
	SupplierName  string `json:"supplierName"`
	ProductCount  int    `json:"productCount"`
	PurchaseCount int    `json:"purchaseCount"`
	TotalCents    int64  `json:"totalCents"`
	PaidCents     int64  `json:"paidCents"`
	PendingCents  int64  `json:"pendingCents"`
}

type SupplierDetail struct {
	SupplierSummary
	Products  []entities.StockItem     `json:"products"`
	Purchases []entities.StockPurchase `json:"purchases"`
}

// Suppliers are views of the commercial history, not a second product catalog.
func (service *StockService) supplierDetails(ctx context.Context) ([]SupplierDetail, error) {
	purchases, err := cashservices.NewCashService(service.repo.DB()).ListPurchases(ctx)
	if err != nil {
		return nil, err
	}
	groups := make(map[string]*SupplierDetail)
	products := make(map[string]map[string]entities.StockItem)
	for _, purchase := range purchases {
		name := strings.TrimSpace(purchase.SupplierName)
		key := strings.ToLower(name)
		detail := groups[key]
		if detail == nil {
			detail = &SupplierDetail{SupplierSummary: SupplierSummary{SupplierName: name}, Products: []entities.StockItem{}, Purchases: []entities.StockPurchase{}}
			groups[key] = detail
			products[key] = make(map[string]entities.StockItem)
		}
		detail.PurchaseCount++
		detail.Purchases = append(detail.Purchases, purchase)
		for _, item := range purchase.Items {
			if item.StockItem.ID != "" {
				products[key][item.StockItemID] = item.StockItem
			}
		}
		if purchase.Status != entities.StockPurchaseConfirmed {
			continue
		}
		detail.TotalCents += purchase.TotalCents
		for _, installment := range purchase.Installments {
			switch installment.Status {
			case entities.PayablePaid:
				detail.PaidCents += installment.AmountCents
			case entities.PayablePending:
				detail.PendingCents += installment.AmountCents
			}
		}
	}
	details := make([]SupplierDetail, 0, len(groups))
	for key, detail := range groups {
		for _, product := range products[key] {
			detail.Products = append(detail.Products, product)
		}
		sort.Slice(detail.Products, func(i, j int) bool { return detail.Products[i].Name < detail.Products[j].Name })
		detail.ProductCount = len(detail.Products)
		details = append(details, *detail)
	}
	sort.Slice(details, func(i, j int) bool {
		return strings.ToLower(details[i].SupplierName) < strings.ToLower(details[j].SupplierName)
	})
	return details, nil
}

func (service *StockService) ListSuppliers(ctx context.Context) ([]SupplierSummary, error) {
	details, err := service.supplierDetails(ctx)
	if err != nil {
		return nil, err
	}
	summaries := make([]SupplierSummary, 0, len(details))
	for _, detail := range details {
		summaries = append(summaries, detail.SupplierSummary)
	}
	return summaries, nil
}

func (service *StockService) GetSupplier(ctx context.Context, name string) (*SupplierDetail, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, apperrors.ErrInvalidInput
	}
	details, err := service.supplierDetails(ctx)
	if err != nil {
		return nil, err
	}
	for _, detail := range details {
		if strings.EqualFold(detail.SupplierName, name) {
			return &detail, nil
		}
	}
	return nil, apperrors.ErrNotFound
}
