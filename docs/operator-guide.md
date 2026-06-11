# Operator Guide

## First Run

```bash
spyber init
spyber version
spyber profiles
spyber find --country KE --sector commerce --segment wholesalers --limit 50
spyber find --country KE --query salon --limit 50
spyber scrape --country KE --limit 50
spyber source add --country GB --type seed --url https://example.com
spyber discover --country GB --domain https://shop.example
spyber discover --country GB --from-sources --limit 100
```

## Web UI

```bash
make run-ui
```

Open `http://127.0.0.1:8091`. For a shared machine, set an admin token:

```bash
SPYBER_ADMIN_TOKEN=change-me make run-ui
```

The browser username is `admin`; the password is the token.

## PostgreSQL

Set `SPYBER_DATABASE_URL` to use PostgreSQL instead of the fallback JSON store:

```bash
export SPYBER_DATABASE_URL='postgres://user:pass@localhost:5432/spyber?sslmode=disable'
spyber init
make run-ui
```

The home screen `Find businesses` action queues a background job. Open `Jobs`
to watch discovery, crawl, contact extraction, and verification progress.
Manual sources and broad ecommerce scraping are secondary paths.

## Crawl

```bash
spyber crawl --country GB
```

The crawler uses safe defaults:

- request timeout
- response size limit
- private host blocking
- source URL retention

## Verify

```bash
spyber contacts verify --country GB
```

Verification checks syntax, source presence, suppression, and optional DNS
signals when implemented.

## Review

```bash
spyber review list --country GB
spyber review approve --contact-id con_...
spyber review reject --contact-id con_... --reason unsuitable
```

## Export

```bash
spyber export --country GB --format csv --only generic
```

Exports exclude suppressed contacts and log the export event.
