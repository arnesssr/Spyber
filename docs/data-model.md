# Data Model

Postgres is the intended durable source of truth. The local JSON store mirrors
the core shape for development.

## Core Entities

- `sources`: allowed discovery inputs per country.
- `companies`: normalized ecommerce business candidates.
- `crawl_jobs`: crawl attempts and failure reasons.
- `contacts`: extracted emails with source evidence.
- `evidence`: country and ecommerce classification evidence.
- `suppression`: contacts excluded from future export.
- `exports`: export events and filters.
- `audit_events`: operator and system actions.

## Contact States

- `needs_review`
- `approved`
- `rejected`
- `suppressed`

## Contact Types

- `generic`: role or mailbox address.
- `named`: likely personal work address.
- `unknown`: not confidently classified.

## Company States

- `candidate`
- `crawled`
- `review`
- `approved`
- `rejected`

## Indexing Priorities

- unique normalized company host
- unique contact email per company
- contact status and type
- country code
- source URL
- suppression email
