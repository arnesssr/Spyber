// SPDX-License-Identifier: AGPL-3.0-only

package pgstore

import (
	"context"
	"database/sql"

	"github.com/arnesssr/Spyber/internal/domain"
)

func (s *Store) UpsertFetchTask(ctx context.Context, task domain.FetchTask) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO fetch_tasks (
			id, find_job_id, company_id, url, purpose, status, attempts, status_code,
			bytes, email_count, link_count, failure_reason, started_at, finished_at,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status,
			attempts = EXCLUDED.attempts,
			status_code = EXCLUDED.status_code,
			bytes = EXCLUDED.bytes,
			email_count = EXCLUDED.email_count,
			link_count = EXCLUDED.link_count,
			failure_reason = EXCLUDED.failure_reason,
			started_at = EXCLUDED.started_at,
			finished_at = EXCLUDED.finished_at,
			updated_at = EXCLUDED.updated_at`,
		task.ID, task.FindJobID, task.CompanyID, task.URL, task.Purpose, task.Status,
		task.Attempts, task.StatusCode, task.Bytes, task.EmailCount, task.LinkCount,
		task.FailureReason, task.StartedAt, task.FinishedAt, task.CreatedAt, task.UpdatedAt)
	return err
}

func (s *Store) ListFetchTasks(ctx context.Context, findJobID domain.ID) ([]domain.FetchTask, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, find_job_id, company_id, url, purpose, status, attempts, status_code,
			bytes, email_count, link_count, failure_reason, started_at, finished_at,
			created_at, updated_at
		FROM fetch_tasks WHERE find_job_id = $1 ORDER BY created_at`, findJobID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.FetchTask
	for rows.Next() {
		item, err := scanFetchTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanFetchTask(row scanner) (domain.FetchTask, error) {
	var item domain.FetchTask
	var started, finished sql.NullTime
	err := row.Scan(
		&item.ID, &item.FindJobID, &item.CompanyID, &item.URL, &item.Purpose,
		&item.Status, &item.Attempts, &item.StatusCode, &item.Bytes,
		&item.EmailCount, &item.LinkCount, &item.FailureReason, &started,
		&finished, &item.CreatedAt, &item.UpdatedAt,
	)
	item.StartedAt = timePtr(started)
	item.FinishedAt = timePtr(finished)
	return item, err
}
