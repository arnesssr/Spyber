// SPDX-License-Identifier: AGPL-3.0-only

package pgstore

import (
	"context"
	"database/sql"

	"github.com/arnesssr/Spyber/internal/domain"
)

func (s *Store) AddSource(ctx context.Context, source domain.Source) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO sources (id, country_code, source_type, url, status, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)
		ON CONFLICT (country_code, url)
		DO UPDATE SET source_type = EXCLUDED.source_type, status = EXCLUDED.status`,
		source.ID, source.CountryCode, source.Type, source.URL, source.Status, source.CreatedAt)
	return err
}

func (s *Store) ListSources(ctx context.Context, countryCode string) ([]domain.Source, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, country_code, source_type, url, status, created_at
		FROM sources WHERE country_code = $1 ORDER BY created_at`, countryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Source
	for rows.Next() {
		item, err := scanSource(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) UpsertCompany(ctx context.Context, company domain.Company) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO companies (
			id, country_code, name, website_url, normalized_host, status,
			ecommerce_score, country_confidence, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
		ON CONFLICT (normalized_host) DO UPDATE SET
			country_code = EXCLUDED.country_code,
			name = EXCLUDED.name,
			website_url = EXCLUDED.website_url,
			status = EXCLUDED.status,
			ecommerce_score = EXCLUDED.ecommerce_score,
			country_confidence = EXCLUDED.country_confidence,
			updated_at = EXCLUDED.updated_at`,
		company.ID, company.CountryCode, company.Name, company.WebsiteURL,
		company.NormalizedHost, company.Status, company.EcommerceScore,
		company.CountryConfidence, company.CreatedAt, company.UpdatedAt)
	return err
}

func (s *Store) ListCompanies(ctx context.Context, countryCode string) ([]domain.Company, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, country_code, name, website_url, normalized_host, status,
			ecommerce_score, country_confidence, created_at, updated_at
		FROM companies WHERE country_code = $1 ORDER BY created_at`, countryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Company
	for rows.Next() {
		item, err := scanCompany(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) GetCompany(ctx context.Context, id domain.ID) (domain.Company, bool, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, country_code, name, website_url, normalized_host, status,
			ecommerce_score, country_confidence, created_at, updated_at
		FROM companies WHERE id = $1`, id)
	item, err := scanCompany(row)
	if err == sql.ErrNoRows {
		return domain.Company{}, false, nil
	}
	if err != nil {
		return domain.Company{}, false, err
	}
	return item, true, nil
}

type scanner interface {
	Scan(dest ...any) error
}

func scanSource(row scanner) (domain.Source, error) {
	var item domain.Source
	err := row.Scan(&item.ID, &item.CountryCode, &item.Type, &item.URL, &item.Status, &item.CreatedAt)
	return item, err
}

func scanCompany(row scanner) (domain.Company, error) {
	var item domain.Company
	err := row.Scan(
		&item.ID, &item.CountryCode, &item.Name, &item.WebsiteURL, &item.NormalizedHost,
		&item.Status, &item.EcommerceScore, &item.CountryConfidence, &item.CreatedAt, &item.UpdatedAt,
	)
	return item, err
}
