// SPDX-License-Identifier: AGPL-3.0-only

package domain

import (
	"strings"
	"time"
)

type Source struct {
	ID          ID           `json:"id"`
	CountryCode string       `json:"country_code"`
	Type        string       `json:"type"`
	URL         string       `json:"url"`
	Status      SourceStatus `json:"status"`
	CreatedAt   time.Time    `json:"created_at"`
}

func NewSource(countryCode, sourceType, rawURL string, now time.Time) (Source, error) {
	country, err := NormalizeCountry(countryCode)
	if err != nil {
		return Source{}, err
	}
	normalizedURL, _, err := NormalizeWebsite(rawURL)
	if err != nil {
		return Source{}, err
	}
	if sourceType == "" {
		sourceType = "seed"
	}
	sourceType = strings.ToLower(strings.TrimSpace(sourceType))
	return Source{
		ID:          NewID("src"),
		CountryCode: country,
		Type:        sourceType,
		URL:         normalizedURL,
		Status:      SourceActive,
		CreatedAt:   now.UTC(),
	}, nil
}
