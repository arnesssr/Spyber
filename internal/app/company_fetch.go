// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"sync"

	"github.com/waymore/spyber/internal/domain"
	"github.com/waymore/spyber/internal/ports"
)

const defaultFetchWorkers = 10

type companyFetchPlan struct {
	company        domain.Company
	candidate      ports.BusinessCandidate
	candidateMatch profileMatch
	tasks          []domain.FetchTask
}

type companyFetchResult struct {
	fetched      int
	contacts     int
	directEmails int
	failures     int
	matched      bool
}

type fetchedPage struct {
	task     domain.FetchTask
	analysis ports.PageAnalysis
	url      string
}

func (a *App) runCompanyFetchPlans(ctx context.Context, profile domain.BusinessProfile, plans []companyFetchPlan, workers int) <-chan companyFetchResult {
	if workers <= 0 {
		workers = defaultFetchWorkers
	}
	if workers > 50 {
		workers = 50
	}
	in := make(chan companyFetchPlan)
	out := make(chan companyFetchResult)
	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for plan := range in {
				out <- a.runCompanyFetchPlan(ctx, profile, plan)
			}
		}()
	}
	go func() {
		defer close(in)
		for _, plan := range plans {
			select {
			case <-ctx.Done():
				return
			case in <- plan:
			}
		}
	}()
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func (a *App) planCompanyFetches(ctx context.Context, jobID domain.ID, company domain.Company, candidate ports.BusinessCandidate, match profileMatch) companyFetchPlan {
	rawTasks := []struct {
		url     string
		purpose domain.FetchPurpose
	}{
		{siteRoot(company.WebsiteURL), domain.FetchRoot},
		{candidate.Website, domain.FetchCandidate},
		{sitePath(company.WebsiteURL, "/contact"), domain.FetchContact},
		{sitePath(company.WebsiteURL, "/contact-us"), domain.FetchContact},
		{sitePath(company.WebsiteURL, "/about"), domain.FetchAbout},
		{sitePath(company.WebsiteURL, "/sitemap.xml"), domain.FetchSitemap},
	}
	seen := map[string]bool{}
	var tasks []domain.FetchTask
	for _, raw := range rawTasks {
		task, err := domain.NewFetchTask(jobID, company.ID, raw.url, raw.purpose, a.now())
		if err != nil || seen[task.URL] {
			continue
		}
		seen[task.URL] = true
		if task.FindJobID != "" {
			_ = a.store.UpsertFetchTask(ctx, task)
		}
		tasks = append(tasks, task)
	}
	return companyFetchPlan{company: company, candidate: candidate, candidateMatch: match, tasks: tasks}
}

func (a *App) runCompanyFetchPlan(ctx context.Context, profile domain.BusinessProfile, plan companyFetchPlan) companyFetchResult {
	best := plan.candidateMatch
	var result companyFetchResult
	var pages []fetchedPage
	seen := map[string]bool{}
	for i := 0; i < len(plan.tasks); i++ {
		task := plan.tasks[i]
		if seen[task.URL] {
			continue
		}
		seen[task.URL] = true
		page, ok := a.runFetchTask(ctx, task)
		if !ok {
			result.failures++
			continue
		}
		result.fetched++
		pages = append(pages, page)
		best = bestProfileMatch(best, scorePageProfile(profile, plan.company, page.analysis))
		for _, link := range page.analysis.ContactLinks {
			if len(plan.tasks) >= 10 {
				break
			}
			next, err := domain.NewFetchTask(task.FindJobID, plan.company.ID, link, domain.FetchContact, a.now())
			if err == nil && !seen[next.URL] {
				if next.FindJobID != "" {
					_ = a.store.UpsertFetchTask(ctx, next)
				}
				plan.tasks = append(plan.tasks, next)
			}
		}
	}
	result.matched = best.Score >= profile.MinScore && !best.Excluded
	a.addProfileEvidence(ctx, plan.company.ID, profile, plan.company.WebsiteURL, best)
	a.updateCompanyMatch(ctx, plan.company, profile, best)
	if !result.matched {
		return result
	}
	for _, page := range pages {
		result.contacts += a.storeContacts(ctx, plan.company.ID, page.url, page.analysis.Emails)
	}
	if plan.candidate.Email != "" && a.storeDirectEmail(ctx, plan.company.ID, plan.candidate) {
		result.directEmails++
	}
	return result
}

func (a *App) runFetchTask(ctx context.Context, task domain.FetchTask) (fetchedPage, bool) {
	task.Attempts++
	started := a.now().UTC()
	task.Status = domain.JobRunning
	task.StartedAt = &started
	task.UpdatedAt = started
	a.saveFetchTask(ctx, task)
	result, err := a.fetcher.Fetch(ctx, task.URL)
	finished := a.now().UTC()
	task.FinishedAt = &finished
	task.UpdatedAt = finished
	if err != nil {
		task.Status = domain.JobFailed
		task.FailureReason = classifyFetchFailure(err)
		a.saveFetchTask(ctx, task)
		return fetchedPage{}, false
	}
	analysis := a.analyzer.Analyze(result.URL, result.Body)
	task.Status = domain.JobSucceeded
	task.StatusCode = result.StatusCode
	task.Bytes = len(result.Body)
	task.EmailCount = len(analysis.Emails)
	task.LinkCount = len(analysis.CandidateLinks) + len(analysis.ContactLinks)
	a.saveFetchTask(ctx, task)
	return fetchedPage{task: task, analysis: analysis, url: result.URL}, true
}

func (a *App) saveFetchTask(ctx context.Context, task domain.FetchTask) {
	if task.FindJobID != "" {
		_ = a.store.UpsertFetchTask(ctx, task)
	}
}

func (a *App) updateFindJobSummary(ctx context.Context, jobID domain.ID, summary FindSummary) {
	if jobID == "" {
		return
	}
	job, found, err := a.store.GetFindJob(ctx, jobID)
	if err != nil || !found {
		return
	}
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
	job.UpdatedAt = a.now().UTC()
	_ = a.store.UpsertFindJob(ctx, job)
}

func classifyFetchFailure(err error) string {
	msg := strings.ToLower(err.Error())
	checks := []struct {
		needle string
		reason string
	}{
		{"resolve host", "dns_failed"},
		{"deadline", "timeout"},
		{"timeout", "timeout"},
		{"tls", "tls_failed"},
		{"private or local", "blocked_private_host"},
		{"status 403", "http_403"},
		{"status 404", "http_404"},
		{"status 429", "http_429"},
		{"status 5", "http_5xx"},
		{"response exceeds", "response_too_large"},
	}
	for _, check := range checks {
		if strings.Contains(msg, check.needle) {
			return check.reason + ": " + err.Error()
		}
	}
	return "fetch_failed: " + err.Error()
}

func siteRoot(raw string) string {
	normalized, _, err := domain.NormalizeWebsite(raw)
	if err != nil {
		return raw
	}
	parsed, err := url.Parse(normalized)
	if err != nil {
		return normalized
	}
	return fmt.Sprintf("%s://%s/", parsed.Scheme, parsed.Host)
}

func sitePath(raw, path string) string {
	root := siteRoot(raw)
	parsed, err := url.Parse(root)
	if err != nil {
		return raw
	}
	parsed.Path = path
	parsed.RawQuery = ""
	parsed.Fragment = ""
	return parsed.String()
}
