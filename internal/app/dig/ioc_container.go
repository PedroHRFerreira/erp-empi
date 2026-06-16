package dig

import (
	"github.com/empi-autocenter/erp-empi/config"
	authservices "github.com/empi-autocenter/erp-empi/internal/domain/auth/services"
	metricservices "github.com/empi-autocenter/erp-empi/internal/domain/metrics/services"
	receiptrepos "github.com/empi-autocenter/erp-empi/internal/domain/receipts/repositories"
	receiptservices "github.com/empi-autocenter/erp-empi/internal/domain/receipts/services"
	stockrepos "github.com/empi-autocenter/erp-empi/internal/domain/stock/repositories"
	stockservices "github.com/empi-autocenter/erp-empi/internal/domain/stock/services"
	userrepos "github.com/empi-autocenter/erp-empi/internal/domain/users/repositories"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
	"gorm.io/gorm"
)

type Container struct {
	Auth     *authservices.AuthService
	Users    *userservices.UserService
	Stock    *stockservices.StockService
	Receipts *receiptservices.ReceiptService
	Metrics  *metricservices.MetricsService
}

func NewContainer(cfg *config.Config, db *gorm.DB) (*Container, error) {
	userRepo := userrepos.NewUserRepository(db)
	stockRepo := stockrepos.NewStockRepository(db)
	receiptRepo := receiptrepos.NewReceiptRepository(db)

	users := userservices.NewUserService(userRepo)
	auth := authservices.NewAuthService(cfg, users)
	stock := stockservices.NewStockService(stockRepo)
	receipts := receiptservices.NewReceiptService(receiptRepo, stockRepo, users)
	metrics := metricservices.NewMetricsService(db)

	return &Container{
		Auth:     auth,
		Users:    users,
		Stock:    stock,
		Receipts: receipts,
		Metrics:  metrics,
	}, nil
}
