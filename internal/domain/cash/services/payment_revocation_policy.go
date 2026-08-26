package services

import (
	"strings"
	"time"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
)

// CanRevokePayment reports whether a payable payment is eligible for its one
// permitted revocation. The caller remains responsible for applying the state
// transition atomically and persisting revokedAt before returning.
func CanRevokePayment(status entities.PayableInstallmentStatus, cashEntryID *string, revokedAt *time.Time) bool {
	return status == entities.PayablePaid &&
		cashEntryID != nil && strings.TrimSpace(*cashEntryID) != "" &&
		revokedAt == nil
}
