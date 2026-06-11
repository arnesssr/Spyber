// SPDX-License-Identifier: AGPL-3.0-only

package localstore

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/waymore/spyber/internal/domain"
)

type Store struct {
	path string
	mu   sync.Mutex
}

type state struct {
	Sources      []domain.Source       `json:"sources"`
	Companies    []domain.Company      `json:"companies"`
	Contacts     []domain.Contact      `json:"contacts"`
	CrawlJobs    []domain.CrawlJob     `json:"crawl_jobs"`
	Evidence     []domain.Evidence     `json:"evidence"`
	Suppressions []domain.Suppression  `json:"suppressions"`
	Exports      []domain.ExportRecord `json:"exports"`
	AuditEvents  []domain.AuditEvent   `json:"audit_events"`
}

func New(path string) *Store {
	if path == "" {
		path = ".spyber/spyber.json"
	}
	return &Store{path: path}
}

func (s *Store) Init(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	if _, err := os.Stat(s.path); err == nil {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	return s.saveLocked(state{})
}

func (s *Store) AddSource(ctx context.Context, source domain.Source) error {
	return s.update(ctx, func(st *state) error {
		for i, existing := range st.Sources {
			if existing.CountryCode == source.CountryCode && existing.URL == source.URL {
				source.ID = existing.ID
				source.CreatedAt = existing.CreatedAt
				st.Sources[i] = source
				return nil
			}
		}
		st.Sources = append(st.Sources, source)
		return nil
	})
}

func (s *Store) ListSources(ctx context.Context, countryCode string) ([]domain.Source, error) {
	st, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	var out []domain.Source
	for _, item := range st.Sources {
		if item.CountryCode == countryCode {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *Store) UpsertCompany(ctx context.Context, company domain.Company) error {
	return s.update(ctx, func(st *state) error {
		for i, existing := range st.Companies {
			if existing.NormalizedHost == company.NormalizedHost {
				company.ID = existing.ID
				company.CreatedAt = existing.CreatedAt
				st.Companies[i] = company
				return nil
			}
		}
		st.Companies = append(st.Companies, company)
		return nil
	})
}

func (s *Store) ListCompanies(ctx context.Context, countryCode string) ([]domain.Company, error) {
	st, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	var out []domain.Company
	for _, item := range st.Companies {
		if item.CountryCode == countryCode {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *Store) GetCompany(ctx context.Context, id domain.ID) (domain.Company, bool, error) {
	st, err := s.read(ctx)
	if err != nil {
		return domain.Company{}, false, err
	}
	for _, item := range st.Companies {
		if item.ID == id {
			return item, true, nil
		}
	}
	return domain.Company{}, false, nil
}

func (s *Store) UpsertContact(ctx context.Context, contact domain.Contact) error {
	return s.update(ctx, func(st *state) error {
		for i, existing := range st.Contacts {
			if existing.CompanyID == contact.CompanyID && existing.Email == contact.Email {
				contact.ID = existing.ID
				contact.FirstSeenAt = existing.FirstSeenAt
				if existing.Status == domain.ContactApproved || existing.Status == domain.ContactRejected || existing.Status == domain.ContactSuppressed {
					contact.Status = existing.Status
				}
				st.Contacts[i] = contact
				return nil
			}
		}
		st.Contacts = append(st.Contacts, contact)
		return nil
	})
}

func (s *Store) ListContacts(ctx context.Context, countryCode string) ([]domain.Contact, error) {
	st, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	companyCountry := map[domain.ID]string{}
	for _, company := range st.Companies {
		companyCountry[company.ID] = company.CountryCode
	}
	var out []domain.Contact
	for _, item := range st.Contacts {
		if companyCountry[item.CompanyID] == countryCode {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *Store) GetContact(ctx context.Context, id domain.ID) (domain.Contact, bool, error) {
	st, err := s.read(ctx)
	if err != nil {
		return domain.Contact{}, false, err
	}
	for _, item := range st.Contacts {
		if item.ID == id {
			return item, true, nil
		}
	}
	return domain.Contact{}, false, nil
}

func (s *Store) AddCrawlJob(ctx context.Context, job domain.CrawlJob) error {
	return s.update(ctx, func(st *state) error {
		for i, existing := range st.CrawlJobs {
			if existing.ID == job.ID {
				st.CrawlJobs[i] = job
				return nil
			}
		}
		st.CrawlJobs = append(st.CrawlJobs, job)
		return nil
	})
}

func (s *Store) ListCrawlJobs(ctx context.Context, countryCode string) ([]domain.CrawlJob, error) {
	st, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	companyCountry := map[domain.ID]string{}
	for _, company := range st.Companies {
		companyCountry[company.ID] = company.CountryCode
	}
	var out []domain.CrawlJob
	for _, item := range st.CrawlJobs {
		if companyCountry[item.CompanyID] == countryCode {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *Store) AddEvidence(ctx context.Context, evidence domain.Evidence) error {
	return s.update(ctx, func(st *state) error {
		st.Evidence = append(st.Evidence, evidence)
		return nil
	})
}

func (s *Store) ListEvidence(ctx context.Context, companyID domain.ID) ([]domain.Evidence, error) {
	st, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	var out []domain.Evidence
	for _, item := range st.Evidence {
		if item.CompanyID == companyID {
			out = append(out, item)
		}
	}
	return out, nil
}

func (s *Store) AddSuppression(ctx context.Context, suppression domain.Suppression) error {
	return s.update(ctx, func(st *state) error {
		for _, existing := range st.Suppressions {
			if existing.Email == suppression.Email {
				return nil
			}
		}
		st.Suppressions = append(st.Suppressions, suppression)
		for i, contact := range st.Contacts {
			if contact.Email == suppression.Email {
				contact.Status = domain.ContactSuppressed
				st.Contacts[i] = contact
			}
		}
		return nil
	})
}

func (s *Store) ListSuppressions(ctx context.Context) ([]domain.Suppression, error) {
	st, err := s.read(ctx)
	if err != nil {
		return nil, err
	}
	return append([]domain.Suppression(nil), st.Suppressions...), nil
}

func (s *Store) AddExport(ctx context.Context, record domain.ExportRecord) error {
	return s.update(ctx, func(st *state) error {
		st.Exports = append(st.Exports, record)
		return nil
	})
}

func (s *Store) AddAuditEvent(ctx context.Context, event domain.AuditEvent) error {
	return s.update(ctx, func(st *state) error {
		st.AuditEvents = append(st.AuditEvents, event)
		return nil
	})
}

func (s *Store) read(ctx context.Context) (state, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return state{}, err
	}
	return s.loadLocked()
}

func (s *Store) update(ctx context.Context, fn func(*state) error) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return err
	}
	st, err := s.loadLocked()
	if err != nil {
		return err
	}
	if err := fn(&st); err != nil {
		return err
	}
	return s.saveLocked(st)
}

func (s *Store) loadLocked() (state, error) {
	data, err := os.ReadFile(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return state{}, nil
	}
	if err != nil {
		return state{}, err
	}
	if len(data) == 0 {
		return state{}, nil
	}
	var st state
	if err := json.Unmarshal(data, &st); err != nil {
		return state{}, err
	}
	return st, nil
}

func (s *Store) saveLocked(st state) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
