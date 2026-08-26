package services

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newCashTestService(t *testing.T) (*CashService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entities.StockItem{}, &entities.CashSession{}, &entities.CashEntry{}, &entities.StockPurchase{}, &entities.StockPurchaseItem{}, &entities.PayableInstallment{}, &entities.Expense{}); err != nil {
		t.Fatal(err)
	}
	return NewCashService(db), db
}

func createCashTestStock(t *testing.T, db *gorm.DB, name string, cost int64, quantity int) entities.StockItem {
	t.Helper()
	item := entities.StockItem{Name: name, CostCents: cost, MarkupPercent: 20, ResalePriceCents: cost * 12 / 10, Quantity: quantity, Active: true}
	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	return item
}

func TestPurchaseUpdatesMultipleItemsAndOnlyPaidInstallmentsCreateOutflow(t *testing.T) {
	service, db := newCashTestService(t)
	ctx := context.Background()
	first := createCashTestStock(t, db, "Filtro", 10_000, 2)
	second := createCashTestStock(t, db, "Óleo", 5_000, 0)
	due := time.Now().AddDate(0, 0, 5).Format("2006-01-02")

	purchase, err := service.CreatePurchase(ctx, PurchaseInput{
		SupplierName: "Fornecedor A",
		Items:        []PurchaseItemInput{{StockItemID: first.ID, Quantity: 2, UnitCostCents: 20_000}, {StockItemID: second.ID, Quantity: 4, UnitCostCents: 5_000}},
		Installments: []PurchaseInstallmentInput{{AmountCents: 30_000, DueDate: due, PlannedMethod: entities.PayableMethodBoleto}, {AmountCents: 30_000, DueDate: due, PlannedMethod: entities.PayableMethodCreditCard}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(purchase.Items) != 2 || len(purchase.Installments) != 2 || purchase.TotalCents != 60_000 {
		t.Fatalf("unexpected purchase: %#v", purchase)
	}

	var updatedFirst, updatedSecond entities.StockItem
	db.First(&updatedFirst, "id = ?", first.ID)
	db.First(&updatedSecond, "id = ?", second.ID)
	if updatedFirst.Quantity != 4 || updatedFirst.CostCents != 15_000 {
		t.Fatalf("expected weighted stock 4 @ 15000, got %d @ %d", updatedFirst.Quantity, updatedFirst.CostCents)
	}
	if updatedSecond.Quantity != 4 || updatedSecond.CostCents != 5_000 {
		t.Fatalf("unexpected second stock: %#v", updatedSecond)
	}

	var entryCount, expenseCount int64
	db.Model(&entities.CashEntry{}).Count(&entryCount)
	db.Model(&entities.Expense{}).Count(&expenseCount)
	if entryCount != 0 || expenseCount != 0 {
		t.Fatalf("pending purchase affected finances: entries=%d expenses=%d", entryCount, expenseCount)
	}

	paidDate := time.Now().AddDate(0, 0, -2).Format("2006-01-02")
	paid, err := service.PayInstallmentAt(ctx, purchase.Installments[0].ID, entities.PaymentMethodPix, paidDate)
	if err != nil {
		t.Fatal(err)
	}
	if paid.Status != entities.PayablePaid || paid.PaymentMethod != entities.PaymentMethodPix {
		t.Fatalf("unexpected paid row: %#v", paid)
	}
	if paid.PaidAt == nil || paid.PaidAt.Format("2006-01-02") != paidDate {
		t.Fatalf("expected informed payment date %s, got %#v", paidDate, paid.PaidAt)
	}
	db.Model(&entities.CashEntry{}).Count(&entryCount)
	db.Model(&entities.Expense{}).Count(&expenseCount)
	if entryCount != 1 || expenseCount != 0 {
		t.Fatalf("payment must create one financial outflow and no expense: entries=%d expenses=%d", entryCount, expenseCount)
	}

	if _, err := service.PayInstallment(ctx, purchase.Installments[1].ID, entities.PaymentMethodCash); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("cash payment without open drawer must conflict, got %v", err)
	}
	if _, err := service.Open(ctx, OpenInput{OpeningCashCents: 10_000}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.PayInstallment(ctx, purchase.Installments[1].ID, entities.PaymentMethodCash); err != nil {
		t.Fatal(err)
	}
	current, err := service.Current(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if current.ExpectedCashCents != -20_000 {
		t.Fatalf("expected physical cash -20000 after payment, got %d", current.ExpectedCashCents)
	}
}

func TestParsePaidAtKeepsCurrentLocalClockForDateOnlyInput(t *testing.T) {
	now := time.Date(2026, time.August, 26, 19, 17, 6, 123, time.Local)
	paidAt, err := parsePaidAtAt("2026-08-20", now)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Date(2026, time.August, 20, 19, 17, 6, 123, time.Local)
	if !paidAt.Equal(want) {
		t.Fatalf("expected selected date with current clock %v, got %v", want, paidAt)
	}
}

func TestInstallmentPaymentCanOnlyBeRevokedOnce(t *testing.T) {
	service, db := newCashTestService(t)
	ctx := context.Background()
	item := createCashTestStock(t, db, "Correia", 9_000, 0)
	purchase, err := service.CreatePurchase(ctx, PurchaseInput{
		SupplierName: "Fornecedor",
		Items:        []PurchaseItemInput{{StockItemID: item.ID, Quantity: 1, UnitCostCents: 9_000}},
		Installments: []PurchaseInstallmentInput{{AmountCents: 9_000, DueDate: "2026-08-26", PlannedMethod: entities.PayableMethodPix}},
	})
	if err != nil {
		t.Fatal(err)
	}
	id := purchase.Installments[0].ID
	if _, err := service.PayInstallment(ctx, id, entities.PaymentMethodPix); err != nil {
		t.Fatal(err)
	}
	revoked, err := service.RevokeInstallmentPayment(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if revoked.Status != entities.PayablePending || revoked.PaymentRevokedAt == nil || revoked.PaidAt != nil || revoked.CashEntryID != nil || revoked.PaymentMethod != "" {
		t.Fatalf("unexpected revoked installment: %#v", revoked)
	}
	var entryCount int64
	if err := db.Model(&entities.CashEntry{}).Count(&entryCount).Error; err != nil {
		t.Fatal(err)
	}
	if entryCount != 0 {
		t.Fatalf("expected revoked payment outflow to be removed, got %d entries", entryCount)
	}
	if _, err := service.PayInstallment(ctx, id, entities.PaymentMethodPix); err != nil {
		t.Fatal(err)
	}
	if _, err := service.RevokeInstallmentPayment(ctx, id); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected second revocation to conflict, got %v", err)
	}
	history, err := service.GetInstallmentHistory(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if history.Purchase == nil || len(history.Installments) != 1 || history.Installment.PaymentRevokedAt == nil {
		t.Fatalf("unexpected payment history: %#v", history)
	}
}

func TestPurchaseValidationAndSafeCancellation(t *testing.T) {
	service, db := newCashTestService(t)
	ctx := context.Background()
	item := createCashTestStock(t, db, "Pastilha", 8_000, 0)
	due := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	input := PurchaseInput{SupplierName: "Fornecedor B", Items: []PurchaseItemInput{{StockItemID: item.ID, Quantity: 3, UnitCostCents: 8_000}}, Installments: []PurchaseInstallmentInput{{AmountCents: 23_999, DueDate: due, PlannedMethod: entities.PayableMethodBoleto}}}
	if _, err := service.CreatePurchase(ctx, input); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected invalid installment sum, got %v", err)
	}
	input.Installments[0].AmountCents = 24_000
	purchase, err := service.CreatePurchase(ctx, input)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CancelPurchase(ctx, purchase.ID); err != nil {
		t.Fatal(err)
	}
	var updated entities.StockItem
	db.First(&updated, "id = ?", item.ID)
	if updated.Quantity != 0 {
		t.Fatalf("expected cancellation to revert stock, got %d", updated.Quantity)
	}
	var installment entities.PayableInstallment
	db.First(&installment, "id = ?", purchase.Installments[0].ID)
	if installment.Status != entities.PayableCancelled {
		t.Fatalf("expected cancelled installment, got %s", installment.Status)
	}
}

func TestCancelPurchaseRestoresExactStockSnapshotAndRejectsChangedStock(t *testing.T) {
	service, db := newCashTestService(t)
	ctx := context.Background()
	item := createCashTestStock(t, db, "Óleo", 10_000, 2)
	item.ResalePriceCents = 13_000
	if err := db.Save(item).Error; err != nil {
		t.Fatal(err)
	}
	due := time.Now().AddDate(0, 0, 10).Format("2006-01-02")
	create := func() *entities.StockPurchase {
		purchase, err := service.CreatePurchase(ctx, PurchaseInput{SupplierName: "Fornecedor snapshot", Items: []PurchaseItemInput{{StockItemID: item.ID, Quantity: 2, UnitCostCents: 20_000}}, Installments: []PurchaseInstallmentInput{{AmountCents: 40_000, DueDate: due, PlannedMethod: entities.PayableMethodBoleto}}})
		if err != nil {
			t.Fatal(err)
		}
		return purchase
	}
	purchase := create()
	if _, err := service.CancelPurchase(ctx, purchase.ID); err != nil {
		t.Fatal(err)
	}
	var restored entities.StockItem
	if err := db.First(&restored, "id = ?", item.ID).Error; err != nil {
		t.Fatal(err)
	}
	if restored.Quantity != 2 || restored.CostCents != 10_000 || restored.ResalePriceCents != 13_000 {
		t.Fatalf("stock snapshot was not restored: %+v", restored)
	}

	purchase = create()
	if err := db.Model(&entities.StockItem{}).Where("id = ?", item.ID).Update("quantity", 3).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.CancelPurchase(ctx, purchase.ID); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected changed stock conflict, got %v", err)
	}
}

func TestPayableAlertsPrioritizeUrgencyAndPreserveOpeningCash(t *testing.T) {
	service, db := newCashTestService(t)
	ctx := context.Background()
	now := time.Now()
	purchase := entities.StockPurchase{SupplierName: "Fornecedor C", TotalCents: 18_000, Status: entities.StockPurchaseConfirmed, PurchasedAt: now}
	purchase.Installments = []entities.PayableInstallment{
		{Number: 1, AmountCents: 3_000, DueDate: now.AddDate(0, 0, -1), Status: entities.PayablePending, PlannedMethod: entities.PayableMethodBoleto},
		{Number: 2, AmountCents: 4_000, DueDate: now, Status: entities.PayablePending, PlannedMethod: entities.PayableMethodPix},
		{Number: 3, AmountCents: 5_000, DueDate: now.AddDate(0, 0, 1), Status: entities.PayablePending, PlannedMethod: entities.PayableMethodBoleto},
		{Number: 4, AmountCents: 6_000, DueDate: now.AddDate(0, 0, 10), Status: entities.PayablePending, PlannedMethod: entities.PayableMethodBoleto},
	}
	if err := db.Create(&purchase).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.Open(ctx, OpenInput{OpeningCashCents: 10_000}); err != nil {
		t.Fatal(err)
	}
	if _, err := service.AddAdjustment(ctx, AdjustmentInput{Direction: entities.CashEntryIn, PaymentMethod: entities.PaymentMethodCash, AmountCents: 6_000, Description: "Entrada", Reason: "Teste"}); err != nil {
		t.Fatal(err)
	}
	alerts, err := service.PayableAlerts(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(alerts) != 4 {
		t.Fatalf("expected 4 alerts, got %#v", alerts)
	}
	if alerts[0].Kind != PayableAlertOverdue || alerts[1].Kind != PayableAlertToday || alerts[2].Kind != PayableAlertTomorrow || alerts[3].Kind != PayableAlertEarly {
		t.Fatalf("unexpected alert order: %#v", alerts)
	}
}

func TestBalancesPreserveNonCashAcrossDrawerCloseAndReopen(t *testing.T) {
	service, db := newCashTestService(t)
	ctx := context.Background()
	opened, err := service.Open(ctx, OpenInput{OpeningCashCents: 5_000})
	if err != nil {
		t.Fatal(err)
	}

	adjustments := []AdjustmentInput{
		{Direction: entities.CashEntryIn, PaymentMethod: entities.PaymentMethodPix, AmountCents: 10_000, Description: "PIX recebido", Reason: "Teste"},
		{Direction: entities.CashEntryIn, PaymentMethod: entities.PaymentMethodDebitCard, AmountCents: 20_000, Description: "Débito recebido", Reason: "Teste"},
		{Direction: entities.CashEntryIn, PaymentMethod: entities.PaymentMethodCreditCard, AmountCents: 30_000, Description: "Crédito recebido", Reason: "Teste"},
		{Direction: entities.CashEntryIn, PaymentMethod: entities.PaymentMethodCash, AmountCents: 4_000, Description: "Dinheiro recebido", Reason: "Teste"},
	}
	for _, input := range adjustments {
		if _, err := service.AddAdjustment(ctx, input); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := service.Close(ctx, CloseInput{ClosingCashCents: 9_000}); err != nil {
		t.Fatal(err)
	}
	// Move the closed fixture to yesterday so Open can model the next business
	// day without relying on the wall clock advancing during the test.
	yesterday := businessDay(time.Now()).AddDate(0, 0, -1)
	if err := db.Model(&entities.CashSession{}).Where("id = ?", opened.ID).Update("business_date", yesterday).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := service.Open(ctx, OpenInput{OpeningCashCents: 9_000}); err != nil {
		t.Fatal(err)
	}

	balances, err := service.Balances(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if balances.PixCents != 10_000 || balances.DebitCardCents != 20_000 || balances.CreditCardCents != 30_000 {
		t.Fatalf("non-cash balances did not survive close and reopen: %#v", balances)
	}
}

func TestBalancesApplySignedOutflowsAndExcludeCash(t *testing.T) {
	service, _ := newCashTestService(t)
	ctx := context.Background()
	if _, err := service.Open(ctx, OpenInput{OpeningCashCents: 50_000}); err != nil {
		t.Fatal(err)
	}

	adjustments := []AdjustmentInput{
		{Direction: entities.CashEntryIn, PaymentMethod: entities.PaymentMethodPix, AmountCents: 12_000, Description: "PIX recebido", Reason: "Teste"},
		{Direction: entities.CashEntryOut, PaymentMethod: entities.PaymentMethodPix, AmountCents: 4_500, Description: "PIX enviado", Reason: "Teste"},
		{Direction: entities.CashEntryIn, PaymentMethod: entities.PaymentMethodCash, AmountCents: 8_000, Description: "Dinheiro recebido", Reason: "Teste"},
		{Direction: entities.CashEntryOut, PaymentMethod: entities.PaymentMethodCash, AmountCents: 3_000, Description: "Dinheiro retirado", Reason: "Teste"},
	}
	for _, input := range adjustments {
		if _, err := service.AddAdjustment(ctx, input); err != nil {
			t.Fatal(err)
		}
	}

	balances, err := service.Balances(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if balances.PixCents != 7_500 {
		t.Fatalf("expected signed PIX balance 7500, got %d", balances.PixCents)
	}
	if balances.DebitCardCents != 0 || balances.CreditCardCents != 0 {
		t.Fatalf("cash movements leaked into non-cash balances: %#v", balances)
	}
}
