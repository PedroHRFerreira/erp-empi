package http

import (
	"errors"
	nethttp "net/http"

	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/labstack/echo/v4"
)

func writeError(c echo.Context, err error) error {
	status := nethttp.StatusInternalServerError
	message := "internal server error"

	switch {
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		status = nethttp.StatusUnauthorized
		message = "invalid credentials"
	case errors.Is(err, apperrors.ErrUnauthorized):
		status = nethttp.StatusUnauthorized
		message = "unauthorized"
	case errors.Is(err, apperrors.ErrForbidden):
		status = nethttp.StatusForbidden
		message = "forbidden"
	case errors.Is(err, apperrors.ErrNotFound):
		status = nethttp.StatusNotFound
		message = "not found"
	case errors.Is(err, apperrors.ErrInvalidInput):
		status = nethttp.StatusBadRequest
		message = "invalid input"
	case errors.Is(err, apperrors.ErrConflict):
		status = nethttp.StatusConflict
		message = "conflict"
	case errors.Is(err, apperrors.ErrInsufficientStock):
		status = nethttp.StatusConflict
		message = "insufficient stock"
	}

	return c.JSON(status, errorResponse{Message: message})
}
