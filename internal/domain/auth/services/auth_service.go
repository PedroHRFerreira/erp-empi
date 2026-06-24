package services

import (
	"context"
	"time"

	"github.com/empi-autocenter/erp-empi/config"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/empi-autocenter/erp-empi/internal/shared/security"
	"github.com/empi-autocenter/erp-empi/internal/shared/validation"
)

type AuthService struct {
	cfg   *config.Config
	users *userservices.UserService
}

type LoginInput struct {
	CPF      string `json:"cpf"`
	Password string `json:"password"`
}

type AuthTokens struct {
	AccessToken string `json:"accessToken"`
	ExpiresIn   int64  `json:"expiresIn"`
}

type LoginResult struct {
	User   *entities.User `json:"user"`
	Tokens AuthTokens     `json:"tokens"`
}

func NewAuthService(cfg *config.Config, users *userservices.UserService) *AuthService {
	return &AuthService{cfg: cfg, users: users}
}

func (service *AuthService) Login(ctx context.Context, input LoginInput) (*LoginResult, error) {
	cpf := validation.OnlyDigits(input.CPF)
	if !validation.IsCPF(cpf) || input.Password == "" {
		return nil, apperrors.ErrInvalidCredentials
	}
	user, err := service.users.FindByCPFAndType(ctx, cpf, entities.UserTypeAdmin)
	if err != nil || user.Type != entities.UserTypeAdmin || user.PasswordHash == nil {
		return nil, apperrors.ErrInvalidCredentials
	}
	if !security.CheckPassword(*user.PasswordHash, input.Password) {
		return nil, apperrors.ErrInvalidCredentials
	}
	tokens, err := service.issueTokens(user)
	if err != nil {
		return nil, err
	}
	return &LoginResult{User: user, Tokens: tokens}, nil
}

func (service *AuthService) ParseAccessToken(token string) (*security.Claims, error) {
	return security.ParseToken(token, service.cfg.JWT.AccessSecret)
}

func (service *AuthService) issueTokens(user *entities.User) (AuthTokens, error) {
	accessTTL := time.Duration(service.cfg.JWT.AccessTTLMinutes) * time.Minute
	access, err := security.GenerateToken(user.ID, string(user.Type), service.cfg.JWT.AccessSecret, accessTTL)
	if err != nil {
		return AuthTokens{}, err
	}
	return AuthTokens{AccessToken: access, ExpiresIn: int64(accessTTL.Seconds())}, nil
}
