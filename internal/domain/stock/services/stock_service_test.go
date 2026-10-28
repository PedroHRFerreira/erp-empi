package services

import (
	"context"
	"errors"
	"fmt"
	cashservices "github.com/empi-autocenter/erp-empi/internal/domain/cash/services"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newStockDeleteTest(t *testing.T) (*StockService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entities.StockItem{}, &entities.StockPurchase{}, &entities.StockPurchaseItem{}, &entities.PayableInstallment{}); err != nil {
		t.Fatal(err)
	}
	return NewStockService(repositories.NewStockRepository(db)), db
}

func createStockDeleteFixture(t *testing.T, db *gorm.DB, status entities.PayableInstallmentStatus) entities.StockItem {
	t.Helper()
	item := entities.StockItem{Name: "Filtro", CostCents: 2_000, ResalePriceCents: 3_000, MarkupPercent: 50, Quantity: 1, Active: true}
	if err := db.Create(&item).Error; err != nil {
		t.Fatal(err)
	}
	purchase := entities.StockPurchase{SupplierName: "Fornecedor", TotalCents: 4_000, Status: entities.StockPurchaseConfirmed, PurchasedAt: time.Now(), Items: []entities.StockPurchaseItem{{StockItemID: item.ID, Quantity: 2, UnitCostCents: 2_000, SubtotalCents: 4_000}}, Installments: []entities.PayableInstallment{{Number: 1, AmountCents: 4_000, DueDate: time.Now(), Status: status, PlannedMethod: entities.PayableMethodPix}}}
	if err := db.Create(&purchase).Error; err != nil {
		t.Fatal(err)
	}
	return item
}

func TestDeleteCancelsUnpaidHistoryEvenWhenStockChanged(t *testing.T) {
	service, db := newStockDeleteTest(t)
	item := createStockDeleteFixture(t, db, entities.PayablePending)

	if err := service.Delete(context.Background(), item.ID); err != nil {
		t.Fatal(err)
	}
	var stored entities.StockItem
	if err := db.First(&stored, "id = ?", item.ID).Error; err != nil {
		t.Fatal(err)
	}
	if stored.Active {
		t.Fatal("expected product to be archived")
	}
	var purchase entities.StockPurchase
	if err := db.First(&purchase).Error; err != nil {
		t.Fatal(err)
	}
	if purchase.Status != entities.StockPurchaseCancelled {
		t.Fatalf("expected cancelled purchase, got %s", purchase.Status)
	}
	var installment entities.PayableInstallment
	if err := db.First(&installment).Error; err != nil {
		t.Fatal(err)
	}
	if installment.Status != entities.PayableCancelled {
		t.Fatalf("expected cancelled installment, got %s", installment.Status)
	}
}

func TestDeleteRejectsProductWithPaidInstallment(t *testing.T) {
	service, db := newStockDeleteTest(t)
	item := createStockDeleteFixture(t, db, entities.PayablePaid)

	if err := service.Delete(context.Background(), item.ID); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	var stored entities.StockItem
	if err := db.First(&stored, "id = ?", item.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.Active {
		t.Fatal("paid product must remain active")
	}
}

func TestDeleteRejectsProductFromSharedPurchase(t *testing.T) {
	service, db := newStockDeleteTest(t)
	item := createStockDeleteFixture(t, db, entities.PayablePending)
	otherItem := entities.StockItem{Name: "Oleo", CostCents: 1_000, ResalePriceCents: 1_500, MarkupPercent: 50, Quantity: 1, Active: true}
	if err := db.Create(&otherItem).Error; err != nil {
		t.Fatal(err)
	}
	var purchase entities.StockPurchase
	if err := db.First(&purchase).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&entities.StockPurchaseItem{StockPurchaseID: purchase.ID, StockItemID: otherItem.ID, Quantity: 1, UnitCostCents: 1_000, SubtotalCents: 1_000}).Error; err != nil {
		t.Fatal(err)
	}

	if err := service.Delete(context.Background(), item.ID); !errors.Is(err, apperrors.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	var stored entities.StockItem
	if err := db.First(&stored, "id = ?", item.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !stored.Active {
		t.Fatal("product from shared purchase must remain active")
	}
}

func entryFixture(quantity int, cost int64, supplier string) *StockEntryInput {
	return &StockEntryInput{SupplierName: supplier, Quantity: quantity, UnitCostCents: cost, Installments: []cashservices.PurchaseInstallmentInput{{AmountCents: int64(quantity) * cost, DueDate: "2026-10-27", PlannedMethod: entities.PayableMethodPix}}}
}

func TestCreateWithEntryIsAtomic(t *testing.T) {
	service, db := newStockDeleteTest(t)
	input := StockInput{Name: "Filtro", CostCents: 1000, MarkupPercent: 50, Quantity: 2, Entry: entryFixture(2, 1000, "Fornecedor")}
	input.Entry.Installments[0].AmountCents = 999
	if _, err := service.Create(context.Background(), input); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected invalid installments, got %v", err)
	}
	for _, model := range []any{&entities.StockItem{}, &entities.StockPurchase{}, &entities.StockPurchaseItem{}, &entities.PayableInstallment{}} {
		var count int64
		if err := db.Model(model).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("failed create left %d rows of %T", count, model)
		}
	}
	input.Entry.Installments[0].AmountCents = 2000
	item, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	if item.Quantity != 2 || item.CostCents != 1000 || item.ResalePriceCents != 1500 {
		t.Fatalf("unexpected item: %+v", item)
	}
}

