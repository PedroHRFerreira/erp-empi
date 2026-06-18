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
	Client          userservices.UpsertClientInput `json:"client"`
	VehicleModel    string                         `json:"vehicleModel"`
	VehicleYear     int                            `json:"vehicleYear"`
	VehiclePlate    string                         `json:"vehiclePlate"`
	Services        string                         `json:"services"`
	LaborPriceCents int64                          `json:"laborPriceCents"`
	PriceCents      int64                          `json:"priceCents"`
	PaymentMethod   entities.PaymentMethod         `json:"paymentMethod"`
	Installments    int                            `json:"installments"`
	Notes           string                         `json:"notes"`
	Items           []ReceiptItemInput             `json:"items"`
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

func (service *ReceiptService) ListByUserID(ctx context.Context, userID string) ([]entities.Receipt, error) {
	return service.repo.ListByUserID(ctx, userID)
}

func (service *ReceiptService) Create(ctx context.Context, adminID string, input ReceiptInput) (*entities.Receipt, error) {
	if err := validateReceiptInput(input); err != nil {
		return nil, err
	}
	admin, err := service.users.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}
	client, err := service.users.UpsertClient(ctx, input.Client)
	if err != nil {
		return nil, err
	}

	requestedQuantities := make(map[string]int)
	stockItems := make(map[string]*entities.StockItem)
	stockItemIDs := make([]string, 0, len(input.Items))

	for _, itemInput := range input.Items {
		if itemInput.StockItemID == "" || itemInput.Quantity <= 0 {
			return nil, apperrors.ErrInvalidInput
		}
		if _, exists := requestedQuantities[itemInput.StockItemID]; !exists {
			stockItemIDs = append(stockItemIDs, itemInput.StockItemID)
		}
		requestedQuantities[itemInput.StockItemID] += itemInput.Quantity
	}

	reservedQuantities, err := service.repo.ReservedQuantitiesByStockItemIDs(ctx, stockItemIDs)
	if err != nil {
		return nil, err
	}

	for stockItemID, quantity := range requestedQuantities {
		stockItem, err := service.stockRepo.FindByID(ctx, stockItemID)
		if err != nil {
			return nil, err
		}
		if stockItem.Quantity < quantity {
			return nil, apperrors.ErrInsufficientStock
		}
		if stockItem.Quantity-reservedQuantities[stockItemID] < quantity {
			return nil, apperrors.ErrReservedStock
		}
		stockItems[stockItemID] = stockItem
	}

	productsTotalCents := int64(0)
	for _, itemInput := range input.Items {
		stockItem := stockItems[itemInput.StockItemID]
		productsTotalCents += stockItem.ResalePriceCents * int64(itemInput.Quantity)
	}
	laborPriceCents := input.LaborPriceCents
	if input.PaymentMethod == "" && input.LaborPriceCents == 0 && input.PriceCents > 0 {
		laborPriceCents = input.PriceCents - productsTotalCents
		if laborPriceCents < 0 {
			laborPriceCents = 0
		}
	}
	subtotalCents := laborPriceCents + productsTotalCents
	paymentMethod := normalizePaymentMethod(input.PaymentMethod)
	installments := normalizeInstallments(paymentMethod, input.Installments)
	cardFeePercent := selectCardFeePercent(paymentMethod, installments, admin)
	cardFeeCents := calculatePercentCents(subtotalCents, cardFeePercent)
	totalCents := subtotalCents + cardFeeCents

	receipt := &entities.Receipt{
		UserID:             client.ID,
		VehicleModel:       strings.TrimSpace(input.VehicleModel),
		VehicleYear:        input.VehicleYear,
		VehiclePlate:       strings.ToUpper(strings.TrimSpace(input.VehiclePlate)),
		Services:           strings.TrimSpace(input.Services),
		LaborPriceCents:    laborPriceCents,
		ProductsTotalCents: productsTotalCents,
		SubtotalCents:      subtotalCents,
		CardFeePercent:     cardFeePercent,
		CardFeeCents:       cardFeeCents,
		PaymentMethod:      paymentMethod,
		Installments:       installments,
		PriceCents:         totalCents,
		Status:             entities.ReceiptStatusPending,
		Notes:              strings.TrimSpace(input.Notes),
	}

	for _, itemInput := range input.Items {
		stockItem := stockItems[itemInput.StockItemID]
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
		input.LaborPriceCents < 0 ||
		input.PriceCents < 0 ||
		!isValidPaymentMethod(input.PaymentMethod) ||
		(input.PaymentMethod == entities.PaymentMethodCreditCard && (input.Installments < 0 || input.Installments > 12)) {
		return apperrors.ErrInvalidInput
	}
	if len(input.Items) == 0 {
		return apperrors.ErrInvalidInput
	}
	return nil
}

func normalizePaymentMethod(method entities.PaymentMethod) entities.PaymentMethod {
	if method == "" {
		return entities.PaymentMethodCash
	}
	return method
}

func normalizeInstallments(method entities.PaymentMethod, installments int) int {
	if method != entities.PaymentMethodCreditCard {
		return 1
	}
	if installments == 0 {
		return 1
	}
	return installments
}

func isValidPaymentMethod(method entities.PaymentMethod) bool {
	switch normalizePaymentMethod(method) {
	case entities.PaymentMethodCreditCard,
		entities.PaymentMethodDebitCard,
		entities.PaymentMethodPix,
		entities.PaymentMethodCash:
		return true
	default:
		return false
	}
}

func calculatePercentCents(value int64, percent float64) int64 {
	if percent <= 0 {
		return 0
	}
	return int64(float64(value) * (percent / 100))
}

func selectCardFeePercent(method entities.PaymentMethod, installments int, admin *entities.User) float64 {
	switch method {
	case entities.PaymentMethodDebitCard:
		return admin.MachineFeePercent
	case entities.PaymentMethodCreditCard:
		if installments > 1 {
			return admin.InstallmentFeePercent
		}
		return admin.MachineFeePercent
	default:
		return 0
	}
}
