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

func TestUpsertClientMatchesByPhoneAndName(t *testing.T) {
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
		Phone: "33999990000",
		CPF:   "52998224725",
		Email: "antigo@example.com",
		Notes: "primeiro",
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Cliente Teste",
		Phone: "33999990000",
		CPF:   "12345678909",
		Email: "novo@example.com",
		Notes: "atualizado",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != first.ID {
		t.Fatalf("expected same client for matching name and phone")
	}
	if updated.CPF != "" || updated.Email != "" {
		t.Fatalf("expected client cpf and email to be ignored, got cpf %q and email %q", updated.CPF, updated.Email)
	}
	if updated.Notes != "atualizado" {
		t.Fatalf("expected matching client notes to be updated")
	}

	samePhoneDifferentName, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Outro Cliente",
		Phone: "33999990000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if samePhoneDifferentName.ID == first.ID {
		t.Fatalf("expected a new client when the phone matches but the name differs")
	}

	sameNameDifferentPhone, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Cliente Teste",
		Phone: "33888880000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if sameNameDifferentPhone.ID == first.ID {
		t.Fatalf("expected a new client when the name matches but the phone differs")
	}

	clients, total, err := userService.ListClients(ctx, 10, 0)
	if err != nil {
		t.Fatal(err)
	}
	if total != 3 || len(clients) != 3 {
		t.Fatalf("expected 3 clients, got total %d and len %d", total, len(clients))
	}
}

func TestUpsertClientMatchesPhoneWithFormatting(t *testing.T) {
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
		Name:  "Cliente Formatado",
		Phone: "(33) 99999-0000",
	})
	if err != nil {
		t.Fatal(err)
	}

	updated, err := userService.UpsertClient(ctx, userservices.UpsertClientInput{
		Name:  "Cliente Formatado",
		Phone: "33999990000",
	})
	if err != nil {
		t.Fatal(err)
	}
	if updated.ID != first.ID {
		t.Fatalf("expected same client for matching normalized phone and name")
	}
}

func TestUpsertClientRequiresPhone(t *testing.T) {
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
