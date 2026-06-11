// SPDX-License-Identifier: AGPL-3.0-only

package countryfinders

import (
	"context"

	"github.com/waymore/spyber/internal/ports"
)

type Multi struct {
	Finders []ports.CountryFinder
}

func New(finders ...ports.CountryFinder) *Multi {
	return &Multi{Finders: finders}
}

func (m *Multi) FindBusinesses(ctx context.Context, countryCode string, limit int) ([]ports.BusinessCandidate, error) {
	seen := map[string]bool{}
	var out []ports.BusinessCandidate
	for _, finder := range m.Finders {
		if finder == nil {
			continue
		}
		remaining := limit - len(out)
		if limit <= 0 {
			remaining = 0
		}
		candidates, err := finder.FindBusinesses(ctx, countryCode, remaining)
		if err != nil {
			continue
		}
		for _, candidate := range candidates {
			if candidate.Website == "" || seen[candidate.Website] {
				continue
			}
			seen[candidate.Website] = true
			out = append(out, candidate)
			if limit > 0 && len(out) >= limit {
				return out, nil
			}
		}
	}
	return out, nil
}
