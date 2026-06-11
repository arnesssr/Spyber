// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"testing"

	"github.com/waymore/spyber/internal/domain"
)

func TestProfileMatchingUsesWholeTerms(t *testing.T) {
	profile, err := domain.FindBusinessProfile("services", "salons")
	if err != nil {
		t.Fatalf("profile: %v", err)
	}
	match := scoreProfileText(profile, 0, "public dialogue on x spaces")
	if match.Score != 0 {
		t.Fatalf("spa should not match inside spaces: %+v", match)
	}
	match = scoreProfileText(profile, 0, "hair-salon booking and beauty services")
	if match.Score < profile.MinScore {
		t.Fatalf("expected salon match, got %+v", match)
	}
}
