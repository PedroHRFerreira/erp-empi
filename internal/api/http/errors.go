package http

import (
	"errors"
	nethttp "net/http"

	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/labstack/echo/v4"
)

func writeError(c echo.Context, err error) error {
	status := nethttp.StatusInternalServerError
	message := "Erro interno do servidor."

	switch {
	case errors.Is(err, apperrors.ErrInvalidCredentials):
		status = nethttp.StatusUnauthorized
		message = "Credenciais inválidas."
	case errors.Is(err, apperrors.ErrUnauthorized):
		status = nethttp.StatusUnauthorized
		message = "Não autorizado."
	case errors.Is(err, apperrors.ErrForbidden):
		status = nethttp.StatusForbidden
		message = "Acesso negado."
	case errors.Is(err, apperrors.ErrNotFound):
		status = nethttp.StatusNotFound
		message = "Registro não encontrado."
	case errors.Is(err, apperrors.ErrInvalidInput):
		status = nethttp.StatusBadRequest
		message = "Dados inválidos."
	case errors.Is(err, apperrors.ErrConflict):
		status = nethttp.StatusConflict
		message = "Conflito ao processar a solicitação."
	case errors.Is(err, apperrors.ErrInsufficientStock):
		status = nethttp.StatusConflict
		message = "Estoque insuficiente."
	case errors.Is(err, apperrors.ErrReservedStock):
		status = nethttp.StatusConflict
		message = "Produto reservado em outro recibo pendente."
	}

	return c.JSON(status, errorResponse{Message: message})
}
