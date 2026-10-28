package repositories

import (
	"context"
	"errors"
	"strings"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockRepository struct {
	db *gorm.DB
}

func NewStockRepository(db *gorm.DB) *StockRepository {
	return &StockRepository{db: db}
}

func (repo *StockRepository) DB() *gorm.DB {
	return repo.db
}

func (repo *StockRepository) Create(ctx context.Context, item *entities.StockItem) error {
	return repo.db.WithContext(ctx).Create(item).Error
}

func (repo *StockRepository) Update(ctx context.Context, item *entities.StockItem) error {
	return repo.db.WithContext(ctx).Save(item).Error
}

func (repo *StockRepository) List(ctx context.Context, limit int, offset int) ([]entities.StockItem, int64, error) {
	return repo.Search(ctx, limit, offset, "", "")
}

func (repo *StockRepository) Search(ctx context.Context, limit, offset int, term, supplier string) ([]entities.StockItem, int64, error) {
	var items []entities.StockItem
	var total int64
	query := repo.db.WithContext(ctx).Model(&entities.StockItem{}).Where("active = ?", true)
	if term = strings.TrimSpace(term); term != "" {
		query = query.Where("LOWER(name) LIKE ?", "%"+strings.ToLower(term)+"%")
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if supplier = strings.TrimSpace(supplier); supplier != "" {
		query = query.Clauses(clause.OrderBy{Expression: clause.Expr{SQL: "CASE WHEN EXISTS (SELECT 1 FROM stock_purchase_items spi JOIN stock_purchases sp ON sp.id = spi.stock_purchase_id WHERE spi.stock_item_id = stock_items.id AND LOWER(TRIM(sp.supplier_name)) = LOWER(?)) THEN 0 ELSE 1 END, stock_items.created_at DESC, stock_items.id DESC", Vars: []any{supplier}, WithoutParentheses: true}})
	} else {
		query = query.Order("created_at desc, id desc")
	}
	err := query.Limit(limit).Offset(offset).Find(&items).Error
	return items, total, err
}

func (repo *StockRepository) FindByID(ctx context.Context, id string) (*entities.StockItem, error) {
	item := new(entities.StockItem)
	err := repo.db.WithContext(ctx).First(item, "id = ? AND active = ?", id, true).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return item, err
}

func (repo *StockRepository) FindByIDForUpdate(tx *gorm.DB, id string) (*entities.StockItem, error) {
	item := new(entities.StockItem)
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(item, "id = ? AND active = ?", id, true).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return item, err
}
