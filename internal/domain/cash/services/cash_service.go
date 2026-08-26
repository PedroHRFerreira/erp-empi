package services

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/empi-autocenter/erp-empi/internal/shared/validation"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type CashService struct{ db *gorm.DB }

type OpenInput struct {
	OpeningCashCents int64 `json:"openingCashCents"`
}
type CloseInput struct {
	ClosingCashCents int64  `json:"closingCashCents"`
	ClosingNotes     string `json:"closingNotes"`
}
type AdjustmentInput struct {
	Direction     entities.CashEntryDirection `json:"direction"`
	PaymentMethod entities.PaymentMethod      `json:"paymentMethod"`
	AmountCents   int64                       `json:"amountCents"`
	Description   string                      `json:"description"`
	Reason        string                      `json:"reason"`
	ReferenceID   string                      `json:"referenceId"`
}

// CashBalances represents the accumulated balances for payment methods that
// are not tied to the lifecycle of a physical cash drawer.
type CashBalances struct {
	PixCents        int64 `json:"pixCents" gorm:"column:pix_cents"`
	DebitCardCents  int64 `json:"debitCardCents" gorm:"column:debit_card_cents"`
	CreditCardCents int64 `json:"creditCardCents" gorm:"column:credit_card_cents"`
}
type PurchaseInput struct {
	SupplierName string                     `json:"supplierName"`
	Items        []PurchaseItemInput        `json:"items"`
	Installments []PurchaseInstallmentInput `json:"installments"`
	// Legacy fields keep the existing endpoint compatible while the UI migrates.
	StockItemID      string `json:"stockItemId"`
	Quantity         int    `json:"quantity"`
	TotalCents       int64  `json:"totalCents"`
	InstallmentCount int    `json:"installmentCount"`
	FirstDueDate     string `json:"firstDueDate"`
}

type PurchaseItemInput struct {
	StockItemID   string `json:"stockItemId"`
	Quantity      int    `json:"quantity"`
	UnitCostCents int64  `json:"unitCostCents"`
}

type PurchaseInstallmentInput struct {
	AmountCents   int64                  `json:"amountCents"`
	DueDate       string                 `json:"dueDate"`
	PlannedMethod entities.PayableMethod `json:"plannedMethod"`
}

type PayableAlertKind string

const (
	PayableAlertOverdue  PayableAlertKind = "overdue"
	PayableAlertToday    PayableAlertKind = "due_today"
	PayableAlertTomorrow PayableAlertKind = "due_tomorrow"
	PayableAlertEarly    PayableAlertKind = "early_payment"
)

type PayableAlert struct {
	InstallmentID string                 `json:"installmentId"`
	Kind          PayableAlertKind       `json:"kind"`
	SupplierName  string                 `json:"supplierName"`
	Number        int                    `json:"number"`
	AmountCents   int64                  `json:"amountCents"`
	DueDate       time.Time              `json:"dueDate"`
	PlannedMethod entities.PayableMethod `json:"plannedMethod"`
}

type PayablePaymentHistory struct {
	Installment  entities.PayableInstallment   `json:"installment"`
	Purchase     *entities.StockPurchase       `json:"purchase,omitempty"`
	Expense      *entities.Expense             `json:"expense,omitempty"`
	Installments []entities.PayableInstallment `json:"installments"`
}

func NewCashService(db *gorm.DB) *CashService { return &CashService{db: db} }
func businessDay(now time.Time) time.Time {
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
}

func (s *CashService) Current(ctx context.Context) (*entities.CashSession, error) {
	var session entities.CashSession
	err := s.db.WithContext(ctx).Preload("Entries", func(db *gorm.DB) *gorm.DB { return db.Order("occurred_at asc, created_at asc") }).Where("status = ?", entities.CashSessionOpen).First(&session).Error
	if err == nil {
		expected, expectedErr := s.cashExpected(s.db.WithContext(ctx), &session)
		if expectedErr != nil {
			return nil, expectedErr
		}
		session.ExpectedCashCents = expected
		return &session, nil
	}
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return nil, err
}

func (s *CashService) ListSessions(ctx context.Context, limit int) ([]entities.CashSession, error) {
	if limit < 1 || limit > 60 {
		limit = 14
	}
	var sessions []entities.CashSession
	err := s.db.WithContext(ctx).Preload("Entries", func(db *gorm.DB) *gorm.DB { return db.Order("occurred_at asc, created_at asc") }).Order("business_date desc").Limit(limit).Find(&sessions).Error
	if err != nil {
		return nil, err
	}
	for i := range sessions {
		expected, expectedErr := s.cashExpected(s.db.WithContext(ctx), &sessions[i])
		if expectedErr != nil {
			return nil, expectedErr
		}
		sessions[i].ExpectedCashCents = expected
	}
	return sessions, nil
}

