package services

import (
	"context"
	"errors"
	"fmt"
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
