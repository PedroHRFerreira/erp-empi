package http

import (
	nethttp "net/http"
	"time"

	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/labstack/echo/v4"
)

type CashHandler struct{ cash *cashservices.CashService }

func NewCashHandler(cash *cashservices.CashService) *CashHandler { return &CashHandler{cash: cash} }

func (h *CashHandler) Current(c echo.Context) error {
	session, err := h.cash.Current(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, session)
}
func (h *CashHandler) ListSessions(c echo.Context) error {
	sessions, err := h.cash.ListSessions(c.Request().Context(), 14)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, sessions)
}
func (h *CashHandler) Balances(c echo.Context) error {
	balances, err := h.cash.Balances(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, balances)
}
func (h *CashHandler) Open(c echo.Context) error {
	input := new(cashservices.OpenInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	session, err := h.cash.Open(c.Request().Context(), *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusCreated, session)
}
func (h *CashHandler) Close(c echo.Context) error {
	input := new(cashservices.CloseInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	session, err := h.cash.Close(c.Request().Context(), *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, session)
}
func (h *CashHandler) AddAdjustment(c echo.Context) error {
	input := new(cashservices.AdjustmentInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	entry, err := h.cash.AddAdjustment(c.Request().Context(), *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusCreated, entry)
}
func (h *CashHandler) PendingInstallments(c echo.Context) error {
	rows, err := h.cash.PendingInstallments(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, rows)
}
func (h *CashHandler) ListPurchases(c echo.Context) error {
	rows, err := h.cash.ListPurchases(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, rows)
}
func (h *CashHandler) GetPurchase(c echo.Context) error {
	row, err := h.cash.GetPurchase(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, row)
}
func (h *CashHandler) CreatePurchase(c echo.Context) error {
	input := new(cashservices.PurchaseInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	purchase, err := h.cash.CreatePurchase(c.Request().Context(), *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusCreated, purchase)
}
func (h *CashHandler) PayInstallment(c echo.Context) error {
	input := struct {
		PaymentMethod string `json:"paymentMethod"`
		PaidAt        string `json:"paidAt"`
	}{}
	if err := c.Bind(&input); err != nil {
		return writeError(c, err)
	}
	row, err := h.cash.PayInstallmentAt(c.Request().Context(), c.Param("id"), entities.PaymentMethod(input.PaymentMethod), input.PaidAt)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, row)
}
func (h *CashHandler) CancelPurchase(c echo.Context) error {
	row, err := h.cash.CancelPurchase(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, row)
}
func (h *CashHandler) DailyEntries(c echo.Context) error {
	day := time.Now()
	if value := c.QueryParam("date"); value != "" {
		parsed, err := time.ParseInLocation("2006-01-02", value, time.Local)
		if err != nil {
			return writeError(c, err)
		}
		day = parsed
	}
	entries, err := h.cash.DailyEntries(c.Request().Context(), day)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, entries)
}
func (h *CashHandler) PayableAlerts(c echo.Context) error {
	alerts, err := h.cash.PayableAlerts(c.Request().Context(), time.Now())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, alerts)
}
