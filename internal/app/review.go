// SPDX-License-Identifier: AGPL-3.0-only

package app

import (
	"context"
	"fmt"

	"github.com/arnesssr/Spyber/internal/domain"
)

func (a *App) ListContacts(ctx context.Context, countryCode string) ([]domain.Contact, error) {
	country, err := domain.NormalizeCountry(countryCode)
	if err != nil {
		return nil, err
	}
	return a.store.ListContacts(ctx, country)
}

func (a *App) VerifyContacts(ctx context.Context, countryCode string) (int, error) {
	contacts, err := a.ListContacts(ctx, countryCode)
	if err != nil {
		return 0, err
	}
	suppressions, err := a.store.ListSuppressions(ctx)
	if err != nil {
		return 0, err
	}
	suppressed := map[string]bool{}
	for _, item := range suppressions {
		suppressed[item.Email] = true
	}
	updated := 0
	for _, contact := range contacts {
		if suppressed[contact.Email] {
			contact.Status = domain.ContactSuppressed
			if err := a.store.UpsertContact(ctx, contact); err == nil {
				updated++
			}
			continue
		}
		if _, err := domain.NormalizeEmail(contact.Email); err != nil {
			contact.Status = domain.ContactRejected
			if err := a.store.UpsertContact(ctx, contact); err == nil {
				updated++
			}
		}
	}
	return updated, nil
}

func (a *App) ApproveContact(ctx context.Context, id domain.ID) (domain.Contact, error) {
	contact, found, err := a.store.GetContact(ctx, id)
	if err != nil {
		return domain.Contact{}, err
	}
	if !found {
		return domain.Contact{}, fmt.Errorf("contact not found: %s", id)
	}
	contact.Status = domain.ContactApproved
	if err := a.store.UpsertContact(ctx, contact); err != nil {
		return domain.Contact{}, err
	}
	a.audit(ctx, "contact.approve", "contact", contact.ID.String(), "{}")
	return contact, nil
}

func (a *App) RejectContact(ctx context.Context, id domain.ID, reason string) (domain.Contact, error) {
	contact, found, err := a.store.GetContact(ctx, id)
	if err != nil {
		return domain.Contact{}, err
	}
	if !found {
		return domain.Contact{}, fmt.Errorf("contact not found: %s", id)
	}
	contact.Status = domain.ContactRejected
	if err := a.store.UpsertContact(ctx, contact); err != nil {
		return domain.Contact{}, err
	}
	if reason == "" {
		reason = "unspecified"
	}
	a.audit(ctx, "contact.reject", "contact", contact.ID.String(), `{"reason":"`+reason+`"}`)
	return contact, nil
}

func (a *App) SuppressEmail(ctx context.Context, email, reason string) (domain.Suppression, error) {
	suppression, err := domain.NewSuppression(email, reason, a.now())
	if err != nil {
		return domain.Suppression{}, err
	}
	if err := a.store.AddSuppression(ctx, suppression); err != nil {
		return domain.Suppression{}, err
	}
	a.audit(ctx, "suppression.add", "suppression", suppression.ID.String(), `{"email":"`+suppression.Email+`"}`)
	return suppression, nil
}
