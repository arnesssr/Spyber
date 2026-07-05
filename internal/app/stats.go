// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"

	"github.com/arnesssr/Spyber/internal/domain"
)

type DashboardStats struct {
	CountryCode       string
	Sources           int
	Companies         int
	ReviewCompanies   int
	Contacts          int
	GenericContacts   int
	NamedContacts     int
	NeedsReview       int
	Approved          int
	Suppressed        int
	CrawlJobs         int
	FailedCrawlJobs   int
	ExportableGeneric int
}

func (a *App) DashboardStats(ctx context.Context, countryCode string) (DashboardStats, error) {
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return DashboardStats{}, err
	}
	sources, err := a.store.ListSources(ctx, country)
	if err != nil {
		return DashboardStats{}, err
	}
	companies, err := a.store.ListCompanies(ctx, country)
	if err != nil {
		return DashboardStats{}, err
	}
	contacts, err := a.store.ListContacts(ctx, country)
	if err != nil {
		return DashboardStats{}, err
	}
	jobs, err := a.store.ListCrawlJobs(ctx, country)
	if err != nil {
		return DashboardStats{}, err
	}
	stats := DashboardStats{CountryCode: country, Sources: len(sources), Companies: len(companies), Contacts: len(contacts), CrawlJobs: len(jobs)}
	for _, company := range companies {
		if company.Status == domain.CompanyReview {
			stats.ReviewCompanies++
		}
	}
	for _, contact := range contacts {
		countContact(&stats, contact)
	}
	for _, job := range jobs {
		if job.Status == domain.JobFailed {
			stats.FailedCrawlJobs++
		}
	}
	return stats, nil
}

func countContact(stats *DashboardStats, contact domain.Contact) {
	switch contact.Type {
	case domain.ContactGeneric:
		stats.GenericContacts++
	case domain.ContactNamed:
		stats.NamedContacts++
	}
	switch contact.Status {
	case domain.ContactNeedsReview:
		stats.NeedsReview++
	case domain.ContactApproved:
		stats.Approved++
	case domain.ContactSuppressed:
		stats.Suppressed++
	}
	if contact.Type == domain.ContactGeneric && contact.Status != domain.ContactSuppressed && contact.Status != domain.ContactRejected {
		stats.ExportableGeneric++
	}
}