// Balances returns the lifetime ledger balance for non-cash payment methods.
// Cash is intentionally excluded because its balance belongs to a drawer
// session and is exposed by Current.
func (s *CashService) Balances(ctx context.Context) (*CashBalances, error) {
	var balances CashBalances
	err := s.db.WithContext(ctx).Model(&entities.CashEntry{}).Select(`
		COALESCE(SUM(CASE WHEN payment_method = ? THEN amount_cents ELSE 0 END), 0) AS pix_cents,
		COALESCE(SUM(CASE WHEN payment_method = ? THEN amount_cents ELSE 0 END), 0) AS debit_card_cents,
		COALESCE(SUM(CASE WHEN payment_method = ? THEN amount_cents ELSE 0 END), 0) AS credit_card_cents
	`, entities.PaymentMethodPix, entities.PaymentMethodDebitCard, entities.PaymentMethodCreditCard).Scan(&balances).Error
	if err != nil {
		return nil, err
	}
	return &balances, nil
}

// EnsureReferenceMutable prevents a source record from rewriting a closed day.
func (s *CashService) EnsureReferenceMutable(ctx context.Context, referenceType, referenceID string) error {
	var count int64
	err := s.db.WithContext(ctx).Table("cash_entries").Joins("JOIN cash_sessions ON cash_sessions.id = cash_entries.cash_session_id").Where("cash_entries.reference_type = ? AND cash_entries.reference_id = ? AND cash_sessions.status = ?", referenceType, referenceID, entities.CashSessionClosed).Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return apperrors.ErrConflict
	}
	return nil
}

func (s *CashService) Open(ctx context.Context, input OpenInput) (*entities.CashSession, error) {
	if input.OpeningCashCents < 0 {
		return nil, apperrors.ErrInvalidInput
	}
	now := time.Now()
	returnSession := new(entities.CashSession)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var openCount int64
		if err := tx.Model(&entities.CashSession{}).Where("status = ?", entities.CashSessionOpen).Count(&openCount).Error; err != nil {
			return err
		}
		if openCount > 0 {
			return apperrors.ErrConflict
		}
		day := businessDay(now)
		var sameDay int64
		if err := tx.Model(&entities.CashSession{}).Where("business_date = ?", day).Count(&sameDay).Error; err != nil {
			return err
		}
		if sameDay > 0 {
			return apperrors.ErrConflict
		}
		*returnSession = entities.CashSession{BusinessDate: day, Status: entities.CashSessionOpen, OpeningCashCents: input.OpeningCashCents}
		return tx.Create(returnSession).Error
	})
	return returnSession, err
}

func (s *CashService) Close(ctx context.Context, input CloseInput) (*entities.CashSession, error) {
	if input.ClosingCashCents < 0 {
		return nil, apperrors.ErrInvalidInput
	}
	var result entities.CashSession
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		session, err := s.openForUpdate(tx)
		if err != nil {
			return err
		}
		expected, err := s.cashExpected(tx, session)
		if err != nil {
			return err
		}
		difference := input.ClosingCashCents - expected
		if difference != 0 && strings.TrimSpace(input.ClosingNotes) == "" {
			return apperrors.ErrInvalidInput
		}
		now := time.Now()
		session.Status = entities.CashSessionClosed
		session.ClosingCashCents = &input.ClosingCashCents
		session.CashDifferenceCents = &difference
		session.ClosingNotes = strings.TrimSpace(input.ClosingNotes)
		session.ClosedAt = &now
		if err := tx.Save(session).Error; err != nil {
			return err
		}
		result = *session
		return nil
	})
	return &result, err
}

func (s *CashService) AddAdjustment(ctx context.Context, input AdjustmentInput) (*entities.CashEntry, error) {
	if input.AmountCents <= 0 || strings.TrimSpace(input.Description) == "" || strings.TrimSpace(input.Reason) == "" || !validMethod(input.PaymentMethod) || (input.Direction != entities.CashEntryIn && input.Direction != entities.CashEntryOut) {
		return nil, apperrors.ErrInvalidInput
	}
	entry := &entities.CashEntry{Direction: input.Direction, Kind: entities.CashEntryAdjust, PaymentMethod: input.PaymentMethod, AmountCents: input.AmountCents, Description: strings.TrimSpace(input.Description), Reason: strings.TrimSpace(input.Reason), ReferenceType: "adjustment", ReferenceID: strings.TrimSpace(input.ReferenceID), OccurredAt: time.Now()}
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return s.record(tx, entry) })
	return entry, err
}

