# Operator Guide

## First Run

```bash
spyber init
spyber scrape --country KE --limit 50
spyber source add --country GB --type seed --url https://example.com
spyber discover --country GB --domain https://shop.example
spyber discover --country GB --from-sources --limit 100
```

## Web UI

```bash
make run-ui
```

Open `http://127.0.0.1:8080`. For a shared machine, set an admin token:

```bash
SPYBER_ADMIN_TOKEN=change-me make run-ui
```

The browser username is `admin`; the password is the token.

The dashboard `Scrape country` action discovers public business websites from
country-level data, crawls those sites, extracts contacts, and verifies results.

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
