// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"strings"

	"github.com/waymore/spyber/internal/domain"
	"github.com/waymore/spyber/internal/ports"
)

type profileMatch struct {
	Score    int
	Terms    []string
	Excluded bool
}

func scoreCandidateProfile(profile domain.BusinessProfile, candidate ports.BusinessCandidate) profileMatch {
	return scoreProfileText(profile, 0, candidate.Name, candidate.Website, candidate.Evidence)
}

func scorePageProfile(profile domain.BusinessProfile, company domain.Company, analysis ports.PageAnalysis) profileMatch {
	return scoreProfileText(profile, analysis.EcommerceScore, company.Name, company.WebsiteURL, analysis.Text)
}

func scoreProfileText(profile domain.BusinessProfile, ecommerceScore int, parts ...string) profileMatch {
	text := strings.ToLower(strings.Join(parts, " "))
	for _, term := range profile.ExcludeTerms {
		if containsTerm(text, term) {
			return profileMatch{Excluded: true}
		}
	}
	seen := map[string]bool{}
	score := 0
	var terms []string
	for _, term := range profile.IncludeTerms {
		if containsTerm(text, term) && !seen[term] {
			score += termWeight(term, 15)
			seen[term] = true
			terms = append(terms, term)
		}
	}
	for _, term := range profile.DiscoveryTerms {
		if containsTerm(text, term) && !seen[term] {
			score += termWeight(term, 10)
			seen[term] = true
			terms = append(terms, term)
		}
	}
	if profile.Sector == "commerce" {
		score += ecommerceScore / 4
	}
	if profile.Segment == "ecommerce" {
		score += ecommerceScore / 3
	}
	return profileMatch{Score: clampScore(score), Terms: terms}
}

func bestProfileMatch(a, b profileMatch) profileMatch {
	if a.Excluded || b.Excluded {
		return profileMatch{Excluded: true}
	}
	if b.Score > a.Score {
		return b
	}
	return a
}

func containsTerm(text, term string) bool {
	term = strings.ToLower(strings.TrimSpace(term))
	if term == "" {
		return false
	}
	if strings.Contains(term, " ") {
		return strings.Contains(text, term)
	}
	for _, token := range textTokens(text) {
		if token == term {
			return true
		}
	}
	return false
}

func termWeight(term string, base int) int {
	if strings.Contains(term, " ") {
		return base + 5
	}
	return base
}

func textTokens(text string) []string {
	return strings.FieldsFunc(text, func(r rune) bool {
		return !(r >= 'a' && r <= 'z' || r >= '0' && r <= '9')
	})
}
