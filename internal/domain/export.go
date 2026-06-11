// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "time"

type ExportRecord struct {
	ID          ID        `json:"id"`
	CountryCode string    `json:"country_code"`
	Format      string    `json:"format"`
	Filters     string    `json:"filters"`
	RowCount    int       `json:"row_count"`
	CreatedAt   time.Time `json:"created_at"`
}

type AuditEvent struct {
	ID         ID        `json:"id"`
	Actor      string    `json:"actor"`
	Action     string    `json:"action"`
	TargetType string    `json:"target_type"`
	TargetID   string    `json:"target_id"`
	Metadata   string    `json:"metadata"`
	CreatedAt  time.Time `json:"created_at"`
}

func NewAuditEvent(actor, action, targetType, targetID, metadata string, now time.Time) AuditEvent {
	if actor == "" {
		actor = "system"
	}
	if metadata == "" {
		metadata = "{}"
	}
	return AuditEvent{
		ID:         NewID("aud"),
		Actor:      actor,
		Action:     action,
		TargetType: targetType,
		TargetID:   targetID,
		Metadata:   metadata,
		CreatedAt:  now.UTC(),
	}
}