func (s *CashService) RecordReceiptPayment(tx *gorm.DB, receipt *entities.Receipt) error {
	return s.record(tx, &entities.CashEntry{Direction: entities.CashEntryIn, Kind: entities.CashEntrySale, PaymentMethod: receipt.PaymentMethod, AmountCents: receipt.PriceCents, Description: "Pagamento de recibo", ReferenceType: "receipt", ReferenceID: receipt.ID, OccurredAt: time.Now()})
}

func (s *CashService) RecordExpense(ctx context.Context, expense *entities.Expense, method entities.PaymentMethod) error {
	if !validMethod(method) {
		return apperrors.ErrInvalidInput
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error { return s.RecordExpenseWithTx(tx, expense, method) })
}
func (s *CashService) RecordExpenseWithTx(tx *gorm.DB, expense *entities.Expense, method entities.PaymentMethod) error {
	if !validMethod(method) {
		return apperrors.ErrInvalidInput
	}
	return s.record(tx, &entities.CashEntry{Direction: entities.CashEntryOut, Kind: entities.CashEntryExpense, PaymentMethod: method, AmountCents: expense.AmountCents, Description: expense.Description, ReferenceType: "expense", ReferenceID: expense.ID, OccurredAt: time.Now()})
}

func (s *CashService) SyncExpenseWithTx(tx *gorm.DB, expense *entities.Expense, method entities.PaymentMethod) error {
	var entries []entities.CashEntry
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("reference_type = ? AND reference_id = ? AND kind = ?", "expense", expense.ID, entities.CashEntryExpense).Find(&entries).Error; err != nil {
		return err
	}
	for _, entry := range entries {
		if err := s.ensureEntryMutable(tx, &entry); err != nil {
			return err
		}
		updates := map[string]any{"amount_cents": -expense.AmountCents, "description": expense.Description, "updated_at": time.Now()}
		if method != "" {
			if !validMethod(method) {
				return apperrors.ErrInvalidInput
			}
			updates["payment_method"] = method
		}
		if err := tx.Model(&entry).Updates(updates).Error; err != nil {
			return err
		}
	}
	return nil
}

func (s *CashService) RemoveExpenseEntriesWithTx(tx *gorm.DB, expenseID string) error {
	var entries []entities.CashEntry
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("reference_type = ? AND reference_id = ? AND kind = ?", "expense", expenseID, entities.CashEntryExpense).Find(&entries).Error; err != nil {
		return err
	}
	for index := range entries {
		if err := s.ensureEntryMutable(tx, &entries[index]); err != nil {
			return err
		}
	}
	if len(entries) == 0 {
		return nil
	}
	return tx.Delete(&entries).Error
}

func (s *CashService) ensureEntryMutable(tx *gorm.DB, entry *entities.CashEntry) error {
	if entry.CashSessionID == nil {
		return nil
	}
	var session entities.CashSession
	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&session, "id = ?", *entry.CashSessionID).Error; err != nil {
		return err
	}
	if session.Status == entities.CashSessionClosed {
		return apperrors.ErrConflict
	}
	return nil
}

