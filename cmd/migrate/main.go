package main

import (
	"log"
	"os"

	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := os.Getenv("DB_WRITE_DSN")
	if dsn == "" {
		log.Fatal("DB_WRITE_DSN is required")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	if err := database.Migrate(db); err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	log.Print("database migration completed")
}