func TestUpdateKeepsLiveBalanceAndReplenishesPaidProduct(t *testing.T) {
	service, db := newStockDeleteTest(t)
	input := StockInput{Name: "Filtro", CostCents: 1000, MarkupPercent: 50, Entry: entryFixture(2, 1000, "Fornecedor A")}
	item, err := service.Create(context.Background(), input)
	if err != nil {
		t.Fatal(err)
	}
	// A purchase can be fully paid and consumed before the next entry.
	if err := db.Model(&entities.PayableInstallment{}).Where("status = ?", entities.PayablePending).Update("status", entities.PayablePaid).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(item).Updates(map[string]any{"quantity": 0, "used_quantity": 2}).Error; err != nil {
		t.Fatal(err)
	}
	staleForm := StockInput{Name: "Filtro atualizado", CostCents: 1000, MarkupPercent: 50, Quantity: 2}
	edited, err := service.Update(context.Background(), item.ID, staleForm)
	if err != nil {
		t.Fatal(err)
	}
	if edited.Quantity != 0 || edited.UsedQuantity != 2 {
		t.Fatalf("stale form overwrote stock balance: %+v", edited)
	}
	staleForm.Entry = entryFixture(3, 1200, "Fornecedor B")
	updated, err := service.Update(context.Background(), item.ID, staleForm)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Quantity != 3 || updated.CostCents != 1200 || updated.ResalePriceCents != 1800 {
		t.Fatalf("unexpected replenished item: %+v", updated)
	}
	var purchases []entities.StockPurchase
	if err := db.Preload("Installments").Order("created_at").Find(&purchases).Error; err != nil {
		t.Fatal(err)
	}
	if len(purchases) != 2 || purchases[0].Installments[0].Status != entities.PayablePaid || purchases[1].Installments[0].Status != entities.PayablePending {
		t.Fatalf("entries are not independent: %+v", purchases)
	}
	detail, err := service.GetSupplier(context.Background(), "fornecedor b")
	if err != nil {
		t.Fatal(err)
	}
	if detail.TotalCents != 3600 || detail.PendingCents != 3600 || detail.ProductCount != 1 {
		t.Fatalf("wrong supplier totals: %+v", detail)
	}
}

func TestExistingProductEntryPreservesCatalogAndWeightedCost(t *testing.T) {
	service, _ := newStockDeleteTest(t)
	item, err := service.Create(context.Background(), StockInput{Name: "Filtro", CostCents: 1000, MarkupPercent: 50, Entry: entryFixture(2, 1000, "Fornecedor")})
	if err != nil {
		t.Fatal(err)
	}
	updated, err := service.Create(context.Background(), StockInput{ExistingStockItemID: item.ID, Name: "Nome indevido", CostCents: 99, MarkupPercent: 0, Entry: entryFixture(2, 2000, "Outro fornecedor")})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != item.ID || updated.Name != "Filtro" || updated.Quantity != 4 || updated.CostCents != 1500 || updated.ResalePriceCents != 2250 {
		t.Fatalf("existing product changed unexpectedly: %+v", updated)
	}
	_, total, err := service.List(context.Background(), 10, 0)
	if err != nil || total != 1 {
		t.Fatalf("duplicate product: total=%d err=%v", total, err)
	}
}

