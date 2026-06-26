package http

import (
	nethttp "net/http"

	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	"github.com/labstack/echo/v4"
)

type ReceiptHandler struct {
	receipts *receiptservices.ReceiptService
}

func NewReceiptHandler(receipts *receiptservices.ReceiptService) *ReceiptHandler {
	return &ReceiptHandler{receipts: receipts}
}

func (handler *ReceiptHandler) List(c echo.Context) error {
	limit, offset := pagination(c)
	receipts, total, err := handler.receipts.List(c.Request().Context(), limit, offset, c.QueryParam("status"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, paginatedResponse{Data: receipts, Total: total, Limit: limit, Offset: offset})
}

func (handler *ReceiptHandler) Get(c echo.Context) error {
	receipt, err := handler.receipts.FindByID(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, receipt)
}

func (handler *ReceiptHandler) Create(c echo.Context) error {
	userID, err := getUserID(c)
	if err != nil {
		return writeError(c, err)
	}
	input := new(receiptservices.ReceiptInput)
	if err := c.Bind(input); err != nil {
		return writeError(c, err)
	}
	receipt, err := handler.receipts.Create(c.Request().Context(), userID, *input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusCreated, receipt)
}

func (handler *ReceiptHandler) MarkPaid(c echo.Context) error {
	receipt, err := handler.receipts.MarkPaid(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, receipt)
}

func (handler *ReceiptHandler) Cancel(c echo.Context) error {
	receipt, err := handler.receipts.Cancel(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, receipt)
}

func (handler *ReceiptHandler) Reopen(c echo.Context) error {
	receipt, err := handler.receipts.Reopen(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, receipt)
}
