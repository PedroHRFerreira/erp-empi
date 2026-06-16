package http

import (
	nethttp "net/http"

	stockservices "github.com/empi-autocenter/erp-empi/internal/domain/stock/services"
	"github.com/labstack/echo/v4"
)

type StockHandler struct {
	stock *stockservices.StockService
}

func NewStockHandler(stock *stockservices.StockService) *StockHandler {
	return &StockHandler{stock: stock}
}

func (handler *StockHandler) List(c echo.Context) error {
	limit, offset := pagination(c)
	items, total, err := handler.stock.List(c.Request().Context(), limit, offset)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, paginatedResponse{Data: items, Total: total, Limit: limit, Offset: offset})
}

func (handler *StockHandler) Create(c echo.Context) error {
	input := new(stockservices.StockInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	item, err := handler.stock.Create(c.Request().Context(), *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusCreated, item)
}

func (handler *StockHandler) Update(c echo.Context) error {
	input := new(stockservices.StockInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	item, err := handler.stock.Update(c.Request().Context(), c.Param("id"), *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, item)
}

func (handler *StockHandler) Delete(c echo.Context) error {
	if err := handler.stock.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return writeError(c, err)
	}
	return c.NoContent(nethttp.StatusNoContent)
}
