package services

import (
	"context"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/gorm"
)

const monthLayout = "2006-01"

// MonthlyGoal is intentionally kept in the goals boundary. The database bootstrap
// must include it in AutoMigrate so a fresh installation creates monthly_goals.
type MonthlyGoal struct {
	Month                time.Time `gorm:"primaryKey;type:date"`
	RevenueTargetCents   int64     `json:"revenueTargetCents" gorm:"not null;default:0"`
	LaborTargetCents     int64     `json:"laborTargetCents" gorm:"not null;default:0"`
	ProductsTargetCents  int64     `json:"productsTargetCents" gorm:"not null;default:0"`
	ClientsTarget        int64     `json:"clientsTarget" gorm:"not null;default:0"`
	NetProfitTargetCents int64     `json:"netProfitTargetCents" gorm:"not null;default:0"`
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

func (MonthlyGoal) TableName() string { return "monthly_goal_targets" }

type GoalInput struct {
	RevenueTargetCents   int64 `json:"revenueTargetCents"`
	LaborTargetCents     int64 `json:"laborTargetCents"`
	ProductsTargetCents  int64 `json:"productsTargetCents"`
	ClientsTarget        int64 `json:"clientsTarget"`
	NetProfitTargetCents int64 `json:"netProfitTargetCents"`
}

type GoalTargets struct {
	Month                string `json:"month"`
	RevenueTargetCents   int64  `json:"revenueTargetCents"`
	LaborTargetCents     int64  `json:"laborTargetCents"`
	ProductsTargetCents  int64  `json:"productsTargetCents"`
	ClientsTarget        int64  `json:"clientsTarget"`
	NetProfitTargetCents int64  `json:"netProfitTargetCents"`
}

type GoalMetrics struct {
	RevenueCents     int64 `json:"revenueCents"`
	LaborCents       int64 `json:"laborCents"`
	ProductsCents    int64 `json:"productsCents"`
	Clients          int64 `json:"clients"`
	Appointments     int64 `json:"appointments"`
	NetProfitCents   int64 `json:"netProfitCents"`
	CardFeesCents    int64 `json:"cardFeesCents"`
	ExpensesCents    int64 `json:"expensesCents"`
	ProductCostCents int64 `json:"productCostCents"`
}

type GoalProjection struct {
	RevenueCents   int64 `json:"revenueCents"`
	LaborCents     int64 `json:"laborCents"`
	ProductsCents  int64 `json:"productsCents"`
	Clients        int64 `json:"clients"`
	NetProfitCents int64 `json:"netProfitCents"`
}

type GoalRequirements struct {
	AverageTicketCents          int64 `json:"averageTicketCents"`
	LaborPerAppointmentCents    int64 `json:"laborPerAppointmentCents"`
	ProductsPerAppointmentCents int64 `json:"productsPerAppointmentCents"`
}

type PricingRecommendation struct {
	StockItemID       string  `json:"stockItemId"`
	Name              string  `json:"name"`
	PreviousQuantity  int64   `json:"previousQuantity"`
	SuggestedQuantity int64   `json:"suggestedQuantity"`
	UnitCostCents     int64   `json:"unitCostCents"`
	MarkupPercent     float64 `json:"markupPercent"`
	CurrentPriceCents int64   `json:"currentPriceCents"`
	MinimumPriceCents int64   `json:"minimumPriceCents"`
	BelowMinimum      bool    `json:"belowMinimum"`
}

type GoalTip struct {
	Kind        string `json:"kind"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

type GoalSummary struct {
	Month                   string                  `json:"month"`
	PeriodStart             string                  `json:"periodStart"`
	PeriodEnd               string                  `json:"periodEnd"`
	Saved                   bool                    `json:"saved"`
	Targets                 GoalTargets             `json:"targets"`
	Previous                GoalMetrics             `json:"previous"`
	Actual                  GoalMetrics             `json:"actual"`
	Projection              GoalProjection          `json:"projection"`
	Requirements            GoalRequirements        `json:"requirements"`
	PendingOpportunityCents int64                   `json:"pendingOpportunityCents"`
	PricingRecommendations  []PricingRecommendation `json:"pricingRecommendations"`
	Tips                    []GoalTip               `json:"tips"`
}

type GoalService struct{ db *gorm.DB }

func NewGoalService(db *gorm.DB) *GoalService { return &GoalService{db: db} }

func ParseMonth(value string) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, apperrors.ErrInvalidInput
	}
	parsed, err := time.ParseInLocation(monthLayout, value, time.Local)
	if err != nil {
		return time.Time{}, apperrors.ErrInvalidInput
	}
	return parsed, nil
}

func (service *GoalService) Get(ctx context.Context, month time.Time, start time.Time, end time.Time) (*GoalSummary, error) {
	month = monthStart(month)
	if start.IsZero() || end.IsZero() {
		start, end = month, month.AddDate(0, 1, 0)
	}
	previousStart := start.AddDate(0, 0, -int(end.Sub(start).Hours()/24))
	actual, err := service.metrics(ctx, start, end)
	if err != nil {
		return nil, err
	}
	previous, err := service.metrics(ctx, previousStart, start)
	if err != nil {
		return nil, err
	}
	pending, err := service.pendingOpportunity(ctx, start, end)
	if err != nil {
		return nil, err
	}

	monthKey := month.Format(monthLayout)
	targets := suggestedTargets(month, previous)
	var stored MonthlyGoal
	err = service.db.WithContext(ctx).First(&stored, "month = ?", month).Error
	saved := err == nil
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	if saved {
		targets = goalTargets(stored)
	}

	recommendations, err := service.pricingRecommendations(ctx, previousStart, start)
	if err != nil {
		return nil, err
	}
	projection := project(actual, month)
	requirements := requirements(targets, actual)
	return &GoalSummary{Month: monthKey, PeriodStart: start.Format("2006-01-02"), PeriodEnd: end.AddDate(0, 0, -1).Format("2006-01-02"), Saved: saved, Targets: targets, Previous: previous, Actual: actual, Projection: projection, Requirements: requirements, PendingOpportunityCents: pending, PricingRecommendations: recommendations, Tips: buildTips(targets, actual, projection, pending, start)}, nil
}

func (service *GoalService) Save(ctx context.Context, month time.Time, input GoalInput) (*GoalSummary, error) {
	if input.RevenueTargetCents < 0 || input.LaborTargetCents < 0 || input.ProductsTargetCents < 0 || input.ClientsTarget < 0 || input.NetProfitTargetCents < 0 {
		return nil, apperrors.ErrInvalidInput
	}
	month = monthStart(month)
	goal := MonthlyGoal{Month: month, RevenueTargetCents: input.RevenueTargetCents, LaborTargetCents: input.LaborTargetCents, ProductsTargetCents: input.ProductsTargetCents, ClientsTarget: input.ClientsTarget, NetProfitTargetCents: input.NetProfitTargetCents}
	if err := service.db.WithContext(ctx).Save(&goal).Error; err != nil {
		return nil, err
	}
	return service.Get(ctx, month, time.Time{}, time.Time{})
}

func (service *GoalService) metrics(ctx context.Context, start, end time.Time) (GoalMetrics, error) {
	var metrics GoalMetrics
	err := service.db.WithContext(ctx).Model(&entities.Receipt{}).Where("status = ?", entities.ReceiptStatusPaid).Where("COALESCE(paid_at, updated_at) >= ? AND COALESCE(paid_at, updated_at) < ?", start, end).Select("COALESCE(SUM(price_cents), 0) AS revenue_cents, COALESCE(SUM(labor_price_cents), 0) AS labor_cents, COALESCE(SUM(products_total_cents), 0) AS products_cents, COUNT(*) AS appointments, COUNT(DISTINCT user_id) AS clients, COALESCE(SUM(card_fee_cents), 0) AS card_fees_cents").Scan(&metrics).Error
	if err != nil {
		return metrics, err
	}
	if err = service.db.WithContext(ctx).Table("receipt_items").Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").Where("receipts.status = ?", entities.ReceiptStatusPaid).Where("COALESCE(receipts.paid_at, receipts.updated_at) >= ? AND COALESCE(receipts.paid_at, receipts.updated_at) < ?", start, end).Select("COALESCE(SUM(receipt_items.unit_cost_cents * receipt_items.quantity), 0)").Scan(&metrics.ProductCostCents).Error; err != nil {
		return metrics, err
	}
	if err = service.db.WithContext(ctx).Model(&entities.Expense{}).Where("archived_at IS NULL").Where("spent_at >= ? AND spent_at < ?", start, end).Select("COALESCE(SUM(amount_cents), 0)").Scan(&metrics.ExpensesCents).Error; err != nil {
		return metrics, err
	}
	metrics.NetProfitCents = metrics.RevenueCents - metrics.ProductCostCents - metrics.ExpensesCents - metrics.CardFeesCents
	return metrics, nil
}

func (service *GoalService) pendingOpportunity(ctx context.Context, start, end time.Time) (int64, error) {
	var result int64
	err := service.db.WithContext(ctx).Model(&entities.Receipt{}).Where("status = ?", entities.ReceiptStatusPending).Where("created_at >= ? AND created_at < ?", start, end).Select("COALESCE(SUM(price_cents), 0)").Scan(&result).Error
	return result, err
}

func (service *GoalService) pricingRecommendations(ctx context.Context, start, end time.Time) ([]PricingRecommendation, error) {
	type row struct {
		StockItemID, Name                                  string
		PreviousQuantity, UnitCostCents, CurrentPriceCents int64
		MarkupPercent                                      float64
	}
	var rows []row
	err := service.db.WithContext(ctx).Table("receipt_items").Select("receipt_items.stock_item_id, stock_items.name, COALESCE(SUM(receipt_items.quantity), 0) AS previous_quantity, stock_items.cost_cents AS unit_cost_cents, stock_items.resale_price_cents AS current_price_cents, stock_items.markup_percent").Joins("JOIN receipts ON receipts.id = receipt_items.receipt_id").Joins("JOIN stock_items ON stock_items.id = receipt_items.stock_item_id").Where("receipts.status = ?", entities.ReceiptStatusPaid).Where("COALESCE(receipts.paid_at, receipts.updated_at) >= ? AND COALESCE(receipts.paid_at, receipts.updated_at) < ?", start, end).Group("receipt_items.stock_item_id, stock_items.name, stock_items.cost_cents, stock_items.resale_price_cents, stock_items.markup_percent").Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	result := make([]PricingRecommendation, 0, len(rows))
	for _, row := range rows {
		minimum := int64(math.Ceil(float64(row.UnitCostCents) * (1 + row.MarkupPercent/100)))
		result = append(result, PricingRecommendation{StockItemID: row.StockItemID, Name: row.Name, PreviousQuantity: row.PreviousQuantity, SuggestedQuantity: grow(row.PreviousQuantity), UnitCostCents: row.UnitCostCents, MarkupPercent: row.MarkupPercent, CurrentPriceCents: row.CurrentPriceCents, MinimumPriceCents: minimum, BelowMinimum: row.CurrentPriceCents < minimum})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].PreviousQuantity > result[j].PreviousQuantity })
	return result, nil
}

func suggestedTargets(month time.Time, previous GoalMetrics) GoalTargets {
	revenue := grow(previous.RevenueCents)
	return GoalTargets{Month: month.Format(monthLayout), RevenueTargetCents: revenue, LaborTargetCents: grow(previous.LaborCents), ProductsTargetCents: grow(previous.ProductsCents), ClientsTarget: grow(previous.Clients), NetProfitTargetCents: int64(math.Ceil(float64(revenue) * .20))}
}
func goalTargets(goal MonthlyGoal) GoalTargets {
	return GoalTargets{Month: goal.Month.Format(monthLayout), RevenueTargetCents: goal.RevenueTargetCents, LaborTargetCents: goal.LaborTargetCents, ProductsTargetCents: goal.ProductsTargetCents, ClientsTarget: goal.ClientsTarget, NetProfitTargetCents: goal.NetProfitTargetCents}
}
func grow(value int64) int64 { return int64(math.Ceil(float64(value) * 1.10)) }
func monthStart(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), 1, 0, 0, 0, 0, time.Local)
}
func project(actual GoalMetrics, month time.Time) GoalProjection {
	elapsed, total := workingDaysElapsed(month), workingDays(month)
	if elapsed == 0 {
		return GoalProjection{}
	}
	scale := float64(total) / float64(elapsed)
	return GoalProjection{RevenueCents: int64(math.Round(float64(actual.RevenueCents) * scale)), LaborCents: int64(math.Round(float64(actual.LaborCents) * scale)), ProductsCents: int64(math.Round(float64(actual.ProductsCents) * scale)), Clients: int64(math.Round(float64(actual.Clients) * scale)), NetProfitCents: int64(math.Round(float64(actual.NetProfitCents) * scale))}
}
func requirements(targets GoalTargets, actual GoalMetrics) GoalRequirements {
	appointments := actual.Appointments
	if appointments < targets.ClientsTarget {
		appointments = targets.ClientsTarget
	}
	if appointments == 0 {
		return GoalRequirements{}
	}
	return GoalRequirements{AverageTicketCents: targets.RevenueTargetCents / appointments, LaborPerAppointmentCents: targets.LaborTargetCents / appointments, ProductsPerAppointmentCents: targets.ProductsTargetCents / appointments}
}
func workingDays(month time.Time) int {
	end := monthStart(month).AddDate(0, 1, 0)
	total := 0
	for day := monthStart(month); day.Before(end); day = day.AddDate(0, 0, 1) {
		if day.Weekday() != time.Saturday && day.Weekday() != time.Sunday {
			total++
		}
	}
	return total
}
func workingDaysElapsed(month time.Time) int {
	month = monthStart(month)
	now := time.Now()
	if now.Before(month) {
		return 0
	}
	end := month.AddDate(0, 1, 0)
	if now.Before(end) {
		end = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, time.Local)
	}
	total := 0
	for day := month; day.Before(end); day = day.AddDate(0, 0, 1) {
		if day.Weekday() != time.Saturday && day.Weekday() != time.Sunday {
			total++
		}
	}
	return total
}
func buildTips(targets GoalTargets, actual GoalMetrics, projection GoalProjection, pending int64, month time.Time) []GoalTip {
	tips := []GoalTip{}
	if pending > 0 {
		tips = append(tips, GoalTip{Kind: "pending", Title: "Recupere recebimentos pendentes", Description: "Há valores pendentes que podem acelerar o caixa do mês."})
	}
	if targets.RevenueTargetCents > 0 && projection.RevenueCents < targets.RevenueTargetCents {
		tips = append(tips, GoalTip{Kind: "pace", Title: "Aumente o ritmo de vendas", Description: "A projeção atual está abaixo da meta de faturamento."})
	}
	if actual.Appointments > 0 && targets.RevenueTargetCents > 0 && actual.RevenueCents/actual.Appointments < requirements(targets, actual).AverageTicketCents {
		tips = append(tips, GoalTip{Kind: "ticket", Title: "Eleve o ticket médio", Description: "Ofereça serviços complementares e produtos compatíveis em cada atendimento."})
	}
	if actual.Appointments > 0 && targets.ProductsTargetCents > 0 && actual.ProductsCents/actual.Appointments < requirements(targets, actual).ProductsPerAppointmentCents {
		tips = append(tips, GoalTip{Kind: "products", Title: "Melhore a venda de produtos", Description: "Revise itens de manutenção preventiva e apresente opções durante o diagnóstico."})
	}
	if targets.NetProfitTargetCents > 0 && projection.NetProfitCents < targets.NetProfitTargetCents {
		tips = append(tips, GoalTip{Kind: "margin", Title: "Proteja sua margem", Description: "Acompanhe custos, despesas e preços para manter o lucro alinhado à meta."})
	}
	return tips
}
