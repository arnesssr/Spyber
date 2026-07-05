// SPDX-License-Identifier: AGPL-3.0-only

package websearch

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/waymore/spyber/internal/domain"
	"github.com/waymore/spyber/internal/ports"
)

const defaultEndpoint = "https://lite.duckduckgo.com/lite/"
const userAgent = "Spyber/0.2.2 (+https://github.com/arnesssr/Spyber)"

type Finder struct {
	Endpoint string
	Client   *http.Client
}

func New(endpoint string) *Finder {
	if endpoint == "" {
		endpoint = defaultEndpoint
	}
	return &Finder{
		Endpoint: endpoint,
		Client:   &http.Client{Timeout: 25 * time.Second},
	}
}

func (f *Finder) FindBusinesses(ctx context.Context, countryCode string, limit int) ([]ports.BusinessCandidate, error) {
	return f.find(ctx, ports.BusinessSearch{CountryCode: countryCode, Limit: limit})
}

func (f *Finder) SearchBusinesses(ctx context.Context, search ports.BusinessSearch) ([]ports.BusinessCandidate, error) {
	return f.find(ctx, search)
}

func (f *Finder) find(ctx context.Context, search ports.BusinessSearch) ([]ports.BusinessCandidate, error) {
	country, err := domain.NormalizeCountry(search.CountryCode)
	if err != nil {
		return nil, err
	}
	limit := search.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	seen := map[string]bool{}
	var pool []scoredCandidate
	for _, query := range searchQueries(country, search.Terms) {
		candidates, err := f.query(ctx, query, perQueryLimit(limit))
		if err != nil {
			return candidatesFromPool(pool, limit), err
		}
		for _, candidate := range candidates {
			host := hostKey(candidate.Website)
			if host == "" || seen[host] || blockedHost(host) {
				continue
			}
			seen[host] = true
			pool = append(pool, scoredCandidate{
				Candidate: candidate,
				Score:     candidateScore(candidate, country, search.Terms),
				Order:     len(pool),
			})
		}
	}
	sort.SliceStable(pool, func(i, j int) bool {
		if pool[i].Score == pool[j].Score {
			return pool[i].Order < pool[j].Order
		}
		return pool[i].Score > pool[j].Score
	})
	return candidatesFromPool(pool, limit), nil
}

func (f *Finder) query(ctx context.Context, query string, limit int) ([]ports.BusinessCandidate, error) {
	searchURL, err := searchURL(f.Endpoint, query)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, searchURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "text/html")
	req.Header.Set("User-Agent", userAgent)
	resp, err := f.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("web search returned status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return nil, err
	}
	return parseResults(string(body), searchURL, query, limit), nil
}

func (f *Finder) client() *http.Client {
	if f.Client != nil {
		return f.Client
	}
	return &http.Client{Timeout: 25 * time.Second}
}

type scoredCandidate struct {
	Candidate ports.BusinessCandidate
	Score     int
	Order     int
}

func candidatesFromPool(pool []scoredCandidate, limit int) []ports.BusinessCandidate {
	var out []ports.BusinessCandidate
	for _, item := range pool {
		out = append(out, item.Candidate)
		if len(out) >= limit {
			break
		}
	}
	return out
}

func perQueryLimit(limit int) int {
	value := limit * 4
	if value < 10 {
		return 10
	}
	if value > 30 {
		return 30
	}
	return value
}

func candidateScore(candidate ports.BusinessCandidate, country string, terms []string) int {
	score := 0
	lowerURL := strings.ToLower(candidate.Website)
	lowerName := strings.ToLower(candidate.Name)
	if strings.Contains(lowerURL, "contact") || strings.Contains(lowerURL, "support") {
		score += 60
	}
	if strings.Contains(lowerURL, "about") {
		score += 20
	}
	if strings.Contains(hostKey(candidate.Website), strings.TrimPrefix(primaryTLD(country), ".")) {
		score += 20
	}
	for _, term := range cleanTerms(terms) {
		if strings.Contains(lowerURL, term) || strings.Contains(lowerName, term) {
			score += 8
		}
	}
	return score
}

func searchURL(endpoint, query string) (string, error) {
	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	values := parsed.Query()
	values.Set("q", query)
	parsed.RawQuery = values.Encode()
	return parsed.String(), nil
}

func searchQueries(country string, terms []string) []string {
	name := countryName(country)
	focused := focusedTerms(terms)
	if len(focused) == 0 {
		return []string{
			"businesses " + name + " contact email",
			"shops " + name + " contact website",
		}
	}
	joined := strings.Join(focused, " ")
	return []string{
		joined + " " + name + " contact email",
		joined + " " + name + " contact website",
		joined + " site:" + primaryTLD(country) + " contact",
	}
}

