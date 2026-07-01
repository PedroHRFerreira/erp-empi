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

type ReceiptExpenseInput struct {
	Description string `json:"description"`
	Category    string `json:"category"`
	AmountCents int64  `json:"amountCents"`
	SpentAt     string `json:"spentAt"`
	Notes       string `json:"notes"`
}

type ReceiptInput struct {
	Client          userservices.UpsertClientInput `json:"client"`
	VehicleModel    string                         `json:"vehicleModel"`
	VehicleYear     int                            `json:"vehicleYear"`
	VehiclePlate    string                         `json:"vehiclePlate"`
	Services        string                         `json:"services"`
	LaborPriceCents int64                          `json:"laborPriceCents"`
	DiscountCents   int64                          `json:"discountCents"`
	PriceCents      int64                          `json:"priceCents"`
	CardFeePercent  *float64                       `json:"cardFeePercent"`
	PaymentMethod   entities.PaymentMethod         `json:"paymentMethod"`
	Installments    int                            `json:"installments"`
	Notes           string                         `json:"notes"`
	Items           []ReceiptItemInput             `json:"items"`
	ServiceExpenses []ReceiptExpenseInput          `json:"serviceExpenses"`
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
	serviceExpenses, err := buildReceiptExpenses(input.ServiceExpenses)
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
	serviceExpensesTotalCents := int64(0)
	for _, expense := range serviceExpenses {
		serviceExpensesTotalCents += expense.AmountCents
	}
	laborPriceCents := input.LaborPriceCents
	if input.PaymentMethod == "" && input.LaborPriceCents == 0 && input.PriceCents > 0 {
		laborPriceCents = input.PriceCents - productsTotalCents - serviceExpensesTotalCents
		if laborPriceCents < 0 {
			laborPriceCents = 0
		}
	}
	discountCents := input.DiscountCents
	subtotalCents := laborPriceCents + productsTotalCents + serviceExpensesTotalCents - discountCents
	paymentMethod := normalizePaymentMethod(input.PaymentMethod)
	installments := normalizeInstallments(paymentMethod, input.Installments)
	cardFeePercent := selectCardFeePercent(paymentMethod, installments, admin, input.CardFeePercent)
	cardFeeCents := calculatePercentCents(subtotalCents, cardFeePercent)
	totalCents := subtotalCents + cardFeeCents

	receipt := &entities.Receipt{
		UserID:             client.ID,
		VehicleModel:       strings.TrimSpace(input.VehicleModel),
		VehicleYear:        input.VehicleYear,
		VehiclePlate:       strings.ToUpper(strings.TrimSpace(input.VehiclePlate)),
		Services:           strings.TrimSpace(input.Services),
		LaborPriceCents:    laborPriceCents,
		DiscountCents:      discountCents,
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

	err = service.repo.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(receipt).Error; err != nil {
			return err
		}

		if len(serviceExpenses) == 0 {
			return nil
		}

		receiptID := receipt.ID
		for index := range serviceExpenses {
			serviceExpenses[index].ReceiptID = &receiptID
		}
		return tx.WithContext(ctx).Create(&serviceExpenses).Error
	})
	return receipt, err
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

func (service *ReceiptService) Cancel(ctx context.Context, id string) (*entities.Receipt, error) {
	var updated *entities.Receipt
	err := service.repo.Transaction(func(tx *gorm.DB) error {
		receipt, err := service.repo.FindByIDForUpdate(tx.WithContext(ctx), id)
		if err != nil {
			return err
		}
		if receipt.Status != entities.ReceiptStatusPending {
			return apperrors.ErrConflict
		}
		receipt.Status = entities.ReceiptStatusCancelled
		receipt.PaidAt = nil
		updated = receipt
		return service.repo.UpdateWithTx(tx.WithContext(ctx), receipt)
	})
	if err != nil {
		return nil, err
	}
	return service.repo.FindByID(ctx, updated.ID)
}

