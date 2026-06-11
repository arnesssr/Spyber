// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"embed"
	"net/http"

	"github.com/waymore/spyber/internal/app"
)

//go:embed templates/*.html static/*.css
var assets embed.FS

type Config struct {
	AdminToken string
}

type Server struct {
	app        *app.App
	adminToken string
	mux        *http.ServeMux
}

func New(application *app.App, cfg Config) http.Handler {
	server := &Server{
		app:        application,
		adminToken: cfg.AdminToken,
		mux:        http.NewServeMux(),
	}
	server.routes()
	return server.security(server.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /", s.dashboard)
	s.mux.HandleFunc("GET /sources", s.sources)
	s.mux.HandleFunc("POST /sources/add", s.addSource)
	s.mux.HandleFunc("POST /discover/domain", s.discoverDomain)
	s.mux.HandleFunc("POST /discover/sources", s.discoverSources)
	s.mux.HandleFunc("POST /crawl", s.crawl)
	s.mux.HandleFunc("POST /scrape", s.scrapeCountry)
	s.mux.HandleFunc("GET /companies", s.companies)
	s.mux.HandleFunc("GET /contacts", s.contacts)
	s.mux.HandleFunc("POST /contacts/verify", s.verifyContacts)
	s.mux.HandleFunc("POST /contacts/approve", s.approveContact)
	s.mux.HandleFunc("POST /contacts/reject", s.rejectContact)
	s.mux.HandleFunc("GET /exports", s.exports)
	s.mux.HandleFunc("POST /exports/download", s.downloadExport)
	s.mux.HandleFunc("GET /suppression", s.suppression)
	s.mux.HandleFunc("POST /suppression/add", s.addSuppression)
	s.mux.Handle("GET /static/", http.FileServerFS(assets))
}