func focusedTerms(terms []string) []string {
	preferred := []string{
		"salon", "hairdresser", "beauty",
		"wholesale", "supplier", "distributor",
		"retail", "shop", "store", "ecommerce",
	}
	available := map[string]bool{}
	for _, term := range cleanTerms(terms) {
		available[term] = true
	}
	var out []string
	for _, term := range preferred {
		if available[term] {
			out = append(out, term)
		}
		if len(out) >= 3 {
			return out
		}
	}
	for _, term := range cleanTerms(terms) {
		if !contains(out, term) {
			out = append(out, term)
		}
		if len(out) >= 3 {
			break
		}
	}
	return out
}

var anchorPattern = regexp.MustCompile(`(?is)<a\s+[^>]*>.*?</a>`)
var hrefPattern = regexp.MustCompile(`(?i)href\s*=\s*["']([^"']+)["']`)
var tagPattern = regexp.MustCompile(`(?is)<[^>]+>`)

func parseResults(text, sourceURL, query string, limit int) []ports.BusinessCandidate {
	seen := map[string]bool{}
	var out []ports.BusinessCandidate
	for _, anchor := range anchorPattern.FindAllString(text, -1) {
		if !strings.Contains(anchor, "result-link") {
			continue
		}
		href := firstHref(anchor)
		raw := decodeResultURL(href)
		title := cleanTitle(anchor)
		normalized, host, err := domain.NormalizeWebsite(raw)
		if err != nil || seen[host] || blockedHost(host) || blockedResult(title, normalized) {
			continue
		}
		seen[host] = true
		out = append(out, ports.BusinessCandidate{
			Name:      title,
			Website:   normalized,
			SourceURL: sourceURL,
			Evidence:  "websearch_lite: " + query,
			Provider:  "websearch",
		})
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func blockedResult(title, rawURL string) bool {
	text := strings.ToLower(title + " " + rawURL)
	blocked := []string{
		"list of ", "top ", "best ", "directory", "business-list",
		"branches", "working hours", "contacts and", "/list/",
		"/tag/", "/category/", "/blog/", "/news/",
	}
	for _, item := range blocked {
		if strings.Contains(text, item) {
			return true
		}
	}
	return false
}

func firstHref(anchor string) string {
	match := hrefPattern.FindStringSubmatch(anchor)
	if len(match) < 2 {
		return ""
	}
	return html.UnescapeString(match[1])
}

func decodeResultURL(raw string) string {
	if strings.HasPrefix(raw, "//") {
		raw = "https:" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if strings.Contains(parsed.Hostname(), "duckduckgo.com") && parsed.Query().Get("uddg") != "" {
		return parsed.Query().Get("uddg")
	}
	return raw
}

func cleanTitle(anchor string) string {
	title := tagPattern.ReplaceAllString(anchor, " ")
	title = html.UnescapeString(title)
	return strings.Join(strings.Fields(title), " ")
}

func blockedHost(host string) bool {
	blocked := []string{
		"duckduckgo.", "bing.", "google.", "yahoo.",
		"facebook.", "instagram.", "linkedin.", "tiktok.", "youtube.", "x.com",
		"wikipedia.", "schema.org", "w3.org",
		"aeroleads.", "africabizinfo.", "primebizlist.", "businesslist.",
		"yellowpages.", "tripadvisor.", "foursquare.", "yelp.", "cybo.",
	}
	for _, item := range blocked {
		if strings.Contains(host, item) {
			return true
		}
	}
	return false
}

func hostKey(raw string) string {
	_, host, err := domain.NormalizeWebsite(raw)
	if err != nil {
		return ""
	}
	return host
}

func countryName(country string) string {
	switch country {
	case "GB":
		return "United Kingdom"
	case "KE":
		return "Kenya"
	case "US":
		return "United States"
	default:
		return country
	}
}

func primaryTLD(country string) string {
	switch country {
	case "GB":
		return ".co.uk"
	case "KE":
		return ".co.ke"
	case "US":
		return ".com"
	default:
		return "." + strings.ToLower(country)
	}
}

func cleanTerms(terms []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, term := range terms {
		term = strings.ToLower(strings.TrimSpace(term))
		if len(term) < 3 || seen[term] {
			continue
		}
		seen[term] = true
		out = append(out, term)
	}
	return out
}

func contains(values []string, needle string) bool {
	for _, value := range values {
		if value == needle {
			return true
		}
	}
	return false
}
