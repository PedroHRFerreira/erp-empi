package repositories

import (
	"context"
	"errors"
	"strings"

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

func (repo *UserRepository) FindActiveClientByID(ctx context.Context, id string) (*entities.User, error) {
	user := new(entities.User)
	err := repo.db.WithContext(ctx).First(user, "id = ? AND type = ? AND archived_at IS NULL", id, entities.UserTypeClient).Error
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

func (repo *UserRepository) FindByCPFAndType(ctx context.Context, cpf string, userType entities.UserType) (*entities.User, error) {
	user := new(entities.User)
	err := repo.db.WithContext(ctx).First(user, "cpf = ? AND type = ?", cpf, userType).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return user, err
}

func (repo *UserRepository) FindClientByPhoneAndName(ctx context.Context, phone string, name string) (*entities.User, error) {
	user := new(entities.User)
	err := repo.db.WithContext(ctx).
		First(user, "phone = ? AND type = ? AND archived_at IS NULL AND LOWER(name) = LOWER(?)", phone, entities.UserTypeClient, strings.TrimSpace(name)).
		Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return user, err
}

func (repo *UserRepository) ListClients(ctx context.Context, limit int, offset int) ([]entities.User, int64, error) {
	var users []entities.User
	var total int64
	query := repo.db.WithContext(ctx).Model(&entities.User{}).Where("type = ? AND archived_at IS NULL", entities.UserTypeClient)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := query.Order("created_at desc").Limit(limit).Offset(offset).Find(&users).Error
	return users, total, err
}

func (repo *UserRepository) UpsertClient(ctx context.Context, user *entities.User) (*entities.User, error) {
	found, err := repo.FindClientByPhoneAndName(ctx, user.Phone, user.Name)
	if err == nil {
		found.Phone = user.Phone
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