func (service *ReceiptService) Reopen(ctx context.Context, id string) (*entities.Receipt, error) {
	var updated *entities.Receipt
	err := service.repo.Transaction(func(tx *gorm.DB) error {
		receipt, err := service.repo.FindByIDForUpdate(tx.WithContext(ctx), id)
		if err != nil {
			return err
		}
		if receipt.Status != entities.ReceiptStatusCancelled {
			return apperrors.ErrConflict
		}
		if err := service.ensureReceiptItemsCanBeReserved(tx.WithContext(ctx), receipt.Items); err != nil {
			return err
		}
		receipt.Status = entities.ReceiptStatusPending
		receipt.PaidAt = nil
		updated = receipt
		return service.repo.UpdateWithTx(tx.WithContext(ctx), receipt)
	})
	if err != nil {
		return nil, err
	}
	return service.repo.FindByID(ctx, updated.ID)
}

func (service *ReceiptService) ensureReceiptItemsCanBeReserved(tx *gorm.DB, items []entities.ReceiptItem) error {
	requestedQuantities := make(map[string]int)
	stockItemIDs := make([]string, 0, len(items))

	for _, item := range items {
		if _, exists := requestedQuantities[item.StockItemID]; !exists {
			stockItemIDs = append(stockItemIDs, item.StockItemID)
		}
		requestedQuantities[item.StockItemID] += item.Quantity
	}

	reservedQuantities, err := service.repo.ReservedQuantitiesByStockItemIDsWithTx(tx, stockItemIDs)
	if err != nil {
		return err
	}

	for stockItemID, quantity := range requestedQuantities {
		stockItem, err := service.stockRepo.FindByIDForUpdate(tx, stockItemID)
		if err != nil {
			return err
		}
		if stockItem.Quantity < quantity {
			return apperrors.ErrInsufficientStock
		}
		if stockItem.Quantity-reservedQuantities[stockItemID] < quantity {
			return apperrors.ErrReservedStock
		}
	}

	return nil
}

func validateReceiptInput(input ReceiptInput) error {
	if strings.TrimSpace(input.VehicleModel) == "" ||
		input.VehicleYear < 1950 ||
		strings.TrimSpace(input.VehiclePlate) == "" ||
		strings.TrimSpace(input.Services) == "" ||
		input.LaborPriceCents < 0 ||
		input.DiscountCents < 0 ||
		input.DiscountCents > input.LaborPriceCents ||
		input.PriceCents < 0 ||
		(input.CardFeePercent != nil && *input.CardFeePercent < 0) ||
		!isValidPaymentMethod(input.PaymentMethod) ||
		(input.PaymentMethod == entities.PaymentMethodCreditCard && (input.Installments < 0 || input.Installments > 12)) {
		return apperrors.ErrInvalidInput
	}
	return nil
}

func buildReceiptExpenses(inputs []ReceiptExpenseInput) ([]entities.Expense, error) {
	expenses := make([]entities.Expense, 0, len(inputs))

	for _, input := range inputs {
		description := strings.TrimSpace(input.Description)
		category := strings.TrimSpace(input.Category)
		notes := strings.TrimSpace(input.Notes)
		spentAt, err := parseReceiptExpenseDate(input.SpentAt)
		if err != nil || description == "" || category == "" || input.AmountCents <= 0 {
			return nil, apperrors.ErrInvalidInput
		}

		expenses = append(expenses, entities.Expense{
			Description: description,
			Category:    category,
			AmountCents: input.AmountCents,
			SpentAt:     spentAt,
			Notes:       notes,
		})
	}

	return expenses, nil
}

func parseReceiptExpenseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, apperrors.ErrInvalidInput
	}
	if parsed, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.Time{}, apperrors.ErrInvalidInput
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

func selectCardFeePercent(method entities.PaymentMethod, installments int, admin *entities.User, customPercent *float64) float64 {
	switch method {
	case entities.PaymentMethodDebitCard:
		if customPercent != nil {
			return *customPercent
		}
		return admin.MachineFeePercent
	case entities.PaymentMethodCreditCard:
		if customPercent != nil {
			return *customPercent
		}
		if installments > 1 {
			return admin.InstallmentFeePercent
		}
		return admin.MachineFeePercent
	default:
		return 0
	}
}
