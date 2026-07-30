package services_test

import (
	"context"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	goalservices "github.com/empi-autocenter/erp-empi/internal/domain/goals/services"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestGoalServiceSuggestsPersistsAndAggregatesMonthlyGoals(t *testing.T) {
	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	month := time.Date(2026, time.July, 1, 0, 0, 0, 0, time.Local)
	previousPaidAt := month.AddDate(0, -1, 5)
	client := &entities.User{Name: "Cliente", Type: entities.UserTypeClient}
	stock := &entities.StockItem{Name: "Filtro", CostCents: 2000, MarkupPercent: 50, ResalePriceCents: 2900, Quantity: 10}
	if err := db.Create(client).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(stock).Error; err != nil {
		t.Fatal(err)
	}
	previous := &entities.Receipt{UserID: &client.ID, Services: "Revisão", LaborPriceCents: 9000, ProductsTotalCents: 1000, SubtotalCents: 10000, PriceCents: 10000, CardFeeCents: 100, Status: entities.ReceiptStatusPaid, PaidAt: &previousPaidAt, Items: []entities.ReceiptItem{{StockItemID: stock.ID, Quantity: 1, UnitCostCents: 2000, UnitResaleCents: 1000, MarkupPercent: 50}}}
	if err := db.Create(previous).Error; err != nil {
		t.Fatal(err)
	}

	service := goalservices.NewGoalService(db)
	summary, err := service.Get(ctx, month, time.Time{}, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if summary.Saved {
		t.Fatal("expected an unsaved suggested goal")
	}
	if summary.Previous.RevenueCents != 10000 || summary.Previous.LaborCents != 9000 || summary.Previous.ProductsCents != 1000 {
		t.Fatalf("unexpected previous metrics: %+v", summary.Previous)
	}
	if summary.Previous.NetProfitCents != 7900 {
		t.Fatalf("expected net profit 7900, got %d", summary.Previous.NetProfitCents)
	}
	if summary.Targets.RevenueTargetCents != 11000 || summary.Targets.NetProfitTargetCents != 2200 {
		t.Fatalf("unexpected targets: %+v", summary.Targets)
	}
	if len(summary.PricingRecommendations) != 1 || !summary.PricingRecommendations[0].BelowMinimum {
		t.Fatalf("expected a below-minimum pricing recommendation, got %+v", summary.PricingRecommendations)
	}

	saved, err := service.Save(ctx, month, goalservices.GoalInput{RevenueTargetCents: 15000, LaborTargetCents: 10000, ProductsTargetCents: 5000, ClientsTarget: 3, NetProfitTargetCents: 3000})
	if err != nil {
		t.Fatal(err)
	}
	if !saved.Saved || saved.Targets.RevenueTargetCents != 15000 || saved.Targets.ClientsTarget != 3 {
		t.Fatalf("expected persisted targets, got %+v", saved)
	}

	updated, err := service.Save(ctx, month, goalservices.GoalInput{RevenueTargetCents: 20000, LaborTargetCents: 13000, ProductsTargetCents: 7000, ClientsTarget: 5, NetProfitTargetCents: 4000})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Targets.RevenueTargetCents != 20000 || updated.Targets.ClientsTarget != 5 {
		t.Fatalf("expected updated targets, got %+v", updated.Targets)
	}
	var count int64
	if err := db.Model(&goalservices.MonthlyGoal{}).Where("month = ?", month).Count(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected one monthly goal, got %d", count)
	}
}

func TestGoalServiceRejectsInvalidInput(t *testing.T) {
	if _, err := goalservices.ParseMonth("2026-13"); err == nil {
		t.Fatal("expected invalid month error")
	}
}
