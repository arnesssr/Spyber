// SPDX-License-Identifier: AGPL-3.0-only

package cli

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/waymore/spyber/internal/app"
	"github.com/waymore/spyber/internal/domain"
	"github.com/waymore/spyber/internal/infra/commoncrawl"
	"github.com/waymore/spyber/internal/infra/countryfinders"
	"github.com/waymore/spyber/internal/infra/htmlparse"
	"github.com/waymore/spyber/internal/infra/httpfetch"
	"github.com/waymore/spyber/internal/infra/overpass"
	"github.com/waymore/spyber/internal/infra/storeconfig"
	"github.com/waymore/spyber/internal/version"
)

func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		usage(stderr)
		return 2
	}
	if args[0] == "version" {
		fmt.Fprintf(stdout, "spyber %s\n", version.Version)
		return 0
	}
	runner, err := newRunner(stdout, stderr)
	if err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	if err := runner.run(ctx, args); err != nil {
		fmt.Fprintln(stderr, "error:", err)
		return 1
	}
	return 0
}

type runner struct {
	out        io.Writer
	err        io.Writer
	app        *app.App
	storeLabel string
}

func newRunner(out, err io.Writer) (*runner, error) {
	store, storeErr := storeconfig.Open()
	if storeErr != nil {
		return nil, storeErr
	}
	return &runner{
		out:        out,
		err:        err,
		app:        app.New(store.Store, httpfetch.New(), htmlparse.New()).WithCountryFinder(countryFinder()),
		storeLabel: store.Label,
	}, nil
}

func overpassEndpoint() string {
	return os.Getenv("SPYBER_OVERPASS_ENDPOINT")
}

func commonCrawlIndex() string {
	return os.Getenv("SPYBER_COMMONCRAWL_INDEX")
}

func countryFinder() *countryfinders.Multi {
	return countryfinders.New(
		overpass.New(overpassEndpoint()),
		commoncrawl.New(commonCrawlIndex()),
	)
}

func (r *runner) run(ctx context.Context, args []string) error {
	switch args[0] {
	case "init":
		return r.init(ctx)
	case "version":
		return r.version()
	case "find":
		return r.find(ctx, args[1:])
	case "profiles":
		return r.profiles(ctx)
	case "scrape":
		return r.scrape(ctx, args[1:])
	case "source":
		return r.source(ctx, args[1:])
	case "discover":
		return r.discover(ctx, args[1:])
	case "crawl":
		return r.crawl(ctx, args[1:])
	case "companies":
		return r.companies(ctx, args[1:])
	case "contacts":
		return r.contacts(ctx, args[1:])
	case "review":
		return r.review(ctx, args[1:])
	case "export":
		return r.export(ctx, args[1:])
	case "suppress":
		return r.suppress(ctx, args[1:])
	default:
		usage(r.err)
		return fmt.Errorf("unknown command %q", args[0])
	}
}

func (r *runner) scrape(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("scrape", flag.ContinueOnError)
	fs.SetOutput(r.err)
	country := fs.String("country", "", "country code")
	limit := fs.Int("limit", 50, "maximum country candidates")
	if err := fs.Parse(args); err != nil {
		return err
	}
	summary, err := r.app.ScrapeCountry(ctx, *country, *limit)
	if err != nil {
		return err
	}
	fmt.Fprintf(r.out, "discovered=%d direct_emails=%d crawled=%d fetched=%d contacts=%d verified=%d failures=%d\n", summary.Discovered, summary.DirectEmails, summary.Crawled, summary.Fetched, summary.Contacts, summary.Verified, summary.Failures)
	return nil
}

func (r *runner) init(ctx context.Context) error {
	if err := r.app.Init(ctx); err != nil {
		return err
	}
	fmt.Fprintln(r.out, "initialized", r.storeLabel)
	return nil
}

func (r *runner) version() error {
	fmt.Fprintf(r.out, "spyber %s\n", version.Version)
	return nil
}