func (s *CashService) CreatePurchase(ctx context.Context, input PurchaseInput) (*entities.StockPurchase, error) {
	input.SupplierName = strings.TrimSpace(input.SupplierName)
	if input.SupplierName == "" || len(input.Items) == 0 || len(input.Installments) == 0 {
		return nil, apperrors.ErrInvalidInput
	}

	purchase := &entities.StockPurchase{SupplierName: input.SupplierName, Status: entities.StockPurchaseConfirmed, PurchasedAt: time.Now()}
	seen := make(map[string]struct{}, len(input.Items))
	for _, item := range input.Items {
		if strings.TrimSpace(item.StockItemID) == "" || item.Quantity < 1 || item.UnitCostCents <= 0 {
			return nil, apperrors.ErrInvalidInput
		}
		if _, exists := seen[item.StockItemID]; exists {
			return nil, apperrors.ErrInvalidInput
		}
		seen[item.StockItemID] = struct{}{}
		subtotal := item.UnitCostCents * int64(item.Quantity)
		purchase.TotalCents += subtotal
		purchase.Items = append(purchase.Items, entities.StockPurchaseItem{StockItemID: item.StockItemID, Quantity: item.Quantity, UnitCostCents: item.UnitCostCents, SubtotalCents: subtotal})
	}
	if len(purchase.Items) > 0 {
		purchase.StockItemID = purchase.Items[0].StockItemID
	}

	var installmentTotal int64
	for index, installment := range input.Installments {
		dueDate, err := time.ParseInLocation("2006-01-02", strings.TrimSpace(installment.DueDate), time.Local)
		if err != nil || installment.AmountCents <= 0 || !validPlannedMethod(installment.PlannedMethod) {
			return nil, apperrors.ErrInvalidInput
		}
		installmentTotal += installment.AmountCents
		purchase.Installments = append(purchase.Installments, entities.PayableInstallment{Number: index + 1, AmountCents: installment.AmountCents, DueDate: dueDate, Status: entities.PayablePending, PlannedMethod: installment.PlannedMethod})
	}
	if installmentTotal != purchase.TotalCents {
		return nil, apperrors.ErrInvalidInput
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for index := range purchase.Items {
			purchaseItem := &purchase.Items[index]
			var stock entities.StockItem
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&stock, "id = ? AND active = ?", purchaseItem.StockItemID, true).Error; err != nil {
				return err
			}
			purchaseItem.PreviousQuantity = stock.Quantity
			purchaseItem.PreviousCostCents = stock.CostCents
			purchaseItem.PreviousResalePriceCents = stock.ResalePriceCents
			previousValue := stock.CostCents * int64(stock.Quantity)
			stock.Quantity += purchaseItem.Quantity
			stock.CostCents = (previousValue + purchaseItem.SubtotalCents) / int64(stock.Quantity)
			stock.ResalePriceCents = validation.CalculateMarkupCents(stock.CostCents, stock.MarkupPercent)
			purchaseItem.ResultingCostCents = stock.CostCents
			purchaseItem.ResultingResalePriceCents = stock.ResalePriceCents
			purchaseItem.HasStockSnapshot = true
			if err := tx.Save(&stock).Error; err != nil {
				return err
			}
		}
		return tx.Create(purchase).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetPurchase(ctx, purchase.ID)
}

func (s *CashService) ListPurchases(ctx context.Context) ([]entities.StockPurchase, error) {
	var purchases []entities.StockPurchase
	err := s.db.WithContext(ctx).Preload("Items.StockItem").Preload("Installments", func(db *gorm.DB) *gorm.DB { return db.Order("number asc") }).Order("purchased_at desc, created_at desc").Find(&purchases).Error
	return purchases, err
}

func (s *CashService) GetPurchase(ctx context.Context, id string) (*entities.StockPurchase, error) {
	var purchase entities.StockPurchase
	err := s.db.WithContext(ctx).Preload("Items.StockItem").Preload("Installments", func(db *gorm.DB) *gorm.DB { return db.Order("number asc") }).First(&purchase, "id = ?", id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperrors.ErrNotFound
	}
	return &purchase, err
}

func (s *CashService) CancelPurchase(ctx context.Context, id string) (*entities.StockPurchase, error) {
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var purchase entities.StockPurchase
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("Items").Preload("Installments").First(&purchase, "id = ?", id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return apperrors.ErrNotFound
			}
			return err
		}
		if purchase.Status != entities.StockPurchaseConfirmed {
			return apperrors.ErrConflict
		}
		for _, installment := range purchase.Installments {
			if installment.Status == entities.PayablePaid {
				return apperrors.ErrConflict
			}
		}
		for _, item := range purchase.Items {
			var stock entities.StockItem
			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&stock, "id = ?", item.StockItemID).Error; err != nil {
				return err
			}
			if !item.HasStockSnapshot || stock.Quantity != item.PreviousQuantity+item.Quantity || stock.CostCents != item.ResultingCostCents || stock.ResalePriceCents != item.ResultingResalePriceCents {
				return apperrors.ErrConflict
			}
			stock.Quantity = item.PreviousQuantity
			stock.CostCents = item.PreviousCostCents
			stock.ResalePriceCents = item.PreviousResalePriceCents
			if err := tx.Save(&stock).Error; err != nil {
				return err
			}
		}
		now := time.Now()
		if err := tx.Model(&entities.PayableInstallment{}).Where("stock_purchase_id = ? AND status = ?", purchase.ID, entities.PayablePending).Updates(map[string]any{"status": entities.PayableCancelled, "updated_at": now}).Error; err != nil {
			return err
		}
		return tx.Model(&purchase).Updates(map[string]any{"status": entities.StockPurchaseCancelled, "cancelled_at": now, "updated_at": now}).Error
	})
	if err != nil {
		return nil, err
	}
	return s.GetPurchase(ctx, id)
}

