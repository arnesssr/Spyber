// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"time"

	"github.com/waymore/spyber/internal/domain"
	"github.com/waymore/spyber/internal/ports"
)

type App struct {
	store         ports.Store
	fetcher       ports.Fetcher
	analyzer      ports.Analyzer
	countryFinder ports.CountryFinder
	now           func() time.Time
}

func New(store ports.Store, fetcher ports.Fetcher, analyzer ports.Analyzer) *App {
	return &App{
		store:    store,
		fetcher:  fetcher,
		analyzer: analyzer,
		now:      time.Now,
	}
}

func (a *App) WithCountryFinder(finder ports.CountryFinder) *App {
	a.countryFinder = finder
	return a
}

func (a *App) Init(ctx context.Context) error {
	return a.store.Init(ctx)
}

func (a *App) AddSource(ctx context.Context, countryCode, sourceType, rawURL string) (domain.Source, error) {
	source, err := domain.NewSource(countryCode, sourceType, rawURL, a.now())
	if err != nil {
		return domain.Source{}, err
	}
	if err := a.store.AddSource(ctx, source); err != nil {
		return domain.Source{}, err
	}
	a.audit(ctx, "source.add", "source", source.ID.String(), `{"country":"`+source.CountryCode+`"}`)
	return source, nil
}

func (a *App) ListSources(ctx context.Context, countryCode string) ([]domain.Source, error) {
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return nil, err
	}
	return a.store.ListSources(ctx, country)
}

func (a *App) DiscoverDomain(ctx context.Context, countryCode, website string) (domain.Company, error) {
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return domain.Company{}, err
	}
	_, host, err := domain.NormalizeWebsite(website)
	if err != nil {
		return domain.Company{}, err
	}
	existing, err := a.store.ListCompanies(ctx, country)
	if err != nil {
		return domain.Company{}, err
	}
	for _, company := range existing {
		if company.NormalizedHost == host {
			return company, nil
		}
	}
	company, err := domain.NewCompany(countryCode, "", website, a.now())
	if err != nil {
		return domain.Company{}, err
	}
	if err := a.store.UpsertCompany(ctx, company); err != nil {
		return domain.Company{}, err
	}
	evidence, err := domain.NewEvidence(company.ID, "discovery", "operator_seed", company.WebsiteURL, 70, a.now())
	if err == nil {
		_ = a.store.AddEvidence(ctx, evidence)
	}
	a.audit(ctx, "company.discover", "company", company.ID.String(), `{"host":"`+company.NormalizedHost+`"}`)
	return company, nil
}

func (a *App) ListSuppressions(ctx context.Context) ([]domain.Suppression, error) {
	return a.store.ListSuppressions(ctx)
}

func (a *App) ListCompanies(ctx context.Context, countryCode string) ([]domain.Company, error) {
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return nil, err
	}
	return a.store.ListCompanies(ctx, country)
}

func (a *App) audit(ctx context.Context, action, targetType, targetID, metadata string) {
	event := domain.NewAuditEvent("cli", action, targetType, targetID, metadata, a.now())
	_ = a.store.AddAuditEvent(ctx, event)
}
