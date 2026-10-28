package maintenance

import (
	_ "embed"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed correct_ss_stock.sql
var correctionSQL string

func TestPostgresSSCorrectionPreservesPaidSaleAndOriginalPurchase(t *testing.T) {
	dsn := os.Getenv("STOCK_TEST_DATABASE_DSN")
	if dsn == "" {
		t.Skip("set STOCK_TEST_DATABASE_DSN to an isolated local validation PostgreSQL database")
	}
	if !strings.Contains(dsn, "host=127.0.0.1") || !strings.Contains(dsn, "dbname=empi_validation_") {
		t.Fatal("requires local validation database")
	}
	admin, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { sql, _ := admin.DB(); sql.Close() }()
	schema := fmt.Sprintf("ss_correction_%d", time.Now().UnixNano())
	if err = admin.Exec("CREATE SCHEMA " + schema).Error; err != nil {
		t.Fatal(err)
	}
	defer admin.Exec("DROP SCHEMA " + schema + " CASCADE")
	db, err := gorm.Open(postgres.Open(dsn+" search_path="+schema+" default_query_exec_mode=simple_protocol"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { sql, _ := db.DB(); sql.Close() }()
	if err = db.AutoMigrate(&entities.User{}, &entities.StockItem{}, &entities.Receipt{}, &entities.ReceiptItem{}, &entities.ReceiptPayment{}, &entities.Expense{}, &entities.CashSession{}, &entities.CashEntry{}, &entities.StockPurchase{}, &entities.StockPurchaseItem{}, &entities.PayableInstallment{}); err != nil {
		t.Fatal(err)
	}
	product := entities.StockItem{ID: "a1fec4a7-ca57-4940-b6a5-36d8737675dd", Name: "RESERVATORIO DE ÁGUA E TAMPA", CostCents: 9100, ResalePriceCents: 13100, Quantity: 1, UsedQuantity: 1, Active: true}
	if err = db.Create(&product).Error; err != nil {
		t.Fatal(err)
	}
	original := entities.StockPurchase{ID: "c1d7d520-bfcf-4ade-8076-47f0c6753b20", StockItemID: product.ID, SupplierName: "SS PEÇAS", TotalCents: 9100, Status: entities.StockPurchaseConfirmed, Items: []entities.StockPurchaseItem{{StockItemID: product.ID, Quantity: 1, UnitCostCents: 9100, SubtotalCents: 9100}}, Installments: []entities.PayableInstallment{{ID: "a9bb9f92-bf83-4e9f-9e93-5f6c5c2634cc", Number: 1, AmountCents: 9100, Status: entities.PayablePaid, DueDate: time.Now(), PlannedMethod: entities.PayableMethodPix}}}
	if err = db.Create(&original).Error; err != nil {
		t.Fatal(err)
	}
	wrong := entities.StockPurchase{ID: "2db5a754-f69a-4a6e-be92-0d363c5043c9", StockItemID: product.ID, SupplierName: "SS PEÇAS", TotalCents: 13100, Status: entities.StockPurchaseConfirmed, Items: []entities.StockPurchaseItem{{StockItemID: product.ID, Quantity: 1, UnitCostCents: 13100, SubtotalCents: 13100}}, Installments: []entities.PayableInstallment{{ID: "aaf6b07c-9b40-448b-9b17-73f1cb8364e2", Number: 1, AmountCents: 13100, Status: entities.PayablePending, DueDate: time.Now(), PlannedMethod: entities.PayableMethodPix}}}
	if err = db.Create(&wrong).Error; err != nil {
		t.Fatal(err)
	}
	receipt := entities.Receipt{ID: "1f3e774d-46d5-417c-8354-4572f4beca68", Services: "Venda preservada", Status: entities.ReceiptStatusPaid, PriceCents: 124149, Items: []entities.ReceiptItem{{StockItemID: product.ID, Quantity: 1, UnitCostCents: 9100, UnitResaleCents: 13100}}}
	if err = db.Create(&receipt).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Exec(correctionSQL).Error; err != nil {
		t.Fatal("preview failed", err)
	}
	if err = db.First(&product, "id = ?", product.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !product.Active || product.Quantity != 1 {
		t.Fatal("preview changed stock")
	}
	if err = db.Exec("SET empi.apply_ss_stock = 'true';\n" + correctionSQL).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.First(&product, "id = ?", product.ID).Error; err != nil {
		t.Fatal(err)
	}
	if product.Active || product.Quantity != 0 || product.UsedQuantity != 1 {
		t.Fatal("incorrect archived balance")
	}
	if err = db.First(&wrong, "id = ?", wrong.ID).Error; err != nil {
		t.Fatal(err)
	}
	if wrong.Status != entities.StockPurchaseCancelled {
		t.Fatal("wrong entry not cancelled")
	}
	var part entities.PayableInstallment
	if err = db.First(&part, "id = ?", original.Installments[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if part.Status != entities.PayablePaid {
		t.Fatal("original payment changed")
	}
	if err = db.First(&receipt, "id = ?", receipt.ID).Error; err != nil {
		t.Fatal(err)
	}
	if receipt.Status != entities.ReceiptStatusPaid || receipt.PriceCents != 124149 {
		t.Fatal("sale changed")
	}
	// Reapplying or encountering an already changed balance must fail atomically.
	if err = db.Exec(correctionSQL).Error; err == nil {
		t.Fatal("second correction unexpectedly accepted")
	}
	db.Exec("ROLLBACK")
}
