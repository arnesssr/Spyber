// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "time"

type Suppression struct {
	ID        ID        `json:"id"`
	Email     string    `json:"email"`
	Reason    string    `json:"reason"`
	CreatedAt time.Time `json:"created_at"`
}

func NewSuppression(email, reason string, now time.Time) (Suppression, error) {
	normalizedEmail, err := NormalizeEmail(email)
	if err != nil {
		return Suppression{}, err
	}
	if reason == "" {
		reason = "unspecified"
	}
	return Suppression{
		ID:        NewID("sup"),
		Email:     normalizedEmail,
		Reason:    reason,
		CreatedAt: now.UTC(),
	}, nil
}
