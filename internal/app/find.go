// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/waymore/spyber/internal/domain"
	"github.com/waymore/spyber/internal/ports"
)

type FindRequest struct {
	CountryCode string
	Sector      string
	Segment     string
	Query       string
	Limit       int
}

type FindSummary struct {
	Profile      domain.BusinessProfile
	Candidates   int
	Created      int
	Duplicates   int
	Matched      int
	Rejected     int
	Fetched      int
	Contacts     int
	DirectEmails int
	Verified     int
	Failures     int
}

func (a *App) FindBusinesses(ctx context.Context, req FindRequest) (FindSummary, error) {
	if a.countryFinder == nil {
		return FindSummary{}, fmt.Errorf("country finder is not configured")
	}
	country, err := domain.NormalizeCountry(req.CountryCode)
	if err != nil {
		return FindSummary{}, err
	}
	profile, err := resolveFindProfile(req)
	if err != nil {
		return FindSummary{}, err
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 50
	}
	existing, err := a.store.ListCompanies(ctx, country)
	if err != nil {
		return FindSummary{}, err
	}
	seen := knownHosts(existing)
	candidates, err := a.searchCandidates(ctx, country, profile, limit)
	if err != nil {
		return FindSummary{}, err
	}
	summary := FindSummary{Profile: profile}
	processed := map[string]bool{}
	for _, candidate := range candidates {
		summary.Candidates++
		if !candidateAllowed(candidate.Website) {
			summary.Rejected++
			continue
		}
		_, host, err := domain.NormalizeWebsite(candidate.Website)
		if err != nil {
			summary.Failures++
			continue
		}
		if processed[host] {
			summary.Duplicates++
			continue
		}
		processed[host] = true
		created, company, err := a.storeBusinessCandidate(ctx, country, candidate, seen)
		if err != nil {
			summary.Failures++
			continue
		}
		if created {
			summary.Created++
		} else {
			summary.Duplicates++
		}
		match := scoreCandidateProfile(profile, candidate)
		a.addProfileEvidence(ctx, company.ID, profile, candidate.SourceURL, match)
		result := a.crawlCompanyForProfile(ctx, company, profile, match)
		summary.Fetched += result.fetched
		summary.Contacts += result.contacts
		if result.failed {
			summary.Failures++
		}
		if result.matched {
			summary.Matched++
			if candidate.Email != "" && a.storeDirectEmail(ctx, company.ID, candidate) {
				summary.DirectEmails++
			}
		} else {
			summary.Rejected++
		}
	}
	verified, err := a.VerifyContacts(ctx, country)
	if err != nil {
		return summary, err
	}
	summary.Verified = verified
	return summary, nil
}

func resolveFindProfile(req FindRequest) (domain.BusinessProfile, error) {
	if strings.TrimSpace(req.Query) != "" {
		return domain.CustomBusinessProfile(req.Query)
	}
	return domain.FindBusinessProfile(req.Sector, req.Segment)
}

func (a *App) searchCandidates(ctx context.Context, country string, profile domain.BusinessProfile, limit int) ([]ports.BusinessCandidate, error) {
	if searcher, ok := a.countryFinder.(ports.BusinessSearcher); ok {
		return searcher.SearchBusinesses(ctx, ports.BusinessSearch{
			CountryCode: country,
			Terms:       profile.DiscoveryTerms,
			Limit:       limit,
		})
	}
	return a.countryFinder.FindBusinesses(ctx, country, limit)
}

type profileCrawlResult struct {
	fetched  int
	contacts int
	matched  bool
	failed   bool
}

func (a *App) crawlCompanyForProfile(ctx context.Context, company domain.Company, profile domain.BusinessProfile, candidateMatch profileMatch) profileCrawlResult {
	job, err := domain.NewCrawlJob(company.ID, company.WebsiteURL, a.now())
	if err != nil {
		return profileCrawlResult{failed: true}
	}
	started := a.now().UTC()
	job.Status = domain.JobRunning
	job.StartedAt = &started
	_ = a.store.AddCrawlJob(ctx, job)

	result, err := a.fetcher.Fetch(ctx, company.WebsiteURL)
	if err != nil {
		matched := candidateMatch.Score >= profile.MinScore && !candidateMatch.Excluded
		a.updateCompanyMatch(ctx, company, profile, candidateMatch)
		a.finishJob(ctx, job, domain.JobFailed, err.Error())
		return profileCrawlResult{matched: matched, failed: true}
	}
	analysis := a.analyzer.Analyze(result.URL, result.Body)
	match := bestProfileMatch(candidateMatch, scorePageProfile(profile, company, analysis))
	a.addProfileEvidence(ctx, company.ID, profile, result.URL, match)
	matched := match.Score >= profile.MinScore && !match.Excluded
	a.updateCompanyMatch(ctx, company, profile, match)
	if !matched {
		a.finishJob(ctx, job, domain.JobSucceeded, "")
		return profileCrawlResult{fetched: 1}
	}
	contacts := a.storeContacts(ctx, company.ID, result.URL, analysis.Emails)
	fetched := 1
	for i, link := range analysis.ContactLinks {
		if i >= 3 {
			break
		}
		extraFetched, extraContacts := a.fetchContactPage(ctx, company.ID, link)
		fetched += extraFetched
		contacts += extraContacts
	}
	a.finishJob(ctx, job, domain.JobSucceeded, "")
	return profileCrawlResult{fetched: fetched, contacts: contacts, matched: true}
}

func (a *App) updateCompanyMatch(ctx context.Context, company domain.Company, profile domain.BusinessProfile, match profileMatch) {
	company.EcommerceScore = match.Score
	company.Status = domain.CompanyRejected
	if match.Score >= profile.MinScore && !match.Excluded {
		company.Status = domain.CompanyReview
	}
	company.UpdatedAt = a.now().UTC()
	_ = a.store.UpsertCompany(ctx, company)
}

func (a *App) addProfileEvidence(ctx context.Context, companyID domain.ID, profile domain.BusinessProfile, sourceURL string, match profileMatch) {
	if match.Score == 0 && len(match.Terms) == 0 && !match.Excluded {
		return
	}
	value := profile.Key() + ": " + strings.Join(match.Terms, ",")
	if match.Excluded {
		value = profile.Key() + ": excluded"
	}
	evidence, err := domain.NewEvidence(companyID, "profile_match", value, sourceURL, match.Score, a.now())
	if err == nil {
		_ = a.store.AddEvidence(ctx, evidence)
	}
}
