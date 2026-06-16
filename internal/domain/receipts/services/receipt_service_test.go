package services_test

import (
	"context"
	"testing"

	"github.com/empi-autocenter/erp-empi/internal/domain/receipts/repositories"
	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	stockrepos "github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	stockservices "github.com/empi-autocenter/erp-empi/internal/domain/stock/services"
	userrepos "github.com/empi-autocenter/erp-empi/internal/domain/users/repositories"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMarkPaidDecreasesStockOnce(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	userRepo := userrepos.NewUserRepository(db)
	stockRepo := stockrepos.NewStockRepository(db)
	receiptRepo := repositories.NewReceiptRepository(db)
	userService := userservices.NewUserService(userRepo)
	stockService := stockservices.NewStockService(stockRepo)
	receiptService := receiptservices.NewReceiptService(receiptRepo, stockRepo, userService)

	stockItem, err := stockService.Create(ctx, stockservices.StockInput{
		Name:          "Filtro de oleo",
		CostCents:     5000,
		MarkupPercent: 10,
		Quantity:      5,
	})
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := receiptService.Create(ctx, receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name: "Cliente Teste",
			CPF:  "52998224725",
		},
		VehicleModel: "Gol",
		VehicleYear:  2020,
		VehiclePlate: "ABC1D23",
		Services:     "Troca de oleo",
		PriceCents:   15000,
		Items: []receiptservices.ReceiptItemInput{
			{StockItemID: stockItem.ID, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if _, err := receiptService.MarkPaid(ctx, receipt.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := receiptService.MarkPaid(ctx, receipt.ID); err != nil {
		t.Fatal(err)
	}

	updated, err := stockService.FindByID(ctx, stockItem.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Quantity != 3 {
		t.Fatalf("expected quantity 3, got %d", updated.Quantity)
	}
	if updated.UsedQuantity != 2 {
		t.Fatalf("expected used quantity 2, got %d", updated.UsedQuantity)
	}
}
