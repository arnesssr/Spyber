# Engine Architecture

This document describes how Spyber turns a country and business intent into
reviewable business contacts. It is for developers and operators evaluating
whether the engine works as claimed.

## Contract

The user should not need to know a URL. The normal flow is:

```text
country + business intent + limit
-> intent expansion
-> web/search-index candidate discovery
-> company dedupe
-> fetch task planning
-> controlled parallel fetching
-> contact sieve
-> contact extraction
-> review/export
```

Manual URL entry is an advanced fallback, not the core product.

## Core Entities

- `find_jobs`: user-visible discovery runs.
- `fetch_tasks`: individual URL fetch attempts for a find job.
- `companies`: deduped business candidates keyed by normalized host.
- `contacts`: extracted public contact emails tied to a company and source URL.
- `evidence`: why a company matched or was rejected.

## Fetch Strategy

Each discovered candidate becomes a company fetch plan. The first task set is:

- root domain
- original candidate URL
- `/contact`
- `/contact-us`
- `/contacts`
- `/support`
- `/customer-service`
- `/help`
- `/locations`
- `/branches`
- `/wholesale`
- `/distributors`
- `/about`
- `/sitemap.xml`
- `/sitemap_index.xml`

The engine may add discovered contact links while processing a company. Crawl
mode controls how far that site-context crawl can expand:

- `standard`: bounded crawl for quick checks.
- `deep`: broader contact-page crawl for normal operator work.
- `exhaustive`: no page-count cap inside the discovered site context.

All modes still keep URL dedupe, request timeouts, private-network blocking,
response-size limits, and per-host delay.

## Discovery Providers

The default provider order is:

- web search through DuckDuckGo Lite or `SPYBER_WEBSEARCH_ENDPOINT`
- OpenStreetMap/Overpass
- Common Crawl country TLD indexes

Provider errors must be visible. A zero-candidate run is treated as a failed
run instead of a successful empty run.

## Throughput Model

Spyber does not fire unlimited requests. It uses controlled parallelism:

- UI find actions create queued background jobs
- the local UI server runs one find job at a time
- each job plans many company fetches
- each job stores its requested crawl mode
- fetch executors process different companies in parallel
- the HTTP fetcher keeps a per-host delay
- each URL attempt is persisted as a fetch task

The user chooses crawl mode, not executor counts. The engine derives internal
parallelism and per-site expansion from that mode.

## Failure Taxonomy

Fetch failures are recorded instead of hidden. Current categories include:

- `dns_failed`
- `timeout`
- `tls_failed`
- `blocked_private_host`
- `http_403`
- `http_404`
- `http_429`
- `http_5xx`
- `response_too_large`
- `fetch_failed`

These categories let operators separate bad discovery from network failures,
rate limits, blocked websites, and true no-contact outcomes.

## Intent Expansion

Custom search terms are expanded before discovery. For example:

```text
shop -> store, products, cart, checkout, delivery
salon -> hairdresser, beauty, barber, spa, booking
wholesale -> supplier, distributor, bulk, trade account
```

This is a deterministic first version. Later versions can add persistent custom
intent dictionaries, embeddings, or optional AI classification, but the engine
must remain explainable and testable.

## Export Standard

Every exportable row must have:

- business website
- contact value
- contact type
- source URL
- match status
- suppression check

Rejected businesses and suppressed contacts must not contribute export rows.

## Current Limits

- PostgreSQL is the normal reliable store.
- Local JSON storage is only an explicit development override.
- Browser automation is not part of the default fetch path yet.
- Phone and WhatsApp extraction are not implemented yet.
- Search provider quality still depends on public result availability and blocking behavior.
