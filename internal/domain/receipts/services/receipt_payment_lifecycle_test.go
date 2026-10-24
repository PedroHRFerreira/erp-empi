package services_test

import (
	"context"
	"testing"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
)

func TestNormalSplitReceiptPreservesPaymentsInClientHistoryAndRecovery(t *testing.T) {
	db, service, _, adminID := splitFixture(t)
	ctx := context.Background()
	input := splitInput([]receiptservices.ReceiptPaymentInput{
		{PaymentMethod: entities.PaymentMethodPix, BaseAmountCents: 6000, Installments: 1},
		{PaymentMethod: entities.PaymentMethodCreditCard, BaseAmountCents: 4000, Installments: 4},
	})
	input.Quick = false
	input.Client = userservices.UpsertClientInput{Name: "Cliente dividido", Phone: "33999990000"}
	input.VehicleModel, input.VehicleYear, input.VehiclePlate = "Gol", 2020, "ABC1D23"
	receipt, err := service.Create(ctx, adminID, input)
	if err != nil {
		t.Fatal(err)
	}
	if receipt.UserID == nil {
		t.Fatal("normal receipt must retain its client")
	}
	history, err := service.ListByUserID(ctx, *receipt.UserID)
	if err != nil || len(history) != 1 || len(history[0].Payments) != 2 {
		t.Fatalf("client history lost payments: %+v err=%v", history, err)
	}
	firstID, secondID := receipt.Payments[0].ID, receipt.Payments[1].ID
	cancelled, err := service.Cancel(ctx, receipt.ID)
	if err != nil || cancelled.Status != entities.ReceiptStatusCancelled || len(cancelled.Payments) != 2 {
		t.Fatalf("cancel lost payments: %+v err=%v", cancelled, err)
	}
	reopened, err := service.Reopen(ctx, receipt.ID)
	if err != nil || reopened.Status != entities.ReceiptStatusPending || len(reopened.Payments) != 2 {
		t.Fatalf("reopen lost payments: %+v err=%v", reopened, err)
	}
	if reopened.Payments[0].ID != firstID || reopened.Payments[1].ID != secondID || reopened.Payments[1].Installments != 4 || reopened.PriceCents != 10320 {
		t.Fatalf("recovery changed payment allocation: %+v", reopened.Payments)
	}
	var entries int64
	if err := db.Model(&entities.CashEntry{}).Count(&entries).Error; err != nil {
		t.Fatal(err)
	}
	if entries != 0 {
		t.Fatalf("pending/cancelled receipt generated %d cash entries", entries)
	}
}
