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
	userHandler := NewUserHandler(container.Users, container.Receipts)
	stockHandler := NewStockHandler(container.Stock)
	receiptHandler := NewReceiptHandler(container.Receipts)
	metricsHandler := NewMetricsHandler(container.Metrics)
	expenseHandler := NewExpenseHandler(container.Expenses, container.Financial)
	goalsHandler := NewGoalsHandler(container.Goals)
	cashHandler := NewCashHandler(container.Cash)

	api := e.Group("/api")
	api.POST("/auth/login", authHandler.Login)

	private := api.Group("", authMiddleware(container.Auth))
	private.GET("/auth/me", userHandler.Me)
	private.GET("/users/profile", userHandler.Me)
	private.PUT("/users/profile", userHandler.UpdateProfile)
	private.GET("/users/clients", userHandler.ListClients)
	private.GET("/users/clients/:id/detail", userHandler.ClientDetail)
	private.DELETE("/users/clients/:id", userHandler.ArchiveClient)
	private.GET("/stock", stockHandler.List)
	private.POST("/stock", stockHandler.Create)
	private.PUT("/stock/:id", stockHandler.Update)
	private.DELETE("/stock/:id", stockHandler.Delete)
	private.GET("/receipts", receiptHandler.List)
	private.POST("/receipts", receiptHandler.Create)
	private.GET("/receipts/:id", receiptHandler.Get)
	private.PUT("/receipts/:id", receiptHandler.Update)
	private.POST("/receipts/:id/pay", receiptHandler.MarkPaid)
	private.POST("/receipts/:id/cancel", receiptHandler.Cancel)
	private.POST("/receipts/:id/reopen", receiptHandler.Reopen)
	private.GET("/metrics/summary", metricsHandler.Summary)
	private.GET("/expenses", expenseHandler.List)
	private.POST("/expenses", expenseHandler.Create)
	private.GET("/expenses/:id", expenseHandler.Get)
	private.PUT("/expenses/:id", expenseHandler.Update)
	private.DELETE("/expenses/:id", expenseHandler.Delete)
	private.GET("/financial/summary", expenseHandler.Summary)
	private.GET("/financial/expenses", expenseHandler.Realized)
	private.GET("/cash/current", cashHandler.Current)
	private.GET("/cash/balances", cashHandler.Balances)
	private.GET("/cash/sessions", cashHandler.ListSessions)
	private.POST("/cash/open", cashHandler.Open)
	private.POST("/cash/close", cashHandler.Close)
	private.POST("/cash/adjustments", cashHandler.AddAdjustment)
	private.GET("/cash/payables", cashHandler.PendingInstallments)
	private.GET("/cash/daily-entries", cashHandler.DailyEntries)
	private.GET("/payables", cashHandler.PendingInstallments)
	private.GET("/payables/alerts", cashHandler.PayableAlerts)
	private.POST("/cash/purchases", cashHandler.CreatePurchase)
	private.POST("/cash/payables/:id/pay", cashHandler.PayInstallment)
	private.POST("/payables/:id/pay", cashHandler.PayInstallment)
	private.GET("/stock/purchases", cashHandler.ListPurchases)
	private.POST("/stock/purchases", cashHandler.CreatePurchase)
	private.GET("/stock/purchases/:id", cashHandler.GetPurchase)
	private.DELETE("/stock/purchases/:id", cashHandler.CancelPurchase)
	private.GET("/goals", goalsHandler.Get)
	private.PUT("/goals", goalsHandler.Save)

	return e
}
