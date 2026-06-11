// SPDX-License-Identifier: AGPL-3.0-only

package ports

import "context"

type BusinessCandidate struct {
	Name      string
	Website   string
	Email     string
	SourceURL string
	Evidence  string
}

type CountryFinder interface {
	FindBusinesses(ctx context.Context, countryCode string, limit int) ([]BusinessCandidate, error)
}
