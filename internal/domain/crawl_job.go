// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "time"

type CrawlJob struct {
	ID            ID         `json:"id"`
	CompanyID     ID         `json:"company_id"`
	URL           string     `json:"url"`
	Status        JobStatus  `json:"status"`
	FailureReason string     `json:"failure_reason"`
	StartedAt     *time.Time `json:"started_at,omitempty"`
	FinishedAt    *time.Time `json:"finished_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
}

func NewCrawlJob(companyID ID, rawURL string, now time.Time) (CrawlJob, error) {
	normalizedURL, _, err := NormalizeWebsite(rawURL)
	if err != nil {
		return CrawlJob{}, err
	}
	return CrawlJob{
		ID:        NewID("job"),
		CompanyID: companyID,
		URL:       normalizedURL,
		Status:    JobQueued,
		CreatedAt: now.UTC(),
	}, nil
}
