package repositories

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/empi-autocenter/erp-empi/internal/shared/validation"
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
	return tx.Omit("Payments").Save(receipt).Error
}

// ReceiptFilters contains validated criteria. EndBefore is an exclusive upper bound.
type ReceiptFilters struct {
	Query         string
	Status        string
	PaymentMethod string
	Type          string
	StartAt       *time.Time
	EndBefore     *time.Time
}

func (repo *ReceiptRepository) List(ctx context.Context, limit int, offset int, status string) ([]entities.Receipt, int64, error) {
	return repo.ListFiltered(ctx, limit, offset, ReceiptFilters{Status: status})
}

func (repo *ReceiptRepository) ListFiltered(ctx context.Context, limit int, offset int, filters ReceiptFilters) ([]entities.Receipt, int64, error) {
	var receipts []entities.Receipt
	var total int64
	query := repo.db.WithContext(ctx).Model(&entities.Receipt{})
	if filters.Query != "" {
		// Escape LIKE metacharacters so names containing % or _ remain literal searches.
		term := "%" + strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(strings.ToLower(filters.Query)) + "%"
		phoneTerm := term
		if strings.Trim(filters.Query, "0123456789 ()+.-\t\r\n") == "" {
			if digits := validation.OnlyDigits(filters.Query); digits != "" {
				phoneTerm = "%" + digits + "%"
			}
		}
		query = query.Where(`(LOWER(receipts.vehicle_plate) LIKE ? ESCAPE '!' OR LOWER(receipts.services) LIKE ? ESCAPE '!' OR EXISTS (SELECT 1 FROM users WHERE users.id = receipts.user_id AND (LOWER(users.name) LIKE ? ESCAPE '!' OR users.phone LIKE ? ESCAPE '!')))`, term, term, term, phoneTerm)
	}
	if filters.Status != "" {
		query = query.Where("receipts.status = ?", filters.Status)
	}
	if filters.StartAt != nil {
		query = query.Where("receipts.created_at >= ?", *filters.StartAt)
	}
	if filters.EndBefore != nil {
		query = query.Where("receipts.created_at < ?", *filters.EndBefore)
	}
	if filters.Type == "quick" {
		query = query.Where("receipts.user_id IS NULL")
	}
	if filters.Type == "normal" {
		query = query.Where("receipts.user_id IS NOT NULL")
	}
	if filters.PaymentMethod == "mixed" {
		query = query.Where(`(receipts.payment_method = ? OR (SELECT COUNT(*) FROM receipt_payments WHERE receipt_payments.receipt_id = receipts.id) > 1)`, entities.PaymentMethodMixed)
	} else if filters.PaymentMethod != "" {
		// Historical receipts have no allocations. Modern receipts match any allocation.
		query = query.Where(`(EXISTS (SELECT 1 FROM receipt_payments WHERE receipt_payments.receipt_id = receipts.id AND receipt_payments.payment_method = ?) OR (NOT EXISTS (SELECT 1 FROM receipt_payments WHERE receipt_payments.receipt_id = receipts.id) AND receipts.payment_method = ?))`, filters.PaymentMethod, filters.PaymentMethod)
	}
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Preload("User").Preload("Items.StockItem").Preload("Expenses").Preload("Payments", func(db *gorm.DB) *gorm.DB { return db.Order("position asc") }).Order("receipts.created_at desc, receipts.id desc").Limit(limit).Offset(offset).Find(&receipts).Error
	for i := range receipts {
		receipts[i].Payments = receipts[i].EffectivePayments()
	}
	return receipts, total, err
}

func (repo *ReceiptRepository) ListByUserID(ctx context.Context, userID string) ([]entities.Receipt, error) {
	var receipts []entities.Receipt
	err := repo.db.WithContext(ctx).
		Preload("User").
		Preload("Items.StockItem").
		Preload("Expenses").
		Preload("Payments", func(db *gorm.DB) *gorm.DB { return db.Order("position asc") }).
		Where("user_id = ?", userID).
		Order("created_at desc").
		Find(&receipts).
		Error
	for i := range receipts {
		receipts[i].Payments = receipts[i].EffectivePayments()
	}
	return receipts, err
}

func (repo *ReceiptRepository) ReservedQuantitiesByStockItemIDs(ctx context.Context, stockItemIDs []string) (map[string]int, error) {
	return reservedQuantitiesByStockItemIDs(ctx, repo.db, stockItemIDs, "")
}

func (repo *ReceiptRepository) ReservedQuantitiesByStockItemIDsExcludingReceipt(ctx context.Context, stockItemIDs []string, receiptID string) (map[string]int, error) {
	return reservedQuantitiesByStockItemIDs(ctx, repo.db, stockItemIDs, receiptID)
}

func (repo *ReceiptRepository) ReservedQuantitiesByStockItemIDsWithTx(tx *gorm.DB, stockItemIDs []string) (map[string]int, error) {
	return reservedQuantitiesByStockItemIDs(tx.Statement.Context, tx, stockItemIDs, "")
}

func (repo *ReceiptRepository) ReservedQuantitiesByStockItemIDsExcludingReceiptWithTx(tx *gorm.DB, stockItemIDs []string, receiptID string) (map[string]int, error) {
	return reservedQuantitiesByStockItemIDs(tx.Statement.Context, tx, stockItemIDs, receiptID)
}

func reservedQuantitiesByStockItemIDs(ctx context.Context, db *gorm.DB, stockItemIDs []string, excludeReceiptID string) (map[string]int, error) {
	type row struct {
		StockItemID string
		Quantity    int
	}

	reserved := make(map[string]int, len(stockItemIDs))
	if len(stockItemIDs) == 0 {
		return reserved, nil
	}

	var rows []row
	query := db.WithContext(ctx).
		Table("receipt_items").
		Select("receipt_items.stock_item_id, COALESCE(SUM(receipt_items.quantity), 0) AS quantity").
		Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").
		Where("receipts.status = ? AND receipt_items.stock_item_id IN ?", entities.ReceiptStatusPending, stockItemIDs)
	if excludeReceiptID != "" {
		query = query.Where("receipts.id <> ?", excludeReceiptID)
	}
	err := query.Group("receipt_items.stock_item_id").Scan(&rows).Error
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
	err := repo.db.WithContext(ctx).Preload("User").Preload("Items.StockItem").Preload("Expenses").Preload("Payments", func(db *gorm.DB) *gorm.DB { return db.Order("position asc") }).First(receipt, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	receipt.Payments = receipt.EffectivePayments()
	return receipt, err
}

func (repo *ReceiptRepository) FindByIDForUpdate(tx *gorm.DB, id string) (*entities.Receipt, error) {
	receipt := new(entities.Receipt)
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").Preload("Payments", func(db *gorm.DB) *gorm.DB { return db.Order("position asc") }).First(receipt, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	receipt.Payments = receipt.EffectivePayments()
	return receipt, err
}
