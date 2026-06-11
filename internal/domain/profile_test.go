// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "testing"

func TestFindBusinessProfile(t *testing.T) {
	profile, err := FindBusinessProfile("commerce", "wholesalers")
	if err != nil {
		t.Fatalf("find profile: %v", err)
	}
	if profile.Key() != "commerce/wholesalers" {
		t.Fatalf("unexpected key: %s", profile.Key())
	}
	if len(profile.DiscoveryTerms) == 0 || profile.MinScore == 0 {
		t.Fatalf("profile should define discovery and scoring rules: %+v", profile)
	}
}

func TestCustomBusinessProfile(t *testing.T) {
	profile, err := CustomBusinessProfile("Hair Salon!")
	if err != nil {
		t.Fatalf("custom profile: %v", err)
	}
	if profile.Key() != "services/salons" {
		t.Fatalf("unexpected key: %s", profile.Key())
	}
	if len(profile.IncludeTerms) < 5 {
		t.Fatalf("unexpected terms: %+v", profile.IncludeTerms)
	}
}

func TestExpandIntentTermsUnderstandsCommerce(t *testing.T) {
	profile, err := CustomBusinessProfile("shop")
	if err != nil {
		t.Fatalf("custom profile: %v", err)
	}
	if profile.Key() != "commerce/retailers" {
		t.Fatalf("unexpected key: %s", profile.Key())
	}
	if !containsProfileTerm(profile.IncludeTerms, "checkout") {
		t.Fatalf("expected commerce expansion, got %+v", profile.IncludeTerms)
	}
}

func containsProfileTerm(terms []string, want string) bool {
	for _, term := range terms {
		if term == want {
			return true
		}
	}
	return false
}
