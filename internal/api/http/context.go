package http

import (
	"strconv"
	"strings"

	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/labstack/echo/v4"
)

const userIDKey = "userID"

type paginatedResponse struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Limit  int         `json:"limit"`
	Offset int         `json:"offset"`
}

type errorResponse struct {
	Message string `json:"message"`
}

func setUserID(c echo.Context, userID string) {
	c.Set(userIDKey, userID)
}

func getUserID(c echo.Context) (string, error) {
	userID, ok := c.Get(userIDKey).(string)
	if !ok || userID == "" {
		return "", apperrors.ErrUnauthorized
	}
	return userID, nil
}

func pagination(c echo.Context) (int, int) {
	limit := parseInt(c.QueryParam("limit"), 10)
	offset := parseInt(c.QueryParam("offset"), 0)
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func parseInt(value string, fallback int) int {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
