package services

import (
	"context"
	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
	"math"
	"strings"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/empi-autocenter/erp-empi/internal/shared/validation"
	"gorm.io/gorm"
)

type StockService struct {
	repo *repositories.StockRepository
}

type StockEntryInput struct {
	SupplierName  string                                  `json:"supplierName"`
	Quantity      int                                     `json:"quantity"`
	UnitCostCents int64                                   `json:"unitCostCents"`
	Installments  []cashservices.PurchaseInstallmentInput `json:"installments"`
}

type StockInput struct {
	ExistingStockItemID string           `json:"existingStockItemId"`
	Entry               *StockEntryInput `json:"entry"`
	Name                string           `json:"name"`
	Description         string           `json:"description"`
	CostCents           int64            `json:"costCents"`
	MarkupPercent       float64          `json:"markupPercent"`
	Quantity            int              `json:"quantity"`
}

func NewStockService(repo *repositories.StockRepository) *StockService {
	return &StockService{repo: repo}
}

func (service *StockService) List(ctx context.Context, limit int, offset int) ([]entities.StockItem, int64, error) {
	return service.repo.List(ctx, limit, offset)
}

func (service *StockService) Search(ctx context.Context, limit, offset int, query, supplier string) ([]entities.StockItem, int64, error) {
	return service.repo.Search(ctx, limit, offset, query, supplier)
}

func (service *StockService) FindByID(ctx context.Context, id string) (*entities.StockItem, error) {
	return service.repo.FindByID(ctx, id)
}

func (service *StockService) Create(ctx context.Context, input StockInput) (*entities.StockItem, error) {
	var item *entities.StockItem
	err := service.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if input.ExistingStockItemID != "" {
			if input.Entry == nil {
				return apperrors.ErrInvalidInput
			}
			var err error
			item, err = service.repo.FindByIDForUpdate(tx, input.ExistingStockItemID)
			if err != nil {
				return err
			}
		} else {
			var err error
			item, err = service.buildItem(input, nil)
			if err != nil {
				return err
			}
			// The entry is the sole source of initial quantity when a purchase is supplied.
			if input.Entry != nil {
				item.Quantity = 0
			}
			if err = tx.Create(item).Error; err != nil {
				return err
			}
		}
		return service.addEntry(tx, item, input.Entry)
	})
	if err != nil {
		return nil, err
	}
	return service.FindByID(ctx, item.ID)
}

func (service *StockService) Update(ctx context.Context, id string, input StockInput) (*entities.StockItem, error) {
	err := service.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		current, err := service.repo.FindByIDForUpdate(tx, id)
		if err != nil {
			return err
		}
		// A stale editor must never replace the actual available balance.
		input.Quantity = current.Quantity
		if input.Entry != nil {
			input.CostCents = current.CostCents
		}
		item, err := service.buildItem(input, current)
		if err != nil {
			return err
		}
		if err = tx.Save(item).Error; err != nil {
			return err
		}
		return service.addEntry(tx, item, input.Entry)
	})
	if err != nil {
		return nil, err
	}
	return service.FindByID(ctx, id)
}

func (service *StockService) addEntry(tx *gorm.DB, item *entities.StockItem, entry *StockEntryInput) error {
	if entry == nil {
		return nil
	}
	_, err := cashservices.NewCashService(tx).CreatePurchaseWithTx(tx, cashservices.PurchaseInput{
		SupplierName: entry.SupplierName,
		Items:        []cashservices.PurchaseItemInput{{StockItemID: item.ID, Quantity: entry.Quantity, UnitCostCents: entry.UnitCostCents}},
		Installments: entry.Installments,
	})
	return err
}

func (service *StockService) Delete(ctx context.Context, id string) error {
	return service.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		item, err := service.repo.FindByIDForUpdate(tx, id)
		if err != nil {
			return err
		}

		var purchaseIDs []string
		if err := tx.Table("stock_purchase_items").
			Select("DISTINCT stock_purchase_items.stock_purchase_id").
			Joins("JOIN stock_purchases ON stock_purchases.id = stock_purchase_items.stock_purchase_id").
			Where("stock_purchase_items.stock_item_id = ? AND stock_purchases.status = ?", id, entities.StockPurchaseConfirmed).
			Scan(&purchaseIDs).Error; err != nil {
			return err
		}

		if len(purchaseIDs) > 0 {
			var sharedPurchases int64
			sharedPurchaseQuery := tx.Table("stock_purchase_items").
				Select("stock_purchase_id").
				Where("stock_purchase_id IN ?", purchaseIDs).
				Group("stock_purchase_id").
				Having("COUNT(*) > 1")
			if err := tx.Table("(?) AS shared_purchases", sharedPurchaseQuery).Count(&sharedPurchases).Error; err != nil {
				return err
			}
			if sharedPurchases > 0 {
				return apperrors.ErrConflict
			}

			var paidInstallments int64
			if err := tx.Model(&entities.PayableInstallment{}).
				Where("stock_purchase_id IN ? AND status = ?", purchaseIDs, entities.PayablePaid).
				Count(&paidInstallments).Error; err != nil {
				return err
			}
			if paidInstallments > 0 {
				return apperrors.ErrConflict
			}

			now := time.Now()
			if err := tx.Model(&entities.PayableInstallment{}).
				Where("stock_purchase_id IN ? AND status = ?", purchaseIDs, entities.PayablePending).
				Updates(map[string]any{"status": entities.PayableCancelled, "updated_at": now}).Error; err != nil {
				return err
			}
			if err := tx.Model(&entities.StockPurchase{}).Where("id IN ?", purchaseIDs).
				Updates(map[string]any{"status": entities.StockPurchaseCancelled, "cancelled_at": now, "updated_at": now}).Error; err != nil {
				return err
			}
		}

		item.Active = false
		return tx.Save(item).Error
	})
}

func (service *StockService) buildItem(input StockInput, current *entities.StockItem) (*entities.StockItem, error) {
	if strings.TrimSpace(input.Name) == "" || input.CostCents <= 0 || input.MarkupPercent < 0 || math.IsNaN(input.MarkupPercent) || math.IsInf(input.MarkupPercent, 0) || input.Quantity < 0 {
		return nil, apperrors.ErrInvalidInput
	}
	item := current
	if item == nil {
		item = &entities.StockItem{Active: true}
	}
	item.Name = strings.TrimSpace(input.Name)
	item.Description = strings.TrimSpace(input.Description)
	item.CostCents = input.CostCents
	item.MarkupPercent = input.MarkupPercent
	item.ResalePriceCents = validation.CalculateMarkupCents(input.CostCents, input.MarkupPercent)
	item.Quantity = input.Quantity
	return item, nil
}
