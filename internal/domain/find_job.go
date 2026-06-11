// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "time"

type FindJob struct {
	ID            ID         `json:"id"`
	CountryCode   string     `json:"country_code"`
	Sector        string     `json:"sector"`
	Segment       string     `json:"segment"`
	Query         string     `json:"query"`
	Limit         int        `json:"limit"`
	Status        JobStatus  `json:"status"`
	ProfileKey    string     `json:"profile_key"`
	Candidates    int        `json:"candidates"`
	Created       int        `json:"created"`
	Duplicates    int        `json:"duplicates"`
	Matched       int        `json:"matched"`
	Rejected      int        `json:"rejected"`
	Fetched       int        `json:"fetched"`
	Contacts      int        `json:"contacts"`
	DirectEmails  int        `json:"direct_emails"`
	Verified      int        `json:"verified"`
	Failures      int        `json:"failures"`
	FailureReason string     `json:"failure_reason"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func NewFindJob(countryCode, sector, segment, query string, limit int, now time.Time) (FindJob, error) {
	country, err := NormalizeCountry(countryCode)
	if err != nil {
		return FindJob{}, err
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 1000 {
		limit = 1000
	}
	now = now.UTC()
	return FindJob{
		ID:          NewID("find"),
		CountryCode: country,
		Sector:      sector,
		Segment:     segment,
		Query:       query,
		Limit:       limit,
		Status:      JobQueued,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func (j FindJob) Request() (sector, segment, query string, limit int) {
	return j.Sector, j.Segment, j.Query, j.Limit
}
