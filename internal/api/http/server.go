package http

import (
	nethttp "net/http"
	"time"

	"github.com/empi-autocenter/erp-empi/config"
	"github.com/empi-autocenter/erp-empi/internal/app/dig"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func NewServer(cfg *config.Config, container *dig.Container) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.Secure())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     []string{cfg.FrontendURL},
		AllowMethods:     []string{nethttp.MethodGet, nethttp.MethodPost, nethttp.MethodPut, nethttp.MethodPatch, nethttp.MethodDelete, nethttp.MethodOptions},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           int((12 * time.Hour).Seconds()),
	}))
	e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(60)))

	e.GET("/health", func(c echo.Context) error {
		return c.JSON(nethttp.StatusOK, map[string]string{"status": "ok"})
	})

	authHandler := NewAuthHandler(container.Auth)
	userHandler := NewUserHandler(container.Users)
	stockHandler := NewStockHandler(container.Stock)
	receiptHandler := NewReceiptHandler(container.Receipts)
	metricsHandler := NewMetricsHandler(container.Metrics)

	api := e.Group("/api")
	api.POST("/auth/login", authHandler.Login)
	api.POST("/auth/refresh", authHandler.Refresh)

	private := api.Group("", authMiddleware(container.Auth))
	private.GET("/auth/me", userHandler.Me)
	private.GET("/users/profile", userHandler.Me)
	private.PUT("/users/profile", userHandler.UpdateProfile)
	private.GET("/users/clients", userHandler.ListClients)
	private.GET("/stock", stockHandler.List)
	private.POST("/stock", stockHandler.Create)
	private.PUT("/stock/:id", stockHandler.Update)
	private.DELETE("/stock/:id", stockHandler.Delete)
	private.GET("/receipts", receiptHandler.List)
	private.POST("/receipts", receiptHandler.Create)
	private.GET("/receipts/:id", receiptHandler.Get)
	private.POST("/receipts/:id/pay", receiptHandler.MarkPaid)
	private.GET("/metrics/summary", metricsHandler.Summary)

	return e
}
