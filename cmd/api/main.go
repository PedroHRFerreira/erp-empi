package main

import (
	"log/slog"
	"os"

	"github.com/empi-autocenter/erp-empi/cmd/api/components"
)

func main() {
	if err := components.Run(); err != nil {
		slog.Error("api stopped", "error", err)
		os.Exit(1)
	}
}
