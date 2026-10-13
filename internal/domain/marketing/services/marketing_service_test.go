package services

import (
	"context"
	"errors"
	"fmt"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/marketing/repositories"
	"github.com/empi-autocenter/erp-empi/internal/domain/marketing/templates"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"math"
	"strings"
	"testing"
)

func marketingFixture(t *testing.T) (*MarketingService, *gorm.DB) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&entities.MarketingBrand{}, &entities.MarketingCreation{}, &entities.MarketingImage{}, &entities.StockItem{}); err != nil {
		t.Fatal(err)
	}
	return NewMarketingService(repositories.NewMarketingRepository(db)), db
}
func content() entities.MarketingContent {
	return entities.MarketingContent{Kind: "service", Objective: "service", Template: "bold", Format: "feed", Title: "Alinhamento", PrimaryColor: "#2563eb", SecondaryColor: "#ffffff", BrandName: "Oficina"}
}
func TestCreationSnapshotsCRUD(t *testing.T) {
	s, db := marketingFixture(t)
	ctx := context.Background()
	stock := entities.StockItem{Name: "Pneu", CostCents: 100, ResalePriceCents: 150, Active: true}
	if err := db.Create(&stock).Error; err != nil {
		t.Fatal(err)
	}
	input := content()
	input.Kind = "product"
	input.ProductID = &stock.ID
	input.Title = stock.Name
	price := int64(150)
	input.PriceCents = &price
	created, err := s.Save(ctx, "", input)
	if err != nil {
		t.Fatal(err)
	}
	if err := db.Model(&stock).Updates(map[string]any{"name": "Alterado", "resale_price_cents": 300, "active": false}).Error; err != nil {
		t.Fatal(err)
	}
	items, total, err := s.List(ctx, 24, 0)
	if err != nil || total != 1 || items[0].Title != "Pneu" || *items[0].PriceCents != 150 {
		t.Fatalf("snapshot lost: %+v %v", items, err)
	}
	input.Caption = "Editada"
	if _, err := s.Save(ctx, created.ID, input); err != nil {
		t.Fatalf("existing archived product should remain editable: %v", err)
	}
	duplicated, err := s.Save(ctx, "", input)
	if err == nil || duplicated != nil {
		t.Fatal("new references to archived stock must fail")
	}
	input.Kind = "offer"
	input.ProductID = nil
	if _, err := s.Save(ctx, "", input); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, created.ID); err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, created.ID); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatal(err)
	}
}
func TestValidationAndReferences(t *testing.T) {
	s, _ := marketingFixture(t)
	ctx := context.Background()
	for _, change := range []func(*entities.MarketingContent){func(in *entities.MarketingContent) { in.Template = "arbitrary" }, func(in *entities.MarketingContent) { in.Title = " " }, func(in *entities.MarketingContent) { in.PrimaryColor = "red" }, func(in *entities.MarketingContent) { p := int64(-1); in.PriceCents = &p }, func(in *entities.MarketingContent) { id := "bad"; in.ImageID = &id }, func(in *entities.MarketingContent) { in.Kind = "product" }} {
		in := content()
		change(&in)
		if _, err := s.Save(ctx, "", in); !errors.Is(err, apperrors.ErrInvalidInput) {
			t.Fatalf("expected invalid input, got %v", err)
		}
	}
	in := content()
	id := "7b46d578-2e67-4514-8ca2-7d5553a11568"
	in.ImageID = &id
	if _, err := s.Save(ctx, "", in); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatal(err)
	}
	if _, err := s.Upload(ctx, []byte("not an image")); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatal(err)
	}
}
func TestBrandSharedUpsert(t *testing.T) {
	s, _ := marketingFixture(t)
	ctx := context.Background()
	b, err := s.Brand(ctx)
	if err != nil || b.PrimaryColor == "" {
		t.Fatal(err)
	}
	b.Name = "Minha oficina"
	if _, err := s.SaveBrand(ctx, *b); err != nil {
		t.Fatal(err)
	}
	b.Name = "Nova oficina"
	if _, err := s.SaveBrand(ctx, *b); err != nil {
		t.Fatal(err)
	}
	read, err := s.Brand(ctx)
	if err != nil || read.Name != "Nova oficina" {
		t.Fatalf("%+v %v", read, err)
	}
}

