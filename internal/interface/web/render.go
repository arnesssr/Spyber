// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/waymore/spyber/internal/app"
	"github.com/waymore/spyber/internal/domain"
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
}

func (s *Server) render(w http.ResponseWriter, r *http.Request, page string, data pageData) {
	data.Notice = r.URL.Query().Get("notice")
	tpl, err := template.ParseFS(assets, "templates/layout.html", "templates/"+page+".html")
	if err != nil {
		http.Error(w, "template error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, "render error", http.StatusInternalServerError)
	}
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
