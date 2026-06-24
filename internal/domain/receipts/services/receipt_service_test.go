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
	admin := createAdmin(t, ctx, userRepo, 0, 0)

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
			Name:  "Cliente Teste",
			Phone: "33999990000",
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

func TestCreateAllowsReceiptWithoutProducts(t *testing.T) {
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
	receiptService := receiptservices.NewReceiptService(receiptRepo, stockRepo, userService)
	admin := createAdmin(t, ctx, userRepo, 0, 0)

	receipt, err := receiptService.Create(ctx, admin.ID, receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name:  "Cliente Sem Produto",
			Phone: "33999990000",
		},
		VehicleModel:    "Gol",
		VehicleYear:     2020,
		VehiclePlate:    "ABC1D23",
		Services:        "Diagnostico",
		LaborPriceCents: 15000,
		PaymentMethod:   entities.PaymentMethodCash,
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Items) != 0 {
		t.Fatalf("expected no receipt items, got %d", len(receipt.Items))
	}
	if receipt.ProductsTotalCents != 0 || receipt.PriceCents != 15000 {
		t.Fatalf("expected total from labor only, got products %d and total %d", receipt.ProductsTotalCents, receipt.PriceCents)
	}
}

func TestCreatePersistsServiceExpensesAndChargesReceiptTotal(t *testing.T) {
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
	receiptService := receiptservices.NewReceiptService(receiptRepo, stockRepo, userService)
	admin := createAdmin(t, ctx, userRepo, 0, 0)

	receipt, err := receiptService.Create(ctx, admin.ID, receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name:  "Cliente Gasto",
			Phone: "33999990000",
		},
		VehicleModel:    "Fiesta",
		VehicleYear:     2021,
		VehiclePlate:    "DEF1D23",
		Services:        "Reparo",
		LaborPriceCents: 20000,
		PaymentMethod:   entities.PaymentMethodCash,
		ServiceExpenses: []receiptservices.ReceiptExpenseInput{
			{
				Description: "Gasolina",
				Category:    "combustível",
				AmountCents: 3000,
				SpentAt:     "2026-06-18",
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if receipt.SubtotalCents != 23000 {
		t.Fatalf("expected service expense in subtotal, got %d", receipt.SubtotalCents)
	}
	if receipt.PriceCents != 23000 {
		t.Fatalf("expected service expense to change receipt total, got %d", receipt.PriceCents)
	}

	var expenses []entities.Expense
	if err := db.Where("receipt_id = ?", receipt.ID).Find(&expenses).Error; err != nil {
		t.Fatal(err)
	}
	if len(expenses) != 1 {
		t.Fatalf("expected one linked expense, got %d", len(expenses))
	}
	if expenses[0].ReceiptID == nil || *expenses[0].ReceiptID != receipt.ID {
		t.Fatalf("expected expense linked to receipt %s, got %+v", receipt.ID, expenses[0].ReceiptID)
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
	admin := createAdmin(t, ctx, userRepo, 0, 0)

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
			Name:  "Cliente Teste",
			Phone: "33999990000",
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
	admin := createAdmin(t, ctx, userRepo, 0, 0)

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

func TestCreateCalculatesReceiptTotalsWithInstallmentFee(t *testing.T) {
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
	admin := createAdmin(t, ctx, userRepo, 5, 8)

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
			Name:  "Cliente Teste",
			Phone: "33999990000",
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
	if receipt.CardFeeCents != 1416 {
		t.Fatalf("expected card fee 1416, got %d", receipt.CardFeeCents)
	}
	if receipt.PriceCents != 19116 {
		t.Fatalf("expected total 19116, got %d", receipt.PriceCents)
	}
	if receipt.Installments != 3 {
		t.Fatalf("expected 3 installments, got %d", receipt.Installments)
	}
}

func TestCreateSelectsSingleCardFeeByPaymentMethod(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		paymentMethod      entities.PaymentMethod
		installments       int
		expectedFeePercent float64
		expectedFeeCents   int64
		expectedTotalCents int64
	}{
		{
			name:               "debit uses machine fee",
			paymentMethod:      entities.PaymentMethodDebitCard,
			installments:       1,
			expectedFeePercent: 5,
			expectedFeeCents:   750,
			expectedTotalCents: 15750,
		},
		{
			name:               "credit in one installment uses machine fee",
			paymentMethod:      entities.PaymentMethodCreditCard,
			installments:       1,
			expectedFeePercent: 5,
			expectedFeeCents:   750,
			expectedTotalCents: 15750,
		},
		{
			name:               "credit in multiple installments uses installment fee",
			paymentMethod:      entities.PaymentMethodCreditCard,
			installments:       4,
			expectedFeePercent: 8,
			expectedFeeCents:   1200,
			expectedTotalCents: 16200,
		},
		{
			name:               "pix has no card fee",
			paymentMethod:      entities.PaymentMethodPix,
			installments:       1,
			expectedFeePercent: 0,
			expectedFeeCents:   0,
			expectedTotalCents: 15000,
		},
		{
			name:               "cash has no card fee",
			paymentMethod:      entities.PaymentMethodCash,
			installments:       1,
			expectedFeePercent: 0,
			expectedFeeCents:   0,
			expectedTotalCents: 15000,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			receipt := createReceiptForFeeTest(t, test.paymentMethod, test.installments, 5, 8)

			if receipt.CardFeePercent != test.expectedFeePercent {
				t.Fatalf("expected fee percent %.2f, got %.2f", test.expectedFeePercent, receipt.CardFeePercent)
			}
			if receipt.CardFeeCents != test.expectedFeeCents {
				t.Fatalf("expected fee cents %d, got %d", test.expectedFeeCents, receipt.CardFeeCents)
			}
			if receipt.PriceCents != test.expectedTotalCents {
				t.Fatalf("expected total %d, got %d", test.expectedTotalCents, receipt.PriceCents)
			}
		})
	}
}

func TestCreateUsesCustomCardFeePercent(t *testing.T) {
	t.Parallel()

	customPercent := 2.5
	receipt := createReceiptForFeeTestWithCustom(t, entities.PaymentMethodCreditCard, 4, 5, 8, &customPercent)

	if receipt.CardFeePercent != customPercent {
		t.Fatalf("expected custom fee percent %.2f, got %.2f", customPercent, receipt.CardFeePercent)
	}
	if receipt.CardFeeCents != 375 {
		t.Fatalf("expected custom card fee 375, got %d", receipt.CardFeeCents)
	}
	if receipt.PriceCents != 15375 {
		t.Fatalf("expected total 15375, got %d", receipt.PriceCents)
	}
}

func TestCreateRejectsNegativeCustomCardFeePercent(t *testing.T) {
	t.Parallel()

	negativePercent := -1.0
	receiptService := receiptservices.NewReceiptService(nil, nil, nil)

	_, err := receiptService.Create(context.Background(), "admin-id", receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name:  "Cliente Taxa",
			Phone: "33999990000",
		},
		VehicleModel:    "Gol",
		VehicleYear:     2020,
		VehiclePlate:    "ABC1D23",
		Services:        "Servico",
		LaborPriceCents: 10000,
		CardFeePercent:  &negativePercent,
		PaymentMethod:   entities.PaymentMethodCreditCard,
		Installments:    1,
	})
	if !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected invalid input, got %v", err)
	}
}

