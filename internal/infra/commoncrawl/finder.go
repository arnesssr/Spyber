// SPDX-License-Identifier: AGPL-3.0-only

package commoncrawl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/waymore/spyber/internal/domain"
	"github.com/waymore/spyber/internal/ports"
)

const collectionsURL = "https://index.commoncrawl.org/collinfo.json"

type Finder struct {
	IndexAPI string
	Client   *http.Client
}

func New(indexAPI string) *Finder {
	return &Finder{
		IndexAPI: indexAPI,
		Client:   &http.Client{Timeout: 35 * time.Second},
	}
}

func (f *Finder) FindBusinesses(ctx context.Context, countryCode string, limit int) ([]ports.BusinessCandidate, error) {
	return f.find(ctx, countryCode, nil, limit)
}

func (f *Finder) SearchBusinesses(ctx context.Context, search ports.BusinessSearch) ([]ports.BusinessCandidate, error) {
	return f.find(ctx, search.CountryCode, search.Terms, search.Limit)
}

func (f *Finder) find(ctx context.Context, countryCode string, terms []string, limit int) ([]ports.BusinessCandidate, error) {
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 250 {
		limit = 100
	}
	indexAPI := f.IndexAPI
	if indexAPI == "" {
		indexAPI, err = f.latestIndex(ctx)
		if err != nil {
			return nil, err
		}
	}
	seen := map[string]bool{}
	var out []ports.BusinessCandidate
	for _, domain := range countryDomains(country) {
		if len(out) >= limit {
			break
		}
		candidates, err := f.query(ctx, indexAPI, domain, terms, limit-len(out))
		if err != nil {
			continue
		}
		for _, candidate := range candidates {
			if seen[candidate.Website] {
				continue
			}
			seen[candidate.Website] = true
			out = append(out, candidate)
			if len(out) >= limit {
				break
			}
		}
	}
	return out, nil
}

func (f *Finder) latestIndex(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, collectionsURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := f.client().Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	var collections []struct {
		API string `json:"cdx-api"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&collections); err != nil {
		return "", err
	}
	if len(collections) == 0 || collections[0].API == "" {
		return "", fmt.Errorf("common crawl index list is empty")
	}
	return collections[0].API, nil
}

func (f *Finder) query(ctx context.Context, indexAPI, domain string, terms []string, limit int) ([]ports.BusinessCandidate, error) {
	rawURL := domain
	if len(cleanTerms(terms)) > 0 {
		rawURL = "*." + domain + "/*" + cleanTerms(terms)[0] + "*"
	}
	values := url.Values{
		"url":    []string{rawURL},
		"output": []string{"json"},
		"filter": []string{"status:200", "mime:text/html"},
		"limit":  []string{"500"},
	}
	if len(cleanTerms(terms)) == 0 {
		values["matchType"] = []string{"domain"}
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, indexAPI+"?"+values.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Spyber/0.1")
	resp, err := f.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("common crawl returned status %d", resp.StatusCode)
	}
	return parseLinesForTerms(resp.Body, limit, terms), nil
}

func (f *Finder) client() *http.Client {
	if f.Client != nil {
		return f.Client
	}
	return &http.Client{Timeout: 35 * time.Second}
}

func countryDomains(country string) []string {
	tlds := countryTLDs(country)
	var domains []string
	for _, tld := range tlds {
		domains = append(domains, strings.TrimPrefix(tld, "."))
	}
	return domains
}

func countryTLDs(country string) []string {
	switch country {
	case "GB":
		return []string{".co.uk", ".uk"}
	case "KE":
		return []string{".co.ke", ".ke"}
	case "US":
		return []string{".us", ".com"}
	default:
		return []string{"." + strings.ToLower(country)}
	}
}

type cdxRecord struct {
	URL string `json:"url"`
}

func parseLines(reader io.Reader, limit int) []ports.BusinessCandidate {
	return parseLinesForTerms(reader, limit, nil)
}

func parseLinesForTerms(reader io.Reader, limit int, terms []string) []ports.BusinessCandidate {
	scanner := bufio.NewScanner(reader)
	seen := map[string]bool{}
	var out []ports.BusinessCandidate
	for scanner.Scan() {
		var record cdxRecord
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			continue
		}
		if !usefulURL(record.URL, terms) {
			continue
		}
		candidate, ok := toCandidate(record.URL)
		host := candidateHost(candidate.Website)
		if !ok || seen[host] {
			continue
		}
		seen[host] = true
		out = append(out, candidate)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func candidateHost(rawURL string) string {
	_, host, err := domain.NormalizeWebsite(rawURL)
	if err != nil {
		return rawURL
	}
	return host
}

func toCandidate(rawURL string) (ports.BusinessCandidate, bool) {
	normalized, host, err := domain.NormalizeWebsite(rawURL)
	if err != nil {
		return ports.BusinessCandidate{}, false
	}
	return ports.BusinessCandidate{
		Name:      host,
		Website:   normalized,
		SourceURL: normalized,
		Evidence:  "commoncrawl_country_tld",
	}, true
}

func usefulURL(rawURL string, terms []string) bool {
	lower := strings.ToLower(rawURL)
	keywords := []string{"shop", "store", "product", "cart", "checkout"}
	blocked := []string{"1-win", "1win", "bet", "casino", "gambl", "login", "register", "sexy", "porn", "adult"}
	for _, item := range blocked {
		if strings.Contains(lower, item) {
			return false
		}
	}
	for _, term := range cleanTerms(terms) {
		if urlContainsTerm(lower, term) {
			return true
		}
	}
	if len(cleanTerms(terms)) > 0 {
		return false
	}
	for _, keyword := range keywords {
		if strings.Contains(lower, keyword) {
			return true
		}
	}
	return false
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

func urlContainsTerm(rawURL, term string) bool {
	for _, token := range strings.FieldsFunc(rawURL, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	}) {
		if token == term {
			return true
		}
	}
	return false
}
