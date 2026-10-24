package services_test

import (
	"context"
	"errors"
	"math"
	"testing"

	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	receiptrepos "github.com/empi-autocenter/erp-empi/internal/domain/receipts/repositories"
	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	stockrepos "github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	userrepos "github.com/empi-autocenter/erp-empi/internal/domain/users/repositories"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func splitFixture(t *testing.T) (*gorm.DB, *receiptservices.ReceiptService, *cashservices.CashService, string) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	users := userrepos.NewUserRepository(db)
	admin := createAdmin(t, context.Background(), users, 5, 8)
	cash := cashservices.NewCashService(db)
	service := receiptservices.NewReceiptService(receiptrepos.NewReceiptRepository(db), stockrepos.NewStockRepository(db), userservices.NewUserService(users), cash)
	return db, service, cash, admin.ID
}

func splitInput(payments []receiptservices.ReceiptPaymentInput) receiptservices.ReceiptInput {
	return receiptservices.ReceiptInput{Quick: true, Services: "Serviço dividido", LaborPriceCents: 11000, DiscountCents: 1000, Payments: payments}
}

func TestSplitPaymentsFeesPersistenceAndUpdate(t *testing.T) {
	for _, tc := range []struct {
		name         string
		methods      [2]entities.PaymentMethod
		installments int
		fee, total   int64
	}{
		{"pix cash", [2]entities.PaymentMethod{entities.PaymentMethodPix, entities.PaymentMethodCash}, 1, 0, 10000},
		{"pix credit installments", [2]entities.PaymentMethod{entities.PaymentMethodPix, entities.PaymentMethodCreditCard}, 4, 320, 10320},
		{"credit debit", [2]entities.PaymentMethod{entities.PaymentMethodCreditCard, entities.PaymentMethodDebitCard}, 4, 680, 10680},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, service, _, adminID := splitFixture(t)
			input := splitInput([]receiptservices.ReceiptPaymentInput{{PaymentMethod: tc.methods[0], BaseAmountCents: 6000, Installments: tc.installments}, {PaymentMethod: tc.methods[1], BaseAmountCents: 4000, Installments: tc.installments}})
			receipt, err := service.Create(context.Background(), adminID, input)
			if err != nil {
				t.Fatal(err)
			}
			if receipt.Status != entities.ReceiptStatusPending || receipt.PaymentMethod != entities.PaymentMethodMixed || receipt.CardFeePercent != 0 || receipt.Installments != 1 || receipt.PriceCents != tc.total || receipt.CardFeeCents != tc.fee {
				t.Fatalf("unexpected receipt %+v", receipt)
			}
			if len(receipt.Payments) != 2 || receipt.Payments[0].BaseAmountCents != 6000 || receipt.Payments[1].BaseAmountCents != 4000 || receipt.Payments[1].Position != 1 {
				t.Fatalf("unexpected payments %+v", receipt.Payments)
			}
			list, _, err := service.List(context.Background(), 10, 0, "")
			if err != nil || len(list) != 1 || len(list[0].Payments) != 2 {
				t.Fatalf("list did not hydrate payments: %+v %v", list, err)
			}
			input.Payments = []receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodPix, BaseAmountCents: 10000, Installments: 1}}
			updated, err := service.Update(context.Background(), adminID, receipt.ID, input)
			if err != nil {
				t.Fatal(err)
			}
			var count int64
			db.Model(&entities.ReceiptPayment{}).Where("receipt_id = ?", receipt.ID).Count(&count)
			if count != 1 || len(updated.Payments) != 1 || updated.PaymentMethod != entities.PaymentMethodPix || updated.PriceCents != 10000 {
				t.Fatalf("update failed %+v count=%d", updated, count)
			}
		})
	}
}

func TestSplitPaymentsRejectInvalidInput(t *testing.T) {
	negative, nan, inf := -1.0, math.NaN(), math.Inf(1)
	valid := receiptservices.ReceiptPaymentInput{PaymentMethod: entities.PaymentMethodPix, BaseAmountCents: 6000, Installments: 1}
	for _, tc := range []struct {
		name     string
		payments []receiptservices.ReceiptPaymentInput
	}{
		{"empty", []receiptservices.ReceiptPaymentInput{}},
		{"three", []receiptservices.ReceiptPaymentInput{valid, valid, valid}},
		{"duplicate", []receiptservices.ReceiptPaymentInput{valid, {PaymentMethod: entities.PaymentMethodPix, BaseAmountCents: 4000}}},
		{"sum", []receiptservices.ReceiptPaymentInput{valid, {PaymentMethod: entities.PaymentMethodCash, BaseAmountCents: 3999}}},
		{"zero split", []receiptservices.ReceiptPaymentInput{valid, {PaymentMethod: entities.PaymentMethodCash}}},
		{"negative", []receiptservices.ReceiptPaymentInput{valid, {PaymentMethod: entities.PaymentMethodCash, BaseAmountCents: -1}}},
		{"mixed part", []receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodMixed, BaseAmountCents: 10000}}},
		{"negative fee", []receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodDebitCard, BaseAmountCents: 10000, CardFeePercent: &negative}}},
		{"nan fee", []receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodDebitCard, BaseAmountCents: 10000, CardFeePercent: &nan}}},
		{"inf fee", []receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodDebitCard, BaseAmountCents: 10000, CardFeePercent: &inf}}},
		{"zero installments", []receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodCreditCard, BaseAmountCents: 10000, Installments: 0}}},
		{"excess installments", []receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodCreditCard, BaseAmountCents: 10000, Installments: 13}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, service, _, adminID := splitFixture(t)
			_, err := service.Create(context.Background(), adminID, splitInput(tc.payments))
			if !errors.Is(err, apperrors.ErrInvalidInput) {
				t.Fatalf("expected invalid input got %v", err)
			}
		})
	}
}

