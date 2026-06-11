// SPDX-License-Identifier: AGPL-3.0-only

package web

import (
	"crypto/subtle"
	"net/http"
	"net/url"
)

func (s *Server) security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "same-origin")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; form-action 'self'; base-uri 'none'")
		if s.adminToken != "" && !s.authorized(r) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Spyber"`)
			http.Error(w, "authentication required", http.StatusUnauthorized)
			return
		}
		if r.Method == http.MethodPost && crossSite(r) {
			http.Error(w, "cross-site form rejected", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) authorized(r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	if !ok || user != "admin" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(pass), []byte(s.adminToken)) == 1
}

func crossSite(r *http.Request) bool {
	if r.Header.Get("Sec-Fetch-Site") == "cross-site" {
		return true
	}
	origin := r.Header.Get("Origin")
	if origin == "" {
		return false
	}
	parsed, err := url.Parse(origin)
	if err != nil {
		return true
	}
	return parsed.Host != r.Host
}
