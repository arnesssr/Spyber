// SPDX-License-Identifier: AGPL-3.0-only

package overpass

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/waymore/spyber/internal/domain"
	"github.com/waymore/spyber/internal/ports"
)

const defaultEndpoint = "https://overpass-api.de/api/interpreter"
const userAgent = "Spyber/0.2.1 (+https://github.com/arnesssr/Spyber)"

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
		Client:   &http.Client{Timeout: 45 * time.Second},
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
	body := url.Values{"data": []string{query(country, terms, limit)}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.Endpoint, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", userAgent)
	client := f.Client
	if client == nil {
		client = &http.Client{Timeout: 45 * time.Second}
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("overpass returned status %d", resp.StatusCode)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024))
	if err != nil {
		return nil, err
	}
	return parse(data, limit), nil
}

func query(country string, terms []string, limit int) string {
	filters := queryFilters(terms)
	var lines []string
	for _, filter := range filters {
		lines = append(lines,
			fmt.Sprintf(`  nwr(area.country)%s["website"];`, filter),
			fmt.Sprintf(`  nwr(area.country)%s["contact:website"];`, filter),
			fmt.Sprintf(`  nwr(area.country)%s["url"];`, filter),
			fmt.Sprintf(`  nwr(area.country)%s["email"];`, filter),
			fmt.Sprintf(`  nwr(area.country)%s["contact:email"];`, filter),
		)
	}
	return fmt.Sprintf(`[out:json][timeout:35];
area["ISO3166-1:alpha2"="%s"][admin_level=2]->.country;
(
%s
);
out tags %d;`, country, strings.Join(lines, "\n"), limit)
}

func queryFilters(terms []string) []string {
	regex := overpassRegex(terms)
	switch {
	case regex == "":
		return []string{`["shop"]`}
	case containsAnyTerm(terms, "salon", "salons", "hairdresser", "beauty", "barber", "spa"):
		return []string{`["shop"~"hairdresser|beauty|beauty_salon"]`, `["name"~"` + regex + `",i]`}
	case containsAnyTerm(terms, "wholesale", "wholesaler", "wholesalers", "supplier", "distributor", "bulk", "trade"):
		return []string{`["shop"~"wholesale|trade"]`, `["name"~"` + regex + `",i]`}
	case containsAnyTerm(terms, "retail", "retailer", "retailers", "shop", "store", "product"):
		return []string{`["shop"]`, `["name"~"` + regex + `",i]`}
	default:
		return []string{`["name"~"` + regex + `",i]`}
	}
}

func overpassRegex(terms []string) string {
	var safe []string
	for _, term := range terms {
		term = strings.TrimSpace(strings.ToLower(term))
		if len(term) < 3 {
			continue
		}
		safe = append(safe, `(^|[^A-Za-z0-9])`+regexp.QuoteMeta(term)+`([^A-Za-z0-9]|$)`)
	}
	return strings.Join(safe, "|")
}

func containsAnyTerm(terms []string, values ...string) bool {
	set := map[string]bool{}
	for _, value := range values {
		set[value] = true
	}
	for _, term := range terms {
		if set[strings.ToLower(strings.TrimSpace(term))] {
			return true
		}
	}
	return false
}

type response struct {
	Elements []element `json:"elements"`
}

type element struct {
	Type string            `json:"type"`
	ID   int64             `json:"id"`
	Tags map[string]string `json:"tags"`
}

func parse(data []byte, limit int) []ports.BusinessCandidate {
	var res response
	if err := json.NewDecoder(bytes.NewReader(data)).Decode(&res); err != nil {
		return nil
	}
	seen := map[string]bool{}
	var out []ports.BusinessCandidate
	for _, item := range res.Elements {
		candidate, ok := toCandidate(item)
		if !ok || seen[candidate.Website] {
			continue
		}
		seen[candidate.Website] = true
		out = append(out, candidate)
		if limit > 0 && len(out) >= limit {
			break
		}
	}
	return out
}

func toCandidate(item element) (ports.BusinessCandidate, bool) {
	tags := item.Tags
	website := first(tags, "website", "contact:website", "url")
	email := first(tags, "email", "contact:email")
	if email != "" {
		normalizedEmail, err := domain.NormalizeEmail(email)
		if err == nil {
			email = normalizedEmail
		} else {
			email = ""
		}
	}
	if website == "" && email == "" {
		return ports.BusinessCandidate{}, false
	}
	if website == "" && email != "" {
		parts := strings.Split(email, "@")
		website = "https://" + parts[1]
	}
	normalized, _, err := domain.NormalizeWebsite(website)
	if err != nil {
		return ports.BusinessCandidate{}, false
	}
	name := first(tags, "name", "operator", "brand")
	return ports.BusinessCandidate{
		Name:      name,
		Website:   normalized,
		Email:     email,
		SourceURL: fmt.Sprintf("https://www.openstreetmap.org/%s/%d", item.Type, item.ID),
		Evidence:  "osm_shop_tags",
	}, true
}

func first(tags map[string]string, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(tags[key]); value != "" {
			return value
		}
	}
	return ""
}