func TestSplitMarkPaidRollsBackAndIsIdempotent(t *testing.T) {
	db, service, cash, adminID := splitFixture(t)
	stock := entities.StockItem{Name: "Peça", CostCents: 1000, ResalePriceCents: 1000, Quantity: 2, Active: true}
	if err := db.Create(&stock).Error; err != nil {
		t.Fatal(err)
	}
	input := splitInput([]receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodPix, BaseAmountCents: 6000, Installments: 1}, {PaymentMethod: entities.PaymentMethodCash, BaseAmountCents: 5000, Installments: 1}})
	input.Items = []receiptservices.ReceiptItemInput{{StockItemID: stock.ID, Quantity: 1}}
	receipt, err := service.Create(context.Background(), adminID, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.MarkPaid(context.Background(), receipt.ID); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected closed drawer conflict got %v", err)
	}
	var count int64
	db.Model(&entities.CashEntry{}).Count(&count)
	db.First(&stock, "id = ?", stock.ID)
	pending, _ := service.FindByID(context.Background(), receipt.ID)
	if count != 0 || stock.Quantity != 2 || stock.UsedQuantity != 0 || pending.Status != entities.ReceiptStatusPending || pending.PaidAt != nil {
		t.Fatalf("partial payment survived rollback entries=%d stock=%+v receipt=%+v", count, stock, pending)
	}
	if _, err = cash.Open(context.Background(), cashservices.OpenInput{OpeningCashCents: 2000}); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if _, err = service.MarkPaid(context.Background(), receipt.ID); err != nil {
			t.Fatal(err)
		}
	}
	db.Model(&entities.CashEntry{}).Count(&count)
	db.First(&stock, "id = ?", stock.ID)
	drawer, err := cash.Current(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	balances, err := cash.Balances(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 || stock.Quantity != 1 || stock.UsedQuantity != 1 || drawer.ExpectedCashCents != 7000 || balances.PixCents != 6000 {
		t.Fatalf("unexpected paid totals entries=%d stock=%+v drawer=%+v balances=%+v", count, stock, drawer, balances)
	}
	var entries []entities.CashEntry
	db.Find(&entries)
	for _, entry := range entries {
		if entry.ReferenceID != receipt.ID || entry.PaymentMethod == entities.PaymentMethodMixed {
			t.Fatalf("invalid cash entry %+v", entry)
		}
	}
}

func TestHistoricalReceiptPaymentsPreserveExactAmounts(t *testing.T) {
	db, service, _, _ := splitFixture(t)
	old := entities.Receipt{Services: "Histórico", PaymentMethod: entities.PaymentMethodCreditCard, PriceCents: 10321, CardFeeCents: 321, CardFeePercent: 3.2, Installments: 3}
	if err := db.Create(&old).Error; err != nil {
		t.Fatal(err)
	}
	receipt, err := service.FindByID(context.Background(), old.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(receipt.Payments) != 1 || receipt.Payments[0].AmountCents != 10321 || receipt.Payments[0].BaseAmountCents != 10000 || receipt.Payments[0].CardFeeCents != 321 {
		t.Fatalf("historical amount changed %+v", receipt)
	}
	if _, err := service.MarkPaid(context.Background(), old.ID); err != nil {
		t.Fatal(err)
	}
	var count int64
	db.Model(&entities.ReceiptPayment{}).Where("receipt_id = ?", old.ID).Count(&count)
	if count != 0 {
		t.Fatalf("historical payment persisted during read/confirmation: %d", count)
	}
}

func TestSplitFeesTruncatePerPartAndAllowZeroSingleReceipt(t *testing.T) {
	_, service, _, adminID := splitFixture(t)
	fee := 2.9
	input := receiptservices.ReceiptInput{Quick: true, Services: "Centavos", LaborPriceCents: 200, Payments: []receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodCreditCard, BaseAmountCents: 101, CardFeePercent: &fee, Installments: 12}, {PaymentMethod: entities.PaymentMethodDebitCard, BaseAmountCents: 99, CardFeePercent: &fee, Installments: 1}}}
	receipt, err := service.Create(context.Background(), adminID, input)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.CardFeeCents != 4 || receipt.PriceCents != 204 || receipt.Payments[0].CardFeeCents != 2 || receipt.Payments[1].CardFeeCents != 2 {
		t.Fatalf("unexpected truncation %+v", receipt)
	}
	zero, err := service.Create(context.Background(), adminID, receiptservices.ReceiptInput{Quick: true, Services: "Cortesia"})
	if err != nil {
		t.Fatal(err)
	}
	if zero.PriceCents != 0 || len(zero.Payments) != 1 {
		t.Fatalf("zero receipt not preserved %+v", zero)
	}
}
