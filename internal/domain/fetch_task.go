// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "time"

type FetchPurpose string

const (
	FetchRoot      FetchPurpose = "root"
	FetchCandidate FetchPurpose = "candidate"
	FetchContact   FetchPurpose = "contact"
	FetchAbout     FetchPurpose = "about"
	FetchSitemap   FetchPurpose = "sitemap"
)

type FetchTask struct {
	ID            ID           `json:"id"`
	FindJobID     ID           `json:"find_job_id"`
	CompanyID     ID           `json:"company_id"`
	URL           string       `json:"url"`
	Purpose       FetchPurpose `json:"purpose"`
	Status        JobStatus    `json:"status"`
	Attempts      int          `json:"attempts"`
	StatusCode    int          `json:"status_code"`
	Bytes         int          `json:"bytes"`
	EmailCount    int          `json:"email_count"`
	LinkCount     int          `json:"link_count"`
	FailureReason string       `json:"failure_reason"`
	StartedAt     *time.Time   `json:"started_at,omitempty"`
	FinishedAt    *time.Time   `json:"finished_at,omitempty"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
}

func NewFetchTask(findJobID, companyID ID, rawURL string, purpose FetchPurpose, now time.Time) (FetchTask, error) {
	normalizedURL, _, err := NormalizeWebsite(rawURL)
	if err != nil {
		return FetchTask{}, err
	}
	now = now.UTC()
	return FetchTask{
		ID:        NewID("ft"),
		FindJobID: findJobID,
		CompanyID: companyID,
		URL:       normalizedURL,
		Purpose:   purpose,
		Status:    JobQueued,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}
