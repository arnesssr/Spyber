# Architecture

Spyber separates domain rules, application workflows, infrastructure adapters,
and interfaces.

Detailed crawl and fetch behavior lives in
[Engine Architecture](engine-architecture.md).

## Boundaries

- `internal/domain`: entities, states, validation, and invariants.
- `internal/app`: use cases such as discovery, crawl, verify, review, export.
- `internal/ports`: repository and adapter interfaces.
- `internal/infra`: concrete adapters for local storage, HTTP, parsing, DNS.
- `internal/interface`: CLI and future server-rendered web handlers.
- `cmd/spyberd`: local operator UI server.

The CLI should not own business logic. It parses user intent, calls app
services, and prints results.

## Scalable Direction

The architecture should grow by strengthening boundaries, not by creating a
larger handler or crawler file.

- Replace local JSON with Postgres repositories behind `internal/ports`.
- Add durable fetch queues before increasing crawl volume further.
- Add durable job state before distributed workers.
- Keep provider integrations behind `CountryFinder` and `BusinessSearcher`.
- Keep scoring/profile rules in app/domain code, not templates.
- Add metrics around discovery, dedupe, crawl success, match rate, and export.
- Keep live-network checks outside `make test`.

## Source Of Truth

Production should use Postgres. The development scaffold uses a local JSON store
so the engine can run before database wiring is added.

## Workflow

1. Receive country, profile or query, and limit.
2. Persist a find job for UI progress and auditability.
3. Discover candidates from country data, source pages, or manual sources.
4. Canonicalize and dedupe companies.
5. Plan fetch tasks for root, candidate, contact, about, and sitemap URLs.
6. Fetch planned tasks with controlled parallelism.
7. Score evidence against the selected business profile.
8. Extract public contacts with source URLs.
9. Reject non-matching businesses before export.
10. Verify, review, suppress, and export compliant records.
11. Maintain suppression and audit history.

## Invariants

- A company must have a normalized website host.
- A contact must have a source URL and company ID.
- A UI find action must create a job before network work starts.
- The local UI worker runs one find job at a time by default.
- Fetch attempts must be persisted with status and failure reason.
- Suppressed contacts must never be exported.
- Rejected companies must not contribute export rows.
- Export records must include filters and timestamps.
- Crawler failures must be recorded instead of hidden.

## Failure Modes

- network timeout
- DNS failure
- TLS failure
- HTTP blocking or rate limiting
- invalid URL
- blocked private host
- large response body
- duplicate company or contact
- source page disappears
- country classification remains uncertain
- DNS verification fails or is unavailable

Failures should be explicit states, not silent skips.
