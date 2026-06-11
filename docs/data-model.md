# Data Model

PostgreSQL is the reliable source of truth when `SPYBER_DATABASE_URL` is set.
The local JSON store mirrors the core shape only as a lightweight fallback.

## Core Entities

- `sources`: allowed discovery inputs per country.
- `companies`: normalized business candidates.
- `find_jobs`: profile-driven discovery and crawl runs.
- `fetch_tasks`: individual URL attempts with status and failure reason.
- `crawl_jobs`: crawl attempts and failure reasons.
- `contacts`: extracted emails with source evidence.
- `evidence`: country, profile, and commerce classification evidence.
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
- find job country and status
- fetch task job and status
- source URL
- suppression email
