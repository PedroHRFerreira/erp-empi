package services

import (
	"context"
	"strings"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/expenses/repositories"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
)

const dateLayout = "2006-01-02"

type ExpenseService struct {
	repo *repositories.ExpenseRepository
}

type ExpenseInput struct {
	Description string  `json:"description"`
	Category    string  `json:"category"`
	AmountCents int64   `json:"amountCents"`
	SpentAt     string  `json:"spentAt"`
	Notes       string  `json:"notes"`
	ReceiptID   *string `json:"receiptId"`
}

func NewExpenseService(repo *repositories.ExpenseRepository) *ExpenseService {
	return &ExpenseService{repo: repo}
}

func (service *ExpenseService) List(ctx context.Context, limit int, offset int, start time.Time, end time.Time) ([]entities.Expense, int64, error) {
	return service.repo.List(ctx, limit, offset, start, end)
}

func (service *ExpenseService) Create(ctx context.Context, input ExpenseInput) (*entities.Expense, error) {
	expense, err := service.buildExpense(ctx, input, nil)
	if err != nil {
		return nil, err
	}
	if err := service.repo.Create(ctx, expense); err != nil {
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
	if err := service.repo.Update(ctx, expense); err != nil {
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
	return service.repo.Update(ctx, expense)
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
		expense = &entities.Expense{}
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
