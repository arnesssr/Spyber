# Spyber

Spyber is a Go-first contact discovery engine. Given a country and search
intent, it discovers related public business sites, crawls contact-heavy pages,
sieves public contact channels, and exports reviewable contacts with source
evidence.

Current version: `0.2.3`

## Scope

Spyber is not a bulk spam tool and does not claim to find every business in a
country. The durable product goal is narrower and testable:

```text
Find a measurable set of public sites related to an operator's search intent,
extract source-backed public contacts, and prevent duplicate or suppressed
contacts from being exported.
```

Current v1 capabilities:

- country-scoped discovery
- search-intent business discovery
- web-search candidate discovery
- source-page candidate discovery
- autonomous country discovery through public indexes
- public website crawling with per-host delay and safe fetch limits
- crawl-mode background find jobs
- contact-sieve filtering for searched sites
- public email extraction
- generic business email preference
- review and suppression workflow
- auditable CSV export
- Go-rendered operator UI

## Search Model

The primary flow is a country plus a search term:

```text
KE + salon
KE + distributor
KE + farm equipment
```

Spyber discovers related public sites, crawls contact-heavy pages, and keeps
only contact rows that retain source evidence.

## Not Done Yet

- Phone extraction is not implemented yet.
- Browser automation is not implemented yet.
- Reviewed precision reporting is not modeled yet.
- The local JSON store is only an explicit development override.

## Stack

- Go only for the CLI, engine, and server-rendered UI
- PostgreSQL as the normal reliable source of truth
- explicit `SPYBER_STORE` JSON override for development and tests only
- no TypeScript or frontend build system in v1

## Quick Start

### Install Without Cloning

Yes. Spyber can be installed without cloning the repository. `go install`
downloads the source, builds native binaries for your current OS and CPU, and
places them in Go's binary directory.

Prerequisites:

- Go `1.26` or newer
- PostgreSQL for normal durable use

Install the latest published version:

```bash
go install github.com/arnesssr/Spyber/cmd/spyber@latest
go install github.com/arnesssr/Spyber/cmd/spyberd@latest
export PATH="$(go env GOPATH)/bin:$PATH"
```

Or install a fixed release:

```bash
go install github.com/arnesssr/Spyber/cmd/spyber@v0.2.3
go install github.com/arnesssr/Spyber/cmd/spyberd@v0.2.3
export PATH="$(go env GOPATH)/bin:$PATH"
```

Check the installed binaries:

```bash
spyber version
spyberd --help
```

### Configure Storage

Spyber requires PostgreSQL for normal use:

```bash
export SPYBER_DATABASE_URL='postgres://user:pass@localhost:5432/spyber?sslmode=disable'
spyber init
```

For a quick local PostgreSQL with Docker:

```bash
docker run --name spyber-postgres \
  -e POSTGRES_USER=spyber \
  -e POSTGRES_PASSWORD=spyber \
  -e POSTGRES_DB=spyber \
  -p 5432:5432 \
  -d postgres:16

export SPYBER_DATABASE_URL='postgres://spyber:spyber@127.0.0.1:5432/spyber?sslmode=disable'
spyber init
```

### Run A Search

```bash
spyber find --country KE --query salon --limit 50 --crawl-mode deep
spyber find --country KE --query distributor --limit 100 --crawl-mode exhaustive
spyber companies list --country KE
spyber contacts list --country KE
spyber export --country KE --format csv --only generic
```

### Run The UI

```bash
spyberd --addr 127.0.0.1:8091
```

Then open:

```text
http://127.0.0.1:8091
```

Manual source workflow:

```bash
spyber source add --country KE --type seed --url https://example.co.ke
spyber discover --country KE --from-sources --limit 100
spyber crawl --country KE
```

## Web UI

From an installed binary:

```bash
spyberd --addr 127.0.0.1:8091
```

Then open:

```text
http://127.0.0.1:8091
```

Set `SPYBER_ADMIN_TOKEN` to require browser Basic Auth with username `admin`:

```bash
SPYBER_ADMIN_TOKEN=change-me spyberd --addr 127.0.0.1:8091
```

The UI is server-rendered Go HTML. It has no TypeScript or frontend build
pipeline.

Set `SPYBER_WEBSEARCH_ENDPOINT` to use a compatible search endpoint. By
default Spyber uses DuckDuckGo Lite for no-key candidate discovery.

Use the country field and `Search contacts` form to enter a search term, set a
limit, choose crawl mode, and queue a background find job. Open `Jobs` to watch
persisted progress while the crawler discovers websites and extracts contacts.
`Broad ecommerce scrape` remains available as a fallback.

## Literal Engine Test

Run a real scrape against a country and inspect whether the output matches the
claim:

```bash
export SPYBER_DATABASE_URL='postgres://user:pass@localhost:5432/spyber?sslmode=disable'
spyber init
spyber find --country KE --query salon --limit 5 --crawl-mode exhaustive
spyber companies list --country KE
spyber contacts list --country KE
spyber export --country KE --format csv --only generic
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

For development from a local clone:

```bash
make install
export PATH="$(go env GOPATH)/bin:$PATH"
make install-check
```

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
- [Install](docs/install.md)
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
