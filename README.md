# Spyber

Spyber is a Go-first business discovery engine. Given a country and business
intent, it discovers candidate businesses, crawls public websites, classifies
evidence, extracts public contact channels, and exports reviewable business
contacts with source evidence.

Current version: `0.2.1`

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
- web-search candidate discovery
- source-page candidate discovery
- autonomous country discovery through public indexes
- public website crawling with per-host delay and safe fetch limits
- segment and ecommerce signal classification
- public email extraction
- generic business email preference
- review and suppression workflow
- auditable CSV export
- Go-rendered operator UI

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
- Browser automation is not implemented yet.
- Reviewed precision reporting is not modeled yet.
- Local JSON is a development store, not the production durability target.

## Stack

- Go only for the CLI, engine, and server-rendered UI
- PostgreSQL as the reliable source of truth when `SPYBER_DATABASE_URL` is set
- local JSON store as a lightweight fallback only
- no TypeScript or frontend build system in v1

## Quick Start

```bash
go test ./...
go run ./cmd/spyber init
go run ./cmd/spyber version
go run ./cmd/spyber profiles
go run ./cmd/spyber find --country KE --sector commerce --segment wholesalers --limit 50
go run ./cmd/spyber find --country KE --query salon --limit 50
go run ./cmd/spyber companies list --country KE
go run ./cmd/spyber contacts list --country KE
go run ./cmd/spyber export --country KE --format csv --only generic
```

The default local store is `.spyber/spyber.json`.

Use PostgreSQL locally or in production:

```bash
export SPYBER_DATABASE_URL='postgres://user:pass@localhost:5432/spyber?sslmode=disable'
go run ./cmd/spyber init
```

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

Set `SPYBER_WEBSEARCH_ENDPOINT` to use a compatible search endpoint. By
default Spyber uses DuckDuckGo Lite for no-key candidate discovery.

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
- country discovery uses web search, public OpenStreetMap/Overpass tags, and Common Crawl country TLD indexes
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
- [Engine Architecture](docs/engine-architecture.md)
- [Compliance](docs/compliance.md)
- [Data Model](docs/data-model.md)
- [Product Engine](docs/product-engine.md)
- [Operator Guide](docs/operator-guide.md)
- [License Policy](docs/license-policy.md)
- [Developers](DEVELOPERS.md)
- [Contributing](CONTRIBUTING.md)
- [Testing](tests/README.md)
- [Changelog](CHANGELOG.md)

---

License: [AGPL-3.0-only](LICENSE)
