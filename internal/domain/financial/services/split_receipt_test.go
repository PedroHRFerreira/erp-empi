package services_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	financialservices "github.com/empi-autocenter/erp-empi/internal/domain/financial/services"
	metricsservices "github.com/empi-autocenter/erp-empi/internal/domain/metrics/services"
	receiptrepos "github.com/empi-autocenter/erp-empi/internal/domain/receipts/repositories"
	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	stockrepos "github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	userrepos "github.com/empi-autocenter/erp-empi/internal/domain/users/repositories"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
)

func TestSplitReceiptCountsOnceAcrossCashFinancialAndMetrics(t *testing.T) {
	db := financialTestDB(t)
	ctx := context.Background()
	admin := entities.User{Name: "Administrador", Type: entities.UserTypeAdmin}
	product := entities.StockItem{Name: "Filtro", Quantity: 5, CostCents: 1000, ResalePriceCents: 2000, Active: true}
	if err := db.Create(&admin).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	cash := cashservices.NewCashService(db)
	if _, err := cash.Open(ctx, cashservices.OpenInput{OpeningCashCents: 1000}); err != nil {
		t.Fatal(err)
	}
	service := receiptservices.NewReceiptService(receiptrepos.NewReceiptRepository(db), stockrepos.NewStockRepository(db), userservices.NewUserService(userrepos.NewUserRepository(db)), cash)
	var input receiptservices.ReceiptInput
	if err := json.Unmarshal([]byte(`{"quick":true,"services":"Revisão com pagamento dividido","laborPriceCents":10000,"discountCents":2000,"payments":[{"paymentMethod":"cash","baseAmountCents":6000,"installments":1},{"paymentMethod":"credit_card","baseAmountCents":4000,"cardFeePercent":5,"installments":4}]}`), &input); err != nil {
		t.Fatal(err)
	}
	input.Items = []receiptservices.ReceiptItemInput{{StockItemID: product.ID, Quantity: 1}}
	receipt, err := service.Create(ctx, admin.ID, input)
	if err != nil {
		t.Fatal(err)
	}
	metrics := metricsservices.NewMetricsService(db)
	pending, err := metrics.Summary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if pending.ReceiptsPending != 1 || pending.RevenuePendingCents != 10200 || pending.RevenuePaidCents != 0 {
		t.Fatalf("unexpected pending totals: %+v", pending)
	}
	for attempt := 0; attempt < 2; attempt++ {
		if _, err := service.MarkPaid(ctx, receipt.ID); err != nil {
			t.Fatal(err)
		}
	}
	var entries []entities.CashEntry
	if err := db.Where("reference_type = ? AND reference_id = ?", "receipt", receipt.ID).Find(&entries).Error; err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected two entries despite repeated payment, got %d", len(entries))
	}
	amounts := map[entities.PaymentMethod]int64{}
	for _, entry := range entries {
		amounts[entry.PaymentMethod] += entry.AmountCents
	}
	if amounts[entities.PaymentMethodCash] != 6000 || amounts[entities.PaymentMethodCreditCard] != 4200 {
		t.Fatalf("unexpected cash distribution: %+v", amounts)
	}
	session, err := cash.Current(ctx)
	if err != nil || session == nil || session.ExpectedCashCents != 7000 {
		t.Fatalf("unexpected drawer: %+v, err=%v", session, err)
	}
	balances, err := cash.Balances(ctx)
	if err != nil || balances.CreditCardCents != 4200 || balances.PixCents != 0 {
		t.Fatalf("unexpected method balances: %+v, err=%v", balances, err)
	}
	if err := db.First(&product, "id = ?", product.ID).Error; err != nil {
		t.Fatal(err)
	}
	if product.Quantity != 4 || product.UsedQuantity != 1 {
		t.Fatalf("expected single stock deduction, got %+v", product)
	}
	now := time.Now()
	financial, err := financialservices.NewFinancialService(db).Summary(ctx, now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if financial.PaidReceiptsCount != 1 || financial.RevenuePaidCents != 10200 || financial.CardFeesCents != 200 || financial.ProductCostCents != 1000 || financial.NetProfitCents != 9000 {
		t.Fatalf("unexpected financial totals: %+v", financial)
	}
	summary, err := metrics.Summary(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if summary.ReceiptsTotal != 1 || summary.ReceiptsPaid != 1 || summary.ReceiptsPending != 0 || summary.RevenuePaidCents != 10200 || summary.AverageTicketPaidCents != 10200 || summary.DiscountsGrantedCents != 2000 {
		t.Fatalf("unexpected metrics totals: %+v", summary)
	}
}
