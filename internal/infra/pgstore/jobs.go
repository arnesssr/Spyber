// SPDX-License-Identifier: AGPL-3.0-only

package pgstore

import (
	"context"
	"database/sql"
	"time"

	"github.com/arnesssr/Spyber/internal/domain"
)

func (s *Store) AddCrawlJob(ctx context.Context, job domain.CrawlJob) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO crawl_jobs (id, company_id, url, status, failure_reason, started_at, finished_at, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			failure_reason = EXCLUDED.failure_reason,
			started_at = EXCLUDED.started_at,
			finished_at = EXCLUDED.finished_at`,
		job.ID, job.CompanyID, job.URL, job.Status, job.FailureReason, job.StartedAt, job.FinishedAt, job.CreatedAt)
	return err
}

func (s *Store) ListCrawlJobs(ctx context.Context, countryCode string) ([]domain.CrawlJob, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT j.id, j.company_id, j.url, j.status, j.failure_reason, j.started_at, j.finished_at, j.created_at
		FROM crawl_jobs j
		JOIN companies c ON c.id = j.company_id
		WHERE c.country_code = $1
		ORDER BY j.created_at`, countryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.CrawlJob
	for rows.Next() {
		item, err := scanCrawlJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) UpsertFindJob(ctx context.Context, job domain.FindJob) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO find_jobs (
			id, country_code, sector, segment, query, limit_count, crawl_mode, status, profile_key,
			candidates, created, duplicates, matched, rejected, fetched, contacts,
			direct_emails, verified, failures, failure_reason,
			started_at, finished_at, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24)
		ON CONFLICT (id) DO UPDATE SET
			crawl_mode = EXCLUDED.crawl_mode,
			status = EXCLUDED.status,
			profile_key = EXCLUDED.profile_key,
			candidates = EXCLUDED.candidates,
			created = EXCLUDED.created,
			duplicates = EXCLUDED.duplicates,
			matched = EXCLUDED.matched,
			rejected = EXCLUDED.rejected,
			fetched = EXCLUDED.fetched,
			contacts = EXCLUDED.contacts,
			direct_emails = EXCLUDED.direct_emails,
			verified = EXCLUDED.verified,
			failures = EXCLUDED.failures,
			failure_reason = EXCLUDED.failure_reason,
			started_at = EXCLUDED.started_at,
			finished_at = EXCLUDED.finished_at,
			updated_at = EXCLUDED.updated_at`,
		findJobArgs(job)...)
	return err
}

func (s *Store) GetFindJob(ctx context.Context, id domain.ID) (domain.FindJob, bool, error) {
	row := s.db.QueryRowContext(ctx, findJobSelect()+` WHERE id = $1`, id)
	item, err := scanFindJob(row)
	if err == sql.ErrNoRows {
		return domain.FindJob{}, false, nil
	}
	if err != nil {
		return domain.FindJob{}, false, err
	}
	return item, true, nil
}

func (s *Store) ListFindJobs(ctx context.Context, countryCode string) ([]domain.FindJob, error) {
	rows, err := s.db.QueryContext(ctx, findJobSelect()+` WHERE country_code = $1 ORDER BY created_at DESC`, countryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.FindJob
	for rows.Next() {
		item, err := scanFindJob(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanCrawlJob(row scanner) (domain.CrawlJob, error) {
	var item domain.CrawlJob
	var started, finished sql.NullTime
	err := row.Scan(&item.ID, &item.CompanyID, &item.URL, &item.Status, &item.FailureReason, &started, &finished, &item.CreatedAt)
	item.StartedAt = timePtr(started)
	item.FinishedAt = timePtr(finished)
	return item, err
}

func findJobSelect() string {
	return `SELECT id, country_code, sector, segment, query, limit_count, crawl_mode, status, profile_key,
		candidates, created, duplicates, matched, rejected, fetched, contacts,
		direct_emails, verified, failures, failure_reason,
		started_at, finished_at, created_at, updated_at FROM find_jobs`
}

func scanFindJob(row scanner) (domain.FindJob, error) {
	var item domain.FindJob
	var started, finished sql.NullTime
	err := row.Scan(
		&item.ID, &item.CountryCode, &item.Sector, &item.Segment, &item.Query, &item.Limit,
		&item.CrawlMode, &item.Status, &item.ProfileKey, &item.Candidates, &item.Created, &item.Duplicates,
		&item.Matched, &item.Rejected, &item.Fetched, &item.Contacts, &item.DirectEmails,
		&item.Verified, &item.Failures, &item.FailureReason, &started, &finished,
		&item.CreatedAt, &item.UpdatedAt,
	)
	item.StartedAt = timePtr(started)
	item.FinishedAt = timePtr(finished)
	return item, err
}

func findJobArgs(job domain.FindJob) []any {
	return []any{
		job.ID, job.CountryCode, job.Sector, job.Segment, job.Query, job.Limit,
		domain.NormalizeCrawlMode(job.CrawlMode), job.Status, job.ProfileKey, job.Candidates, job.Created, job.Duplicates,
		job.Matched, job.Rejected, job.Fetched, job.Contacts, job.DirectEmails,
		job.Verified, job.Failures, job.FailureReason, job.StartedAt, job.FinishedAt,
		job.CreatedAt, job.UpdatedAt,
	}
}

func timePtr(value sql.NullTime) *time.Time {
	if !value.Valid {
		return nil
	}
	return &value.Time
}
