package repositories

import (
	"context"
	"errors"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (repo *UserRepository) Create(ctx context.Context, user *entities.User) error {
	return repo.db.WithContext(ctx).Create(user).Error
}

func (repo *UserRepository) Update(ctx context.Context, user *entities.User) error {
	return repo.db.WithContext(ctx).Save(user).Error
}

func (repo *UserRepository) FindByID(ctx context.Context, id string) (*entities.User, error) {
	user := new(entities.User)
	err := repo.db.WithContext(ctx).First(user, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return user, err
}

func (repo *UserRepository) FindByCPF(ctx context.Context, cpf string) (*entities.User, error) {
	user := new(entities.User)
	err := repo.db.WithContext(ctx).First(user, "cpf = ?", cpf).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return user, err
}

func (repo *UserRepository) ListClients(ctx context.Context, limit int, offset int) ([]entities.User, int64, error) {
	var users []entities.User
	var total int64
	query := repo.db.WithContext(ctx).Model(&entities.User{}).Where("type = ?", entities.UserTypeClient)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&users).Error
	return users, total, err
}

func (repo *UserRepository) UpsertClientByCPF(ctx context.Context, user *entities.User) (*entities.User, error) {
	found, err := repo.FindByCPF(ctx, user.CPF)
	if err == nil {
		found.Name = user.Name
		found.Phone = user.Phone
		found.Email = user.Email
		found.Address = user.Address
		found.Notes = user.Notes
		if user.MarkupPercent > 0 {
			found.MarkupPercent = user.MarkupPercent
		}
		return found, repo.Update(ctx, found)
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	user.Type = entities.UserTypeClient
	if user.MarkupPercent == 0 {
		user.MarkupPercent = 10
	}
	return user, repo.Create(ctx, user)
}
