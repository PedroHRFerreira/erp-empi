package http

import (
	nethttp "net/http"

	authservices "github.com/empi-autocenter/erp-empi/internal/domain/auth/services"
	"github.com/labstack/echo/v4"
)

type AuthHandler struct {
	auth *authservices.AuthService
}

func NewAuthHandler(auth *authservices.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

func (handler *AuthHandler) Login(c echo.Context) error {
	input := new(authservices.LoginInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	result, err := handler.auth.Login(c.Request().Context(), *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, result)
}
