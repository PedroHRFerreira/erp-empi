package validation

import (
	"bytes"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"

	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
)

const MaxImageBytes = 5 * 1024 * 1024
const maxImagePixels = 20_000_000

// ValidateImage validates the actual image bytes before they are persisted.
// Inspect dimensions before decoding to bound the decoder's memory allocation.
func ValidateImage(data []byte) (string, error) {
	if len(data) == 0 || len(data) > MaxImageBytes {
		return "", apperrors.ErrInvalidInput
	}
	mime := http.DetectContentType(data)
	expectedFormat := ""
	switch mime {
	case "image/png":
		expectedFormat = "png"
	case "image/jpeg":
		expectedFormat = "jpeg"
	default:
		return "", apperrors.ErrInvalidInput
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil || format != expectedFormat || config.Width <= 0 || config.Height <= 0 {
		return "", apperrors.ErrInvalidInput
	}
	if config.Width > maxImagePixels/config.Height {
		return "", apperrors.ErrInvalidInput
	}
	decoded, format, err := image.Decode(bytes.NewReader(data))
	if err != nil || format != expectedFormat || decoded.Bounds().Dx() != config.Width || decoded.Bounds().Dy() != config.Height {
		return "", apperrors.ErrInvalidInput
	}
	return mime, nil
}
