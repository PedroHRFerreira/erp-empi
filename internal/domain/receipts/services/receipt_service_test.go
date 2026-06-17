package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/receipts/repositories"
	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	stockrepos "github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	stockservices "github.com/empi-autocenter/erp-empi/internal/domain/stock/services"
	userrepos "github.com/empi-autocenter/erp-empi/internal/domain/users/repositories"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
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
	admin := createAdmin(t, ctx, userRepo, 0)

	stockItem, err := stockService.Create(ctx, stockservices.StockInput{
		Name:          "Filtro de oleo",
		CostCents:     5000,
		MarkupPercent: 10,
		Quantity:      5,
	})
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := receiptService.Create(ctx, admin.ID, receiptservices.ReceiptInput{
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

func TestCreateRejectsQuantityAboveStock(t *testing.T) {
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
	admin := createAdmin(t, ctx, userRepo, 0)

	stockItem, err := stockService.Create(ctx, stockservices.StockInput{
		Name:          "Pastilha de freio",
		CostCents:     8000,
		MarkupPercent: 15,
		Quantity:      3,
	})
	if err != nil {
		t.Fatal(err)
	}

	_, err = receiptService.Create(ctx, admin.ID, receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name: "Cliente Teste",
			CPF:  "52998224725",
		},
		VehicleModel: "Gol",
		VehicleYear:  2020,
		VehiclePlate: "ABC1D23",
		Services:     "Troca de pastilhas",
		PriceCents:   25000,
		Items: []receiptservices.ReceiptItemInput{
			{StockItemID: stockItem.ID, Quantity: 2},
			{StockItemID: stockItem.ID, Quantity: 2},
		},
	})
	if !errors.Is(err, apperrors.ErrInsufficientStock) {
		t.Fatalf("expected insufficient stock, got %v", err)
	}
}

func TestCreateRejectsReservedPendingStock(t *testing.T) {
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
	admin := createAdmin(t, ctx, userRepo, 0)

	stockItem, err := stockService.Create(ctx, stockservices.StockInput{
		Name:          "Oleo unico",
		CostCents:     3500,
		MarkupPercent: 10,
		Quantity:      1,
	})
	if err != nil {
		t.Fatal(err)
	}

	input := receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name:  "Cliente Reserva",
			Phone: "33999990000",
		},
		VehicleModel:    "Gol",
		VehicleYear:     2020,
		VehiclePlate:    "ABC1D23",
		Services:        "Troca de oleo",
		LaborPriceCents: 10000,
		Items: []receiptservices.ReceiptItemInput{
			{StockItemID: stockItem.ID, Quantity: 1},
		},
	}
	if _, err := receiptService.Create(ctx, admin.ID, input); err != nil {
		t.Fatal(err)
	}

	input.Client.Name = "Outro Cliente"
	input.Client.Phone = "33888880000"
	_, err = receiptService.Create(ctx, admin.ID, input)
	if !errors.Is(err, apperrors.ErrReservedStock) {
		t.Fatalf("expected reserved stock, got %v", err)
	}
}

func TestCreateCalculatesReceiptTotalsWithCreditCardFee(t *testing.T) {
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
	admin := createAdmin(t, ctx, userRepo, 5)

	stockItem, err := stockService.Create(ctx, stockservices.StockInput{
		Name:          "Oleo",
		CostCents:     3500,
		MarkupPercent: 10,
		Quantity:      50,
	})
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := receiptService.Create(ctx, admin.ID, receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name: "Cliente Teste",
			CPF:  "52998224725",
		},
		VehicleModel:    "Gol",
		VehicleYear:     2020,
		VehiclePlate:    "ABC1D23",
		Services:        "Troca de oleo",
		LaborPriceCents: 10000,
		PaymentMethod:   entities.PaymentMethodCreditCard,
		Installments:    3,
		Items: []receiptservices.ReceiptItemInput{
			{StockItemID: stockItem.ID, Quantity: 2},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	if receipt.ProductsTotalCents != 7700 {
		t.Fatalf("expected products total 7700, got %d", receipt.ProductsTotalCents)
	}
	if receipt.SubtotalCents != 17700 {
		t.Fatalf("expected subtotal 17700, got %d", receipt.SubtotalCents)
	}
	if receipt.CardFeeCents != 885 {
		t.Fatalf("expected card fee 885, got %d", receipt.CardFeeCents)
	}
	if receipt.PriceCents != 18585 {
		t.Fatalf("expected total 18585, got %d", receipt.PriceCents)
	}
	if receipt.Installments != 3 {
		t.Fatalf("expected 3 installments, got %d", receipt.Installments)
	}
}

func createAdmin(t *testing.T, ctx context.Context, repo *userrepos.UserRepository, machineFeePercent float64) *entities.User {
	t.Helper()

	admin := &entities.User{
		Name:              "Admin",
		CPF:               "52998224725",
		Type:              entities.UserTypeAdmin,
		MarkupPercent:     10,
		MachineFeePercent: machineFeePercent,
	}
	if err := repo.Create(ctx, admin); err != nil {
		t.Fatal(err)
	}
	return admin
}
