// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "time"

type Company struct {
	ID                ID            `json:"id"`
	CountryCode       string        `json:"country_code"`
	Name              string        `json:"name"`
	WebsiteURL        string        `json:"website_url"`
	NormalizedHost    string        `json:"normalized_host"`
	Status            CompanyStatus `json:"status"`
	EcommerceScore    int           `json:"ecommerce_score"`
	CountryConfidence int           `json:"country_confidence"`
	CreatedAt         time.Time     `json:"created_at"`
	UpdatedAt         time.Time     `json:"updated_at"`
}

func NewCompany(countryCode, name, website string, now time.Time) (Company, error) {
	country, err := NormalizeCountry(countryCode)
	if err != nil {
		return Company{}, err
	}
	normalizedURL, host, err := NormalizeWebsite(website)
	if err != nil {
		return Company{}, err
	}
	if name == "" {
		name = host
	}
	now = now.UTC()
	return Company{
		ID:                NewID("cmp"),
		CountryCode:       country,
		Name:              name,
		WebsiteURL:        normalizedURL,
		NormalizedHost:    host,
		Status:            CompanyCandidate,
		CountryConfidence: 50,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}
