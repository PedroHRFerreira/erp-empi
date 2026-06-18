package http

import (
	nethttp "net/http"

	expenseservices "github.com/empi-autocenter/erp-empi/internal/domain/expenses/services"
	financialservices "github.com/empi-autocenter/erp-empi/internal/domain/financial/services"
	"github.com/labstack/echo/v4"
)

type ExpenseHandler struct {
	expenses  *expenseservices.ExpenseService
	financial *financialservices.FinancialService
}

func NewExpenseHandler(expenses *expenseservices.ExpenseService, financial *financialservices.FinancialService) *ExpenseHandler {
	return &ExpenseHandler{expenses: expenses, financial: financial}
}

func (handler *ExpenseHandler) List(c echo.Context) error {
	limit, offset := pagination(c)
	start, end, err := queryDateRange(c)
	if err != nil {
		return writeError(c, err)
	}
	expenses, total, err := handler.expenses.List(c.Request().Context(), limit, offset, start, end)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, paginatedResponse{Data: expenses, Total: total, Limit: limit, Offset: offset})
}

func (handler *ExpenseHandler) Create(c echo.Context) error {
	input := new(expenseservices.ExpenseInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	expense, err := handler.expenses.Create(c.Request().Context(), *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusCreated, expense)
}

func (handler *ExpenseHandler) Update(c echo.Context) error {
	input := new(expenseservices.ExpenseInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	expense, err := handler.expenses.Update(c.Request().Context(), c.Param("id"), *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, expense)
}

func (handler *ExpenseHandler) Delete(c echo.Context) error {
	if err := handler.expenses.Archive(c.Request().Context(), c.Param("id")); err != nil {
		return writeError(c, err)
	}
	return c.NoContent(nethttp.StatusNoContent)
}

func (handler *ExpenseHandler) Summary(c echo.Context) error {
	start, end, err := queryDateRange(c)
	if err != nil {
		return writeError(c, err)
	}
	summary, err := handler.financial.Summary(c.Request().Context(), start, end)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, summary)
}
