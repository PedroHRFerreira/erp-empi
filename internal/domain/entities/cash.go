package entities

import (
	"time"

	"gorm.io/gorm"
)

type CashSessionStatus string
type CashEntryDirection string
type CashEntryKind string
type PayableInstallmentStatus string
type StockPurchaseStatus string
type PayableMethod string

const (
	CashSessionOpen   CashSessionStatus = "open"
	CashSessionClosed CashSessionStatus = "closed"

	CashEntryIn      CashEntryDirection = "in"
	CashEntryOut     CashEntryDirection = "out"
	CashEntrySale    CashEntryKind      = "receipt_payment"
	CashEntryExpense CashEntryKind      = "expense"
	CashEntryStock   CashEntryKind      = "stock_payment"
	CashEntryAdjust  CashEntryKind      = "adjustment"

	PayablePending   PayableInstallmentStatus = "pending"
	PayablePaid      PayableInstallmentStatus = "paid"
	PayableCancelled PayableInstallmentStatus = "cancelled"

	StockPurchaseConfirmed StockPurchaseStatus = "confirmed"
	StockPurchaseCancelled StockPurchaseStatus = "cancelled"

	PayableMethodBoleto     PayableMethod = "boleto"
	PayableMethodCash       PayableMethod = "cash"
	PayableMethodPix        PayableMethod = "pix"
	PayableMethodDebitCard  PayableMethod = "debit_card"
	PayableMethodCreditCard PayableMethod = "credit_card"
	PayableMethodLegacy     PayableMethod = "legacy"
)

// CashSession is the immutable daily reconciliation record. BusinessDate is
// intentionally separate from timestamps so a late close keeps its original day.
type CashSession struct {
	ID                  string            `json:"id" gorm:"type:char(36);primaryKey"`
	BusinessDate        time.Time         `json:"businessDate" gorm:"uniqueIndex;not null"`
	Status              CashSessionStatus `json:"status" gorm:"size:20;not null;index"`
	OpeningCashCents    int64             `json:"openingCashCents" gorm:"not null;default:0"`
	ClosingCashCents    *int64            `json:"closingCashCents,omitempty"`
	CashDifferenceCents *int64            `json:"cashDifferenceCents,omitempty"`
	ClosingNotes        string            `json:"closingNotes" gorm:"size:700"`
	ClosedAt            *time.Time        `json:"closedAt,omitempty"`
	ExpectedCashCents   int64             `json:"expectedCashCents" gorm:"-"`
	Entries             []CashEntry       `json:"entries,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Timestamps
}

type CashEntry struct {
	ID string `json:"id" gorm:"type:char(36);primaryKey"`
	// Type is retained for the append-only cash ledger that existed before the
	// daily session UI. AmountCents is signed in persistence (+ incoming, - outgoing).
	Type string `json:"type" gorm:"size:30;not null;index"`
	// Nullable for legacy entries that existed before daily cash sessions. The
	// service assigns it for every movement created by this feature.
	CashSessionID *string `json:"cashSessionId,omitempty" gorm:"type:char(36);index"`
	// These are nullable at the database boundary to keep old cash_entries
	// rows readable. New entries are fully validated by CashService.
	Direction     CashEntryDirection `json:"direction" gorm:"size:10"`
	Kind          CashEntryKind      `json:"kind" gorm:"size:40;index"`
	PaymentMethod PaymentMethod      `json:"paymentMethod" gorm:"size:30"`
	AmountCents   int64              `json:"amountCents"`
	Description   string             `json:"description" gorm:"size:180"`
	ReferenceType string             `json:"referenceType" gorm:"size:40"`
	ReferenceID   string             `json:"referenceId" gorm:"type:char(36);index"`
	Reason        string             `json:"reason" gorm:"size:700"`
	OccurredAt    time.Time          `json:"occurredAt" gorm:"index"`
	Timestamps
}

type StockPurchase struct {
	ID string `json:"id" gorm:"type:char(36);primaryKey"`
	// StockItemID is retained for compatibility with purchases created before
	// multi-item purchases. New code uses Items as the source of truth.
	StockItemID  string               `json:"stockItemId,omitempty" gorm:"type:char(36);index"`
	SupplierName string               `json:"supplierName" gorm:"size:140;not null"`
	TotalCents   int64                `json:"totalCents" gorm:"not null"`
	Status       StockPurchaseStatus  `json:"status" gorm:"size:20;not null;default:confirmed;index"`
	PurchasedAt  time.Time            `json:"purchasedAt" gorm:"not null;default:CURRENT_TIMESTAMP;index"`
	CancelledAt  *time.Time           `json:"cancelledAt,omitempty"`
	Items        []StockPurchaseItem  `json:"items" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Installments []PayableInstallment `json:"installments" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Timestamps
}

type StockPurchaseItem struct {
	ID                        string    `json:"id" gorm:"type:char(36);primaryKey"`
	StockPurchaseID           string    `json:"stockPurchaseId" gorm:"type:char(36);not null;index"`
	StockItemID               string    `json:"stockItemId" gorm:"type:char(36);not null;index"`
	StockItem                 StockItem `json:"stockItem" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Quantity                  int       `json:"quantity" gorm:"not null"`
	UnitCostCents             int64     `json:"unitCostCents" gorm:"not null"`
	SubtotalCents             int64     `json:"subtotalCents" gorm:"not null"`
	PreviousQuantity          int       `json:"-" gorm:"not null;default:0"`
	PreviousCostCents         int64     `json:"-" gorm:"not null;default:0"`
	PreviousResalePriceCents  int64     `json:"-" gorm:"not null;default:0"`
	ResultingCostCents        int64     `json:"-" gorm:"not null;default:0"`
	ResultingResalePriceCents int64     `json:"-" gorm:"not null;default:0"`
	HasStockSnapshot          bool      `json:"-" gorm:"not null;default:false"`
	Timestamps
}

