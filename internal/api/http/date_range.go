package http

import (
	"strings"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/labstack/echo/v4"
)

const queryDateLayout = "2006-01-02"

func queryDateRange(c echo.Context) (time.Time, time.Time, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	end := start.AddDate(0, 1, 0)

	if value := strings.TrimSpace(c.QueryParam("startDate")); value != "" {
		parsed, err := time.ParseInLocation(queryDateLayout, value, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.ErrInvalidInput
		}
		start = parsed
	}

	if value := strings.TrimSpace(c.QueryParam("endDate")); value != "" {
		parsed, err := time.ParseInLocation(queryDateLayout, value, time.Local)
		if err != nil {
			return time.Time{}, time.Time{}, apperrors.ErrInvalidInput
		}
		end = parsed.AddDate(0, 0, 1)
	}

	if !end.After(start) {
		return time.Time{}, time.Time{}, apperrors.ErrInvalidInput
	}
	return start, end, nil
}