func (r *runner) source(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("source requires add or list")
	}
	switch args[0] {
	case "add":
		fs := flag.NewFlagSet("source add", flag.ContinueOnError)
		fs.SetOutput(r.err)
		country := fs.String("country", "", "country code")
		sourceType := fs.String("type", "seed", "source type")
		rawURL := fs.String("url", "", "source URL")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		source, err := r.app.AddSource(ctx, *country, *sourceType, *rawURL)
		if err != nil {
			return err
		}
		fmt.Fprintf(r.out, "%s %s %s\n", source.ID, source.CountryCode, source.URL)
		return nil
	case "list":
		fs := flag.NewFlagSet("source list", flag.ContinueOnError)
		fs.SetOutput(r.err)
		country := fs.String("country", "", "country code")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		sources, err := r.app.ListSources(ctx, *country)
		if err != nil {
			return err
		}
		for _, source := range sources {
			fmt.Fprintf(r.out, "%s\t%s\t%s\t%s\n", source.ID, source.CountryCode, source.Type, source.URL)
		}
		return nil
	default:
		return fmt.Errorf("unknown source command %q", args[0])
	}
}

func (r *runner) discover(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	fs.SetOutput(r.err)
	country := fs.String("country", "", "country code")
	domainFlag := fs.String("domain", "", "website domain or URL")
	fromSources := fs.Bool("from-sources", false, "discover candidate companies from active sources")
	limit := fs.Int("limit", 100, "maximum companies to create from sources")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *fromSources {
		summary, err := r.app.DiscoverFromSources(ctx, *country, *limit)
		if err != nil {
			return err
		}
		fmt.Fprintf(r.out, "sources=%d fetched=%d candidates=%d created=%d skipped=%d failures=%d\n", summary.Sources, summary.Fetched, summary.Candidates, summary.Created, summary.Skipped, summary.Failures)
		return nil
	}
	company, err := r.app.DiscoverDomain(ctx, *country, *domainFlag)
	if err != nil {
		return err
	}
	fmt.Fprintf(r.out, "%s %s %s\n", company.ID, company.CountryCode, company.WebsiteURL)
	return nil
}

func (r *runner) crawl(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("crawl", flag.ContinueOnError)
	fs.SetOutput(r.err)
	country := fs.String("country", "", "country code")
	if err := fs.Parse(args); err != nil {
		return err
	}
	summary, err := r.app.CrawlCountry(ctx, *country)
	if err != nil {
		return err
	}
	fmt.Fprintf(r.out, "companies=%d fetched=%d contacts=%d failures=%d\n", summary.Companies, summary.Fetched, summary.Contacts, summary.Failures)
	return nil
}

func (r *runner) companies(ctx context.Context, args []string) error {
	if len(args) == 0 || args[0] != "list" {
		return fmt.Errorf("companies requires list")
	}
	fs := flag.NewFlagSet("companies list", flag.ContinueOnError)
	fs.SetOutput(r.err)
	country := fs.String("country", "", "country code")
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	companies, err := r.app.ListCompanies(ctx, *country)
	if err != nil {
		return err
	}
	for _, company := range companies {
		fmt.Fprintf(r.out, "%s\t%s\t%s\t%d\t%s\n", company.ID, company.CountryCode, company.Status, company.EcommerceScore, company.WebsiteURL)
	}
	return nil
}

func (r *runner) contacts(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("contacts requires list or verify")
	}
	switch args[0] {
	case "list":
		return r.listContacts(ctx, args[1:])
	case "verify":
		return r.verifyContacts(ctx, args[1:])
	default:
		return fmt.Errorf("unknown contacts command %q", args[0])
	}
}

func (r *runner) listContacts(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("contacts list", flag.ContinueOnError)
	fs.SetOutput(r.err)
	country := fs.String("country", "", "country code")
	if err := fs.Parse(args); err != nil {
		return err
	}
	contacts, err := r.app.ListContacts(ctx, *country)
	if err != nil {
		return err
	}
	for _, contact := range contacts {
		fmt.Fprintf(r.out, "%s\t%s\t%s\t%s\t%s\n", contact.ID, contact.Email, contact.Type, contact.Status, contact.SourceURL)
	}
	return nil
}

func (r *runner) verifyContacts(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("contacts verify", flag.ContinueOnError)
	fs.SetOutput(r.err)
	country := fs.String("country", "", "country code")
	if err := fs.Parse(args); err != nil {
		return err
	}
	updated, err := r.app.VerifyContacts(ctx, *country)
	if err != nil {
		return err
	}
	fmt.Fprintf(r.out, "updated=%d\n", updated)
	return nil
}

