// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"strings"
	"testing"

	"github.com/waymore/spyber/internal/infra/localstore"
	"github.com/waymore/spyber/internal/ports"
)

type fakeFetcher struct{}

func (fakeFetcher) Fetch(ctx context.Context, rawURL string) (ports.FetchResult, error) {
	return ports.FetchResult{URL: rawURL, StatusCode: 200, Body: []byte("ok")}, nil
}

type fakeAnalyzer struct{}

func (fakeAnalyzer) Analyze(baseURL string, body []byte) ports.PageAnalysis {
	return ports.PageAnalysis{
		Emails:           []string{"sales@example.com", "jane.doe@example.com"},
		EcommerceSignals: []string{"checkout", "cart_path"},
		EcommerceScore:   40,
	}
}

type discoveryAnalyzer struct{}

func (discoveryAnalyzer) Analyze(baseURL string, body []byte) ports.PageAnalysis {
	return ports.PageAnalysis{
		CandidateLinks: []string{"https://shop-a.example", "https://facebook.com/shop"},
	}
}

func TestCrawlReviewAndExportWorkflow(t *testing.T) {
	ctx := context.Background()
	store := localstore.New(t.TempDir() + "/spyber.json")
	app := New(store, fakeFetcher{}, fakeAnalyzer{})
	if err := app.Init(ctx); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := app.DiscoverDomain(ctx, "GB", "https://shop.example"); err != nil {
		t.Fatalf("discover: %v", err)
	}
	summary, err := app.CrawlCountry(ctx, "GB")
	if err != nil {
		t.Fatalf("crawl: %v", err)
	}
	if summary.Contacts != 2 {
		t.Fatalf("expected 2 contacts, got %d", summary.Contacts)
	}
	contacts, err := app.ListContacts(ctx, "GB")
	if err != nil {
		t.Fatalf("contacts: %v", err)
	}
	for _, contact := range contacts {
		if contact.Email == "sales@example.com" {
			if _, err := app.ApproveContact(ctx, contact.ID); err != nil {
				t.Fatalf("approve: %v", err)
			}
		}
	}
	data, rows, err := app.ExportContacts(ctx, ExportOptions{CountryCode: "GB", Format: "csv", Only: "generic"})
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected 1 exported row, got %d", rows)
	}
	if !strings.Contains(string(data), "sales@example.com") {
		t.Fatalf("expected sales@example.com in export: %s", string(data))
	}
	if strings.Contains(string(data), "jane.doe@example.com") {
		t.Fatalf("named email should not be in generic export: %s", string(data))
	}
}

func TestDiscoverFromSourcesCreatesFilteredCompanies(t *testing.T) {
	ctx := context.Background()
	store := localstore.New(t.TempDir() + "/spyber.json")
	app := New(store, fakeFetcher{}, discoveryAnalyzer{})
	if err := app.Init(ctx); err != nil {
		t.Fatalf("init: %v", err)
	}
	if _, err := app.AddSource(ctx, "GB", "directory", "https://directory.example"); err != nil {
		t.Fatalf("source: %v", err)
	}
	summary, err := app.DiscoverFromSources(ctx, "GB", 10)
	if err != nil {
		t.Fatalf("discover from sources: %v", err)
	}
	if summary.Created != 1 {
		t.Fatalf("expected 1 created company, got %d", summary.Created)
	}
	companies, err := app.ListCompanies(ctx, "GB")
	if err != nil {
		t.Fatalf("companies: %v", err)
	}
	if len(companies) != 1 || companies[0].NormalizedHost != "shop-a.example" {
		t.Fatalf("unexpected companies: %+v", companies)
	}
}
