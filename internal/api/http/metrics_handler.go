package http

import (
	nethttp "net/http"

	metricservices "github.com/empi-autocenter/erp-empi/internal/domain/metrics/services"
	"github.com/labstack/echo/v4"
)

type MetricsHandler struct {
	metrics *metricservices.MetricsService
}

func NewMetricsHandler(metrics *metricservices.MetricsService) *MetricsHandler {
	return &MetricsHandler{metrics: metrics}
}

func (handler *MetricsHandler) Summary(c echo.Context) error {
	summary, err := handler.metrics.Summary(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, summary)
}
