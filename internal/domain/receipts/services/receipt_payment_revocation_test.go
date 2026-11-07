package services_test

import (
	"context"
	"errors"
	"testing"

	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
)

func TestRevokeReceiptPaymentRestoresPendingBalancesAndStock(t *testing.T) {
	db, service, cash, adminID := splitFixture(t)
	ctx := context.Background()
	stock := entities.StockItem{Name: "Peça revogada", CostCents: 1000, ResalePriceCents: 1000, Quantity: 2, Active: true}
	if err := db.Create(&stock).Error; err != nil {
		t.Fatal(err)
	}
	input := splitInput([]receiptservices.ReceiptPaymentInput{
		{PaymentMethod: entities.PaymentMethodPix, BaseAmountCents: 6000, Installments: 1},
		{PaymentMethod: entities.PaymentMethodCash, BaseAmountCents: 5000, Installments: 1},
	})
	input.Items = []receiptservices.ReceiptItemInput{{StockItemID: stock.ID, Quantity: 1}}
	receipt, err := service.Create(ctx, adminID, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cash.Open(ctx, cashservices.OpenInput{OpeningCashCents: 2000}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.MarkPaid(ctx, receipt.ID); err != nil {
		t.Fatal(err)
	}
	revoked, err := service.RevokePayment(ctx, receipt.ID)
	if err != nil {
		t.Fatal(err)
	}
	if revoked.Status != entities.ReceiptStatusPending || revoked.PaidAt != nil {
		t.Fatalf("expected pending unpaid receipt, got %+v", revoked)
	}
	if len(revoked.Payments) != 2 || revoked.Payments[0].ID != receipt.Payments[0].ID || revoked.Payments[1].ID != receipt.Payments[1].ID || revoked.PriceCents != receipt.PriceCents {
		t.Fatalf("revocation changed payment allocation: %+v", revoked)
	}
	var entries int64
	if err := db.Model(&entities.CashEntry{}).Where("reference_id = ?", receipt.ID).Count(&entries).Error; err != nil {
		t.Fatal(err)
	}
	if entries != 0 {
		t.Fatalf("expected receipt cash entries removed, got %d", entries)
	}
	if err := db.First(&stock, "id = ?", stock.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stock.Quantity != 2 || stock.UsedQuantity != 0 {
		t.Fatalf("revocation did not restore stock: %+v", stock)
	}
	drawer, err := cash.Current(ctx)
	if err != nil {
		t.Fatal(err)
	}
	balances, err := cash.Balances(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if drawer.ExpectedCashCents != 2000 || balances.PixCents != 0 {
		t.Fatalf("revocation left paid balances: drawer=%+v balances=%+v", drawer, balances)
	}
	if _, err := service.RevokePayment(ctx, receipt.ID); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected pending revocation conflict, got %v", err)
	}
	if _, err := service.MarkPaid(ctx, receipt.ID); err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&entities.CashEntry{}).Where("reference_id = ?", receipt.ID).Count(&entries).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.First(&stock, "id = ?", stock.ID).Error; err != nil {
		t.Fatal(err)
	}
	drawer, err = cash.Current(ctx)
	if err != nil {
		t.Fatal(err)
	}
	balances, err = cash.Balances(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if entries != 2 || stock.Quantity != 1 || stock.UsedQuantity != 1 || drawer.ExpectedCashCents != 7000 || balances.PixCents != 6000 {
		t.Fatalf("repayment duplicated or lost effects: entries=%d stock=%+v drawer=%+v balances=%+v", entries, stock, drawer, balances)
	}
}

func TestRevokeReceiptPaymentRejectsUnpaidReceipts(t *testing.T) {
	for _, cancelled := range []bool{false, true} {
		t.Run(map[bool]string{false: "pending", true: "cancelled"}[cancelled], func(t *testing.T) {
			_, service, _, adminID := splitFixture(t)
			ctx := context.Background()
			receipt, err := service.Create(ctx, adminID, splitInput([]receiptservices.ReceiptPaymentInput{{PaymentMethod: entities.PaymentMethodPix, BaseAmountCents: 10000, Installments: 1}}))
			if err != nil {
				t.Fatal(err)
			}
			if cancelled {
				if _, err := service.Cancel(ctx, receipt.ID); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := service.RevokePayment(ctx, receipt.ID); !errors.Is(err, apperrors.ErrConflict) {
				t.Fatalf("expected unpaid receipt conflict, got %v", err)
			}
		})
	}
}
