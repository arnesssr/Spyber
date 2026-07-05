// SPDX-License-Identifier: AGPL-3.0-only

package countryfinders

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/arnesssr/Spyber/internal/ports"
)

type failingFinder struct{}

func (failingFinder) FindBusinesses(ctx context.Context, countryCode string, limit int) ([]ports.BusinessCandidate, error) {
	return nil, fmt.Errorf("provider down")
}

type duplicateFinder struct{}

func (duplicateFinder) FindBusinesses(ctx context.Context, countryCode string, limit int) ([]ports.BusinessCandidate, error) {
	return []ports.BusinessCandidate{
		{Website: "https://shop.example/contact"},
		{Website: "https://shop.example/about"},
	}, nil
}

func TestMultiReportsProviderErrorsWhenEmpty(t *testing.T) {
	_, err := New(failingFinder{}).FindBusinesses(context.Background(), "KE", 5)
	if err == nil || !strings.Contains(err.Error(), "provider down") {
		t.Fatalf("expected provider error, got %v", err)
	}
}

func TestMultiDedupesByHost(t *testing.T) {
	candidates, err := New(duplicateFinder{}).FindBusinesses(context.Background(), "KE", 5)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if len(candidates) != 1 {
		t.Fatalf("expected 1 candidate, got %+v", candidates)
	}
}
