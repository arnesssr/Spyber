# Engine Architecture

This document describes how Spyber turns a country and business intent into
reviewable business contacts. It is for developers and operators evaluating
whether the engine works as claimed.

## Contract

The user should not need to know a URL. The normal flow is:

```text
country + business intent + limit
-> intent expansion
-> candidate discovery
-> company dedupe
-> fetch task planning
-> controlled parallel fetching
-> profile scoring
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

Each matched candidate becomes a company fetch plan. The first task set is:

- root domain
- original candidate URL
- `/contact`
- `/contact-us`
- `/about`
- `/sitemap.xml`

The engine may add discovered contact links while processing a company, capped
per company so a single site cannot explode the crawl.

## Throughput Model

Spyber does not fire unlimited requests. It uses controlled parallelism:

- UI find actions create queued background jobs
- the local UI server runs one find job at a time
- each job plans many company fetches
- workers process different companies in parallel
- the HTTP fetcher keeps a per-host delay
- each URL attempt is persisted as a fetch task

The current default worker count is conservative. It can be raised later after
the fetch task table and failure metrics show where the real bottlenecks are.

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

Custom search terms are expanded before discovery and scoring. For example:

```text
shop -> commerce/retailers + store, products, cart, checkout, delivery
salon -> services/salons + hairdresser, beauty, barber, spa, booking
wholesale -> commerce/wholesalers + supplier, distributor, bulk, trade account
```

This is a deterministic first version. Later versions can add persistent custom
business profiles, embeddings, or optional AI classification, but the engine
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

- PostgreSQL is the reliable store when `SPYBER_DATABASE_URL` is set.
- Local JSON storage is for development, not high-volume production.
- Browser automation is not part of the default fetch path yet.
- Phone and WhatsApp extraction are not implemented yet.
- Search provider coverage is still limited.
