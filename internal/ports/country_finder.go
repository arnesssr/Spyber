// SPDX-License-Identifier: AGPL-3.0-only

package ports

import "context"

type BusinessCandidate struct {
	Name      string
	Website   string
	Email     string
	SourceURL string
	Evidence  string
	Provider  string
}

type BusinessSearch struct {
	CountryCode string
	Terms       []string
	Limit       int
}

type CountryFinder interface {
	FindBusinesses(ctx context.Context, countryCode string, limit int) ([]BusinessCandidate, error)
}

type BusinessSearcher interface {
	SearchBusinesses(ctx context.Context, search BusinessSearch) ([]BusinessCandidate, error)
}