func (r *runner) review(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("review requires list, approve, or reject")
	}
	switch args[0] {
	case "list":
		return r.listContacts(ctx, args[1:])
	case "approve":
		return r.reviewUpdate(ctx, args[1:], true)
	case "reject":
		return r.reviewUpdate(ctx, args[1:], false)
	default:
		return fmt.Errorf("unknown review command %q", args[0])
	}
}

func (r *runner) reviewUpdate(ctx context.Context, args []string, approve bool) error {
	fs := flag.NewFlagSet("review update", flag.ContinueOnError)
	fs.SetOutput(r.err)
	contactID := fs.String("contact-id", "", "contact ID")
	reason := fs.String("reason", "", "reject reason")
	if err := fs.Parse(args); err != nil {
		return err
	}
	if *contactID == "" {
		return fmt.Errorf("contact-id is required")
	}
	if approve {
		contact, err := r.app.ApproveContact(ctx, domain.ID(*contactID))
		if err != nil {
			return err
		}
		fmt.Fprintf(r.out, "approved %s\n", contact.ID)
		return nil
	}
	contact, err := r.app.RejectContact(ctx, domain.ID(*contactID), *reason)
	if err != nil {
		return err
	}
	fmt.Fprintf(r.out, "rejected %s\n", contact.ID)
	return nil
}

func (r *runner) export(ctx context.Context, args []string) error {
	fs := flag.NewFlagSet("export", flag.ContinueOnError)
	fs.SetOutput(r.err)
	country := fs.String("country", "", "country code")
	format := fs.String("format", "csv", "export format")
	only := fs.String("only", "generic", "generic or all")
	if err := fs.Parse(args); err != nil {
		return err
	}
	data, rows, err := r.app.ExportContacts(ctx, app.ExportOptions{
		CountryCode: *country,
		Format:      *format,
		Only:        *only,
	})
	if err != nil {
		return err
	}
	fmt.Fprint(r.out, string(data))
	fmt.Fprintf(r.err, "exported rows=%d\n", rows)
	return nil
}

func (r *runner) suppress(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("suppress requires add or list")
	}
	switch args[0] {
	case "add":
		fs := flag.NewFlagSet("suppress add", flag.ContinueOnError)
		fs.SetOutput(r.err)
		email := fs.String("email", "", "email address")
		reason := fs.String("reason", "opt_out", "suppression reason")
		if err := fs.Parse(args[1:]); err != nil {
			return err
		}
		item, err := r.app.SuppressEmail(ctx, *email, *reason)
		if err != nil {
			return err
		}
		fmt.Fprintf(r.out, "%s %s\n", item.ID, item.Email)
		return nil
	case "list":
		items, err := r.app.ListSuppressions(ctx)
		if err != nil {
			return err
		}
		for _, item := range items {
			fmt.Fprintf(r.out, "%s\t%s\t%s\n", item.ID, item.Email, item.Reason)
		}
		return nil
	default:
		return fmt.Errorf("unknown suppress command %q", args[0])
	}
}

func usage(w io.Writer) {
	lines := []string{
		"usage: spyber <command> [args]",
		"",
		"commands:",
		"  init",
		"  version",
		"  profiles",
		"  find --country KE --sector commerce --segment wholesalers --limit 50",
		"  find --country KE --query salon --limit 50",
		"  scrape --country KE --limit 50",
		"  source add --country GB --type seed --url https://example.com",
		"  source list --country GB",
		"  discover --country GB --domain https://shop.example",
		"  discover --country GB --from-sources --limit 100",
		"  crawl --country GB",
		"  companies list --country GB",
		"  contacts list --country GB",
		"  contacts verify --country GB",
		"  review list --country GB",
		"  review approve --contact-id con_...",
		"  review reject --contact-id con_... --reason unsuitable",
		"  export --country GB --format csv --only generic",
		"  suppress add --email user@example.com --reason opt_out",
	}
	fmt.Fprintln(w, strings.Join(lines, "\n"))
}
