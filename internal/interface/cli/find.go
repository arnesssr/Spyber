// SPDX-License-Identifier: AGPL-3.0-only

package cli

import (
	"context"
	"flag"
	"fmt"

	"github.com/waymore/spyber/internal/app"
	"github.com/waymore/spyber/internal/domain"
)

func (r *runner) find(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("find", flag.ContinueOnError)
	fs.SetOutput(r.err)
	country := fs.String("country", "", "country code")
	sector := fs.String("sector", "commerce", "business sector")
	segment := fs.String("segment", "wholesalers", "business segment")
	query := fs.String("query", "", "custom search term")
	limit := fs.Int("limit", 50, "maximum candidates")
	if err := fs.Parse(args); err != nil {
		return err
	}
	summary, err := r.app.FindBusinesses(ctx, app.FindRequest{
		CountryCode: *country,
		Sector:      *sector,
		Segment:     *segment,
		Query:       *query,
		Limit:       *limit,
	})
	if err != nil {
		return err
	}
	fmt.Fprintf(
		r.out,
		"profile=%s candidates=%d created=%d duplicates=%d matched=%d rejected=%d fetched=%d contacts=%d direct_emails=%d verified=%d failures=%d\n",
		summary.Profile.Key(),
		summary.Candidates,
		summary.Created,
		summary.Duplicates,
		summary.Matched,
		summary.Rejected,
		summary.Fetched,
		summary.Contacts,
		summary.DirectEmails,
		summary.Verified,
		summary.Failures,
	)
	return nil
}

func (r *runner) profiles(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	for _, profile := range domain.BusinessProfiles() {
		fmt.Fprintf(r.out, "%s\t%s\n", profile.Key(), profile.Description)
	}
	return nil
}
