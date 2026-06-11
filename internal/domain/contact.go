// SPDX-License-Identifier: AGPL-3.0-only

package domain

import (
	"strings"
	"time"
)

type Contact struct {
	ID          ID            `json:"id"`
	CompanyID   ID            `json:"company_id"`
	Email       string        `json:"email"`
	Type        ContactType   `json:"type"`
	Status      ContactStatus `json:"status"`
	SourceURL   string        `json:"source_url"`
	FirstSeenAt time.Time     `json:"first_seen_at"`
	LastSeenAt  time.Time     `json:"last_seen_at"`
}

func NewContact(companyID ID, email, sourceURL string, now time.Time) (Contact, error) {
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil {
		return Contact{}, err
	}
	normalizedSource, _, err := NormalizeWebsite(sourceURL)
	if err != nil {
		return Contact{}, err
	}
	now = now.UTC()
	return Contact{
		ID:          NewID("con"),
		CompanyID:   companyID,
		Email:       normalizedEmail,
		Type:        ClassifyContactType(normalizedEmail),
		Status:      ContactNeedsReview,
		SourceURL:   normalizedSource,
		FirstSeenAt: now,
		LastSeenAt:  now,
	}, nil
}

func ClassifyContactType(email string) ContactType {
	local := strings.Split(strings.ToLower(email), "@")[0]
	generic := map[string]bool{
		"admin": true, "contact": true, "hello": true, "hi": true,
		"info": true, "sales": true, "support": true, "team": true,
		"wholesale": true, "orders": true, "customerservice": true,
		"service": true, "help": true, "care": true,
	}
	if generic[local] {
		return ContactGeneric
	}
	if strings.Contains(local, ".") || strings.Contains(local, "_") {
		return ContactNamed
	}
	return ContactUnknown
}
