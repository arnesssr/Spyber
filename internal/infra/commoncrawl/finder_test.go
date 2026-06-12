// SPDX-License-Identifier: AGPL-3.0-only

package commoncrawl

import (
	"strings"
	"testing"
)

func TestParseLinesDedupesHosts(t *testing.T) {
	data := `{"url":"https://shop-a.co.ke/contact"}
{"url":"https://shop-a.co.ke/products"}
{"url":"https://shop-b.co.ke/cart"}`
	candidates := parseLines(strings.NewReader(data), 10)
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
	if candidates[0].Website != "https://shop-a.co.ke/contact" {
		t.Fatalf("unexpected website: %s", candidates[0].Website)
	}
}

func TestParseLinesForTermsFiltersBySearchTerms(t *testing.T) {
	data := `{"url":"https://a.co.ke/salon"}
{"url":"https://b.co.ke/cart"}
{"url":"https://c.co.ke/hair-salon-contact"}`
	candidates := parseLinesForTerms(strings.NewReader(data), 10, []string{"salon"})
	if len(candidates) != 2 {
		t.Fatalf("expected 2 salon candidates, got %d", len(candidates))
	}
	if candidates[0].Website != "https://a.co.ke/salon" {
		t.Fatalf("unexpected first website: %s", candidates[0].Website)
	}
}

func TestCommonCrawlQueriesUseMultipleTerms(t *testing.T) {
	queries := commonCrawlQueries("co.ke", []string{"salon", "hairdresser", "beauty"})
	joined := strings.Join(queries, "\n")
	for _, expected := range []string{"*salon*", "*hairdresser*", "*beauty*", "*contact*"} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected %s in queries: %+v", expected, queries)
		}
	}
}