func TestDeletedCreationCannotBeRecreatedByStaleUpdate(t *testing.T) {
	s, db := marketingFixture(t)
	ctx := context.Background()
	item, err := s.Save(ctx, "", content())
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, item.ID); err != nil {
		t.Fatal(err)
	}
	item.Caption = "Edição em outra aba"
	if err := repositories.NewMarketingRepository(db).Update(ctx, item); !errors.Is(err, apperrors.ErrNotFound) {
		t.Fatalf("stale update should report not found: %v", err)
	}
	_, total, err := s.List(ctx, 24, 0)
	if err != nil || total != 0 {
		t.Fatalf("deleted creation was recreated: total=%d err=%v", total, err)
	}
}

func TestUpdateClearsOptionalContent(t *testing.T) {
	s, _ := marketingFixture(t)
	ctx := context.Background()
	in := content()
	price := int64(1990)
	in.PriceCents = &price
	in.Description = "Descrição inicial"
	in.CTA = "Agende agora"
	in.Caption = "Legenda inicial"
	item, err := s.Save(ctx, "", in)
	if err != nil {
		t.Fatal(err)
	}
	in.PriceCents = nil
	in.Description, in.CTA, in.Caption = "", "", ""
	if _, err := s.Save(ctx, item.ID, in); err != nil {
		t.Fatal(err)
	}
	items, _, err := s.List(ctx, 24, 0)
	if err != nil || len(items) != 1 || items[0].PriceCents != nil || items[0].Description != "" || items[0].CTA != "" || items[0].Caption != "" {
		t.Fatalf("optional values were not cleared: %+v %v", items, err)
	}
}

func TestCatalogTemplateSquarePersistence(t *testing.T) {
	s, _ := marketingFixture(t)
	catalog := templates.Catalog()
	if len(catalog) == 0 {
		t.Fatal("expected bundled templates")
	}
	in := content()
	in.Template = catalog[0].ID
	in.Format = "square"
	saved, err := s.Save(context.Background(), "", in)
	if err != nil {
		t.Fatal(err)
	}
	items, _, err := s.List(context.Background(), 24, 0)
	if err != nil || len(items) != 1 || items[0].Template != in.Template || items[0].Format != "square" {
		t.Fatalf("template or format lost: %+v %v", items, err)
	}
	in.Caption = "Nova legenda"
	if _, err := s.Save(context.Background(), saved.ID, in); err != nil {
		t.Fatal(err)
	}
	in.Template = "ig-unknown"
	if err := ValidateContent(in); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("unknown catalog ID accepted: %v", err)
	}
	in.Template = catalog[0].ID
	in.Format = "other"
	if err := ValidateContent(in); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("unknown format accepted: %v", err)
	}
}

func editableLayer() entities.MarketingLayer {
	return entities.MarketingLayer{ID: "headline", Type: "text", X: 0.1, Y: 0.1, Width: 0.7, Height: 0.2, Text: "Revisão", FontFamily: "Inter", FontSize: 48, Color: "#ffffff", FontWeight: 700, Align: "left", Visible: true}
}

func TestEditableLayersPersistDuplicateAndClear(t *testing.T) {
	s, _ := marketingFixture(t)
	ctx := context.Background()
	in := content()
	in.Format = "original"
	in.Layers = []entities.MarketingLayer{editableLayer()}
	item, err := s.Save(ctx, "", in)
	if err != nil {
		t.Fatal(err)
	}
	items, _, err := s.List(ctx, 24, 0)
	if err != nil || len(items) != 1 || len(items[0].Layers) != 1 || items[0].Layers[0].Text != "Revisão" {
		t.Fatalf("layers lost: %+v %v", items, err)
	}
	clone, err := s.Save(ctx, "", items[0].MarketingContent)
	if err != nil || clone.ID == item.ID || len(clone.Layers) != 1 {
		t.Fatalf("duplicate failed: %+v %v", clone, err)
	}
	in.Layers = []entities.MarketingLayer{}
	if _, err := s.Save(ctx, item.ID, in); err != nil {
		t.Fatal(err)
	}
	read, err := s.repo.Find(ctx, item.ID)
	if err != nil || len(read.Layers) != 0 {
		t.Fatalf("layers not cleared: %+v %v", read, err)
	}
}

