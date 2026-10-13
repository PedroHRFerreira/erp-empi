package http

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/empi-autocenter/erp-empi/config"
	"github.com/empi-autocenter/erp-empi/internal/app/dig"
	authservices "github.com/empi-autocenter/erp-empi/internal/domain/auth/services"
	"github.com/empi-autocenter/erp-empi/internal/domain/marketing/templates"
	"github.com/empi-autocenter/erp-empi/internal/shared/security"
	"github.com/labstack/echo/v4"
)

func TestMarketingCatalogAvailableToAdmin(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{AccessSecret: "marketing-test-secret"}}
	server := NewServer(cfg, &dig.Container{Auth: authservices.NewAuthService(cfg, nil)})
	token, err := security.GenerateToken("admin-id", "admin", cfg.JWT.AccessSecret, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", "/api/marketing/templates", nil)
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	server.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("want catalog, got %d: %s", response.Code, response.Body.String())
	}
	var items []templates.Template
	if err := json.Unmarshal(response.Body.Bytes(), &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != len(templates.Catalog()) || len(items) == 0 {
		t.Fatalf("incomplete catalog: %d", len(items))
	}
}

func TestMarketingRoutesRequireAuthentication(t *testing.T) {
	server := NewServer(&config.Config{}, &dig.Container{})
	for _, path := range []string{"/api/marketing/templates", "/api/marketing/brand", "/api/marketing/creations", "/api/marketing/images/example"} {
		response := httptest.NewRecorder()
		server.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 401 {
			t.Errorf("%s: want unauthorized, got %d", path, response.Code)
		}
	}
}

func TestMarketingRoutesRejectClientTokens(t *testing.T) {
	cfg := &config.Config{JWT: config.JWTConfig{AccessSecret: "marketing-test-secret"}}
	server := NewServer(cfg, &dig.Container{Auth: authservices.NewAuthService(cfg, nil)})
	token, err := security.GenerateToken("client-id", "client", cfg.JWT.AccessSecret, time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	for _, route := range []struct{ method, path string }{
		{"GET", "/api/marketing/templates"}, {"GET", "/api/marketing/brand"}, {"PUT", "/api/marketing/brand"},
		{"GET", "/api/marketing/creations"}, {"POST", "/api/marketing/creations"},
		{"PUT", "/api/marketing/creations/example"}, {"DELETE", "/api/marketing/creations/example"},
		{"POST", "/api/marketing/images"}, {"GET", "/api/marketing/images/example"},
	} {
		request := httptest.NewRequest(route.method, route.path, nil)
		request.Header.Set("Authorization", "Bearer "+token)
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		if response.Code != 401 {
			t.Errorf("%s %s: client token accepted (%d)", route.method, route.path, response.Code)
		}
	}
}

func TestMarketingUploadRejectsOversizedAndMissingFiles(t *testing.T) {
	for _, oversized := range []bool{false, true} {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		if oversized {
			part, err := writer.CreateFormFile("file", "photo.png")
			if err != nil {
				t.Fatal(err)
			}
			if _, err = part.Write(make([]byte, 5*1024*1024+1)); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		request := httptest.NewRequest("POST", "/api/marketing/images", &body)
		request.Header.Set("Content-Type", writer.FormDataContentType())
		response := httptest.NewRecorder()
		ctx := echo.New().NewContext(request, response)
		if err := NewMarketingHandler(nil).Upload(ctx); err != nil {
			t.Fatal(err)
		}
		if response.Code != 400 {
			t.Errorf("oversized=%v: want bad request, got %d", oversized, response.Code)
		}
	}
}
