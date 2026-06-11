// SPDX-License-Identifier: AGPL-3.0-only

package web

import "net/http"

func (s *Server) dashboard(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	stats, err := s.app.DashboardStats(r.Context(), country)
	data := pageData{Title: "Dashboard", Active: "dashboard", Country: country, Stats: stats}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, r, "dashboard", data)
}

func (s *Server) sources(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	stats, _ := s.app.DashboardStats(r.Context(), country)
	sources, err := s.app.ListSources(r.Context(), country)
	data := pageData{Title: "Sources", Active: "sources", Country: country, Stats: stats, Sources: sources}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, r, "sources", data)
}

func (s *Server) companies(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	stats, _ := s.app.DashboardStats(r.Context(), country)
	companies, err := s.app.ListCompanies(r.Context(), country)
	data := pageData{Title: "Companies", Active: "companies", Country: country, Stats: stats, Companies: companies}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, r, "companies", data)
}

func (s *Server) contacts(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	stats, _ := s.app.DashboardStats(r.Context(), country)
	contacts, err := s.app.ListContacts(r.Context(), country)
	data := pageData{Title: "Contacts", Active: "contacts", Country: country, Stats: stats, Contacts: contacts}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, r, "contacts", data)
}

func (s *Server) exports(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	stats, err := s.app.DashboardStats(r.Context(), country)
	data := pageData{Title: "Exports", Active: "exports", Country: country, Stats: stats}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, r, "exports", data)
}

func (s *Server) suppression(w http.ResponseWriter, r *http.Request) {
	country := currentCountry(r)
	stats, _ := s.app.DashboardStats(r.Context(), country)
	items, err := s.app.ListSuppressions(r.Context())
	data := pageData{Title: "Suppression", Active: "suppression", Country: country, Stats: stats, Suppressions: items}
	if err != nil {
		data.Error = err.Error()
	}
	s.render(w, r, "suppression", data)
}
