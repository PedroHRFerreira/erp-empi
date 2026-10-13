// Package templates serves the bundled, versioned marketing template catalog.
package templates

import (
	_ "embed"
	"encoding/json"
	"github.com/empi-autocenter/erp-empi/internal/domain/entities"
)

type Defaults struct {
	Title          string `json:"title"`
	Description    string `json:"description"`
	CTA            string `json:"cta"`
	Caption        string `json:"caption"`
	PrimaryColor   string `json:"primaryColor"`
	SecondaryColor string `json:"secondaryColor"`
	Kind           string `json:"kind"`
	Objective      string `json:"objective"`
}

type PhotoTreatment struct {
	Side    string  `json:"side"`
	Color   string  `json:"color"`
	Opacity float64 `json:"opacity"`
}

type Template struct {
	Cutouts        map[string]string         `json:"cutouts,omitempty"`
	PhotoTreatment *PhotoTreatment           `json:"photoTreatment,omitempty"`
	NativeWidth    int                       `json:"nativeWidth"`
	NativeHeight   int                       `json:"nativeHeight"`
	Layers         []entities.MarketingLayer `json:"layers"`
	ID             string                    `json:"id"`
	Name           string                    `json:"name"`
	Category       string                    `json:"category"`
	Layout         string                    `json:"layout"`
	SourceURL      string                    `json:"sourceUrl"`
	PreviewURL     string                    `json:"previewUrl"`
	SourceKind     string                    `json:"sourceKind"`
	Slide          int                       `json:"slide"`
	BackgroundURL  string                    `json:"backgroundUrl"`
	Defaults       Defaults                  `json:"defaults"`
}

//go:embed catalog.json
var catalogJSON []byte

var catalog = loadCatalog()

func loadCatalog() []Template {
	var items []Template
	if err := json.Unmarshal(catalogJSON, &items); err != nil {
		panic("invalid bundled marketing catalog: " + err.Error())
	}
	seen := make(map[string]bool, len(items))
	for _, item := range items {
		if item.ID == "" || len(item.ID) > 16 || seen[item.ID] || item.ID == "bold" || item.ID == "minimal" {
			panic("invalid or duplicated marketing template ID: " + item.ID)
		}
		seen[item.ID] = true
	}
	return items
}

// List returns a copy so callers cannot alter the bundled registry.
func Catalog() []Template {
	data, _ := json.Marshal(catalog)
	var copied []Template
	_ = json.Unmarshal(data, &copied)
	return copied
}

// ValidID also accepts legacy templates to keep saved creations editable.
func ValidID(id string) bool {
	if id == "bold" || id == "minimal" {
		return true
	}
	for _, item := range catalog {
		if item.ID == id {
			return true
		}
	}
	return false
}
