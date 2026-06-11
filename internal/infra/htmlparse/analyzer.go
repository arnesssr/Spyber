// SPDX-License-Identifier: AGPL-3.0-only

package htmlparse

import (
	"bytes"
	"html"
	"net/url"
	"regexp"
	"strings"

	"github.com/waymore/spyber/internal/ports"
)

type Analyzer struct{}

var (
	emailPattern = regexp.MustCompile(`[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}`)
	hrefPattern  = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)
)

func New() *Analyzer {
	return &Analyzer{}
}

func (a *Analyzer) Analyze(baseURL string, body []byte) ports.PageAnalysis {
	text := string(body)
	signals := ecommerceSignals(text)
	return ports.PageAnalysis{
		Emails:           extractEmails(text),
		ContactLinks:     extractContactLinks(baseURL, text),
		CandidateLinks:   extractCandidateLinks(baseURL, text),
		EcommerceSignals: signals,
		EcommerceScore:   scoreSignals(signals),
	}
}

func extractEmails(text string) []string {
	matches := emailPattern.FindAllString(text, -1)
	seen := map[string]bool{}
	var out []string
	for _, match := range matches {
		email := strings.ToLower(strings.Trim(match, ".,;:()[]{}<>\"'"))
		if seen[email] {
			continue
		}
		seen[email] = true
		out = append(out, email)
	}
	return out
}

func extractContactLinks(baseURL, text string) []string {
	return filterLinks(extractLinks(baseURL, text), looksLikeContactLink)
}

func extractCandidateLinks(baseURL, text string) []string {
	return filterLinks(extractLinks(baseURL, text), func(link string) bool {
		return strings.HasPrefix(link, "http://") || strings.HasPrefix(link, "https://")
	})
}

func extractLinks(baseURL, text string) []string {
	base, err := url.Parse(baseURL)
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []string
	for _, match := range hrefPattern.FindAllStringSubmatch(text, -1) {
		raw := html.UnescapeString(match[1])
		parsed, err := url.Parse(raw)
		if err != nil {
			continue
		}
		resolved := base.ResolveReference(parsed)
		if resolved.Scheme != "http" && resolved.Scheme != "https" {
			continue
		}
		link := resolved.String()
		if seen[link] {
			continue
		}
		seen[link] = true
		out = append(out, link)
	}
	return out
}

func filterLinks(links []string, keep func(string) bool) []string {
	var out []string
	for _, link := range links {
		if keep(link) {
			out = append(out, link)
		}
	}
	return out
}

func looksLikeContactLink(raw string) bool {
	raw = strings.ToLower(raw)
	keywords := []string{"contact", "about", "support", "customer-service", "help", "wholesale"}
	for _, keyword := range keywords {
		if strings.Contains(raw, keyword) {
			return true
		}
	}
	return false
}

func ecommerceSignals(text string) []string {
	lower := string(bytes.ToLower([]byte(text)))
	checks := map[string]string{
		"add_to_cart": "add to cart",
		"cart_path":   "/cart",
		"checkout":    "checkout",
		"shopify":     "shopify",
		"woocommerce": "woocommerce",
		"magento":     "magento",
		"product":     "product",
		"product_cat": "product-category",
		"sku":         "sku",
		"wishlist":    "add to wishlist",
		"out_stock":   "out of stock",
		"buy_now":     "buy now",
		"ksh_price":   "ksh",
	}
	var out []string
	for signal, needle := range checks {
		if strings.Contains(lower, needle) {
			out = append(out, signal)
		}
	}
	return out
}

func scoreSignals(signals []string) int {
	score := len(signals) * 20
	if score > 100 {
		return 100
	}
	return score
}
