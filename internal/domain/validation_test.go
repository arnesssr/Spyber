// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "testing"

func TestNormalizeCountry(t *testing.T) {
	got, err := NormalizeCountry("gb")
	if err != nil {
		t.Fatalf("expected country to normalize: %v", err)
	}
	if got != "GB" {
		t.Fatalf("expected GB, got %s", got)
	}
	if _, err := NormalizeCountry("gbr"); err == nil {
		t.Fatal("expected invalid country error")
	}
}

func TestNormalizeWebsite(t *testing.T) {
	raw, host, err := NormalizeWebsite("shop.example/products")
	if err != nil {
		t.Fatalf("expected website to normalize: %v", err)
	}
	if host != "shop.example" {
		t.Fatalf("expected host shop.example, got %s", host)
	}
	if raw != "https://shop.example/products" {
		t.Fatalf("unexpected normalized url: %s", raw)
	}
	if _, _, err := NormalizeWebsite("file:///etc/passwd"); err == nil {
		t.Fatal("expected invalid URL scheme")
	}
}

func TestClassifyContactType(t *testing.T) {
	cases := map[string]ContactType{
		"sales@example.com":     ContactGeneric,
		"jane.doe@example.com":  ContactNamed,
		"founder@example.com":   ContactUnknown,
		"support@example.com":   ContactGeneric,
		"first_last@example.io": ContactNamed,
	}
	for email, want := range cases {
		if got := ClassifyContactType(email); got != want {
			t.Fatalf("%s: expected %s, got %s", email, want, got)
		}
	}
}

func TestNormalizeFindJobLimits(t *testing.T) {
	if got := NormalizeFindLimit(0); got != DefaultFindLimit {
		t.Fatalf("expected default find limit, got %d", got)
	}
	if got := NormalizeFindLimit(MaxFindLimit + 1); got != MaxFindLimit {
		t.Fatalf("expected max find limit, got %d", got)
	}
	if got := NormalizeCrawlMode(""); got != DefaultCrawlMode {
		t.Fatalf("expected default crawl mode, got %s", got)
	}
	settings := CrawlSettingsForMode(CrawlModeExhaustive)
	if settings.FetchParallelism != 100 || settings.MaxPagesPerCompany != 0 {
		t.Fatalf("unexpected exhaustive settings: %+v", settings)
	}
}
