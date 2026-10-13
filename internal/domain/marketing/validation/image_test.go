package validation

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/jpeg"
	"image/png"
	"testing"

	"github.com/empi-autocenter/erp-empi/internal/shared/apperrors"
)

func TestValidateImage(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 32, 24))
	var pngBytes, jpegBytes bytes.Buffer
	if err := png.Encode(&pngBytes, img); err != nil {
		t.Fatal(err)
	}
	if err := jpeg.Encode(&jpegBytes, img, nil); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		data []byte
		mime string
	}{
		{"png", pngBytes.Bytes(), "image/png"},
		{"jpeg", jpegBytes.Bytes(), "image/jpeg"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mime, err := ValidateImage(tc.data)
			if err != nil || mime != tc.mime {
				t.Fatalf("got %q, %v", mime, err)
			}
		})
	}

	// Rewrite the PNG IHDR with a valid CRC, so dimension rejection happens
	// before pixel decoding rather than merely rejecting a corrupt header.
	huge := append([]byte(nil), pngBytes.Bytes()...)
	binary.BigEndian.PutUint32(huge[16:20], 100_000)
	binary.BigEndian.PutUint32(huge[20:24], 100_000)
	binary.BigEndian.PutUint32(huge[29:33], crc32.ChecksumIEEE(huge[12:29]))
	zero := append([]byte(nil), pngBytes.Bytes()...)
	binary.BigEndian.PutUint32(zero[16:20], 0)
	binary.BigEndian.PutUint32(zero[29:33], crc32.ChecksumIEEE(zero[12:29]))
	for _, tc := range []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"oversized", make([]byte, MaxImageBytes+1)},
		{"unsupported", []byte("GIF89a")},
		{"corrupt png pixels", pngBytes.Bytes()[:len(pngBytes.Bytes())-20]},
		{"corrupt jpeg pixels", jpegBytes.Bytes()[:len(jpegBytes.Bytes())/2]},
		{"excessive dimensions", huge},
		{"zero dimensions", zero},
		{"png signature with jpeg body", append(append([]byte(nil), pngBytes.Bytes()[:8]...), jpegBytes.Bytes()...)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			mime, err := ValidateImage(tc.data)
			if mime != "" || !errors.Is(err, apperrors.ErrInvalidInput) {
				t.Fatalf("got %q, %v", mime, err)
			}
		})
	}
}