func TestUpdateWithInvalidEntryRollsBackMetadataAndQuantity(t *testing.T) {
	service, _ := newStockDeleteTest(t)
	item, err := service.Create(context.Background(), StockInput{Name: "Filtro", CostCents: 1000, MarkupPercent: 50, Quantity: 2})
	if err != nil {
		t.Fatal(err)
	}
	entry := entryFixture(2, 2000, "Fornecedor")
	entry.Installments[0].DueDate = "invalid"
	if _, err = service.Update(context.Background(), item.ID, StockInput{Name: "Renomeado", CostCents: 1000, MarkupPercent: 75, Entry: entry}); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("expected invalid entry, got %v", err)
	}
	stored, err := service.FindByID(context.Background(), item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Name != "Filtro" || stored.Quantity != 2 || stored.MarkupPercent != 50 {
		t.Fatalf("failed update partially persisted: %+v", stored)
	}
}

func TestPurchasePersistenceFailureRollsBackProduct(t *testing.T) {
	service, db := newStockDeleteTest(t)
	if err := db.Exec("CREATE TRIGGER fail_installment BEFORE INSERT ON payable_installments BEGIN SELECT RAISE(ABORT, 'simulated write failure'); END;").Error; err != nil {
		t.Fatal(err)
	}
	_, err := service.Create(context.Background(), StockInput{Name: "Filtro", CostCents: 1000, MarkupPercent: 50, Entry: entryFixture(2, 1000, "Fornecedor")})
	if err == nil {
		t.Fatal("expected database failure")
	}
	for _, model := range []any{&entities.StockItem{}, &entities.StockPurchase{}, &entities.StockPurchaseItem{}, &entities.PayableInstallment{}} {
		var count int64
		if err := db.Model(model).Count(&count).Error; err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Fatalf("failed transaction left %d rows of %T", count, model)
		}
	}
}

func TestSearchIncludesZeroBalanceBeyondFirstPageAndPrioritizesSupplier(t *testing.T) {
	service, _ := newStockDeleteTest(t)
	first, err := service.Create(context.Background(), StockInput{Name: "Filtro primeiro", CostCents: 1000, MarkupPercent: 50, Entry: entryFixture(1, 1000, "Fornecedor A")})
	if err != nil {
		t.Fatal(err)
	}
	if err := service.repo.DB().Model(first).Update("quantity", 0).Error; err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 12; i++ {
		if _, err := service.Create(context.Background(), StockInput{Name: fmt.Sprintf("Filtro %02d", i), CostCents: 1000, MarkupPercent: 50}); err != nil {
			t.Fatal(err)
		}
	}
	matches, total, err := service.Search(context.Background(), 10, 0, "filtro", "Fornecedor A")
	if err != nil {
		t.Fatal(err)
	}
	if total != 13 || len(matches) != 10 || matches[0].ID != first.ID || matches[0].Quantity != 0 {
		t.Fatalf("search failed supplier priority/zero balance: total=%d matches=%+v", total, matches)
	}
}

func TestSupplierTotalsIgnoreCanceledPurchasesButRetainHistory(t *testing.T) {
	service, db := newStockDeleteTest(t)
	item, err := service.Create(context.Background(), StockInput{Name: "Filtro", CostCents: 1000, MarkupPercent: 50, Entry: entryFixture(2, 1000, "Fornecedor")})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = service.Create(context.Background(), StockInput{ExistingStockItemID: item.ID, Entry: entryFixture(1, 1000, " fornecedor ")}); err != nil {
		t.Fatal(err)
	}
	var purchase entities.StockPurchase
	if err := db.Order("created_at desc").First(&purchase).Error; err != nil {
		t.Fatal(err)
	}
	if _, err = cashservices.NewCashService(db).CancelPurchase(context.Background(), purchase.ID); err != nil {
		t.Fatal(err)
	}
	detail, err := service.GetSupplier(context.Background(), "FORNECEDOR")
	if err != nil {
		t.Fatal(err)
	}
	if detail.PurchaseCount != 2 || detail.ProductCount != 1 || detail.TotalCents != 2000 || detail.PendingCents != 2000 || len(detail.Purchases) != 2 {
		t.Fatalf("incorrect supplier history: %+v", detail)
	}
}
