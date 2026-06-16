package services

import (
	"context"
	"errors"
	"strings"

	"github.com/empi-autocenter/erp-empi/config"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/users/repositories"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/empi-autocenter/erp-empi/internal/shared/security"
	"github.com/empi-autocenter/erp-empi/internal/shared/validation"
)

type UserService struct {
	repo *repositories.UserRepository
}

type UpsertClientInput struct {
	Name          string  `json:"name"`
	CPF           string  `json:"cpf"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	Address       string  `json:"address"`
	Notes         string  `json:"notes"`
	MarkupPercent float64 `json:"markupPercent"`
}

type UpdateProfileInput struct {
	Name          string  `json:"name"`
	CPF           string  `json:"cpf"`
	Phone         string  `json:"phone"`
	Email         string  `json:"email"`
	Address       string  `json:"address"`
	Notes         string  `json:"notes"`
	MarkupPercent float64 `json:"markupPercent"`
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (service *UserService) SeedAdmin(ctx context.Context, admin config.AdminConfig) error {
	cpf := validation.OnlyDigits(admin.CPF)
	if cpf == "" || admin.Password == "" {
		return errors.New("admin cpf and password are required")
	}
	existing, err := service.repo.FindByCPF(ctx, cpf)
	if err == nil {
		if existing.Type != entities.UserTypeAdmin {
			return errors.New("configured admin cpf already belongs to a client")
		}
		return nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return err
	}
	hash, err := security.HashPassword(admin.Password)
	if err != nil {
		return err
	}
	user := &entities.User{
		Name:          fallback(admin.Name, "Administrador EMPI"),
		CPF:           cpf,
		PasswordHash:  &hash,
		Type:          entities.UserTypeAdmin,
		Email:         admin.Email,
		Phone:         validation.OnlyDigits(admin.Phone),
		MarkupPercent: fallbackFloat(admin.MarkupPercent, 10),
	}
	return service.repo.Create(ctx, user)
}

func (service *UserService) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return service.repo.FindByID(ctx, id)
}

func (service *UserService) FindByCPF(ctx context.Context, cpf string) (*entities.User, error) {
	return service.repo.FindByCPF(ctx, validation.OnlyDigits(cpf))
}

func (service *UserService) ListClients(ctx context.Context, limit int, offset int) ([]entities.User, int64, error) {
	return service.repo.ListClients(ctx, limit, offset)
}

func (service *UserService) UpsertClient(ctx context.Context, input UpsertClientInput) (*entities.User, error) {
	cpf := validation.OnlyDigits(input.CPF)
	if strings.TrimSpace(input.Name) == "" || !validation.IsCPF(cpf) {
		return nil, apperrors.ErrInvalidInput
	}
	user := &entities.User{
		Name:          strings.TrimSpace(input.Name),
		CPF:           cpf,
		Type:          entities.UserTypeClient,
		Phone:         validation.OnlyDigits(input.Phone),
		Email:         strings.TrimSpace(input.Email),
		Address:       strings.TrimSpace(input.Address),
		Notes:         strings.TrimSpace(input.Notes),
		MarkupPercent: fallbackFloat(input.MarkupPercent, 10),
	}
	return service.repo.UpsertClientByCPF(ctx, user)
}

func (service *UserService) UpdateProfile(ctx context.Context, userID string, input UpdateProfileInput) (*entities.User, error) {
	user, err := service.repo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if user.Type != entities.UserTypeAdmin {
		return nil, apperrors.ErrForbidden
	}
	cpf := validation.OnlyDigits(input.CPF)
	if strings.TrimSpace(input.Name) == "" || !validation.IsCPF(cpf) || input.MarkupPercent < 0 {
		return nil, apperrors.ErrInvalidInput
	}
	user.Name = strings.TrimSpace(input.Name)
	user.CPF = cpf
	user.Phone = validation.OnlyDigits(input.Phone)
	user.Email = strings.TrimSpace(input.Email)
	user.Address = strings.TrimSpace(input.Address)
	user.Notes = strings.TrimSpace(input.Notes)
	user.MarkupPercent = input.MarkupPercent
	return user, service.repo.Update(ctx, user)
}

func fallback(value string, defaultValue string) string {
	if strings.TrimSpace(value) == "" {
		return defaultValue
	}
	return strings.TrimSpace(value)
}

func fallbackFloat(value float64, defaultValue float64) float64 {
	if value == 0 {
		return defaultValue
	}
	return value
}
