package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	metricservices "github.com/empi-autocenter/erp-empi/internal/domain/metrics/services"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMetricsSummarySeparatesCancelledReceiptsFromActiveMovement(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	activeClient := &entities.User{
		Name: "Cliente Ativo",
		Type: entities.UserTypeClient,
	}
	cancelledClient := &entities.User{
		Name: "Cliente Cancelado",
		Type: entities.UserTypeClient,
	}
	if err := db.WithContext(ctx).Create(activeClient).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Create(cancelledClient).Error; err != nil {
		t.Fatal(err)
	}

	activeReceipt := &entities.Receipt{
		UserID:             stringPointer(activeClient.ID),
		VehicleModel:       "Gol",
		VehicleYear:        2020,
		VehiclePlate:       "ABC1D23",
		Services:           "Servico ativo",
		LaborPriceCents:    10000,
		DiscountCents:      1500,
		ProductsTotalCents: 0,
		SubtotalCents:      10000,
		PriceCents:         10000,
		Status:             entities.ReceiptStatusPending,
		Timestamps: entities.Timestamps{
			CreatedAt: time.Now().Add(-2 * time.Hour),
		},
	}
	cancelledReceipt := &entities.Receipt{
		UserID:             stringPointer(cancelledClient.ID),
		VehicleModel:       "Uno",
		VehicleYear:        2021,
		VehiclePlate:       "DEF1D23",
		Services:           "Servico cancelado",
		LaborPriceCents:    20000,
		DiscountCents:      5000,
		ProductsTotalCents: 0,
		SubtotalCents:      20000,
		PriceCents:         20000,
		Status:             entities.ReceiptStatusCancelled,
		Timestamps: entities.Timestamps{
			CreatedAt: time.Now(),
		},
	}
	if err := db.WithContext(ctx).Create(activeReceipt).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.WithContext(ctx).Create(cancelledReceipt).Error; err != nil {
		t.Fatal(err)
	}

	summary, err := metricservices.NewMetricsService(db).Summary(ctx)
	if err != nil {
		t.Fatal(err)
	}

	if summary.ReceiptsTotal != 1 {
		t.Fatalf("expected one active receipt, got %d", summary.ReceiptsTotal)
	}
	if summary.ReceiptsCancelled != 1 {
		t.Fatalf("expected one cancelled receipt, got %d", summary.ReceiptsCancelled)
	}
	if summary.LastReceipt == nil || summary.LastReceipt.ID != activeReceipt.ID {
		t.Fatalf("expected last active receipt, got %+v", summary.LastReceipt)
	}
	if summary.DiscountsGrantedCents != 1500 {
		t.Fatalf("expected active discounts 1500, got %d", summary.DiscountsGrantedCents)
	}
	if summary.ReceiptsActiveTotalCents != 10000 {
		t.Fatalf("expected active receipt total 10000, got %d", summary.ReceiptsActiveTotalCents)
	}
	if len(summary.RecentClients) != 1 || summary.RecentClients[0].ID != activeClient.ID {
		t.Fatalf("expected only active client in recent clients, got %+v", summary.RecentClients)
	}
}

func stringPointer(value string) *string {
	return &value
}
