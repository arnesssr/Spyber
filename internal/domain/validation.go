// SPDX-License-Identifier: AGPL-3.0-only

package domain

import (
	"errors"
	"net/mail"
	"net/url"
	"strings"
)

var (
	ErrInvalidCountry = errors.New("country code must be two letters")
	ErrInvalidEmail   = errors.New("email is invalid")
	ErrInvalidURL     = errors.New("url must be http or https")
)

func NormalizeCountry(code string) (string, error) {
	code = strings.ToUpper(strings.TrimSpace(code))
	if len(code) != 2 {
		return "", ErrInvalidCountry
	}
	for _, r := range code {
		if r < 'A' || r > 'Z' {
			return "", ErrInvalidCountry
		}
	}
	return code, nil
}

func NormalizeEmail(email string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return "", ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(email)
	if err != nil || addr.Address != email || !strings.Contains(email, "@") {
		return "", ErrInvalidEmail
	}
	return email, nil
}

func NormalizeWebsite(raw string) (normalizedURL string, host string, err error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "", "", ErrInvalidURL
	}
	if !strings.Contains(raw, "://") {
		raw = "https://" + raw
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", ErrInvalidURL
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", "", ErrInvalidURL
	}
	if parsed.Hostname() == "" {
		return "", "", ErrInvalidURL
	}
	parsed.Fragment = ""
	host = strings.ToLower(parsed.Hostname())
	return parsed.String(), host, nil
}
