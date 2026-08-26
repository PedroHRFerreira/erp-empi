package services

import (
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
)

func TestCanRevokePayment(t *testing.T) {
	cashEntryID := "cash-entry-id"
	emptyCashEntryID := "  "
	revokedAt := time.Now()

	tests := []struct {
		name        string
		status      entities.PayableInstallmentStatus
		cashEntryID *string
		revokedAt   *time.Time
		want        bool
	}{
		{name: "paid payment with cash entry and no prior revocation", status: entities.PayablePaid, cashEntryID: &cashEntryID, want: true},
		{name: "pending payment", status: entities.PayablePending, cashEntryID: &cashEntryID, want: false},
		{name: "cancelled payment", status: entities.PayableCancelled, cashEntryID: &cashEntryID, want: false},
		{name: "paid payment without cash entry", status: entities.PayablePaid, want: false},
		{name: "paid payment with blank cash entry", status: entities.PayablePaid, cashEntryID: &emptyCashEntryID, want: false},
		{name: "paid payment already revoked once", status: entities.PayablePaid, cashEntryID: &cashEntryID, revokedAt: &revokedAt, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanRevokePayment(tt.status, tt.cashEntryID, tt.revokedAt); got != tt.want {
				t.Fatalf("CanRevokePayment() = %v, want %v", got, tt.want)
			}
		})
	}
}
