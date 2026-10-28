package services

import (
	"context"
	"math"
	"strings"
	"time"
	_ "time/tzdata"

	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
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
	cash      *cashservices.CashService
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

type ReceiptPaymentInput struct {
	PaymentMethod   entities.PaymentMethod `json:"paymentMethod"`
	BaseAmountCents int64                  `json:"baseAmountCents"`
	CardFeePercent  *float64               `json:"cardFeePercent"`
	Installments    int                    `json:"installments"`
}

type ReceiptInput struct {
	Payments        []ReceiptPaymentInput          `json:"payments"`
	Quick           bool                           `json:"quick"`
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

type preparedReceipt struct {
	UserID          *string
	Receipt         entities.Receipt
	Items           []entities.ReceiptItem
	ServiceExpenses []entities.Expense
}

func NewReceiptService(repo *receiptrepos.ReceiptRepository, stockRepo *stockrepos.StockRepository, users *userservices.UserService, cash ...*cashservices.CashService) *ReceiptService {
	service := &ReceiptService{repo: repo, stockRepo: stockRepo, users: users}
	if len(cash) > 0 {
		service.cash = cash[0]
	}
	return service
}

type ReceiptFilters struct {
	Query         string
	StartDate     string
	EndDate       string
	Status        string
	PaymentMethod string
	Type          string
}

func (service *ReceiptService) List(ctx context.Context, limit int, offset int, status string) ([]entities.Receipt, int64, error) {
	return service.ListFiltered(ctx, limit, offset, ReceiptFilters{Status: status})
}

func (service *ReceiptService) ListFiltered(ctx context.Context, limit int, offset int, filters ReceiptFilters) ([]entities.Receipt, int64, error) {
	criteria, err := prepareReceiptFilters(filters)
	if err != nil {
		return nil, 0, err
	}
	return service.repo.ListFiltered(ctx, limit, offset, criteria)
}

func prepareReceiptFilters(filters ReceiptFilters) (receiptrepos.ReceiptFilters, error) {
	criteria := receiptrepos.ReceiptFilters{Query: strings.TrimSpace(filters.Query), Status: filters.Status, PaymentMethod: filters.PaymentMethod, Type: filters.Type}
	switch filters.Status {
	case "", string(entities.ReceiptStatusPending), string(entities.ReceiptStatusPaid), string(entities.ReceiptStatusCancelled):
	default:
		return criteria, apperrors.ErrInvalidInput
	}
	switch filters.PaymentMethod {
	case "", string(entities.PaymentMethodCash), string(entities.PaymentMethodPix), string(entities.PaymentMethodCreditCard), string(entities.PaymentMethodDebitCard), string(entities.PaymentMethodMixed), string(entities.PaymentMethodLegacy):
	default:
		return criteria, apperrors.ErrInvalidInput
	}
	switch filters.Type {
	case "", "normal", "quick":
	default:
		return criteria, apperrors.ErrInvalidInput
	}
	location, err := time.LoadLocation("America/Sao_Paulo")
	if err != nil {
		return criteria, err
	}
	if filters.StartDate != "" {
		date, err := time.ParseInLocation("2006-01-02", filters.StartDate, location)
		if err != nil {
			return criteria, apperrors.ErrInvalidInput
		}
		date = date.UTC()
		criteria.StartAt = &date
	}
	if filters.EndDate != "" {
		date, err := time.ParseInLocation("2006-01-02", filters.EndDate, location)
		if err != nil {
			return criteria, apperrors.ErrInvalidInput
		}
		date = date.AddDate(0, 0, 1).UTC()
		criteria.EndBefore = &date
	}
	if criteria.StartAt != nil && criteria.EndBefore != nil && !criteria.StartAt.Before(*criteria.EndBefore) {
		return criteria, apperrors.ErrInvalidInput
	}
	return criteria, nil
}

func (service *ReceiptService) FindByID(ctx context.Context, id string) (*entities.Receipt, error) {
	return service.repo.FindByID(ctx, id)
}

func (service *ReceiptService) ListByUserID(ctx context.Context, userID string) ([]entities.Receipt, error) {
	return service.repo.ListByUserID(ctx, userID)
}

func (service *ReceiptService) Create(ctx context.Context, adminID string, input ReceiptInput) (*entities.Receipt, error) {
	prepared, err := service.prepareReceipt(ctx, adminID, input, nil, "")
	if err != nil {
		return nil, err
	}

	receipt := prepared.Receipt
	receipt.UserID = prepared.UserID
	receipt.Status = entities.ReceiptStatusPending
	receipt.Items = prepared.Items
	err = service.repo.Transaction(func(tx *gorm.DB) error {
		if err := tx.WithContext(ctx).Create(&receipt).Error; err != nil {
			return err
		}

		if len(prepared.ServiceExpenses) == 0 {
			return nil
		}

		receiptID := receipt.ID
		for index := range prepared.ServiceExpenses {
			prepared.ServiceExpenses[index].ReceiptID = &receiptID
		}
		return tx.WithContext(ctx).Create(&prepared.ServiceExpenses).Error
	})
	if err != nil {
		return nil, err
	}
	return service.repo.FindByID(ctx, receipt.ID)
}

func (service *ReceiptService) Update(ctx context.Context, adminID string, id string, input ReceiptInput) (*entities.Receipt, error) {
	current, err := service.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if current.Status != entities.ReceiptStatusPending {
		return nil, apperrors.ErrConflict
	}

	prepared, err := service.prepareReceipt(ctx, adminID, input, nil, id)
	if err != nil {
		return nil, err
	}

	var updated *entities.Receipt
	err = service.repo.Transaction(func(tx *gorm.DB) error {
		receipt, err := service.repo.FindByIDForUpdate(tx.WithContext(ctx), id)
		if err != nil {
			return err
		}
		if receipt.Status != entities.ReceiptStatusPending {
			return apperrors.ErrConflict
		}

		receipt.UserID = prepared.UserID
		receipt.VehicleModel = prepared.Receipt.VehicleModel
		receipt.VehicleYear = prepared.Receipt.VehicleYear
		receipt.VehiclePlate = prepared.Receipt.VehiclePlate
		receipt.Services = prepared.Receipt.Services
		receipt.LaborPriceCents = prepared.Receipt.LaborPriceCents
		receipt.DiscountCents = prepared.Receipt.DiscountCents
		receipt.ProductsTotalCents = prepared.Receipt.ProductsTotalCents
		receipt.SubtotalCents = prepared.Receipt.SubtotalCents
		receipt.CardFeePercent = prepared.Receipt.CardFeePercent
		receipt.CardFeeCents = prepared.Receipt.CardFeeCents
		receipt.PaymentMethod = prepared.Receipt.PaymentMethod
		receipt.Installments = prepared.Receipt.Installments
		receipt.PriceCents = prepared.Receipt.PriceCents
		receipt.Notes = prepared.Receipt.Notes
		receipt.PaidAt = nil
		receipt.Items = nil
		receipt.Payments = nil
		if err := tx.Where("receipt_id = ?", id).Delete(&entities.ReceiptPayment{}).Error; err != nil {
			return err
		}

		if err := tx.WithContext(ctx).Where("receipt_id = ?", id).Delete(&entities.ReceiptItem{}).Error; err != nil {
			return err
		}
		if err := tx.WithContext(ctx).Where("receipt_id = ?", id).Delete(&entities.Expense{}).Error; err != nil {
			return err
		}
		if err := service.repo.UpdateWithTx(tx.WithContext(ctx), receipt); err != nil {
			return err
		}

		for index := range prepared.Receipt.Payments {
			prepared.Receipt.Payments[index].ReceiptID = id
		}
		if err := tx.Create(&prepared.Receipt.Payments).Error; err != nil {
			return err
		}
		for index := range prepared.Items {
			prepared.Items[index].ReceiptID = id
		}
		if len(prepared.Items) > 0 {
			if err := tx.WithContext(ctx).Create(&prepared.Items).Error; err != nil {
				return err
			}
		}

		for index := range prepared.ServiceExpenses {
			prepared.ServiceExpenses[index].ReceiptID = &id
		}
		if len(prepared.ServiceExpenses) > 0 {
			if err := tx.WithContext(ctx).Create(&prepared.ServiceExpenses).Error; err != nil {
				return err
			}
		}

		updated = receipt
		return nil
	})
	if err != nil {
		return nil, err
	}
	return service.repo.FindByID(ctx, updated.ID)
}

func (service *ReceiptService) prepareReceipt(ctx context.Context, adminID string, input ReceiptInput, tx *gorm.DB, excludeReceiptID string) (*preparedReceipt, error) {
	if err := validateReceiptInput(input); err != nil {
		return nil, err
	}
	serviceExpenses, err := buildReceiptExpenses(input.ServiceExpenses)
	if err != nil {
		return nil, err
	}
	serviceExpensesTotalCents := serviceExpensesTotal(serviceExpenses)
	if len(input.Items) == 0 && input.DiscountCents > input.LaborPriceCents+serviceExpensesTotalCents {
		return nil, apperrors.ErrInvalidInput
	}

	admin, err := service.users.FindByID(ctx, adminID)
	if err != nil {
		return nil, err
	}

	var userID *string
	if !input.Quick {
		client, err := service.users.UpsertClient(ctx, input.Client)
		if err != nil {
			return nil, err
		}
		clientID := client.ID
		userID = &clientID
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

	var reservedQuantities map[string]int
	if tx != nil && excludeReceiptID != "" {
		reservedQuantities, err = service.repo.ReservedQuantitiesByStockItemIDsExcludingReceiptWithTx(tx, stockItemIDs, excludeReceiptID)
	} else if excludeReceiptID != "" {
		reservedQuantities, err = service.repo.ReservedQuantitiesByStockItemIDsExcludingReceipt(ctx, stockItemIDs, excludeReceiptID)
	} else {
		reservedQuantities, err = service.repo.ReservedQuantitiesByStockItemIDs(ctx, stockItemIDs)
	}
	if err != nil {
		return nil, err
	}

	for stockItemID, quantity := range requestedQuantities {
		var stockItem *entities.StockItem
		if tx != nil {
			stockItem, err = service.stockRepo.FindByIDForUpdate(tx, stockItemID)
		} else {
			stockItem, err = service.stockRepo.FindByID(ctx, stockItemID)
		}
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
	items := make([]entities.ReceiptItem, 0, len(input.Items))
	for _, itemInput := range input.Items {
		stockItem := stockItems[itemInput.StockItemID]
		productsTotalCents += stockItem.ResalePriceCents * int64(itemInput.Quantity)
		items = append(items, entities.ReceiptItem{
			StockItemID:     stockItem.ID,
			Quantity:        itemInput.Quantity,
			UnitCostCents:   stockItem.CostCents,
			UnitResaleCents: stockItem.ResalePriceCents,
			MarkupPercent:   stockItem.MarkupPercent,
		})
	}

	laborPriceCents := input.LaborPriceCents
	if input.Payments == nil && input.PaymentMethod == "" && input.LaborPriceCents == 0 && input.PriceCents > 0 {
		laborPriceCents = input.PriceCents - productsTotalCents - serviceExpensesTotalCents
		if laborPriceCents < 0 {
			laborPriceCents = 0
		}
	}
	discountCents := input.DiscountCents
	grossSubtotalCents := laborPriceCents + productsTotalCents + serviceExpensesTotalCents
	if discountCents > grossSubtotalCents {
		return nil, apperrors.ErrInvalidInput
	}
	subtotalCents := grossSubtotalCents - discountCents
	payments, err := preparePayments(input, subtotalCents, admin)
	if err != nil {
		return nil, err
	}
	paymentMethod := payments[0].PaymentMethod
	installments := payments[0].Installments
	cardFeePercent := payments[0].CardFeePercent
	var cardFeeCents, totalCents int64
	for _, payment := range payments {
		if payment.AmountCents > math.MaxInt64-totalCents {
			return nil, apperrors.ErrInvalidInput
		}
		cardFeeCents += payment.CardFeeCents
		totalCents += payment.AmountCents
	}
	if len(payments) == 2 {
		paymentMethod = entities.PaymentMethodMixed
		installments = 1
		cardFeePercent = 0
	}

	return &preparedReceipt{
		UserID: userID,
		Receipt: entities.Receipt{
			Payments:           payments,
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
			Notes:              strings.TrimSpace(input.Notes),
		},
		Items:           items,
		ServiceExpenses: serviceExpenses,
	}, nil
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
		if service.cash != nil {
			if err := service.cash.RecordReceiptPayment(tx.WithContext(ctx), receipt); err != nil {
				return err
			}
		}
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
	if strings.TrimSpace(input.Services) == "" ||
		input.LaborPriceCents < 0 ||
		input.DiscountCents < 0 ||
		input.PriceCents < 0 ||
		(input.CardFeePercent != nil && !validFeePercent(*input.CardFeePercent)) ||
		(input.Payments == nil && !isValidPaymentMethod(input.PaymentMethod)) ||
		(input.PaymentMethod == entities.PaymentMethodCreditCard && (input.Installments < 0 || input.Installments > 12)) {
		return apperrors.ErrInvalidInput
	}
	if input.Payments != nil {
		if len(input.Payments) < 1 || len(input.Payments) > 2 {
			return apperrors.ErrInvalidInput
		}
		seen := make(map[entities.PaymentMethod]bool)
		for _, payment := range input.Payments {
			if payment.PaymentMethod == "" ||
				!isValidPaymentMethod(payment.PaymentMethod) ||
				seen[payment.PaymentMethod] ||
				payment.BaseAmountCents < 0 ||
				(len(input.Payments) == 2 && payment.BaseAmountCents == 0) ||
				(payment.CardFeePercent != nil && !validFeePercent(*payment.CardFeePercent)) ||
				(payment.PaymentMethod == entities.PaymentMethodCreditCard && (payment.Installments < 1 || payment.Installments > 12)) {
				return apperrors.ErrInvalidInput
			}
			seen[payment.PaymentMethod] = true
		}
	}
	if !input.Quick && (strings.TrimSpace(input.VehicleModel) == "" ||
		input.VehicleYear < 1950 ||
		strings.TrimSpace(input.VehiclePlate) == "") {
		return apperrors.ErrInvalidInput
	}
	return nil
}

func serviceExpensesTotal(expenses []entities.Expense) int64 {
	total := int64(0)
	for _, expense := range expenses {
		total += expense.AmountCents
	}
	return total
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

func validFeePercent(percent float64) bool {
	return !math.IsNaN(percent) && !math.IsInf(percent, 0) && percent >= 0
}

func preparePayments(input ReceiptInput, subtotalCents int64, admin *entities.User) ([]entities.ReceiptPayment, error) {
	inputs := input.Payments
	if inputs == nil {
		method := normalizePaymentMethod(input.PaymentMethod)
		inputs = []ReceiptPaymentInput{{PaymentMethod: method, BaseAmountCents: subtotalCents, CardFeePercent: input.CardFeePercent, Installments: normalizeInstallments(method, input.Installments)}}
	}
	payments := make([]entities.ReceiptPayment, 0, len(inputs))
	var baseTotal int64
	for index, input := range inputs {
		if input.BaseAmountCents > math.MaxInt64-baseTotal {
			return nil, apperrors.ErrInvalidInput
		}
		baseTotal += input.BaseAmountCents
		installments := normalizeInstallments(input.PaymentMethod, input.Installments)
		percent := selectCardFeePercent(input.PaymentMethod, installments, admin, input.CardFeePercent)
		if !validFeePercent(percent) || float64(input.BaseAmountCents)*(percent/100) >= float64(math.MaxInt64-input.BaseAmountCents) {
			return nil, apperrors.ErrInvalidInput
		}
		fee := calculatePercentCents(input.BaseAmountCents, percent)
		payments = append(payments, entities.ReceiptPayment{
			Position: index, PaymentMethod: input.PaymentMethod,
			BaseAmountCents: input.BaseAmountCents, CardFeePercent: percent,
			CardFeeCents: fee, AmountCents: input.BaseAmountCents + fee,
			Installments: installments,
		})
	}
	if baseTotal != subtotalCents {
		return nil, apperrors.ErrInvalidInput
	}
	return payments, nil
}
