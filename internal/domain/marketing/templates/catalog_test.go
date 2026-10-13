package templates

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestCatalogCoverageAndEditableDefaults(t *testing.T) {
	var items []Template
	if err := json.Unmarshal(catalogJSON, &items); err != nil {
		t.Fatal(err)
	}
	if len(items) != 54 {
		t.Fatalf("got %d templates; want all 54 slides and covers", len(items))
	}
	ids := map[string]bool{}
	sources := map[string][]int{}
	carousel := map[string][]int{}
	colors := regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
	layouts := map[string]bool{"institutional": true, "photo-left": true, "photo-top": true, "checklist": true, "promotion": true, "technical": true, "brand": true}
	for _, item := range items {
		if item.ID == "" || len(item.ID) > 16 || ids[item.ID] {
			t.Errorf("invalid/duplicate ID %q", item.ID)
		}
		ids[item.ID] = true
		source, err := url.Parse(item.SourceURL)
		if err != nil || source.Scheme != "https" || source.Hostname() != "www.instagram.com" {
			t.Errorf("%s invalid source URL", item.ID)
		}
		sources[item.SourceURL] = append(sources[item.SourceURL], item.Slide)
		if item.SourceKind == "carousel" {
			carousel[item.SourceURL] = append(carousel[item.SourceURL], item.Slide)
		}
		if item.SourceKind != "post" && item.SourceKind != "carousel" && item.SourceKind != "reel" {
			t.Errorf("%s invalid source kind", item.ID)
		}
		if item.Slide < 1 || item.Name == "" || item.Category == "" || !layouts[item.Layout] {
			t.Errorf("%s incomplete descriptor", item.ID)
		}
		if utf8.RuneCountInString(item.Defaults.Title) > 140 || item.Defaults.Title == "" || utf8.RuneCountInString(item.Defaults.Description) > 500 || utf8.RuneCountInString(item.Defaults.Caption) > 2200 || utf8.RuneCountInString(item.Defaults.CTA) > 100 {
			t.Errorf("%s defaults exceed field limits", item.ID)
		}
		if !colors.MatchString(item.Defaults.PrimaryColor) || !colors.MatchString(item.Defaults.SecondaryColor) {
			t.Errorf("%s invalid palette", item.ID)
		}
		if item.Defaults.Kind != "service" && item.Defaults.Kind != "offer" {
			t.Errorf("%s invalid kind", item.ID)
		}
		if item.Defaults.Objective != "promotion" && item.Defaults.Objective != "service" && item.Defaults.Objective != "launch" {
			t.Errorf("%s invalid objective", item.ID)
		}
		for _, asset := range []string{item.PreviewURL, item.BackgroundURL} {
			if !strings.HasPrefix(asset, "/marketing/") || strings.Contains(asset, "..") || strings.ContainsAny(asset, "?\\#") {
				t.Errorf("%s unsafe local asset %q", item.ID, asset)
				continue
			}
			path := filepath.Join("..", "..", "..", "..", "frontend", "public", strings.TrimPrefix(asset, "/"))
			info, err := os.Stat(path)
			if err != nil || !info.Mode().IsRegular() {
				t.Errorf("%s missing asset %s", item.ID, asset)
			}
		}
	}
	if len(sources) != 46 {
		t.Errorf("got %d source publications; want 46", len(sources))
	}
	counts := []int{}
	for source, slides := range carousel {
		sort.Ints(slides)
		counts = append(counts, len(slides))
		for i, slide := range slides {
			if slide != i+1 {
				t.Errorf("%s incomplete carousel: %v", source, slides)
				break
			}
		}
	}
	sort.Ints(counts)
	if len(counts) != 3 || counts[0] != 3 || counts[1] != 3 || counts[2] != 5 {
		t.Errorf("carousel coverage = %v; want [3 3 5]", counts)
	}
}

func TestLegacyAndUnknownTemplateIDs(t *testing.T) {
	if !ValidID("bold") || !ValidID("minimal") {
		t.Fatal("legacy saved creations must remain editable")
	}
	if ValidID("") || ValidID("unknown") {
		t.Fatal("unknown templates must be rejected")
	}
}

func TestCatalogLayersAreIndependent(t *testing.T) {
	first := Catalog()
	original := first[0].Layers[0].Text
	first[0].Layers[0].Text = "Changed"
	if Catalog()[0].Layers[0].Text != original {
		t.Fatal("caller can mutate bundled layers")
	}
}
