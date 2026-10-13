package repositories

import (
	"context"
	"errors"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type MarketingRepository struct{ db *gorm.DB }

func NewMarketingRepository(db *gorm.DB) *MarketingRepository { return &MarketingRepository{db: db} }
func (r *MarketingRepository) Brand(ctx context.Context) (*entities.MarketingBrand, error) {
	b := &entities.MarketingBrand{ID: 1, PrimaryColor: "#2563eb", SecondaryColor: "#ffffff"}
	err := r.db.WithContext(ctx).First(b, 1).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return b, nil
	}
	return b, err
}
func (r *MarketingRepository) SaveBrand(ctx context.Context, b *entities.MarketingBrand) error {
	b.ID = 1
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(b).Error
}
func (r *MarketingRepository) List(ctx context.Context, limit, offset int) ([]entities.MarketingCreation, int64, error) {
	items := []entities.MarketingCreation{}
	var total int64
	q := r.db.WithContext(ctx).Model(&entities.MarketingCreation{})
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := q.Order("updated_at desc, id desc").Limit(limit).Offset(offset).Find(&items).Error
	return items, total, err
}
func (r *MarketingRepository) Find(ctx context.Context, id string) (*entities.MarketingCreation, error) {
	item := new(entities.MarketingCreation)
	err := r.db.WithContext(ctx).First(item, "id = ?", id).Error
	return item, mapError(err)
}
func (r *MarketingRepository) Create(ctx context.Context, item *entities.MarketingCreation) error {
	return r.db.WithContext(ctx).Create(item).Error
}
func (r *MarketingRepository) Update(ctx context.Context, item *entities.MarketingCreation) error {
	// An update must not recreate a creation deleted in another browser tab.
	result := r.db.WithContext(ctx).Model(item).Select("*").Updates(item)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
func (r *MarketingRepository) Delete(ctx context.Context, id string) error {
	result := r.db.WithContext(ctx).Delete(&entities.MarketingCreation{}, "id = ?", id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
func (r *MarketingRepository) SaveImage(ctx context.Context, image *entities.MarketingImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}
func (r *MarketingRepository) Image(ctx context.Context, id string) (*entities.MarketingImage, error) {
	image := new(entities.MarketingImage)
	err := r.db.WithContext(ctx).First(image, "id = ?", id).Error
	return image, mapError(err)
}
func (r *MarketingRepository) HasImage(ctx context.Context, id string) error {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.MarketingImage{}).Where("id = ?", id).Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return apperrors.ErrInvalidInput
	}
	return nil
}
func (r *MarketingRepository) HasProduct(ctx context.Context, id string, active bool) error {
	var count int64
	q := r.db.WithContext(ctx).Model(&entities.StockItem{}).Where("id = ?", id)
	if active {
		q = q.Where("active = ?", true)
	}
	if err := q.Count(&count).Error; err != nil {
		return err
	}
	if count == 0 {
		return apperrors.ErrInvalidInput
	}
	return nil
}
func mapError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return apperrors.ErrNotFound
	}
	return err
}
