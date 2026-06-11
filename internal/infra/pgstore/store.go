// SPDX-License-Identifier: AGPL-3.0-only

package pgstore

import (
	"context"
	"database/sql"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type Store struct {
	db *sql.DB
}

func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, err
	}
	return &Store{db: db}, nil
}

func (s *Store) Init(ctx context.Context) error {
	if err := s.db.PingContext(ctx); err != nil {
		return err
	}
	for _, stmt := range schemaStatements {
		if _, err := s.db.ExecContext(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) Close() error {
	return s.db.Close()
}
