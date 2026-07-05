// SPDX-License-Identifier: AGPL-3.0-only

package countryfinders

import (
	"context"
	"fmt"
	"strings"

	"github.com/arnesssr/Spyber/internal/domain"
	"github.com/arnesssr/Spyber/internal/ports"
)

type Multi struct {
	Finders []ports.CountryFinder
}

func New(finders ...ports.CountryFinder) *Multi {
	return &Multi{Finders: finders}
}

func (m *Multi) FindBusinesses(ctx context.Context, countryCode string, limit int) ([]ports.BusinessCandidate, error) {
	return m.search(ctx, ports.BusinessSearch{CountryCode: countryCode, Limit: limit}, false)
}

func (m *Multi) SearchBusinesses(ctx context.Context, search ports.BusinessSearch) ([]ports.BusinessCandidate, error) {
	return m.search(ctx, search, true)
}

func (m *Multi) search(ctx context.Context, search ports.BusinessSearch, preferSearch bool) ([]ports.BusinessCandidate, error) {
	seen := map[string]bool{}
	var out []ports.BusinessCandidate
	var providerErrors []string
	for _, finder := range m.Finders {
		if finder == nil {
			continue
		}
		if search.Limit > 0 && len(out) >= search.Limit {
			return out, nil
		}
		remaining := search.Limit - len(out)
		if search.Limit <= 0 {
			remaining = 0
		}
		candidates, err := findWithProvider(ctx, finder, search, remaining, preferSearch)
		if err != nil {
			providerErrors = append(providerErrors, fmt.Sprintf("%T: %v", finder, err))
			continue
		}
		for _, candidate := range candidates {
			host := candidateHost(candidate.Website)
			if host == "" || seen[host] {
				continue
			}
			seen[host] = true
			out = append(out, candidate)
			if search.Limit > 0 && len(out) >= search.Limit {
				return out, nil
			}
		}
	}
	if len(out) == 0 && len(providerErrors) > 0 {
		return nil, fmt.Errorf("candidate providers failed: %s", strings.Join(providerErrors, "; "))
	}
	return out, nil
}

func findWithProvider(ctx context.Context, finder ports.CountryFinder, search ports.BusinessSearch, limit int, preferSearch bool) ([]ports.BusinessCandidate, error) {
	if preferSearch {
		if searcher, ok := finder.(ports.BusinessSearcher); ok {
			search.Limit = limit
			return searcher.SearchBusinesses(ctx, search)
		}
	}
	return finder.FindBusinesses(ctx, search.CountryCode, limit)
}

func candidateHost(raw string) string {
	_, host, err := domain.NormalizeWebsite(raw)
	if err != nil {
		return ""
	}
	return host
}
