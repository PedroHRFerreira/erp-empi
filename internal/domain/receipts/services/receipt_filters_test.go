package services_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/receipts/repositories"
	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestReceiptFiltersCombineBeforePaginationAndIncludeSplitPayments(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}
	client := entities.User{Name: "Cliente Alice", Phone: "33999990000", Type: entities.UserTypeClient}
	if err := db.Create(&client).Error; err != nil {
		t.Fatal(err)
	}
	start := time.Date(2026, 10, 26, 3, 0, 0, 0, time.UTC)
	end := start.AddDate(0, 0, 1)
	seeds := []entities.Receipt{
		{ID: "before", Services: "Outside before", PaymentMethod: entities.PaymentMethodCash, Timestamps: entities.Timestamps{CreatedAt: start.Add(-time.Second)}},
		{ID: "mixed-start", UserID: &client.ID, Services: "Troca de filtro 100%", VehiclePlate: "ABC1D23", Status: entities.ReceiptStatusPaid, PaymentMethod: entities.PaymentMethodMixed, Timestamps: entities.Timestamps{CreatedAt: start}, Payments: []entities.ReceiptPayment{{Position: 1, PaymentMethod: entities.PaymentMethodPix}, {Position: 2, PaymentMethod: entities.PaymentMethodCash}}},
		{ID: "mixed-end", UserID: &client.ID, Services: "Troca de filtro", VehiclePlate: "XYZ1D23", Status: entities.ReceiptStatusPaid, PaymentMethod: entities.PaymentMethodMixed, Timestamps: entities.Timestamps{CreatedAt: end.Add(-time.Microsecond)}, Payments: []entities.ReceiptPayment{{Position: 1, PaymentMethod: entities.PaymentMethodPix}, {Position: 2, PaymentMethod: entities.PaymentMethodCreditCard}}},
		{ID: "quick", Services: "Venda rapida", Status: entities.ReceiptStatusCancelled, PaymentMethod: entities.PaymentMethodPix, Timestamps: entities.Timestamps{CreatedAt: start.Add(time.Hour)}},
		{ID: "after", Services: "Outside after", PaymentMethod: entities.PaymentMethodCash, Timestamps: entities.Timestamps{CreatedAt: end}},
	}
	for i := range seeds {
		if err := db.Create(&seeds[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	service := receiptservices.NewReceiptService(repositories.NewReceiptRepository(db), nil, nil)
	cases := []struct {
		name    string
		filters receiptservices.ReceiptFilters
		total   int64
	}{
		{"all", receiptservices.ReceiptFilters{}, 5},
		{"client case insensitive", receiptservices.ReceiptFilters{Query: "  ALICE  "}, 2},
		{"phone", receiptservices.ReceiptFilters{Query: "3399999"}, 2},
		{"formatted phone", receiptservices.ReceiptFilters{Query: "(33) 99999-0000"}, 2},
		{"plate", receiptservices.ReceiptFilters{Query: "abc1d23"}, 1},
		{"service", receiptservices.ReceiptFilters{Query: "filtro"}, 2},
		{"literal wildcard", receiptservices.ReceiptFilters{Query: "100%"}, 1},
		{"inclusive local day", receiptservices.ReceiptFilters{StartDate: "2026-10-26", EndDate: "2026-10-26"}, 3},
		{"quick", receiptservices.ReceiptFilters{Type: "quick"}, 3},
		{"normal", receiptservices.ReceiptFilters{Type: "normal"}, 2},
		{"cancelled", receiptservices.ReceiptFilters{Status: "cancelled"}, 1},
		{"pix includes both mixed and historical", receiptservices.ReceiptFilters{PaymentMethod: "pix"}, 3},
		{"cash includes mixed and historical", receiptservices.ReceiptFilters{PaymentMethod: "cash"}, 3},
		{"credit allocation", receiptservices.ReceiptFilters{PaymentMethod: "credit_card"}, 1},
		{"mixed", receiptservices.ReceiptFilters{PaymentMethod: "mixed"}, 2},
		{"combined", receiptservices.ReceiptFilters{Query: "Alice", Status: "paid", Type: "normal", PaymentMethod: "pix", StartDate: "2026-10-26", EndDate: "2026-10-26"}, 2},
		{"none", receiptservices.ReceiptFilters{Query: "no matching receipt"}, 0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rows, total, err := service.ListFiltered(context.Background(), 1, 0, tc.filters)
			if err != nil {
				t.Fatal(err)
			}
			if total != tc.total {
				t.Fatalf("total = %d, want %d", total, tc.total)
			}
			expectedRows := 1
			if total == 0 {
				expectedRows = 0
			}
			if len(rows) != expectedRows {
				t.Fatalf("page length %d, want %d", len(rows), expectedRows)
			}
		})
	}
	rows, total, err := service.ListFiltered(context.Background(), 1, 1, receiptservices.ReceiptFilters{PaymentMethod: "mixed"})
	if err != nil || total != 2 || len(rows) != 1 || rows[0].ID != "mixed-start" || len(rows[0].Payments) != 2 {
		t.Fatalf("second filtered page: %+v total=%d err=%v", rows, total, err)
	}
}

func TestReceiptFiltersRejectInvalidCriteriaBeforeDatabaseAccess(t *testing.T) {
	service := receiptservices.NewReceiptService(nil, nil, nil)
	for _, filters := range []receiptservices.ReceiptFilters{
		{Status: "unknown"}, {PaymentMethod: "unknown"}, {Type: "unknown"},
		{StartDate: "2026-02-30"}, {EndDate: "not-date"}, {StartDate: "2026-10-27", EndDate: "2026-10-26"},
	} {
		_, _, err := service.ListFiltered(context.Background(), 10, 0, filters)
		if !errors.Is(err, apperrors.ErrInvalidInput) {
			t.Fatalf("filters %+v: got %v", filters, err)
		}
	}
}