func (s *CashService) PendingInstallments(ctx context.Context) ([]entities.PayableInstallment, error) {
	var rows []entities.PayableInstallment
	err := s.db.WithContext(ctx).Preload("StockPurchase.Items.StockItem").Preload("StockPurchase.Installments").Preload("Expense").Where("status = ?", entities.PayablePending).Order("due_date asc, number asc").Find(&rows).Error
	return rows, err
}

func (s *CashService) GetInstallmentHistory(ctx context.Context, id string) (*PayablePaymentHistory, error) {
	var row entities.PayableInstallment
	err := s.db.WithContext(ctx).
		Preload("StockPurchase.Items.StockItem").
		Preload("StockPurchase.Installments", func(db *gorm.DB) *gorm.DB { return db.Order("number asc") }).
		Preload("Expense.Installments", func(db *gorm.DB) *gorm.DB { return db.Order("number asc") }).
		First(&row, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	history := &PayablePaymentHistory{Installment: row, Purchase: row.StockPurchase, Expense: row.Expense}
	if row.StockPurchase != nil {
		history.Installments = row.StockPurchase.Installments
	} else if row.Expense != nil {
		history.Installments = row.Expense.Installments
	}
	return history, nil
}
func (s *CashService) PayInstallment(ctx context.Context, id string, method entities.PaymentMethod) (*entities.PayableInstallment, error) {
	return s.PayInstallmentAt(ctx, id, method, "")
}

func (s *CashService) PayInstallmentAt(ctx context.Context, id string, method entities.PaymentMethod, paidAtInput string) (*entities.PayableInstallment, error) {
	if !validMethod(method) {
		return nil, apperrors.ErrInvalidInput
	}
	paidAt, err := parsePaidAt(paidAtInput)
	if err != nil {
		return nil, err
	}
	var result entities.PayableInstallment
	err = s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row entities.PayableInstallment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Preload("StockPurchase").Preload("Expense").First(&row, "id = ?", id).Error; err != nil {
			return err
		}
		if row.Status != entities.PayablePending {
			return apperrors.ErrConflict
		}
		kind := entities.CashEntryStock
		description := ""
		if row.StockPurchase != nil {
			description = "Compra de estoque: " + row.StockPurchase.SupplierName
		} else if row.Expense != nil {
			kind = entities.CashEntryExpense
			description = row.Expense.Description
		} else {
			return apperrors.ErrInvalidInput
		}
		entry := &entities.CashEntry{Direction: entities.CashEntryOut, Kind: kind, PaymentMethod: method, AmountCents: row.AmountCents, Description: description, ReferenceType: "payable_installment", ReferenceID: row.ID, OccurredAt: paidAt}
		if err := s.record(tx, entry); err != nil {
			return err
		}
		row.Status = entities.PayablePaid
		row.PaidAt = &paidAt
		row.PaymentMethod = method
		row.CashEntryID = &entry.ID
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		result = row
		return nil
	})
	return &result, err
}

func (s *CashService) RevokeInstallmentPayment(ctx context.Context, id string) (*entities.PayableInstallment, error) {
	var result entities.PayableInstallment
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var row entities.PayableInstallment
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&row, "id = ?", id).Error; err != nil {
			return err
		}
		if !CanRevokePayment(row.Status, row.CashEntryID, row.PaymentRevokedAt) {
			return apperrors.ErrConflict
		}
		if err := tx.Delete(&entities.CashEntry{}, "id = ?", *row.CashEntryID).Error; err != nil {
			return err
		}
		now := time.Now()
		row.Status = entities.PayablePending
		row.PaymentRevokedAt = &now
		row.PaidAt = nil
		row.PaymentMethod = ""
		row.CashEntryID = nil
		if err := tx.Save(&row).Error; err != nil {
			return err
		}
		result = row
		return nil
	})
	return &result, err
}

func parsePaidAt(value string) (time.Time, error) {
	return parsePaidAtAt(value, time.Now())
}

