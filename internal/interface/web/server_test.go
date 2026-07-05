// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/arnesssr/Spyber/internal/app"
	"github.com/arnesssr/Spyber/internal/infra/htmlparse"
	"github.com/arnesssr/Spyber/internal/infra/httpfetch"
	"github.com/arnesssr/Spyber/internal/infra/localstore"
	"github.com/arnesssr/Spyber/internal/ports"
)

func TestDashboardRenders(t *testing.T) {
	store := localstore.New(t.TempDir() + "/spyber.json")
	application := app.New(store, httpfetch.New(), htmlparse.New())
	if err := application.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}
	server := New(application, Config{})
	req := httptest.NewRequest(http.MethodGet, "/?country=GB", nil)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if !strings.Contains(res.Body.String(), "Search contacts") {
		t.Fatalf("expected find body, got %s", res.Body.String())
	}
	if !strings.Contains(res.Body.String(), "Crawl mode") {
		t.Fatalf("expected crawl mode selector, got %s", res.Body.String())
	}
}

func TestAdminTokenRequiresBasicAuth(t *testing.T) {
	store := localstore.New(t.TempDir() + "/spyber.json")
	application := app.New(store, httpfetch.New(), htmlparse.New())
	if err := application.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}
	server := New(application, Config{AdminToken: "secret"})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	if res.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", res.Code)
	}
}

func TestFindQueuesBackgroundJob(t *testing.T) {
	store := localstore.New(t.TempDir() + "/spyber.json")
	application := app.New(store, webFakeFetcher{}, htmlparse.New()).WithCountryFinder(webFakeFinder{})
	if err := application.Init(context.Background()); err != nil {
		t.Fatalf("init: %v", err)
	}
	server := New(application, Config{})
	form := url.Values{
		"country":    {"KE"},
		"query":      {"shop"},
		"limit":      {"1"},
		"crawl_mode": {"exhaustive"},
	}
	req := httptest.NewRequest(http.MethodPost, "/find", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res := httptest.NewRecorder()
	server.ServeHTTP(res, req)
	if res.Code != http.StatusSeeOther {
		t.Fatalf("expected redirect, got %d", res.Code)
	}
	if location := res.Header().Get("Location"); !strings.HasPrefix(location, "/jobs?") {
		t.Fatalf("expected jobs redirect, got %s", location)
	}
	waitForJob(t, application, store)
}

func waitForJob(t *testing.T, application *app.App, store *localstore.Store) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		jobs, err := application.ListFindJobs(context.Background(), "KE")
		if err != nil {
			t.Fatalf("jobs: %v", err)
		}
		if len(jobs) > 0 && jobs[0].Status == "succeeded" {
			if jobs[0].Matched != 1 || jobs[0].Contacts != 1 {
				t.Fatalf("unexpected job summary: %+v", jobs[0])
			}
			if jobs[0].CrawlMode != "exhaustive" {
				t.Fatalf("expected exhaustive crawl mode, got %s", jobs[0].CrawlMode)
			}
			tasks, err := store.ListFetchTasks(context.Background(), jobs[0].ID)
			if err != nil {
				t.Fatalf("fetch tasks: %v", err)
			}
			if len(tasks) == 0 {
				t.Fatal("expected persisted fetch tasks")
			}
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("background find job did not finish")
}

type webFakeFinder struct{}

func (webFakeFinder) FindBusinesses(ctx context.Context, countryCode string, limit int) ([]ports.BusinessCandidate, error) {
	return nil, nil
}

func (webFakeFinder) SearchBusinesses(ctx context.Context, search ports.BusinessSearch) ([]ports.BusinessCandidate, error) {
	return []ports.BusinessCandidate{{
		Name:      "Shop",
		Website:   "https://shop.example",
		SourceURL: "https://source.example",
		Evidence:  "shop product checkout",
	}}, nil
}

type webFakeFetcher struct{}

func (webFakeFetcher) Fetch(ctx context.Context, rawURL string) (ports.FetchResult, error) {
	body := `<html><body>shop product checkout info@shop.example</body></html>`
	return ports.FetchResult{URL: rawURL, StatusCode: 200, Body: []byte(body)}, nil
}
