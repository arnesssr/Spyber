// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "time"

type Evidence struct {
	ID         ID        `json:"id"`
	CompanyID  ID        `json:"company_id"`
	Type       string    `json:"type"`
	Value      string    `json:"value"`
	SourceURL  string    `json:"source_url"`
	Confidence int       `json:"confidence"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewEvidence(companyID ID, evidenceType, value, sourceURL string, confidence int, now time.Time) (Evidence, error) {
	normalizedSource, _, err := NormalizeWebsite(sourceURL)
	if err != nil {
		return Evidence{}, err
	}
	if confidence < 0 {
		confidence = 0
	}
	if confidence > 100 {
		confidence = 100
	}
	return Evidence{
		ID:         NewID("evd"),
		CompanyID:  companyID,
		Type:       evidenceType,
		Value:      value,
		SourceURL:  normalizedSource,
		Confidence: confidence,
		CreatedAt:  now.UTC(),
	}, nil
}
