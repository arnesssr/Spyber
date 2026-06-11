# Architecture

Spyber separates domain rules, application workflows, infrastructure adapters,
and interfaces.

## Boundaries

- `internal/domain`: entities, states, validation, and invariants.
- `internal/app`: use cases such as discovery, crawl, verify, review, export.
- `internal/ports`: repository and adapter interfaces.
- `internal/infra`: concrete adapters for local storage, HTTP, parsing, DNS.
- `internal/interface`: CLI and future server-rendered web handlers.
- `cmd/spyberd`: local operator UI server.

The CLI should not own business logic. It parses user intent, calls app
services, and prints results.

## Source Of Truth

Production should use Postgres. The development scaffold uses a local JSON store
so the engine can run before database wiring is added.

## Workflow

1. Receive country, business intent, and limit.
2. Discover candidates from country data, source pages, or manual sources.
3. Canonicalize and dedupe companies.
4. Crawl public pages with safe defaults.
5. Score evidence against the business intent.
6. Extract public contacts with source URLs.
7. Reject non-matching businesses before export.
8. Verify, review, suppress, and export compliant records.
9. Maintain suppression and audit history.

## Invariants

- A company must have a normalized website host.
- A contact must have a source URL and company ID.
- Suppressed contacts must never be exported.
- Rejected companies must not contribute export rows.
- Export records must include filters and timestamps.
- Crawler failures must be recorded instead of hidden.

## Failure Modes

- network timeout
- invalid URL
- blocked private host
- large response body
- duplicate company or contact
- source page disappears
- country classification remains uncertain
- DNS verification fails or is unavailable

Failures should be explicit states, not silent skips.
