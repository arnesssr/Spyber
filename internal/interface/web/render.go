// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"html/template"
	"net/http"
	"strconv"
	"strings"

	"github.com/arnesssr/Spyber/internal/app"
	"github.com/arnesssr/Spyber/internal/domain"
)

type pageData struct {
	Title        string
	Active       string
	Country      string
	Notice       string
	Error        string
	Stats        app.DashboardStats
	Sources      []domain.Source
	Companies    []domain.Company
	Contacts     []domain.Contact
	Suppressions []domain.Suppression
	Profiles     []domain.BusinessProfile
	FindJobs     []domain.FindJob
	AutoRefresh  bool
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, page string, data pageData) {
	data.Notice = r.URL.Query().Get("notice")
	tpl, err := template.New("layout.html").Funcs(template.FuncMap{
		"jobProgress":      jobProgress,
		"jobProcessed":     jobProcessed,
		"jobProgressTotal": jobProgressTotal,
		"jobProgressText":  jobProgressText,
		"jobCrawlMode":     jobCrawlMode,
	}).ParseFS(assets, "templates/layout.html", "templates/"+page+".html")
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
}

func jobProgress(job domain.FindJob) int {
	if job.Status == domain.JobSucceeded || job.Status == domain.JobFailed {
		return 100
	}
	total := jobProgressTotal(job)
	if total <= 0 {
		if job.Status == domain.JobRunning {
			return 5
		}
		return 0
	}
	processed := jobProcessed(job)
	if processed <= 0 {
		if job.Status == domain.JobRunning {
			return 5
		}
		return 0
	}
	if processed > total {
		processed = total
	}
	progress := processed * 100 / total
	if job.Status == domain.JobRunning && progress >= 100 {
		return 99
	}
	return progress
}

func jobProcessed(job domain.FindJob) int {
	processed := job.Matched + job.Rejected + job.Duplicates
	if processed < 0 {
		return 0
	}
	return processed
}

func jobProgressTotal(job domain.FindJob) int {
	if job.Candidates > 0 {
		return job.Candidates
	}
	return job.Limit
}

func jobProgressText(job domain.FindJob) string {
	progress := jobProgress(job)
	if job.Candidates > 0 {
		return template.HTMLEscapeString(
			strconv.Itoa(progress) + "% - " +
				strconv.Itoa(jobProcessed(job)) + "/" +
				strconv.Itoa(job.Candidates) + " candidates",
		)
	}
	if job.Status == domain.JobRunning {
		return strconv.Itoa(progress) + "% - discovering"
	}
	if job.Status == domain.JobQueued {
		return "0% - queued"
	}
	return strconv.Itoa(progress) + "% - complete"
}

func jobCrawlMode(job domain.FindJob) string {
	return domain.NormalizeCrawlMode(job.CrawlMode)
}

func currentCountry(r *http.Request) string {
	country := r.FormValue("country")
	if country == "" {
		country = r.URL.Query().Get("country")
	}
	if country == "" {
		country = "GB"
	}
	return strings.ToUpper(strings.TrimSpace(country))
}

func redirectBack(w http.ResponseWriter, r *http.Request, path, country, notice string) {
	query := "?country=" + template.URLQueryEscaper(country)
	if notice != "" {
		query += "&notice=" + template.URLQueryEscaper(notice)
	}
	http.Redirect(w, r, path+query, http.StatusSeeOther)
}
