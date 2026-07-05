// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"fmt"
	"time"

	"github.com/arnesssr/Spyber/internal/domain"
)

type CrawlSummary struct {
	Companies int
	Fetched   int
	Contacts  int
	Failures  int
}

func (a *App) CrawlCountry(ctx context.Context, countryCode string) (CrawlSummary, error) {
	if a.fetcher == nil || a.analyzer == nil {
		return CrawlSummary{}, fmt.Errorf("crawler dependencies are not configured")
	}
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return CrawlSummary{}, err
	}
	companies, err := a.store.ListCompanies(ctx, country)
	if err != nil {
		return CrawlSummary{}, err
	}
	var summary CrawlSummary
	for _, company := range companies {
		summary.Companies++
		fetched, contacts, failed := a.crawlCompany(ctx, company)
		summary.Fetched += fetched
		summary.Contacts += contacts
		if failed {
			summary.Failures++
		}
	}
	return summary, nil
}

func (a *App) crawlCompany(ctx context.Context, company domain.Company) (int, int, bool) {
	job, err := domain.NewCrawlJob(company.ID, company.WebsiteURL, a.now())
	if err != nil {
		return 0, 0, true
	}
	started := a.now().UTC()
	job.Status = domain.JobRunning
	job.StartedAt = &started
	_ = a.store.AddCrawlJob(ctx, job)

	fetched, contacts, signals, err := a.fetchAndAnalyze(ctx, company, company.WebsiteURL)
	if err != nil {
		a.finishJob(ctx, job, domain.JobFailed, err.Error())
		return fetched, contacts, true
	}
	company.Status = domain.CompanyCrawled
	company.EcommerceScore = clampScore(signals * 20)
	company.UpdatedAt = a.now().UTC()
	if company.EcommerceScore >= 40 {
		company.Status = domain.CompanyReview
	} else {
		company.Status = domain.CompanyRejected
	}
	_ = a.store.UpsertCompany(ctx, company)
	a.finishJob(ctx, job, domain.JobSucceeded, "")
	return fetched, contacts, false
}

func (a *App) fetchAndAnalyze(ctx context.Context, company domain.Company, rawURL string) (int, int, int, error) {
	result, err := a.fetcher.Fetch(ctx, rawURL)
	if err != nil {
		return 0, 0, 0, err
	}
	analysis := a.analyzer.Analyze(result.URL, result.Body)
	for _, signal := range analysis.EcommerceSignals {
		evidence, err := domain.NewEvidence(company.ID, "ecommerce", signal, result.URL, 70, a.now())
		if err == nil {
			_ = a.store.AddEvidence(ctx, evidence)
		}
	}
	if analysis.EcommerceScore < 40 {
		return 1, 0, len(analysis.EcommerceSignals), nil
	}
	contacts := a.storeContacts(ctx, company.ID, result.URL, analysis.Emails)
	totalFetched := 1
	totalContacts := contacts
	for i, link := range analysis.ContactLinks {
		if i >= 3 {
			break
		}
		extraFetched, extraContacts := a.fetchContactPage(ctx, company.ID, link)
		totalFetched += extraFetched
		totalContacts += extraContacts
	}
	return totalFetched, totalContacts, len(analysis.EcommerceSignals), nil
}

func (a *App) fetchContactPage(ctx context.Context, companyID domain.ID, link string) (int, int) {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	result, err := a.fetcher.Fetch(ctx, link)
	if err != nil {
		return 0, 0
	}
	analysis := a.analyzer.Analyze(result.URL, result.Body)
	return 1, a.storeContacts(ctx, companyID, result.URL, analysis.Emails)
}

func (a *App) storeContacts(ctx context.Context, companyID domain.ID, sourceURL string, emails []string) int {
	seen := a.knownCompanyEmails(ctx, companyID)
	count := 0
	for _, email := range emails {
		contact, err := domain.NewContact(companyID, email, sourceURL, a.now())
		if err != nil {
			continue
		}
		if seen[contact.Email] {
			continue
		}
		if err := a.store.UpsertContact(ctx, contact); err == nil {
			seen[contact.Email] = true
			count++
		}
	}
	return count
}

func (a *App) knownCompanyEmails(ctx context.Context, companyID domain.ID) map[string]bool {
	seen := map[string]bool{}
	contacts, err := a.store.ListCompanyContacts(ctx, companyID)
	if err != nil {
		return seen
	}
	for _, contact := range contacts {
		seen[contact.Email] = true
	}
	return seen
}

func (a *App) finishJob(ctx context.Context, job domain.CrawlJob, status domain.JobStatus, reason string) {
	finished := a.now().UTC()
	job.Status = status
	job.FailureReason = reason
	job.FinishedAt = &finished
	_ = a.store.AddCrawlJob(ctx, job)
}

func clampScore(score int) int {
	if score < 0 {
		return 0
	}
	if score > 100 {
		return 100
	}
	return score
}
