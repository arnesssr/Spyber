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
	"strings"
	"time"

	"github.com/waymore/spyber/internal/domain"
	"github.com/waymore/spyber/internal/ports"
)

const defaultEndpoint = "https://overpass-api.de/api/interpreter"

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
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return nil, err
	}
	if limit <= 0 || limit > 250 {
		limit = 100
	}
	body := url.Values{"data": []string{query(country, limit)}}.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, f.Endpoint, strings.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("User-Agent", "Spyber/0.1")
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

func query(country string, limit int) string {
	return fmt.Sprintf(`[out:json][timeout:35];
area["ISO3166-1:alpha2"="%s"][admin_level=2]->.country;
(
  nwr(area.country)["shop"]["website"];
  nwr(area.country)["shop"]["contact:website"];
  nwr(area.country)["shop"]["url"];
  nwr(area.country)["shop"]["email"];
  nwr(area.country)["shop"]["contact:email"];
);
out tags %d;`, country, limit)
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
