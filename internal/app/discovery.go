// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"net/url"
	"strings"

	"github.com/waymore/spyber/internal/domain"
)

type DiscoverySummary struct {
	Sources    int
	Fetched    int
	Created    int
	Skipped    int
	Failures   int
	Candidates int
}

func (a *App) DiscoverFromSources(ctx context.Context, countryCode string, limit int) (DiscoverySummary, error) {
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return DiscoverySummary{}, err
	}
	sources, err := a.store.ListSources(ctx, country)
	if err != nil {
		return DiscoverySummary{}, err
	}
	existing, err := a.store.ListCompanies(ctx, country)
	if err != nil {
		return DiscoverySummary{}, err
	}
	seen := knownHosts(existing)
	var summary DiscoverySummary
	for _, source := range sources {
		if source.Status != domain.SourceActive {
			continue
		}
		summary.Sources++
		if source.Type == "seed" {
			created, err := a.discoverCandidate(ctx, country, source.URL, source.URL, seen)
			countDiscoveryResult(&summary, created, err)
			if limitReached(summary.Created, limit) {
				break
			}
		}
		result, err := a.fetcher.Fetch(ctx, source.URL)
		if err != nil {
			summary.Failures++
			continue
		}
		summary.Fetched++
		analysis := a.analyzer.Analyze(result.URL, result.Body)
		var profileLinks []string
		for _, link := range analysis.CandidateLinks {
			if sameHost(source.URL, link) {
				if source.Type != "seed" {
					profileLinks = append(profileLinks, link)
				}
				continue
			}
			a.discoverSourceLink(ctx, country, link, result.URL, seen, &summary)
			if limitReached(summary.Created, limit) {
				return summary, nil
			}
		}
		for i, link := range profileLinks {
			if i >= 10 {
				break
			}
			a.discoverFromProfile(ctx, country, source.URL, link, seen, &summary)
			if limitReached(summary.Created, limit) {
				return summary, nil
			}
		}
	}
	return summary, nil
}

func (a *App) discoverFromProfile(ctx context.Context, country, sourceURL, profileURL string, seen map[string]bool, summary *DiscoverySummary) {
	result, err := a.fetcher.Fetch(ctx, profileURL)
	if err != nil {
		summary.Failures++
		return
	}
	summary.Fetched++
	analysis := a.analyzer.Analyze(result.URL, result.Body)
	for _, link := range analysis.CandidateLinks {
		if sameHost(sourceURL, link) {
			continue
		}
		a.discoverSourceLink(ctx, country, link, result.URL, seen, summary)
	}
}

func (a *App) discoverSourceLink(ctx context.Context, country, link, sourceURL string, seen map[string]bool, summary *DiscoverySummary) {
	if !candidateAllowed(link) {
		summary.Skipped++
		return
	}
	summary.Candidates++
	created, err := a.discoverCandidate(ctx, country, link, sourceURL, seen)
	countDiscoveryResult(summary, created, err)
}

func (a *App) discoverCandidate(ctx context.Context, country, rawURL, sourceURL string, seen map[string]bool) (bool, error) {
	_, host, err := domain.NormalizeWebsite(rawURL)
	if err != nil {
		return false, err
	}
	if seen[host] {
		return false, nil
	}
	company, err := domain.NewCompany(country, "", rawURL, a.now())
	if err != nil {
		return false, err
	}
	if err := a.store.UpsertCompany(ctx, company); err != nil {
		return false, err
	}
	seen[host] = true
	evidence, err := domain.NewEvidence(company.ID, "discovery", "source_link", sourceURL, 60, a.now())
	if err == nil {
		_ = a.store.AddEvidence(ctx, evidence)
	}
	a.audit(ctx, "company.discover_from_source", "company", company.ID.String(), `{"host":"`+company.NormalizedHost+`"}`)
	return true, nil
}

func knownHosts(companies []domain.Company) map[string]bool {
	seen := map[string]bool{}
	for _, company := range companies {
		seen[company.NormalizedHost] = true
	}
	return seen
}

func sameHost(a, b string) bool {
	_, hostA, errA := domain.NormalizeWebsite(a)
	_, hostB, errB := domain.NormalizeWebsite(b)
	return errA == nil && errB == nil && hostA == hostB
}

func candidateAllowed(raw string) bool {
	host := hostOf(raw)
	blocked := []string{
		"facebook.", "instagram.", "linkedin.", "tiktok.", "twitter.",
		"x.com", "youtube.", "google.", "schema.org", "w3.org",
		"cloudflare.", "shopify.com", "wordpress.org",
		"1-win", "1win", "bet", "casino", "gambl",
		"sexy", "porn", "adult",
	}
	for _, item := range blocked {
		if strings.Contains(host, item) {
			return false
		}
	}
	return true
}

func limitReached(created, limit int) bool {
	return limit > 0 && created >= limit
}

func countDiscoveryResult(summary *DiscoverySummary, created bool, err error) {
	if err != nil {
		summary.Skipped++
		return
	}
	if created {
		summary.Created++
		return
	}
	summary.Skipped++
}

func hostOf(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if parsed.Hostname() == "" {
		return raw
	}
	return parsed.Hostname()
}
