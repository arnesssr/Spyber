// SPDX-License-Identifier: AGPL-3.0-only

package domain

import "time"

const (
	DefaultFindLimit    = 50
	MaxFindLimit        = 1000
	CrawlModeStandard   = "standard"
	CrawlModeDeep       = "deep"
	CrawlModeExhaustive = "exhaustive"
	DefaultCrawlMode    = CrawlModeDeep
)

type CrawlSettings struct {
	Mode               string
	FetchParallelism   int
	MaxPagesPerCompany int
}

type FindJob struct {
	ID            ID         `json:"id"`
	CountryCode   string     `json:"country_code"`
	Sector        string     `json:"sector"`
	Segment       string     `json:"segment"`
	Query         string     `json:"query"`
	Limit         int        `json:"limit"`
	CrawlMode     string     `json:"crawl_mode"`
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
	now = now.UTC()
	return FindJob{
		ID:          NewID("find"),
		CountryCode: country,
		Sector:      sector,
		Segment:     segment,
		Query:       query,
		Limit:       NormalizeFindLimit(limit),
		CrawlMode:   DefaultCrawlMode,
		Status:      JobQueued,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

func NormalizeFindLimit(limit int) int {
	if limit <= 0 {
		return DefaultFindLimit
	}
	if limit > MaxFindLimit {
		return MaxFindLimit
	}
	return limit
}

func NormalizeCrawlMode(mode string) string {
	switch mode {
	case CrawlModeStandard, CrawlModeDeep, CrawlModeExhaustive:
		return mode
	default:
		return DefaultCrawlMode
	}
}

func CrawlSettingsForMode(mode string) CrawlSettings {
	mode = NormalizeCrawlMode(mode)
	switch mode {
	case CrawlModeStandard:
		return CrawlSettings{Mode: mode, FetchParallelism: 10, MaxPagesPerCompany: 20}
	case CrawlModeExhaustive:
		return CrawlSettings{Mode: mode, FetchParallelism: 100, MaxPagesPerCompany: 0}
	default:
		return CrawlSettings{Mode: mode, FetchParallelism: 50, MaxPagesPerCompany: 100}
	}
}

func (j FindJob) Request() (sector, segment, query string, limit int) {
	return j.Sector, j.Segment, j.Query, j.Limit
}