func TestEditableLayerValidation(t *testing.T) {
	for _, change := range []func(*entities.MarketingLayer){
		func(l *entities.MarketingLayer) { l.ID = "" }, func(l *entities.MarketingLayer) { l.Type = "script" },
		func(l *entities.MarketingLayer) { l.X = math.NaN() }, func(l *entities.MarketingLayer) { l.Y = math.Inf(1) },
		func(l *entities.MarketingLayer) { l.Width = 0 }, func(l *entities.MarketingLayer) { l.Height = 1 },
		func(l *entities.MarketingLayer) { l.FontFamily = "unknown" }, func(l *entities.MarketingLayer) { l.FontSize = 401 },
		func(l *entities.MarketingLayer) { l.FontWeight = 450 }, func(l *entities.MarketingLayer) { l.Align = "justify" },
		func(l *entities.MarketingLayer) { l.Color = "red" }, func(l *entities.MarketingLayer) { l.BackgroundColor = "red" },
		func(l *entities.MarketingLayer) { l.Text = strings.Repeat("a", 501) }, func(l *entities.MarketingLayer) { id := "bad"; l.ImageID = &id },
	} {
		layer := editableLayer()
		change(&layer)
		in := content()
		in.Layers = []entities.MarketingLayer{layer}
		if err := ValidateContent(in); !errors.Is(err, apperrors.ErrInvalidInput) {
			t.Fatalf("invalid layer accepted: %+v %v", layer, err)
		}
	}
	in := content()
	in.Layers = []entities.MarketingLayer{editableLayer(), editableLayer()}
	if err := ValidateContent(in); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatal("duplicate ids accepted")
	}
	in.Layers = make([]entities.MarketingLayer, 61)
	if err := ValidateContent(in); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatal("too many layers accepted")
	}
	s, _ := marketingFixture(t)
	layer := editableLayer()
	id := "7b46d578-2e67-4514-8ca2-7d5553a11568"
	layer.ImageID = &id
	in.Layers = []entities.MarketingLayer{layer}
	if _, err := s.Save(context.Background(), "", in); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatal("missing layer image accepted")
	}
}

func TestEditableImageLayerLimit(t *testing.T) {
	in := content()
	for i := 0; i < 12; i++ {
		in.Layers = append(in.Layers, entities.MarketingLayer{ID: fmt.Sprintf("photo-%d", i), Type: "image", Width: 0.5, Height: 0.5, Fit: "cover", Visible: true})
	}
	if err := ValidateContent(in); err != nil {
		t.Fatalf("twelve image layers should be supported: %v", err)
	}
	in.Layers = append(in.Layers, entities.MarketingLayer{ID: "photo-13", Type: "image", Width: 0.5, Height: 0.5, Fit: "contain", Visible: true})
	if err := ValidateContent(in); !errors.Is(err, apperrors.ErrInvalidInput) {
		t.Fatalf("thirteen image layers must be rejected: %v", err)
	}
}

func TestEditableLayerRotationValidation(t *testing.T) {
	in := content()
	for _, rotation := range []float64{-30, 0, 30} {
		layer := editableLayer()
		layer.Rotation = rotation
		in.Layers = []entities.MarketingLayer{layer}
		if err := ValidateContent(in); err != nil {
			t.Fatalf("valid rotation %v rejected: %v", rotation, err)
		}
	}
	for _, rotation := range []float64{-30.01, 30.01, math.NaN(), math.Inf(1)} {
		layer := editableLayer()
		layer.Rotation = rotation
		in.Layers = []entities.MarketingLayer{layer}
		if err := ValidateContent(in); !errors.Is(err, apperrors.ErrInvalidInput) {
			t.Fatalf("invalid rotation %v accepted: %v", rotation, err)
		}
	}
}

func TestEditableLayerSupportedFonts(t *testing.T) {
	for _, family := range []string{"Inter", "Roboto Condensed", "Oswald", "Bebas Neue", "Anton", "Dancing Script", "Knewave", "Kalam"} {
		in := content()
		layer := editableLayer()
		layer.FontFamily = family
		in.Layers = []entities.MarketingLayer{layer}
		if err := ValidateContent(in); err != nil {
			t.Fatalf("supported font %q rejected: %v", family, err)
		}
	}
}

func TestBundledLayersCanBeSaved(t *testing.T) {
	for _, template := range templates.Catalog() {
		in := content()
		in.Template = template.ID
		in.Format = "original"
		in.Layers = template.Layers
		if err := ValidateContent(in); err != nil {
			t.Errorf("%s contains unsavable layers: %v", template.ID, err)
		}
	}
}
