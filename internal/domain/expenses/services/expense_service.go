package services

import (
	"context"
	"strings"
	"time"

	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/expenses/repositories"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/gorm"
)

const dateLayout = "2006-01-02"

type ExpenseService struct {
	repo *repositories.ExpenseRepository
	cash *cashservices.CashService
}

type ExpenseInput struct {
	Description   string                    `json:"description"`
	Category      string                    `json:"category"`
	AmountCents   int64                     `json:"amountCents"`
	SpentAt       string                    `json:"spentAt"`
	Notes         string                    `json:"notes"`
	ReceiptID     *string                   `json:"receiptId"`
	PaymentMethod entities.PaymentMethod    `json:"paymentMethod"`
	Installments  []ExpenseInstallmentInput `json:"installments"`
}

type ExpenseInstallmentInput struct {
	AmountCents   int64                  `json:"amountCents"`
	DueDate       string                 `json:"dueDate"`
	PlannedMethod entities.PayableMethod `json:"plannedMethod"`
}

func NewExpenseService(repo *repositories.ExpenseRepository, cash ...*cashservices.CashService) *ExpenseService {
	service := &ExpenseService{repo: repo}
	if len(cash) > 0 {
		service.cash = cash[0]
	}
	return service
}

func (service *ExpenseService) List(ctx context.Context, limit int, offset int, start time.Time, end time.Time) ([]entities.Expense, int64, error) {
	return service.repo.List(ctx, limit, offset, start, end)
}

func (service *ExpenseService) Get(ctx context.Context, id string) (*entities.Expense, error) {
	return service.repo.FindByID(ctx, strings.TrimSpace(id))
}

func (service *ExpenseService) Create(ctx context.Context, input ExpenseInput) (*entities.Expense, error) {
	expense, err := service.buildExpense(ctx, input, nil)
	if err != nil {
		return nil, err
	}
	installments, err := buildInstallments(input, expense)
	if err != nil {
		return nil, err
	}
	expense.Installments = installments
	if err := service.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error { return tx.Create(expense).Error }); err != nil {
		return nil, err
	}
	return service.repo.FindByID(ctx, expense.ID)
}

func (service *ExpenseService) Update(ctx context.Context, id string, input ExpenseInput) (*entities.Expense, error) {
	current, err := service.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	expense, err := service.buildExpense(ctx, input, current)
	if err != nil {
		return nil, err
	}
	installments, err := buildInstallments(input, expense)
	if err != nil {
		return nil, err
	}
	err = service.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var paid int64
		if err := tx.Model(&entities.PayableInstallment{}).Where("expense_id = ? AND status = ?", id, entities.PayablePaid).Count(&paid).Error; err != nil {
			return err
		}
		if paid > 0 {
			return apperrors.ErrConflict
		}
		if err := tx.Where("expense_id = ?", id).Delete(&entities.PayableInstallment{}).Error; err != nil {
			return err
		}
		expense.Installments = installments
		return tx.Session(&gorm.Session{FullSaveAssociations: true}).Save(expense).Error
	})
	if err != nil {
		return nil, err
	}
	return service.repo.FindByID(ctx, expense.ID)
}

func (service *ExpenseService) Archive(ctx context.Context, id string) error {
	expense, err := service.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	now := time.Now()
	expense.ArchivedAt = &now
	return service.repo.DB().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var paid int64
		if err := tx.Model(&entities.PayableInstallment{}).Where("expense_id = ? AND status = ?", id, entities.PayablePaid).Count(&paid).Error; err != nil {
			return err
		}
		if paid > 0 {
			return apperrors.ErrConflict
		}
		if err := tx.Where("expense_id = ?", id).Delete(&entities.PayableInstallment{}).Error; err != nil {
			return err
		}
		expense.Installments = nil
		return tx.Save(expense).Error
	})
}

func buildInstallments(input ExpenseInput, expense *entities.Expense) ([]entities.PayableInstallment, error) {
	rows := input.Installments
	if len(rows) == 0 {
		rows = []ExpenseInstallmentInput{{AmountCents: expense.AmountCents, DueDate: expense.SpentAt.Format(dateLayout), PlannedMethod: entities.PayableMethodBoleto}}
	}
	result := make([]entities.PayableInstallment, 0, len(rows))
	var total int64
	for index, row := range rows {
		dueDate, err := ParseExpenseDate(row.DueDate)
		if err != nil || row.AmountCents <= 0 || !validPlannedMethod(row.PlannedMethod) {
			return nil, apperrors.ErrInvalidInput
		}
		total += row.AmountCents
		result = append(result, entities.PayableInstallment{ExpenseID: &expense.ID, Number: index + 1, AmountCents: row.AmountCents, DueDate: dueDate, Status: entities.PayablePending, PlannedMethod: row.PlannedMethod})
	}
	if total != expense.AmountCents {
		return nil, apperrors.ErrInvalidInput
	}
	return result, nil
}

func validPlannedMethod(method entities.PayableMethod) bool {
	switch method {
	case entities.PayableMethodBoleto, entities.PayableMethodCash, entities.PayableMethodPix, entities.PayableMethodDebitCard, entities.PayableMethodCreditCard:
		return true
	default:
		return false
	}
}

func (service *ExpenseService) buildExpense(ctx context.Context, input ExpenseInput, current *entities.Expense) (*entities.Expense, error) {
	description := strings.TrimSpace(input.Description)
	category := strings.TrimSpace(input.Category)
	notes := strings.TrimSpace(input.Notes)
	receiptID, err := service.normalizeReceiptID(ctx, input.ReceiptID)
	if err != nil {
		return nil, err
	}
	spentAt, err := ParseExpenseDate(input.SpentAt)
	if err != nil || description == "" || category == "" || input.AmountCents <= 0 {
		return nil, apperrors.ErrInvalidInput
	}

	expense := current
	if expense == nil {
		expense = &entities.Expense{AffectsProfit: true}
	}
	expense.Description = description
	expense.Category = category
	expense.AmountCents = input.AmountCents
	expense.SpentAt = spentAt
	expense.Notes = notes
	expense.ReceiptID = receiptID
	expense.Receipt = nil
	return expense, nil
}

func (service *ExpenseService) normalizeReceiptID(ctx context.Context, receiptID *string) (*string, error) {
	if receiptID == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*receiptID)
	if trimmed == "" {
		return nil, nil
	}
	exists, err := service.repo.ReceiptExists(ctx, trimmed)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, apperrors.ErrInvalidInput
	}
	return &trimmed, nil
}

func ParseExpenseDate(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, apperrors.ErrInvalidInput
	}
	if parsed, err := time.ParseInLocation(dateLayout, value, time.Local); err == nil {
		return parsed, nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.Time{}, apperrors.ErrInvalidInput
}
