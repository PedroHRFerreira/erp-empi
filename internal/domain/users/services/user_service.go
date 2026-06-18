package services

import (
	"context"
	"errors"
	"strings"
	"time"

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
	Name                  string  `json:"name"`
	CPF                   string  `json:"cpf"`
	Phone                 string  `json:"phone"`
	Email                 string  `json:"email"`
	Address               string  `json:"address"`
	Notes                 string  `json:"notes"`
	MarkupPercent         float64 `json:"markupPercent"`
	MachineFeePercent     float64 `json:"machineFeePercent"`
	InstallmentFeePercent float64 `json:"installmentFeePercent"`
}

type UpdateProfileInput struct {
	Name                  string  `json:"name"`
	CPF                   string  `json:"cpf"`
	Phone                 string  `json:"phone"`
	Email                 string  `json:"email"`
	Address               string  `json:"address"`
	Notes                 string  `json:"notes"`
	MarkupPercent         float64 `json:"markupPercent"`
	MachineFeePercent     float64 `json:"machineFeePercent"`
	InstallmentFeePercent float64 `json:"installmentFeePercent"`
}

func NewUserService(repo *repositories.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (service *UserService) SeedAdmin(ctx context.Context, admin config.AdminConfig) error {
	cpf := validation.OnlyDigits(admin.CPF)
	if cpf == "" || admin.Password == "" {
		return errors.New("admin cpf and password are required")
	}
	existing, err := service.repo.FindByCPFAndType(ctx, cpf, entities.UserTypeAdmin)
	if err == nil {
		hash, err := security.HashPassword(admin.Password)
		if err != nil {
			return err
		}
		existing.Name = fallback(admin.Name, "Administrador EMPI")
		existing.PasswordHash = &hash
		existing.Email = strings.TrimSpace(admin.Email)
		existing.Phone = validation.OnlyDigits(admin.Phone)
		existing.MarkupPercent = fallbackFloat(admin.MarkupPercent, 10)
		existing.MachineFeePercent = admin.MachineFeePercent
		existing.InstallmentFeePercent = admin.InstallmentFeePercent
		return service.repo.Update(ctx, existing)
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return err
	}
	hash, err := security.HashPassword(admin.Password)
	if err != nil {
		return err
	}
	user := &entities.User{
		Name:                  fallback(admin.Name, "Administrador EMPI"),
		CPF:                   cpf,
		PasswordHash:          &hash,
		Type:                  entities.UserTypeAdmin,
		Email:                 admin.Email,
		Phone:                 validation.OnlyDigits(admin.Phone),
		MarkupPercent:         fallbackFloat(admin.MarkupPercent, 10),
		MachineFeePercent:     admin.MachineFeePercent,
		InstallmentFeePercent: admin.InstallmentFeePercent,
	}
	return service.repo.Create(ctx, user)
}

func (service *UserService) FindByID(ctx context.Context, id string) (*entities.User, error) {
	return service.repo.FindByID(ctx, id)
}

func (service *UserService) FindByCPF(ctx context.Context, cpf string) (*entities.User, error) {
	return service.repo.FindByCPF(ctx, validation.OnlyDigits(cpf))
}

func (service *UserService) FindByCPFAndType(ctx context.Context, cpf string, userType entities.UserType) (*entities.User, error) {
	return service.repo.FindByCPFAndType(ctx, validation.OnlyDigits(cpf), userType)
}

func (service *UserService) ListClients(ctx context.Context, limit int, offset int) ([]entities.User, int64, error) {
	return service.repo.ListClients(ctx, limit, offset)
}

func (service *UserService) FindActiveClientByID(ctx context.Context, id string) (*entities.User, error) {
	return service.repo.FindActiveClientByID(ctx, id)
}

func (service *UserService) ArchiveClient(ctx context.Context, id string) (*entities.User, error) {
	user, err := service.repo.FindActiveClientByID(ctx, id)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	user.Name = "Cliente removido"
	user.CPF = ""
	user.Phone = ""
	user.Email = ""
	user.Address = ""
	user.Notes = ""
	user.ArchivedAt = &now
	return user, service.repo.Update(ctx, user)
}

func (service *UserService) UpsertClient(ctx context.Context, input UpsertClientInput) (*entities.User, error) {
	phone := validation.OnlyDigits(input.Phone)
	if strings.TrimSpace(input.Name) == "" || !isValidClientPhone(phone) {
		return nil, apperrors.ErrInvalidInput
	}
	user := &entities.User{
		Name:          strings.TrimSpace(input.Name),
		Type:          entities.UserTypeClient,
		Phone:         phone,
		Address:       strings.TrimSpace(input.Address),
		Notes:         strings.TrimSpace(input.Notes),
		MarkupPercent: fallbackFloat(input.MarkupPercent, 10),
	}
	return service.repo.UpsertClient(ctx, user)
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
	if strings.TrimSpace(input.Name) == "" || !validation.IsCPF(cpf) || input.MarkupPercent < 0 || input.MachineFeePercent < 0 || input.InstallmentFeePercent < 0 {
		return nil, apperrors.ErrInvalidInput
	}
	user.Name = strings.TrimSpace(input.Name)
	user.CPF = cpf
	user.Phone = validation.OnlyDigits(input.Phone)
	user.Email = strings.TrimSpace(input.Email)
	user.Address = strings.TrimSpace(input.Address)
	user.Notes = strings.TrimSpace(input.Notes)
	user.MarkupPercent = input.MarkupPercent
	user.MachineFeePercent = input.MachineFeePercent
	user.InstallmentFeePercent = input.InstallmentFeePercent
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

func isValidClientPhone(phone string) bool {
	return len(phone) == 10 || len(phone) == 11
}
