// SPDX-License-Identifier: AGPL-3.0-only

package pgstore

import (
	"context"
	"database/sql"

	"github.com/waymore/spyber/internal/domain"
)

func (s *Store) UpsertContact(ctx context.Context, contact domain.Contact) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO contacts (
			id, company_id, email, contact_type, status, source_url, first_seen_at, last_seen_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (company_id, email) DO UPDATE SET
			contact_type = EXCLUDED.contact_type,
			status = CASE
				WHEN contacts.status IN ('approved','rejected','suppressed') THEN contacts.status
				ELSE EXCLUDED.status
			END,
			source_url = EXCLUDED.source_url,
			last_seen_at = EXCLUDED.last_seen_at`,
		contact.ID, contact.CompanyID, contact.Email, contact.Type, contact.Status,
		contact.SourceURL, contact.FirstSeenAt, contact.LastSeenAt)
	return err
}

func (s *Store) ListContacts(ctx context.Context, countryCode string) ([]domain.Contact, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT c.id, c.company_id, c.email, c.contact_type, c.status, c.source_url, c.first_seen_at, c.last_seen_at
		FROM contacts c
		JOIN companies co ON co.id = c.company_id
		WHERE co.country_code = $1
		ORDER BY c.first_seen_at`, countryCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContacts(rows)
}

func (s *Store) ListCompanyContacts(ctx context.Context, companyID domain.ID) ([]domain.Contact, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, company_id, email, contact_type, status, source_url, first_seen_at, last_seen_at
		FROM contacts WHERE company_id = $1 ORDER BY first_seen_at`, companyID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanContacts(rows)
}

func (s *Store) GetContact(ctx context.Context, id domain.ID) (domain.Contact, bool, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT id, company_id, email, contact_type, status, source_url, first_seen_at, last_seen_at
		FROM contacts WHERE id = $1`, id)
	item, err := scanContact(row)
	if err == sql.ErrNoRows {
		return domain.Contact{}, false, nil
	}
	if err != nil {
		return domain.Contact{}, false, err
	}
	return item, true, nil
}

func scanContacts(rows *sql.Rows) ([]domain.Contact, error) {
	var out []domain.Contact
	for rows.Next() {
		item, err := scanContact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func scanContact(row scanner) (domain.Contact, error) {
	var item domain.Contact
	err := row.Scan(&item.ID, &item.CompanyID, &item.Email, &item.Type, &item.Status, &item.SourceURL, &item.FirstSeenAt, &item.LastSeenAt)
	return item, err
}
