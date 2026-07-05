// SPDX-License-Identifier: AGPL-3.0-only

package pgstore

import (
	"context"

	"github.com/arnesssr/Spyber/internal/domain"
)

func (s *Store) AddEvidence(ctx context.Context, evidence domain.Evidence) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO evidence (id, company_id, evidence_type, value, source_url, confidence, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		evidence.ID, evidence.CompanyID, evidence.Type, evidence.Value,
		evidence.SourceURL, evidence.Confidence, evidence.CreatedAt)
	return err
}

func (s *Store) ListEvidence(ctx context.Context, companyID domain.ID) ([]domain.Evidence, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, company_id, evidence_type, value, source_url, confidence, created_at
		FROM evidence WHERE company_id = $1 ORDER BY created_at`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Evidence
	for rows.Next() {
		var item domain.Evidence
		if err := rows.Scan(&item.ID, &item.CompanyID, &item.Type, &item.Value, &item.SourceURL, &item.Confidence, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) AddSuppression(ctx context.Context, suppression domain.Suppression) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO suppression (id, email, reason, created_at)
		VALUES ($1,$2,$3,$4)
		ON CONFLICT (email) DO NOTHING`,
		suppression.ID, suppression.Email, suppression.Reason, suppression.CreatedAt); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `UPDATE contacts SET status = 'suppressed' WHERE email = $1`, suppression.Email); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Store) ListSuppressions(ctx context.Context) ([]domain.Suppression, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, email, reason, created_at FROM suppression ORDER BY created_at`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Suppression
	for rows.Next() {
		var item domain.Suppression
		if err := rows.Scan(&item.ID, &item.Email, &item.Reason, &item.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (s *Store) AddExport(ctx context.Context, record domain.ExportRecord) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO exports (id, country_code, format, filters, row_count, created_at)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		record.ID, record.CountryCode, record.Format, record.Filters, record.RowCount, record.CreatedAt)
	return err
}

func (s *Store) AddAuditEvent(ctx context.Context, event domain.AuditEvent) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO audit_events (id, actor, action, target_type, target_id, metadata, created_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		event.ID, event.Actor, event.Action, event.TargetType, event.TargetID, event.Metadata, event.CreatedAt)
	return err
}
