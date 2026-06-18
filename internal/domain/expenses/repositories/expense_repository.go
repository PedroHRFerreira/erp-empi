package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/gorm"
)

type ExpenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) *ExpenseRepository {
	return &ExpenseRepository{db: db}
}

func (repo *ExpenseRepository) List(ctx context.Context, limit int, offset int, start time.Time, end time.Time) ([]entities.Expense, int64, error) {
	var expenses []entities.Expense
	var total int64
	query := repo.db.WithContext(ctx).
		Model(&entities.Expense{}).
		Where("archived_at IS NULL").
		Where("spent_at >= ? AND spent_at < ?", start, end)

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Order("spent_at desc, created_at desc").
		Limit(limit).
		Offset(offset).
		Find(&expenses).
		Error
	return expenses, total, err
}

func (repo *ExpenseRepository) FindByID(ctx context.Context, id string) (*entities.Expense, error) {
	expense := new(entities.Expense)
	err := repo.db.WithContext(ctx).
		Where("id = ? AND archived_at IS NULL", id).
		First(expense).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return expense, err
}

func (repo *ExpenseRepository) Create(ctx context.Context, expense *entities.Expense) error {
	return repo.db.WithContext(ctx).Create(expense).Error
}

func (repo *ExpenseRepository) Update(ctx context.Context, expense *entities.Expense) error {
	return repo.db.WithContext(ctx).Save(expense).Error
}