func parsePaidAtAt(value string, now time.Time) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return now, nil
	}
	if parsed, err := time.ParseInLocation("2006-01-02", value, time.Local); err == nil {
		return time.Date(parsed.Year(), parsed.Month(), parsed.Day(), now.Hour(), now.Minute(), now.Second(), now.Nanosecond(), time.Local), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed, nil
	}
	return time.Time{}, apperrors.ErrInvalidInput
}

func (s *CashService) DailyEntries(ctx context.Context, day time.Time) ([]entities.CashEntry, error) {
	start := businessDay(day)
	end := start.AddDate(0, 0, 1)
	var entries []entities.CashEntry
	err := s.db.WithContext(ctx).Where("occurred_at >= ? AND occurred_at < ?", start, end).Order("occurred_at asc, created_at asc").Find(&entries).Error
	return entries, err
}

func (s *CashService) PayableAlerts(ctx context.Context, now time.Time) ([]PayableAlert, error) {
	rows, err := s.PendingInstallments(ctx)
	if err != nil {
		return nil, err
	}
	today := businessDay(now)
	tomorrow := today.AddDate(0, 0, 1)
	windowEnd := today.AddDate(0, 0, 31)
	available := int64(0)
	if session, currentErr := s.Current(ctx); currentErr != nil {
		return nil, currentErr
	} else if session != nil {
		available = session.ExpectedCashCents - session.OpeningCashCents
	}
	alerts := make([]PayableAlert, 0)
	for _, row := range rows {
		due := businessDay(row.DueDate)
		kind := PayableAlertKind("")
		switch {
		case due.Before(today):
			kind = PayableAlertOverdue
		case due.Equal(today):
			kind = PayableAlertToday
		case due.Equal(tomorrow):
			kind = PayableAlertTomorrow
		case due.Before(windowEnd) && row.AmountCents <= available:
			kind = PayableAlertEarly
			available -= row.AmountCents
		}
		if kind != "" {
			name := ""
			if row.StockPurchase != nil {
				name = row.StockPurchase.SupplierName
			} else if row.Expense != nil {
				name = row.Expense.Description
			}
			alerts = append(alerts, PayableAlert{InstallmentID: row.ID, Kind: kind, SupplierName: name, Number: row.Number, AmountCents: row.AmountCents, DueDate: row.DueDate, PlannedMethod: row.PlannedMethod})
		}
	}
	return alerts, nil
}

func (s *CashService) openForUpdate(tx *gorm.DB) (*entities.CashSession, error) {
	var session entities.CashSession
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("status = ?", entities.CashSessionOpen).First(&session).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrConflict
	}
	return &session, err
}
func (s *CashService) record(tx *gorm.DB, entry *entities.CashEntry) error {
	if entry.PaymentMethod == entities.PaymentMethodCash {
		session, err := s.openForUpdate(tx)
		if err != nil {
			return err
		}
		entry.CashSessionID = &session.ID
	} else {
		var session entities.CashSession
		if err := tx.Where("status = ?", entities.CashSessionOpen).First(&session).Error; err == nil {
			entry.CashSessionID = &session.ID
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
	}
	entry.Type = ledgerType(entry.Kind)
	if entry.Direction == entities.CashEntryOut {
		entry.AmountCents = -entry.AmountCents
	}
	return tx.Create(entry).Error
}
func (s *CashService) cashExpected(tx *gorm.DB, session *entities.CashSession) (int64, error) {
	var movement int64
	if err := tx.Model(&entities.CashEntry{}).Where("cash_session_id = ? AND payment_method = ?", session.ID, entities.PaymentMethodCash).Select("COALESCE(SUM(amount_cents), 0)").Scan(&movement).Error; err != nil {
		return 0, err
	}
	return session.OpeningCashCents + movement, nil
}

func ledgerType(kind entities.CashEntryKind) string {
	switch kind {
	case entities.CashEntrySale:
		return "receipt_payment"
	case entities.CashEntryExpense:
		return "expense"
	case entities.CashEntryStock:
		return "stock_purchase"
	default:
		return "adjustment"
	}
}
func validMethod(method entities.PaymentMethod) bool {
	return method == entities.PaymentMethodCash || method == entities.PaymentMethodPix || method == entities.PaymentMethodDebitCard || method == entities.PaymentMethodCreditCard
}

func validPlannedMethod(method entities.PayableMethod) bool {
	return method == entities.PayableMethodBoleto || method == entities.PayableMethodCash || method == entities.PayableMethodPix || method == entities.PayableMethodDebitCard || method == entities.PayableMethodCreditCard
}
