// SPDX-License-Identifier: AGPL-3.0-only

package websearch

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/arnesssr/Spyber/internal/ports"
)

func TestParseResultsDecodesAndFilters(t *testing.T) {
	sourceURL := "https://lite.duckduckgo.com/lite/?q=salon"
	html := `<a rel="nofollow" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fsalon.example%2Fcontact&amp;rut=abc" class='result-link'>Salon Example</a>
<a rel="nofollow" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fwww.facebook.com%2Fsalon&amp;rut=abc" class='result-link'>Facebook</a>
<a rel="nofollow" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fnews.example%2Ftop-salons-kenya&amp;rut=abc" class='result-link'>Top salons in Kenya</a>
<a rel="nofollow" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fsalon.example%2Fabout&amp;rut=abc" class='result-link'>Duplicate</a>`
	candidates := parseResults(html, sourceURL, "salon", 10)
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %+v", candidates)
	}
	if candidates[0].Website != "https://salon.example/contact" {
		t.Fatalf("unexpected website: %s", candidates[0].Website)
	}
	if candidates[0].Name != "Salon Example" {
		t.Fatalf("unexpected name: %s", candidates[0].Name)
	}
	if candidates[0].Provider != "websearch" {
		t.Fatalf("unexpected provider: %s", candidates[0].Provider)
	}
}

func TestSearchBusinessesQueriesEndpoint(t *testing.T) {
	var queries []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.Query().Get("q"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<a rel="nofollow" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Furbanhair.co.ke%2F&amp;rut=abc" class='result-link'>Urban Hair Kenya</a>`))
	}))
	defer server.Close()

	finder := New(server.URL)
	candidates, err := finder.SearchBusinesses(context.Background(), ports.BusinessSearch{
		CountryCode: "KE",
		Terms:       []string{"salon", "hairdresser", "beauty"},
		Limit:       2,
	})
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if !hasQueryContaining(queries, "salon", "Kenya") {
		t.Fatalf("unexpected queries: %+v", queries)
	}
	if len(candidates) != 1 || candidates[0].Website != "https://urbanhair.co.ke/" {
		t.Fatalf("unexpected candidates: %+v", candidates)
	}
}

func hasQueryContaining(queries []string, terms ...string) bool {
	for _, query := range queries {
		matched := true
		for _, term := range terms {
			if !strings.Contains(query, term) {
				matched = false
			}
		}
		if matched {
			return true
		}
	}
	return false
}
