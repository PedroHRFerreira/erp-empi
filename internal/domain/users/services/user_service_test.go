package services_test

import (
	"context"
	"testing"

	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/users/repositories"
	userservices "github.com/empi-autocenter/erp-empi/internal/domain/users/services"
	"github.com/empi-autocenter/erp-empi/internal/infra/database"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUpsertClientMatchesByCPFAndName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	userService := userservices.NewUserService(repositories.NewUserRepository(db))

	first, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Cliente Teste",
		CPF:   "52998224725",
		Phone: "33999990000",
		Email: "antigo@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Cliente Teste",
		CPF:   "52998224725",
		Phone: "33888880000",
		Email: "novo@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != first.ID {
		t.Fatalf("expected same client for matching name and cpf")
	}
	if updated.Phone != "33888880000" || updated.Email != "novo@example.com" {
		t.Fatalf("expected matching client contact data to be updated")
	}

	other, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Outro Cliente",
		CPF:   "52998224725",
		Phone: "33777770000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if other.ID == first.ID {
		t.Fatalf("expected a new client when the cpf matches but the name differs")
	}

	clients, total, err := userService.ListClients(ctx, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 2 || len(clients) != 2 {
		t.Fatalf("expected 2 clients, got total %d and len %d", total, len(clients))
	}
}

func TestUpsertClientWithoutCPFMatchesByPhoneAndName(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	userService := userservices.NewUserService(repositories.NewUserRepository(db))

	first, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Cliente Sem CPF",
		Phone: "33999990000",
		Email: "antigo@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Cliente Sem CPF",
		Phone: "33999990000",
		Email: "novo@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != first.ID {
		t.Fatalf("expected same client for matching name and phone without cpf")
	}
	if updated.Email != "novo@example.com" {
		t.Fatalf("expected matching client data to be updated")
	}

	other, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Outro Cliente",
		Phone: "33999990000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if other.ID == first.ID {
		t.Fatalf("expected a new client when the phone matches but the name differs")
	}
}

func TestUpsertClientRequiresPhoneWhenCPFIsEmpty(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	userService := userservices.NewUserService(repositories.NewUserRepository(db))
	if _, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{Name: "Cliente Sem Identificador"}); err == nil {
		t.Fatal("expected an error when cpf and phone are empty")
	}
}

func TestArchiveClientHidesAndAnonymizesClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	userService := userservices.NewUserService(repositories.NewUserRepository(db))

	client, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Cliente Para Remover",
		Phone: "33999990000",
		Email: "cliente@example.com",
	})
	if err != nil {
		t.Fatal(err)
	}

	archived, err := userService.ArchiveClient(ctx, client.ID)
	if err != nil {
		t.Fatal(err)
	}
	if archived.Name != "Cliente removido" || archived.Phone != "" || archived.Email != "" || archived.ArchivedAt == nil {
		t.Fatalf("expected anonymized archived client, got %+v", archived)
	}

	clients, total, err := userService.ListClients(ctx, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 0 || len(clients) != 0 {
		t.Fatalf("expected archived client to be hidden, got total %d and len %d", total, len(clients))
	}
}

func TestUpdateProfileSavesCardFees(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := database.AutoMigrate(db); err != nil {
		t.Fatal(err)
	}

	repo := repositories.NewUserRepository(db)
	userService := userservices.NewUserService(repo)
	admin := &entities.User{
		Name:          "Admin",
		CPF:           "52998224725",
		Type:          entities.UserTypeAdmin,
		MarkupPercent: 10,
	}
	if err := repo.Create(ctx, admin); err != nil {
		t.Fatal(err)
	}

	updated, err := userService.UpdateProfile(ctx, admin.ID, userservices.UpdateProfileInput{
		Name:                  "Admin Atualizado",
		CPF:                   "52998224725",
		MarkupPercent:         12,
		MachineFeePercent:     4.5,
		InstallmentFeePercent: 8.25,
	})
	if err != nil {
		t.Fatal(err)
	}

	if updated.MachineFeePercent != 4.5 {
		t.Fatalf("expected machine fee 4.5, got %.2f", updated.MachineFeePercent)
	}
	if updated.InstallmentFeePercent != 8.25 {
		t.Fatalf("expected installment fee 8.25, got %.2f", updated.InstallmentFeePercent)
	}
}
