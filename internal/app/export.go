// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"

	"github.com/waymore/spyber/internal/domain"
)

type ExportOptions struct {
	CountryCode string
	Only        string
	Format      string
}

func (a *App) ExportContacts(ctx context.Context, opts ExportOptions) ([]byte, int, error) {
	country, err := domain.NormalizeCountry(opts.CountryCode)
	if err != nil {
		return nil, 0, err
	}
	if opts.Format == "" {
		opts.Format = "csv"
	}
	if opts.Format != "csv" {
		return nil, 0, fmt.Errorf("unsupported export format: %s", opts.Format)
	}
	contacts, err := a.store.ListContacts(ctx, country)
	if err != nil {
		return nil, 0, err
	}
	companies, err := a.store.ListCompanies(ctx, country)
	if err != nil {
		return nil, 0, err
	}
	eligibleCompanies := exportableCompanies(companies)
	suppressions, err := a.store.ListSuppressions(ctx)
	if err != nil {
		return nil, 0, err
	}
	suppressed := map[string]bool{}
	for _, item := range suppressions {
		suppressed[item.Email] = true
	}
	var out bytes.Buffer
	writer := csv.NewWriter(&out)
	_ = writer.Write([]string{"email", "type", "status", "source_url", "company_id"})
	rows := 0
	for _, contact := range contacts {
		if !exportable(contact, opts.Only, suppressed, eligibleCompanies) {
			continue
		}
		_ = writer.Write([]string{
			contact.Email,
			string(contact.Type),
			string(contact.Status),
			contact.SourceURL,
			contact.CompanyID.String(),
		})
		rows++
	}
	writer.Flush()
	record := domain.ExportRecord{
		ID:          domain.NewID("exp"),
		CountryCode: country,
		Format:      opts.Format,
		Filters:     "only=" + opts.Only,
		RowCount:    rows,
		CreatedAt:   a.now().UTC(),
	}
	_ = a.store.AddExport(ctx, record)
	a.audit(ctx, "export.create", "export", record.ID.String(), fmt.Sprintf(`{"rows":%d}`, rows))
	return out.Bytes(), rows, writer.Error()
}

func exportable(contact domain.Contact, only string, suppressed map[string]bool, eligibleCompanies map[domain.ID]bool) bool {
	if suppressed[contact.Email] || contact.Status == domain.ContactSuppressed {
		return false
	}
	if !eligibleCompanies[contact.CompanyID] {
		return false
	}
	if contact.Status != domain.ContactApproved && contact.Status != domain.ContactNeedsReview {
		return false
	}
	if only == "generic" && contact.Type != domain.ContactGeneric {
		return false
	}
	return true
}

func exportableCompanies(companies []domain.Company) map[domain.ID]bool {
	eligible := map[domain.ID]bool{}
	for _, company := range companies {
		if company.EcommerceScore >= 40 || company.Status == domain.CompanyApproved || company.Status == domain.CompanyReview {
			eligible[company.ID] = true
		}
	}
	return eligible
}
