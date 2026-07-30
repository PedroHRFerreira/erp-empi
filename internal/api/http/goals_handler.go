package http

import (
	nethttp "net/http"
	"strings"
	"time"

	goalservices "github.com/empi-autocenter/erp-empi/internal/domain/goals/services"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/labstack/echo/v4"
)

type GoalsHandler struct{ goals *goalservices.GoalService }

func NewGoalsHandler(goals *goalservices.GoalService) *GoalsHandler {
	return &GoalsHandler{goals: goals}
}
func (handler *GoalsHandler) Get(c echo.Context) error {
	month, start, end, err := goalPeriod(c)
	if err != nil {
		return writeError(c, err)
	}
	summary, err := handler.goals.Get(c.Request().Context(), month, start, end)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, summary)
}
func (handler *GoalsHandler) Save(c echo.Context) error {
	month, _, _, err := goalPeriod(c)
	if err != nil {
		return writeError(c, err)
	}
	input := new(goalservices.GoalInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	summary, err := handler.goals.Save(c.Request().Context(), month, *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, summary)
}
func goalPeriod(c echo.Context) (time.Time, time.Time, time.Time, error) {
	value := strings.TrimSpace(c.QueryParam("month"))
	if value == "" {
		now := time.Now()
		value = now.AddDate(0, -1, 0).Format("2006-01")
	}
	month, err := goalservices.ParseMonth(value)
	if err != nil {
		return time.Time{}, time.Time{}, time.Time{}, apperrors.ErrInvalidInput
	}
	startValue, endValue := strings.TrimSpace(c.QueryParam("startDate")), strings.TrimSpace(c.QueryParam("endDate"))
	if startValue == "" && endValue == "" {
		return month, time.Time{}, time.Time{}, nil
	}
	start, startErr := time.ParseInLocation("2006-01-02", startValue, time.Local)
	end, endErr := time.ParseInLocation("2006-01-02", endValue, time.Local)
	if startErr != nil || endErr != nil || end.Before(start) {
		return time.Time{}, time.Time{}, time.Time{}, apperrors.ErrInvalidInput
	}
	return month, start, end.AddDate(0, 0, 1), nil
}
