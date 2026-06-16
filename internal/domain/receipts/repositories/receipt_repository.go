package repositories

import (
	"context"
	"errors"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ReceiptRepository struct {
	db *gorm.DB
}

func NewReceiptRepository(db *gorm.DB) *ReceiptRepository {
	return &ReceiptRepository{db: db}
}

func (repo *ReceiptRepository) Transaction(fn func(tx *gorm.DB) error) error {
	return repo.db.Transaction(fn)
}

func (repo *ReceiptRepository) Create(ctx context.Context, receipt *entities.Receipt) error {
	return repo.db.WithContext(ctx).Create(receipt).Error
}

func (repo *ReceiptRepository) UpdateWithTx(tx *gorm.DB, receipt *entities.Receipt) error {
	return tx.Save(receipt).Error
}

func (repo *ReceiptRepository) List(ctx context.Context, limit int, offset int, status string) ([]entities.Receipt, int64, error) {
	var receipts []entities.Receipt
	var total int64
	query := repo.db.WithContext(ctx).Model(&entities.Receipt{}).Preload("User").Preload("Items.StockItem")
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&receipts).Error
	return receipts, total, err
}

func (repo *ReceiptRepository) FindByID(ctx context.Context, id string) (*entities.Receipt, error) {
	receipt := new(entities.Receipt)
	err := repo.db.WithContext(ctx).Preload("User").Preload("Items.StockItem").First(receipt, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return receipt, err
}

func (repo *ReceiptRepository) FindByIDForUpdate(tx *gorm.DB, id string) (*entities.Receipt, error) {
	receipt := new(entities.Receipt)
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").First(receipt, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return receipt, err
}
