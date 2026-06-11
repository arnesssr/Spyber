// SPDX-License-Identifier: AGPL-3.0-only

package overpass

import (
	"strings"
	"testing"
)

func TestParseBusinessCandidates(t *testing.T) {
	data := []byte(`{
		"elements": [
			{"type":"node","id":1,"tags":{"name":"Shop A","shop":"clothes","website":"https://shop-a.example","email":"sales@shop-a.example"}},
			{"type":"way","id":2,"tags":{"shop":"electronics","contact:website":"shop-b.example"}},
			{"type":"node","id":3,"tags":{"shop":"books"}}
		]
	}`)
	candidates := parse(data, 10)
	if len(candidates) != 2 {
		t.Fatalf("expected 2 candidates, got %d", len(candidates))
	}
	if candidates[0].Website != "https://shop-a.example" {
		t.Fatalf("unexpected first website: %s", candidates[0].Website)
	}
	if candidates[1].Website != "https://shop-b.example" {
		t.Fatalf("unexpected second website: %s", candidates[1].Website)
	}
}

func TestQueryUsesSearchTerms(t *testing.T) {
	got := query("KE", []string{"salon"}, 5)
	if !strings.Contains(got, `["shop"~"hairdresser|beauty|beauty_salon"]`) {
		t.Fatalf("expected salon shop filter, got %s", got)
	}
	if !strings.Contains(got, `salon`) {
		t.Fatalf("expected salon name filter, got %s", got)
	}
	if strings.Contains(got, `"spa",i`) {
		t.Fatalf("expected bounded term regex, got %s", got)
	}
}
