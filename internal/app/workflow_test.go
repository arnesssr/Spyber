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

type fakeCountryFinder struct{}

func (fakeCountryFinder) FindBusinesses(ctx context.Context, countryCode string, limit int) ([]ports.BusinessCandidate, error) {
	return []ports.BusinessCandidate{
		{Name: "Shop A", Website: "https://shop-a.example", Email: "sales@shop-a.example", SourceURL: "https://www.openstreetmap.org/node/1", Evidence: "test"},
	}, nil
}

type profileAnalyzer struct{}

func (profileAnalyzer) Analyze(baseURL string, body []byte) ports.PageAnalysis {
	return ports.PageAnalysis{
		Emails:       []string{"sales@wholesale.example"},
		ContactLinks: []string{"https://wholesale.example/contact"},
		Text:         "wholesale supplier trade account bulk orders",
	}
}

type profileCountryFinder struct{}

func (profileCountryFinder) FindBusinesses(ctx context.Context, countryCode string, limit int) ([]ports.BusinessCandidate, error) {
	return nil, nil
}

func (profileCountryFinder) SearchBusinesses(ctx context.Context, search ports.BusinessSearch) ([]ports.BusinessCandidate, error) {
	return []ports.BusinessCandidate{
		{Name: "Wholesale Example", Website: "https://wholesale.example", Email: "info@wholesale.example", SourceURL: "https://www.openstreetmap.org/node/2", Evidence: "wholesale supplier"},
		{Name: "Wholesale Example Duplicate", Website: "https://wholesale.example/about", SourceURL: "https://example.invalid", Evidence: "duplicate"},
	}, nil
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

func TestScrapeCountryDiscoversCrawlsAndStoresContacts(t *testing.T) {
	ctx := context.Background()
	store := localstore.New(t.TempDir() + "/spyber.json")
	app := New(store, fakeFetcher{}, fakeAnalyzer{}).WithCountryFinder(fakeCountryFinder{})
	if err := app.Init(ctx); err != nil {
		t.Fatalf("init: %v", err)
	}
	summary, err := app.ScrapeCountry(ctx, "KE", 10)
	if err != nil {
		t.Fatalf("scrape country: %v", err)
	}
	if summary.Discovered != 1 || summary.Crawled != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	contacts, err := app.ListContacts(ctx, "KE")
	if err != nil {
		t.Fatalf("contacts: %v", err)
	}
	if len(contacts) < 2 {
		t.Fatalf("expected scraped contacts, got %+v", contacts)
	}
}

func TestFindBusinessesUsesProfileAndDedupesCompanies(t *testing.T) {
	ctx := context.Background()
	store := localstore.New(t.TempDir() + "/spyber.json")
	app := New(store, fakeFetcher{}, profileAnalyzer{}).WithCountryFinder(profileCountryFinder{})
	if err := app.Init(ctx); err != nil {
		t.Fatalf("init: %v", err)
	}
	summary, err := app.FindBusinesses(ctx, FindRequest{
		CountryCode: "KE",
		Sector:      "commerce",
		Segment:     "wholesalers",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("find businesses: %v", err)
	}
	if summary.Created != 1 || summary.Duplicates != 1 || summary.Matched != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if summary.Contacts != 1 || summary.DirectEmails != 1 {
		t.Fatalf("expected unique page and direct contacts, got %+v", summary)
	}
	companies, err := app.ListCompanies(ctx, "KE")
	if err != nil {
		t.Fatalf("companies: %v", err)
	}
	if len(companies) != 1 || companies[0].Status != "review" {
		t.Fatalf("expected one review company, got %+v", companies)
	}
	contacts, err := app.ListContacts(ctx, "KE")
	if err != nil {
		t.Fatalf("contacts: %v", err)
	}
	if len(contacts) != 2 {
		t.Fatalf("expected page and direct contacts, got %+v", contacts)
	}
}
