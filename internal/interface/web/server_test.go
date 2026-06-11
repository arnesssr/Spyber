// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/waymore/spyber/internal/app"
	"github.com/waymore/spyber/internal/infra/htmlparse"
	"github.com/waymore/spyber/internal/infra/httpfetch"
	"github.com/waymore/spyber/internal/infra/localstore"
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
	if !strings.Contains(res.Body.String(), "Dashboard") {
		t.Fatalf("expected dashboard body, got %s", res.Body.String())
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
