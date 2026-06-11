// SPDX-License-Identifier: AGPL-3.0-only

package htmlparse

import "testing"

func TestAnalyzeExtractsSignalsEmailsAndContactLinks(t *testing.T) {
	html := []byte(`
		<html>
			<body>
				<a href="/contact">Contact</a>
				<a href="mailto:sales@example.com">Email</a>
				<button>Add to cart</button>
				<a href="/cart">Cart</a>
				<p>Wholesale: wholesale@example.com</p>
			</body>
		</html>
	`)
	analysis := New().Analyze("https://shop.example", html)
	if len(analysis.Emails) != 2 {
		t.Fatalf("expected 2 emails, got %d", len(analysis.Emails))
	}
	if len(analysis.ContactLinks) != 1 {
		t.Fatalf("expected 1 contact link, got %d", len(analysis.ContactLinks))
	}
	if len(analysis.CandidateLinks) == 0 {
		t.Fatal("expected candidate links")
	}
	if analysis.ContactLinks[0] != "https://shop.example/contact" {
		t.Fatalf("unexpected contact link: %s", analysis.ContactLinks[0])
	}
	if analysis.EcommerceScore == 0 {
		t.Fatal("expected ecommerce score")
	}
}
