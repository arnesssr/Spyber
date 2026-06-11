// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"fmt"

	"github.com/waymore/spyber/internal/domain"
)

func (a *App) CreateFindJob(ctx context.Context, req FindRequest) (domain.FindJob, error) {
	sector := req.Sector
	segment := req.Segment
	if sector == "" && req.Query == "" {
		sector = "commerce"
	}
	if segment == "" && req.Query == "" {
		segment = "wholesalers"
	}
	job, err := domain.NewFindJob(req.CountryCode, sector, segment, req.Query, req.Limit, a.now())
	if err != nil {
		return domain.FindJob{}, err
	}
	if err := a.store.UpsertFindJob(ctx, job); err != nil {
		return domain.FindJob{}, err
	}
	a.audit(ctx, "find.create", "find_job", job.ID.String(), `{"country":"`+job.CountryCode+`"}`)
	return job, nil
}

func (a *App) RunFindJob(ctx context.Context, id domain.ID) error {
	job, found, err := a.store.GetFindJob(ctx, id)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("find job %s not found", id)
	}
	if job.Status == domain.JobRunning || job.Status == domain.JobSucceeded {
		return nil
	}
	started := a.now().UTC()
	job.Status = domain.JobRunning
	job.StartedAt = &started
	job.UpdatedAt = started
	if err := a.store.UpsertFindJob(ctx, job); err != nil {
		return err
	}
	summary, runErr := a.FindBusinesses(ctx, FindRequest{
		CountryCode: job.CountryCode,
		Sector:      job.Sector,
		Segment:     job.Segment,
		Query:       job.Query,
		Limit:       job.Limit,
		JobID:       job.ID,
	})
	finished := a.now().UTC()
	job.Status = domain.JobSucceeded
	job.FinishedAt = &finished
	job.UpdatedAt = finished
	job.ProfileKey = summary.Profile.Key()
	job.Candidates = summary.Candidates
	job.Created = summary.Created
	job.Duplicates = summary.Duplicates
	job.Matched = summary.Matched
	job.Rejected = summary.Rejected
	job.Fetched = summary.Fetched
	job.Contacts = summary.Contacts
	job.DirectEmails = summary.DirectEmails
	job.Verified = summary.Verified
	job.Failures = summary.Failures
	if runErr != nil {
		job.Status = domain.JobFailed
		job.FailureReason = runErr.Error()
	}
	if err := a.store.UpsertFindJob(ctx, job); err != nil {
		return err
	}
	if runErr != nil {
		return runErr
	}
	return nil
}

func (a *App) ListFindJobs(ctx context.Context, countryCode string) ([]domain.FindJob, error) {
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return nil, err
	}
	return a.store.ListFindJobs(ctx, country)
}
