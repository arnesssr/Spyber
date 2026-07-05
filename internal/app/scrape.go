// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"fmt"

	"github.com/arnesssr/Spyber/internal/domain"
	"github.com/arnesssr/Spyber/internal/ports"
)

type CountryScrapeSummary struct {
	Discovered   int
	DirectEmails int
	Crawled      int
	Fetched      int
	Contacts     int
	Failures     int
	Verified     int
}

func (a *App) DiscoverCountry(ctx context.Context, countryCode string, limit int) (CountryScrapeSummary, error) {
	if a.countryFinder == nil {
		return CountryScrapeSummary{}, fmt.Errorf("country finder is not configured")
	}
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return CountryScrapeSummary{}, err
	}
	existing, err := a.store.ListCompanies(ctx, country)
	if err != nil {
		return CountryScrapeSummary{}, err
	}
	seen := knownHosts(existing)
	candidates, err := a.countryFinder.FindBusinesses(ctx, country, limit)
	if err != nil {
		return CountryScrapeSummary{}, err
	}
	var summary CountryScrapeSummary
	if len(candidates) == 0 {
		return summary, fmt.Errorf("%w in %s", ErrNoCandidates, country)
	}
	for _, candidate := range candidates {
		created, company, err := a.storeBusinessCandidate(ctx, country, candidate, seen)
		if err != nil {
			summary.Failures++
			continue
		}
		if created {
			summary.Discovered++
		}
		if candidate.Email != "" && company.ID != "" {
			if a.storeDirectEmail(ctx, company.ID, candidate) {
				summary.DirectEmails++
			}
		}
	}
	return summary, nil
}

func (a *App) ScrapeCountry(ctx context.Context, countryCode string, limit int) (CountryScrapeSummary, error) {
	summary, err := a.DiscoverCountry(ctx, countryCode, limit)
	if err != nil {
		return summary, err
	}
	crawl, err := a.CrawlCountry(ctx, countryCode)
	if err != nil {
		return summary, err
	}
	summary.Crawled = crawl.Companies
	summary.Fetched = crawl.Fetched
	summary.Contacts += crawl.Contacts
	summary.Failures += crawl.Failures
	verified, err := a.VerifyContacts(ctx, countryCode)
	if err != nil {
		return summary, err
	}
	summary.Verified = verified
	return summary, nil
}

func (a *App) storeBusinessCandidate(ctx context.Context, country string, candidate ports.BusinessCandidate, seen map[string]bool) (bool, domain.Company, error) {
	if !candidateAllowed(candidate.Website) {
		return false, domain.Company{}, nil
	}
	_, host, err := domain.NormalizeWebsite(candidate.Website)
	if err != nil {
		return false, domain.Company{}, err
	}
	if seen[host] {
		companies, err := a.store.ListCompanies(ctx, country)
		if err != nil {
			return false, domain.Company{}, err
		}
		for _, company := range companies {
			if company.NormalizedHost == host {
				return false, company, nil
			}
		}
	}
	company, err := domain.NewCompany(country, candidate.Name, candidate.Website, a.now())
	if err != nil {
		return false, domain.Company{}, err
	}
	if err := a.store.UpsertCompany(ctx, company); err != nil {
		return false, domain.Company{}, err
	}
	seen[host] = true
	evidence, err := domain.NewEvidence(company.ID, "country_discovery", candidate.Evidence, candidate.SourceURL, 80, a.now())
	if err == nil {
		_ = a.store.AddEvidence(ctx, evidence)
	}
	a.audit(ctx, "company.discover_country", "company", company.ID.String(), `{"host":"`+company.NormalizedHost+`"}`)
	return true, company, nil
}

func (a *App) storeDirectEmail(ctx context.Context, companyID domain.ID, candidate ports.BusinessCandidate) bool {
	contact, err := domain.NewContact(companyID, candidate.Email, candidate.SourceURL, a.now())
	if err != nil {
		return false
	}
	if a.knownCompanyEmails(ctx, companyID)[contact.Email] {
		return false
	}
	return a.store.UpsertContact(ctx, contact) == nil
}
