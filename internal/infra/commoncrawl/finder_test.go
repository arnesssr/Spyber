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
