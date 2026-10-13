package services

import (
	"context"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
	"github.com/empi-autocenter/erp-empi/internal/domain/marketing/repositories"
	"github.com/empi-autocenter/erp-empi/internal/domain/marketing/templates"
	imagevalidation "github.com/empi-autocenter/erp-empi/internal/domain/marketing/validation"
	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
	"github.com/google/uuid"
	"math"
	"regexp"
	"strings"
	"unicode/utf8"
)

type MarketingService struct {
	repo *repositories.MarketingRepository
}

func NewMarketingService(repo *repositories.MarketingRepository) *MarketingService {
	return &MarketingService{repo: repo}
}

var colorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func validText(s string, max int) bool {
	return utf8.ValidString(s) && utf8.RuneCountInString(s) <= max
}
func validID(id *string) bool {
	if id == nil {
		return true
	}
	_, err := uuid.Parse(*id)
	return err == nil
}
func ValidateContent(in entities.MarketingContent) error {
	if !oneOf(in.Kind, "product", "service", "offer") || !oneOf(in.Objective, "promotion", "service", "launch") || !templates.ValidID(in.Template) || !oneOf(in.Format, "feed", "story", "square", "original") {
		return apperrors.ErrInvalidInput
	}
	if strings.TrimSpace(in.Title) == "" || !validText(in.Title, 140) || !validText(in.Description, 500) || !validText(in.CTA, 100) || !validText(in.Caption, 2200) || !validText(in.BrandName, 140) || !validText(in.BrandContact, 200) {
		return apperrors.ErrInvalidInput
	}
	if !colorPattern.MatchString(in.PrimaryColor) || !colorPattern.MatchString(in.SecondaryColor) || in.PriceCents != nil && (*in.PriceCents < 0 || *in.PriceCents > 99999999999) {
		return apperrors.ErrInvalidInput
	}
	if !validID(in.ImageID) || !validID(in.LogoImageID) || !validID(in.ProductID) || in.Kind == "product" && in.ProductID == nil || in.Kind != "product" && in.ProductID != nil {
		return apperrors.ErrInvalidInput
	}
	return validateLayers(in.Layers)
}

func validateLayers(layers []entities.MarketingLayer) error {
	if len(layers) > 60 {
		return apperrors.ErrInvalidInput
	}
	seen := make(map[string]bool, len(layers))
	total := 0
	images := 0
	for _, layer := range layers {
		if strings.TrimSpace(layer.ID) == "" || !validText(layer.ID, 80) || seen[layer.ID] || !oneOf(layer.Type, "text", "image") {
			return apperrors.ErrInvalidInput
		}
		seen[layer.ID] = true
		for _, value := range []float64{layer.X, layer.Y, layer.Width, layer.Height, layer.FontSize, layer.Rotation} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				return apperrors.ErrInvalidInput
			}
		}
		if layer.Rotation < -30 || layer.Rotation > 30 || layer.X < 0 || layer.Y < 0 || layer.Width <= 0 || layer.Height <= 0 || layer.X+layer.Width > 1+1e-9 || layer.Y+layer.Height > 1+1e-9 || !validID(layer.ImageID) || !validText(layer.Text, 500) {
			return apperrors.ErrInvalidInput
		}
		total += utf8.RuneCountInString(layer.Text)
		if total > 12000 || layer.BackgroundColor != "" && !colorPattern.MatchString(layer.BackgroundColor) {
			return apperrors.ErrInvalidInput
		}
		if layer.Type == "text" {
			if !oneOf(layer.FontFamily, "Inter", "Roboto Condensed", "Oswald", "Bebas Neue", "Anton", "Dancing Script", "Knewave", "Kalam") || layer.FontSize < 8 || layer.FontSize > 400 || !colorPattern.MatchString(layer.Color) || !oneOf(layer.Align, "left", "center", "right") || layer.FontWeight < 400 || layer.FontWeight > 900 || layer.FontWeight%100 != 0 {
				return apperrors.ErrInvalidInput
			}
		} else {
			images++
			if images > 12 || !oneOf(layer.Fit, "cover", "contain") {
				return apperrors.ErrInvalidInput
			}
		}
	}
	return nil
}
func oneOf(s string, values ...string) bool {
	for _, v := range values {
		if s == v {
			return true
		}
	}
	return false
}
func (s *MarketingService) Brand(ctx context.Context) (*entities.MarketingBrand, error) {
	return s.repo.Brand(ctx)
}
func (s *MarketingService) SaveBrand(ctx context.Context, in entities.MarketingBrand) (*entities.MarketingBrand, error) {
	if !validText(in.Name, 140) || !validText(in.Contact, 200) || !colorPattern.MatchString(in.PrimaryColor) || !colorPattern.MatchString(in.SecondaryColor) || !validID(in.LogoImageID) {
		return nil, apperrors.ErrInvalidInput
	}
	if in.LogoImageID != nil {
		if err := s.repo.HasImage(ctx, *in.LogoImageID); err != nil {
			return nil, err
		}
	}
	in.Name = strings.TrimSpace(in.Name)
	in.Contact = strings.TrimSpace(in.Contact)
	return &in, s.repo.SaveBrand(ctx, &in)
}
func (s *MarketingService) List(ctx context.Context, limit, offset int) ([]entities.MarketingCreation, int64, error) {
	return s.repo.List(ctx, limit, offset)
}
func (s *MarketingService) Save(ctx context.Context, id string, in entities.MarketingContent) (*entities.MarketingCreation, error) {
	if err := ValidateContent(in); err != nil {
		return nil, err
	}
	item := &entities.MarketingCreation{}
	if id != "" {
		var err error
		item, err = s.repo.Find(ctx, id)
		if err != nil {
			return nil, err
		}
	}
	for _, imageID := range []*string{in.ImageID, in.LogoImageID} {
		if imageID != nil {
			if err := s.repo.HasImage(ctx, *imageID); err != nil {
				return nil, err
			}
		}
	}
	for _, layer := range in.Layers {
		if layer.ImageID != nil {
			if err := s.repo.HasImage(ctx, *layer.ImageID); err != nil {
				return nil, err
			}
		}
	}
	if in.ProductID != nil {
		active := item.ProductID == nil || *item.ProductID != *in.ProductID
		if err := s.repo.HasProduct(ctx, *in.ProductID, active); err != nil {
			return nil, err
		}
	}
	in.Title = strings.TrimSpace(in.Title)
	item.MarketingContent = in
	if id == "" {
		return item, s.repo.Create(ctx, item)
	}
	return item, s.repo.Update(ctx, item)
}
func (s *MarketingService) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
func (s *MarketingService) Upload(ctx context.Context, data []byte) (*entities.MarketingImage, error) {
	mime, err := imagevalidation.ValidateImage(data)
	if err != nil {
		return nil, apperrors.ErrInvalidInput
	}
	image := &entities.MarketingImage{MimeType: mime, Data: data}
	return image, s.repo.SaveImage(ctx, image)
}
func (s *MarketingService) Image(ctx context.Context, id string) (*entities.MarketingImage, error) {
	return s.repo.Image(ctx, id)
}
