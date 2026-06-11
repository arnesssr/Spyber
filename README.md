# Spyber

Spyber is a Go-first ecommerce business intelligence crawler. It discovers
public ecommerce businesses, records evidence, extracts public business contact
channels, and keeps audit controls around review, suppression, and export.

The project is AGPL-3.0-only. Modified network services built from this project
must make their corresponding source available under the license terms.

## Scope

Spyber is not a bulk spam tool. The durable product goal is a verified database
of ecommerce companies with source evidence and compliance controls.

V1 focuses on:

- country-scoped discovery
- source-page candidate discovery
- public website crawling with per-host delay and safe fetch limits
- ecommerce signal classification
- public email extraction
- generic business email preference
- review and suppression workflow
- auditable CSV export
- Go-rendered operator UI

## Stack

- Go only for the CLI, engine, and server-rendered UI
- Postgres as the intended production source of truth
- local JSON store for early development
- no TypeScript or frontend build system in v1

## Quick Start

```bash
go test ./...
go run ./cmd/spyber init
go run ./cmd/spyber source add --country GB --type seed --url https://example.com
go run ./cmd/spyber discover --country GB --domain https://shop.example
go run ./cmd/spyber crawl --country GB
go run ./cmd/spyber contacts verify --country GB
go run ./cmd/spyber review list --country GB
go run ./cmd/spyber export --country GB --format csv --only generic
```

The default local store is `.spyber/spyber.json`.

## Web UI

Run the operator UI:

```bash
make run-ui
```

Then open:

```text
http://127.0.0.1:8080
```

Set `SPYBER_ADMIN_TOKEN` to require browser Basic Auth with username `admin`:

```bash
SPYBER_ADMIN_TOKEN=change-me make run-ui
```

The UI is server-rendered Go HTML. It has no TypeScript or frontend build
pipeline.

## Safety Defaults

- only `http` and `https` URLs are accepted
- private, loopback, and link-local hosts are blocked by the fetcher by default
- every contact must keep its source URL
- exports exclude suppressed contacts
- source and export actions are audit logged
- named personal emails are classified separately from generic role addresses
- the web UI binds to `127.0.0.1:8080` by default

## Documentation

- [Architecture](docs/architecture.md)
- [Compliance](docs/compliance.md)
- [Data Model](docs/data-model.md)
- [Operator Guide](docs/operator-guide.md)
- [License Policy](docs/license-policy.md)