func createReceiptForFeeTest(
	t *testing.T,
	paymentMethod entities.PaymentMethod,
	installments int,
	machineFeePercent float64,
	installmentFeePercent float64,
) *entities.Receipt {
	return createReceiptForFeeTestWithCustom(t, paymentMethod, installments, machineFeePercent, installmentFeePercent, nil)
}

func createReceiptForFeeTestWithCustom(
	t *testing.T,
	paymentMethod entities.PaymentMethod,
	installments int,
	machineFeePercent float64,
	installmentFeePercent float64,
	customPercent *float64,
) *entities.Receipt {
	t.Helper()

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
	admin := createAdmin(t, ctx, userRepo, machineFeePercent, installmentFeePercent)

	stockItem, err := stockService.Create(ctx, stockservices.StockInput{
		Name:          "Filtro",
		CostCents:     5000,
		MarkupPercent: 0,
		Quantity:      10,
	})
	if err != nil {
		t.Fatal(err)
	}

	receipt, err := receiptService.Create(ctx, admin.ID, receiptservices.ReceiptInput{
		Client: userservices.UpsertClientInput{
			Name:  "Cliente Taxa",
			Phone: "33999990000",
		},
		VehicleModel:    "Gol",
		VehicleYear:     2020,
		VehiclePlate:    "ABC1D23",
		Services:        "Servico",
		LaborPriceCents: 10000,
		CardFeePercent:  customPercent,
		PaymentMethod:   paymentMethod,
		Installments:    installments,
		Items: []receiptservices.ReceiptItemInput{
			{StockItemID: stockItem.ID, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	return receipt
}

func createAdmin(
	t *testing.T,
	ctx context.Context,
	repo *userrepos.UserRepository,
	machineFeePercent float64,
	installmentFeePercent float64,
) *entities.User {
	t.Helper()

	admin := &entities.User{
		Name:                  "Admin",
		CPF:                   "52998224725",
		Type:                  entities.UserTypeAdmin,
		MarkupPercent:         10,
		MachineFeePercent:     machineFeePercent,
		InstallmentFeePercent: installmentFeePercent,
	}
	if err := repo.Create(ctx, admin); err != nil {
		t.Fatal(err)
	}
	return admin
}
