package services

import (
	"context"
	"strings"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	receiptrepos "github.com/empi-autocenter/erp-empi/internal/domain/receipts/repositories"
	stockrepos "github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/gorm"
)

type ReceiptService struct {
	repo      *receiptrepos.ReceiptRepository
	stockRepo *stockrepos.StockRepository
	users     *userservices.UserService
}

type ReceiptItemInput struct {
	StockItemID string `json:"stockItemId"`
	Quantity    int    `json:"quantity"`
}

type ReceiptInput struct {
	Client       userservices.UpsertClientInput `json:"client"`
	VehicleModel string                         `json:"vehicleModel"`
	VehicleYear  int                            `json:"vehicleYear"`
	VehiclePlate string                         `json:"vehiclePlate"`
	Services     string                         `json:"services"`
	PriceCents   int64                          `json:"priceCents"`
	Notes        string                         `json:"notes"`
	Items        []ReceiptItemInput             `json:"items"`
}

func NewReceiptService(repo *receiptrepos.ReceiptRepository, stockRepo *stockrepos.StockRepository, users *userservices.UserService) *ReceiptService {
	return &ReceiptService{repo: repo, stockRepo: stockRepo, users: users}
}

func (service *ReceiptService) List(ctx context.Context, limit int, offset int, status string) ([]entities.Receipt, int64, error) {
	return service.repo.List(ctx, limit, offset, status)
}

func (service *ReceiptService) FindByID(ctx context.Context, id string) (*entities.Receipt, error) {
	return service.repo.FindByID(ctx, id)
}

func (service *ReceiptService) Create(ctx context.Context, input ReceiptInput) (*entities.Receipt, error) {
	if err := validateReceiptInput(input); err != nil {
		return nil, err
	}
	client, err := service.users.UpsertClient(ctx, input.Client)
	if err != nil {
		return nil, err
	}

	receipt := &entities.Receipt{
		UserID:       client.ID,
		VehicleModel: strings.TrimSpace(input.VehicleModel),
		VehicleYear:  input.VehicleYear,
		VehiclePlate: strings.ToUpper(strings.TrimSpace(input.VehiclePlate)),
		Services:     strings.TrimSpace(input.Services),
		PriceCents:   input.PriceCents,
		Status:       entities.ReceiptStatusPending,
		Notes:        strings.TrimSpace(input.Notes),
	}

	for _, itemInput := range input.Items {
		stockItem, err := service.stockRepo.FindByID(ctx, itemInput.StockItemID)
		if err != nil {
			return nil, err
		}
		if itemInput.Quantity <= 0 {
			return nil, apperrors.ErrInvalidInput
		}
		receipt.Items = append(receipt.Items, entities.ReceiptItem{
			StockItemID:     stockItem.ID,
			Quantity:        itemInput.Quantity,
			UnitCostCents:   stockItem.CostCents,
			UnitResaleCents: stockItem.ResalePriceCents,
			MarkupPercent:   stockItem.MarkupPercent,
		})
	}

	return receipt, service.repo.Create(ctx, receipt)
}

func (service *ReceiptService) MarkPaid(ctx context.Context, id string) (*entities.Receipt, error) {
	var updated *entities.Receipt
	err := service.repo.Transaction(func(tx *gorm.DB) error {
		receipt, err := service.repo.FindByIDForUpdate(tx.WithContext(ctx), id)
		if err != nil {
			return err
		}
		if receipt.Status == entities.ReceiptStatusPaid {
			updated = receipt
			return nil
		}
		if receipt.Status == entities.ReceiptStatusCancelled {
			return apperrors.ErrConflict
		}
		for _, item := range receipt.Items {
			stockItem, err := service.stockRepo.FindByIDForUpdate(tx.WithContext(ctx), item.StockItemID)
			if err != nil {
				return err
			}
			if stockItem.Quantity < item.Quantity {
				return apperrors.ErrInsufficientStock
			}
			stockItem.Quantity -= item.Quantity
			stockItem.UsedQuantity += item.Quantity
			if err := tx.Save(stockItem).Error; err != nil {
				return err
			}
		}
		now := time.Now()
		receipt.Status = entities.ReceiptStatusPaid
		receipt.PaidAt = &now
		updated = receipt
		return service.repo.UpdateWithTx(tx.WithContext(ctx), receipt)
	})
	if err != nil {
		return nil, err
	}
	return service.repo.FindByID(ctx, updated.ID)
}

func validateReceiptInput(input ReceiptInput) error {
	if strings.TrimSpace(input.VehicleModel) == "" ||
		input.VehicleYear < 1950 ||
		strings.TrimSpace(input.VehiclePlate) == "" ||
		strings.TrimSpace(input.Services) == "" ||
		input.PriceCents <= 0 {
		return apperrors.ErrInvalidInput
	}
	return nil
}
