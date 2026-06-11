# Spyber

Spyber is a Go-first business discovery engine. Given a country and business
intent, it discovers candidate businesses, crawls public websites, classifies
evidence, extracts public contact channels, and exports reviewable business
contacts with source evidence.

## Scope

Spyber is not a bulk spam tool and does not claim to find every business in a
country. The durable product goal is narrower and testable:

```text
Find a measurable set of public businesses matching an operator's intent,
prove why each result matched, extract public contacts, and prevent duplicate
or suppressed contacts from being exported.
```

Current v1 capabilities:

- country-scoped discovery
- profile-driven business search
- source-page candidate discovery
- autonomous country discovery through public indexes
- public website crawling with per-host delay and safe fetch limits
- segment and ecommerce signal classification
- public email extraction
- generic business email preference
- review and suppression workflow
- auditable CSV export
- Go-rendered operator UI

## Engine Contract

The user should not need to know a URL. The normal workflow is:

```text
country + business intent + limit
-> candidate discovery
-> dedupe
-> crawl
-> segment/ecommerce scoring
-> contact extraction
-> review/export
```

Manual URL entry is an advanced fallback, not the core product.

Every exportable row must have:

- business website
- contact value
- contact type
- source URL
- match status
- suppression check

## Current Algorithms

- **Country discovery:** OpenStreetMap/Overpass shop tags and Common Crawl
  country TLD indexes.
- **Canonicalization:** normalized URLs, hosts, countries, and emails.
- **Deduplication:** companies by normalized host; contacts by company and
  email.
- **Crawl planning:** company entry page plus bounded contact/about/support
  links.
- **Ecommerce scoring:** storefront keywords, platform markers, product/cart
  paths, pricing and catalog signals.
- **Contact extraction:** public email extraction, generic/named/unknown
  classification, source URL retention.
- **Export gating:** suppressed contacts and non-matching businesses are
  excluded.

## Business Profiles

Spyber ships with a small public profile catalog:

```text
Commerce -> Wholesalers
Commerce -> Retailers
Commerce -> Ecommerce
Services -> Salons
```

Each profile defines discovery terms, include terms, exclude terms, and an
acceptance threshold. Custom search terms are also supported for early
exploration, for example `--query salon`.

## Not Done Yet

- Phone extraction is not implemented yet.
- Search-job history is not modeled yet.
- Reviewed precision reporting is not modeled yet.
- Local JSON is a development store, not the production durability target.

## Stack

- Go only for the CLI, engine, and server-rendered UI
- Postgres as the intended production source of truth
- local JSON store for early development
- no TypeScript or frontend build system in v1

## Quick Start

```bash
go test ./...
go run ./cmd/spyber init
go run ./cmd/spyber profiles
go run ./cmd/spyber find --country KE --sector commerce --segment wholesalers --limit 50
go run ./cmd/spyber find --country KE --query salon --limit 50
go run ./cmd/spyber companies list --country KE
go run ./cmd/spyber contacts list --country KE
go run ./cmd/spyber export --country KE --format csv --only generic
```

The default local store is `.spyber/spyber.json`.

Manual source workflow:

```bash
go run ./cmd/spyber source add --country KE --type seed --url https://example.co.ke
go run ./cmd/spyber discover --country KE --from-sources --limit 100
go run ./cmd/spyber crawl --country KE
```

## Web UI

Run the operator UI:

```bash
make run-ui
```

Then open:

```text
http://127.0.0.1:8091
```

Set `SPYBER_ADMIN_TOKEN` to require browser Basic Auth with username `admin`:

```bash
SPYBER_ADMIN_TOKEN=change-me make run-ui
```

The UI is server-rendered Go HTML. It has no TypeScript or frontend build
pipeline.

Use the country field and `Find businesses` form to choose a business type,
set a limit, and queue a background find job. Open `Jobs` to watch the run
complete while the crawler discovers websites and extracts contacts. `Broad
ecommerce scrape` remains available as a fallback.

## Literal Engine Test

Run a real scrape against a country and inspect whether the output matches the
claim:

```bash
rm -f /tmp/spyber-ke.json
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber init
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber find --country KE --sector commerce --segment wholesalers --limit 5
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber companies list --country KE
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber contacts list --country KE
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber export --country KE --format csv --only generic
```

The outcome is acceptable only if exported rows are public business contacts,
deduped, source-backed, and tied to matched businesses.

## Safety Defaults

- only `http` and `https` URLs are accepted
- private, loopback, and link-local hosts are blocked by the fetcher by default
- country discovery uses public OpenStreetMap/Overpass shop tags and Common Crawl country TLD indexes
- every contact must keep its source URL
- exports exclude suppressed contacts
- source and export actions are audit logged
- named personal emails are classified separately from generic role addresses
- the web UI binds to `127.0.0.1:8091` by default

## Verification

```bash
make test
make vet
make check-build
make lines
```

`make lines` enforces the project rule that every file stays under 700 lines.

## Documentation

- [Architecture](docs/architecture.md)
- [Compliance](docs/compliance.md)
- [Data Model](docs/data-model.md)
- [Product Engine](docs/product-engine.md)
- [Operator Guide](docs/operator-guide.md)
- [License Policy](docs/license-policy.md)
- [Developers](DEVELOPERS.md)
- [Contributing](CONTRIBUTING.md)
- [Testing](tests/README.md)

---

License: [AGPL-3.0-only](LICENSE)
