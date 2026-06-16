package services

import (
	"context"
	"strings"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/empi-autocenter/erp-empi/internal/shared/validation"
)

type StockService struct {
	repo *repositories.StockRepository
}

type StockInput struct {
	Name          string  `json:"name"`
	Description   string  `json:"description"`
	CostCents     int64   `json:"costCents"`
	MarkupPercent float64 `json:"markupPercent"`
	Quantity      int     `json:"quantity"`
}

func NewStockService(repo *repositories.StockRepository) *StockService {
	return &StockService{repo: repo}
}

func (service *StockService) List(ctx context.Context, limit int, offset int) ([]entities.StockItem, int64, error) {
	return service.repo.List(ctx, limit, offset)
}

func (service *StockService) FindByID(ctx context.Context, id string) (*entities.StockItem, error) {
	return service.repo.FindByID(ctx, id)
}

func (service *StockService) Create(ctx context.Context, input StockInput) (*entities.StockItem, error) {
	item, err := service.buildItem(input, nil)
	if err != nil {
		return nil, err
	}
	return item, service.repo.Create(ctx, item)
}

func (service *StockService) Update(ctx context.Context, id string, input StockInput) (*entities.StockItem, error) {
	current, err := service.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	item, err := service.buildItem(input, current)
	if err != nil {
		return nil, err
	}
	return item, service.repo.Update(ctx, item)
}

func (service *StockService) Delete(ctx context.Context, id string) error {
	item, err := service.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	item.Active = false
	return service.repo.Update(ctx, item)
}

func (service *StockService) buildItem(input StockInput, current *entities.StockItem) (*entities.StockItem, error) {
	if strings.TrimSpace(input.Name) == "" || input.CostCents <= 0 || input.MarkupPercent < 0 || input.Quantity < 0 {
		return nil, apperrors.ErrInvalidInput
	}
	item := current
	if item == nil {
		item = &entities.StockItem{Active: true}
	}
	item.Name = strings.TrimSpace(input.Name)
	item.Description = strings.TrimSpace(input.Description)
	item.CostCents = input.CostCents
	item.MarkupPercent = input.MarkupPercent
	item.ResalePriceCents = validation.CalculateMarkupCents(input.CostCents, input.MarkupPercent)
	item.Quantity = input.Quantity
	return item, nil
}
