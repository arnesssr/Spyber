// SPDX-License-Identifier: AGPL-3.0-only

package ports

import (
	"context"

	"github.com/waymore/spyber/internal/domain"
)

type Store interface {
	Init(ctx context.Context) error
	AddSource(ctx context.Context, source domain.Source) error
	ListSources(ctx context.Context, countryCode string) ([]domain.Source, error)
	UpsertCompany(ctx context.Context, company domain.Company) error
	ListCompanies(ctx context.Context, countryCode string) ([]domain.Company, error)
	GetCompany(ctx context.Context, id domain.ID) (domain.Company, bool, error)
	UpsertContact(ctx context.Context, contact domain.Contact) error
	ListContacts(ctx context.Context, countryCode string) ([]domain.Contact, error)
	ListCompanyContacts(ctx context.Context, companyID domain.ID) ([]domain.Contact, error)
	GetContact(ctx context.Context, id domain.ID) (domain.Contact, bool, error)
	AddCrawlJob(ctx context.Context, job domain.CrawlJob) error
	ListCrawlJobs(ctx context.Context, countryCode string) ([]domain.CrawlJob, error)
	UpsertFindJob(ctx context.Context, job domain.FindJob) error
	GetFindJob(ctx context.Context, id domain.ID) (domain.FindJob, bool, error)
	ListFindJobs(ctx context.Context, countryCode string) ([]domain.FindJob, error)
	AddEvidence(ctx context.Context, evidence domain.Evidence) error
	ListEvidence(ctx context.Context, companyID domain.ID) ([]domain.Evidence, error)
	AddSuppression(ctx context.Context, suppression domain.Suppression) error
	ListSuppressions(ctx context.Context) ([]domain.Suppression, error)
	AddExport(ctx context.Context, record domain.ExportRecord) error
	AddAuditEvent(ctx context.Context, event domain.AuditEvent) error
}
