package components

import (
	"testing"
	"time"
)

func TestConfigureTimezoneUsesSaoPauloBusinessDay(t *testing.T) {
	previous := time.Local
	t.Cleanup(func() { time.Local = previous })

	if err := configureTimezone("America/Sao_Paulo"); err != nil {
		t.Fatalf("configure timezone: %v", err)
	}

	utcNow := time.Date(2026, time.August, 25, 0, 35, 0, 0, time.UTC)
	localNow := utcNow.In(time.Local)
	if localNow.Year() != 2026 || localNow.Month() != time.August || localNow.Day() != 24 {
		t.Fatalf("expected local business date 2026-08-24, got %s", localNow.Format("2006-01-02"))
	}
}

func TestConfigureTimezoneRejectsInvalidName(t *testing.T) {
	if err := configureTimezone("invalid/timezone"); err == nil {
		t.Fatal("expected invalid timezone error")
	}
}
