package http

import (
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	marketingservices "github.com/empi-autocenter/erp-empi/internal/domain/marketing/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/marketing/templates"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/labstack/echo/v4"
	"io"
	nethttp "net/http"
)

type MarketingHandler struct {
	service *marketingservices.MarketingService
}

func NewMarketingHandler(service *marketingservices.MarketingService) *MarketingHandler {
	return &MarketingHandler{service: service}
}
func (h *MarketingHandler) Brand(c echo.Context) error {
	b, err := h.service.Brand(c.Request().Context())
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, b)
}
func (h *MarketingHandler) SaveBrand(c echo.Context) error {
	var input entities.MarketingBrand
	if err := c.Bind(&input); err != nil {
		return writeError(c, apperrors.ErrInvalidInput)
	}
	b, err := h.service.SaveBrand(c.Request().Context(), input)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, b)
}
func (h *MarketingHandler) List(c echo.Context) error {
	limit, offset := pagination(c)
	items, total, err := h.service.List(c.Request().Context(), limit, offset)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusOK, paginatedResponse{Data: items, Total: total, Limit: limit, Offset: offset})
}
func (h *MarketingHandler) Save(c echo.Context) error {
	var input entities.MarketingContent
	if err := c.Bind(&input); err != nil {
		return writeError(c, apperrors.ErrInvalidInput)
	}
	item, err := h.service.Save(c.Request().Context(), c.Param("id"), input)
	if err != nil {
		return writeError(c, err)
	}
	status := nethttp.StatusOK
	if c.Param("id") == "" {
		status = nethttp.StatusCreated
	}
	return c.JSON(status, item)
}
func (h *MarketingHandler) Delete(c echo.Context) error {
	if err := h.service.Delete(c.Request().Context(), c.Param("id")); err != nil {
		return writeError(c, err)
	}
	return c.NoContent(nethttp.StatusNoContent)
}
func (h *MarketingHandler) Upload(c echo.Context) error {
	const maxBytes = 5 * 1024 * 1024
	c.Request().Body = nethttp.MaxBytesReader(c.Response(), c.Request().Body, maxBytes+64*1024)
	file, err := c.FormFile("file")
	if err != nil {
		return writeError(c, apperrors.ErrInvalidInput)
	}
	defer func() {
		if c.Request().MultipartForm != nil {
			_ = c.Request().MultipartForm.RemoveAll()
		}
	}()
	if file.Size > maxBytes {
		return writeError(c, apperrors.ErrInvalidInput)
	}
	stream, err := file.Open()
	if err != nil {
		return writeError(c, err)
	}
	defer stream.Close()
	data, err := io.ReadAll(io.LimitReader(stream, maxBytes+1))
	if err != nil {
		return writeError(c, err)
	}
	if len(data) > maxBytes {
		return writeError(c, apperrors.ErrInvalidInput)
	}
	image, err := h.service.Upload(c.Request().Context(), data)
	if err != nil {
		return writeError(c, err)
	}
	return c.JSON(nethttp.StatusCreated, image)
}
func (h *MarketingHandler) Image(c echo.Context) error {
	image, err := h.service.Image(c.Request().Context(), c.Param("id"))
	if err != nil {
		return writeError(c, err)
	}
	c.Response().Header().Set("Cache-Control", "private, no-store")
	return c.Blob(nethttp.StatusOK, image.MimeType, image.Data)
}

func (h *MarketingHandler) Templates(c echo.Context) error {
	return c.JSON(nethttp.StatusOK, templates.Catalog())
}
