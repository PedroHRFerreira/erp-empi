package http

import (
	"strings"

	authservices "github.com/empi-autocenter/erp-empi/internal/domain/auth/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/labstack/echo/v4"
)

func authMiddleware(auth *authservices.AuthService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			header := c.Request().Header.Get("Authorization")
			token := strings.TrimPrefix(header, "Bearer ")
			if token == "" || token == header {
				return writeError(c, apperrors.ErrUnauthorized)
			}
			claims, err := auth.ParseAccessToken(token)
			if err != nil || claims.Type != string(entities.UserTypeAdmin) {
				return writeError(c, apperrors.ErrUnauthorized)
			}
			setUserID(c, claims.UserID)
			return next(c)
		}
	}
}