type PayableInstallment struct {
	ID               string                   `json:"id" gorm:"type:char(36);primaryKey"`
	StockPurchaseID  *string                  `json:"stockPurchaseId,omitempty" gorm:"type:char(36);index"`
	ExpenseID        *string                  `json:"expenseId,omitempty" gorm:"type:char(36);index"`
	Number           int                      `json:"number" gorm:"not null"`
	AmountCents      int64                    `json:"amountCents" gorm:"not null"`
	DueDate          time.Time                `json:"dueDate" gorm:"not null;index"`
	Status           PayableInstallmentStatus `json:"status" gorm:"size:20;not null;index"`
	PlannedMethod    PayableMethod            `json:"plannedMethod" gorm:"size:30;not null;default:boleto"`
	PaymentMethod    PaymentMethod            `json:"paymentMethod,omitempty" gorm:"size:30"`
	PaidAt           *time.Time               `json:"paidAt,omitempty"`
	PaymentRevokedAt *time.Time               `json:"paymentRevokedAt,omitempty"`
	CashEntryID      *string                  `json:"cashEntryId,omitempty" gorm:"type:char(36);index"`
	StockPurchase    *StockPurchase           `json:"stockPurchase,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Expense          *Expense                 `json:"expense,omitempty" gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Timestamps
}

func (session *CashSession) BeforeCreate(_ *gorm.DB) error { assignID(&session.ID); return nil }
func (entry *CashEntry) BeforeCreate(_ *gorm.DB) error     { assignID(&entry.ID); return nil }
func (purchase *StockPurchase) BeforeCreate(_ *gorm.DB) error {
	assignID(&purchase.ID)
	if purchase.Status == "" {
		purchase.Status = StockPurchaseConfirmed
	}
	if purchase.PurchasedAt.IsZero() {
		purchase.PurchasedAt = time.Now()
	}
	return nil
}
func (item *StockPurchaseItem) BeforeCreate(_ *gorm.DB) error { assignID(&item.ID); return nil }
func (installment *PayableInstallment) BeforeCreate(_ *gorm.DB) error {
	assignID(&installment.ID)
	if installment.Status == "" {
		installment.Status = PayablePending
	}
	if installment.PlannedMethod == "" {
		installment.PlannedMethod = PayableMethodBoleto
	}
	return nil
}
