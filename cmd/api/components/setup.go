package components

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/empi-autocenter/erp-empi/config"
	httpapi "github.com/empi-autocenter/erp-empi/internal/api/http"
	"github.com/empi-autocenter/erp-empi/internal/app/dig"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	dbClient, err := database.NewPostgresClient(cfg.Database.PostgresDSN)
	if err != nil {
		return err
	}

	container, err := dig.NewContainer(cfg, dbClient.DB)
	if err != nil {
		return err
	}

	if err := container.Users.SeedAdmin(context.Background(), cfg.Admin); err != nil {
		return err
	}

	server := httpapi.NewServer(cfg, container)
	go func() {
		addr := ":" + cfg.APIPort
		slog.Info("api listening", "addr", addr)
		if err := server.Start(addr); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("http server error", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return server.Shutdown(ctx)
}
