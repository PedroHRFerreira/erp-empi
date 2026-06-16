package http

import (
	nethttp "net/http"

	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	users *userservices.UserService
}

func NewUserHandler(users *userservices.UserService) *UserHandler {
	return &UserHandler{users: users}
}

func (handler *UserHandler) Me(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return writeError(c, err)
	}
	user, err := handler.users.FindByID(c.Request().Context(), userID)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, user)
}

func (handler *UserHandler) UpdateProfile(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return writeError(c, err)
	}
	input := new(userservices.UpdateProfileInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	user, err := handler.users.UpdateProfile(c.Request().Context(), userID, *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, user)
}

func (handler *UserHandler) ListClients(c echo.Context) error {
	limit, offset := pagination(c)
	users, total, err := handler.users.ListClients(c.Request().Context(), limit, offset)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, paginatedResponse{Data: users, Total: total, Limit: limit, Offset: offset})
}
