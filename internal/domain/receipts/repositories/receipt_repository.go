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

func (repo *ReceiptRepository) ListByUserID(ctx context.Context, userID string) ([]entities.Receipt, error) {
	var receipts []entities.Receipt
	err := repo.db.WithContext(ctx).
		Preload("User").
		Preload("Items.StockItem").
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&receipts).
		Error
	return receipts, err
}

func (repo *ReceiptRepository) ReservedQuantitiesByStockItemIDs(ctx context.Context, stockItemIDs []string) (map[string]int, error) {
	type row struct {
		StockItemID string
		Quantity    int
	}

	reserved := make(map[string]int, len(stockItemIDs))
	if len(stockItemIDs) == 0 {
		return reserved, nil
	}

	var rows []row
	err := repo.db.WithContext(ctx).
		Table("receipt_items").
		Select("receipt_items.stock_item_id, COALESCE(SUM(receipt_items.quantity), 0) AS quantity").
		Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipts.status = ? AND receipt_items.stock_item_id IN ?", entities.ReceiptStatusPending, stockItemIDs).
		Group("receipt_items.stock_item_id").
		Scan(&rows).
		Error
	if err != nil {
		return nil, err
	}
	for _, item := range rows {
		reserved[item.StockItemID] = item.Quantity
	}
	return reserved, nil
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
