// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/waymore/spyber/internal/app"
	"github.com/waymore/spyber/internal/domain"
)

func (s *Server) addSource(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	_, err := s.app.AddSource(r.Context(), country, r.FormValue("type"), r.FormValue("url"))
	notice := "source added"
	if err != nil {
		notice = err.Error()
	}
	redirectBack(w, r, "/sources", country, notice)
}

func (s *Server) discoverDomain(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	_, err := s.app.DiscoverDomain(r.Context(), country, r.FormValue("domain"))
	notice := "company discovered"
	if err != nil {
		notice = err.Error()
	}
	redirectBack(w, r, "/companies", country, notice)
}

func (s *Server) discoverSources(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	summary, err := s.app.DiscoverFromSources(r.Context(), country, 100)
	notice := fmt.Sprintf("created %d companies from %d sources", summary.Created, summary.Sources)
	if err != nil {
		notice = err.Error()
	}
	redirectBack(w, r, "/sources", country, notice)
}

func (s *Server) crawl(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	summary, err := s.app.CrawlCountry(r.Context(), country)
	notice := fmt.Sprintf("fetched %d pages, found %d contacts", summary.Fetched, summary.Contacts)
	if err != nil {
		notice = err.Error()
	}
	redirectBack(w, r, "/", country, notice)
}

func (s *Server) scrapeCountry(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	limit := formInt(r, "limit", 20)
	summary, err := s.app.ScrapeCountry(r.Context(), country, limit)
	notice := fmt.Sprintf("discovered %d, fetched %d pages, found %d contacts", summary.Discovered, summary.Fetched, summary.Contacts+summary.DirectEmails)
	if err != nil {
		notice = err.Error()
	}
	redirectBack(w, r, "/", country, notice)
}

func (s *Server) findBusinesses(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	sector, segment := profileParts(r.FormValue("profile"))
	job, err := s.app.CreateFindJob(r.Context(), app.FindRequest{
		CountryCode: country,
		Sector:      sector,
		Segment:     segment,
		Query:       r.FormValue("query"),
		Limit:       formInt(r, "limit", 50),
	})
	notice := "find job queued"
	if err != nil {
		notice = err.Error()
		redirectBack(w, r, "/", country, notice)
		return
	}
	if !s.enqueueFindJob(job.ID) {
		notice = "find queue is full"
	}
	redirectBack(w, r, "/jobs", country, notice)
}

func (s *Server) enqueueFindJob(id domain.ID) bool {
	select {
	case s.findQueue <- id:
		return true
	default:
		return false
	}
}

func profileParts(value string) (string, string) {
	parts := strings.SplitN(value, "/", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "commerce", "wholesalers"
}

func formInt(r *http.Request, key string, fallback int) int {
	value, err := strconv.Atoi(r.FormValue(key))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func (s *Server) verifyContacts(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	updated, err := s.app.VerifyContacts(r.Context(), country)
	notice := fmt.Sprintf("verified contacts, updated %d", updated)
	if err != nil {
		notice = err.Error()
	}
	redirectBack(w, r, "/contacts", country, notice)
}

func (s *Server) approveContact(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	_, err := s.app.ApproveContact(r.Context(), domain.ID(r.FormValue("contact_id")))
	notice := "contact approved"
	if err != nil {
		notice = err.Error()
	}
	redirectBack(w, r, "/contacts", country, notice)
}

func (s *Server) rejectContact(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	_, err := s.app.RejectContact(r.Context(), domain.ID(r.FormValue("contact_id")), "operator_rejected")
	notice := "contact rejected"
	if err != nil {
		notice = err.Error()
	}
	redirectBack(w, r, "/contacts", country, notice)
}

func (s *Server) addSuppression(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	_, err := s.app.SuppressEmail(r.Context(), r.FormValue("email"), r.FormValue("reason"))
	notice := "email suppressed"
	if err != nil {
		notice = err.Error()
	}
	redirectBack(w, r, "/suppression", country, notice)
}

func (s *Server) downloadExport(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	data, _, err := s.app.ExportContacts(r.Context(), app.ExportOptions{
		CountryCode: country,
		Format:      "csv",
		Only:        r.FormValue("only"),
	})
	if err != nil {
		redirectBack(w, r, "/exports", country, err.Error())
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", `attachment; filename="spyber-contacts.csv"`)
	_, _ = w.Write(data)
}
