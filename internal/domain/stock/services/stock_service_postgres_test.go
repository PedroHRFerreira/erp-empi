package services

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func TestPostgresAtomicStockEntryAndConcurrentReplenishment(t *testing.T) {
	dsn := os.Getenv("STOCK_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set STOCK_TEST_DATABASE_DSN to an isolated local test PostgreSQL database")
	}
	if (!strings.Contains(dsn, "host=127.0.0.1") && !strings.Contains(dsn, "host=localhost")) || (!strings.Contains(dsn, "dbname=empi_validation_") && !strings.Contains(dsn, "dbname=empi_test")) {
		t.Fatal("stock integration tests require an explicitly named local test database")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	adminSQL, err := admin.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer adminSQL.Close()
	schema := fmt.Sprintf("stock_integration_%d", time.Now().UnixNano())
	if err := admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	defer admin.Exec("DROP SCHEMA " + schema + " CASCADE")
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	defer sqlDB.Close()
	if err := db.AutoMigrate(&entities.StockItem{}, &entities.StockPurchase{}, &entities.StockPurchaseItem{}, &entities.PayableInstallment{}); err != nil {
		t.Fatal(err)
	}
	service := NewStockService(repositories.NewStockRepository(db))
	ctx := context.Background()
	invalid := StockInput{Name: "Atomic", CostCents: 1000, MarkupPercent: 50, Entry: entryFixture(2, 1000, "Fornecedor A")}
	invalid.Entry.Installments[0].AmountCents = 100
	if _, err := service.Create(ctx, invalid); err == nil {
		t.Fatal("expected rejected installments")
	}
	var count int64
	if err := db.Model(&entities.StockItem{}).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 0 {
		t.Fatal("invalid purchase persisted a product")
	}
	invalid.Entry.Installments[0].AmountCents = 2000
	item, err := service.Create(ctx, invalid)
	if err != nil {
		t.Fatal(err)
	}
	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, quantity := range []int{3, 2} {
		wg.Add(1)
		go func(quantity int) {
			defer wg.Done()
			_, err := service.Update(ctx, item.ID, StockInput{Name: "Atomic", CostCents: 1000, MarkupPercent: 50, Quantity: 2, Entry: entryFixture(quantity, 2000, "Fornecedor B")})
			results <- err
		}(quantity)
	}
	wg.Wait()
	close(results)
	for err := range results {
		if err != nil {
			t.Fatal(err)
		}
	}
	stored, err := service.FindByID(ctx, item.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Quantity != 7 {
		t.Fatalf("lost concurrent stock entry: quantity=%d", stored.Quantity)
	}
	var purchaseCount, partCount int64
	if err := db.Model(&entities.StockPurchase{}).Count(&purchaseCount).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&entities.PayableInstallment{}).Count(&partCount).Error; err != nil {
		t.Fatal(err)
	}
	if purchaseCount != 3 || partCount != 3 {
		t.Fatalf("entry records missing: purchases=%d installments=%d", purchaseCount, partCount)
	}
	detail, err := service.GetSupplier(ctx, "Fornecedor B")
	if err != nil {
		t.Fatal(err)
	}
	if detail.ProductCount != 1 || detail.TotalCents != 10000 || detail.PendingCents != 10000 {
		t.Fatalf("incorrect supplier view: %+v", detail)
	}
}
