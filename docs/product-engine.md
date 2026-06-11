# Product Engine

This document is for users, developers, and contributors who need to understand
what Spyber is supposed to do and how we judge whether it works.

## User Job

The user wants:

```text
Find businesses of type X in country Y and return usable public contacts.
```

The user should not need to know business URLs.

## Current Product Claim

Spyber can discover a measurable set of public businesses for a country, crawl
their public websites, classify commerce evidence, extract public contact
emails, and export source-backed rows after suppression checks.

It does not claim to find every business in a country.

## Engine Flow

```text
country + business intent + limit
-> candidate discovery
-> canonicalization and dedupe
-> public website crawl
-> evidence scoring
-> contact extraction
-> review, suppression, export
```

## Current Algorithms

- **Candidate discovery:** OpenStreetMap/Overpass shop tags, Common Crawl
  country TLD indexes, and optional manual sources.
- **Canonicalization:** normalize country codes, URLs, hosts, and emails.
- **Deduplication:** avoid adding the same company host or contact email twice.
- **Crawl planning:** fetch the candidate page and bounded contact/about/support
  links.
- **Evidence scoring:** use storefront, product, cart, checkout, pricing,
  catalog, and platform signals.
- **Contact extraction:** extract public emails, classify generic/named/unknown,
  and keep source URLs.
- **Export gating:** exclude suppressed contacts and rejected businesses.

## Outcome Metrics

Every serious run should be judged by:

- candidate businesses discovered
- unique businesses after dedupe
- crawl success rate
- businesses matched
- businesses rejected
- contacts found
- generic business contacts found
- exportable rows
- duplicate rate
- reviewed precision

## Current Gaps

- Segment profiles are not implemented yet.
- Phone extraction is not implemented yet.
- Business-name extraction is basic.
- Search-job history is not modeled yet.
- Reviewed precision reports are not modeled yet.
- Local JSON is a development store, not the production durability target.

## Acceptance Standard

A run is useful only when:

- exported rows have source URLs
- duplicate companies are not re-added
- suppressed emails do not export
- rejected businesses do not export
- contact rows are tied to matched businesses
- the operator can inspect why a business matched

## Literal Local Test

```bash
make live-test-ke
```

This runs a small live Kenya scrape and prints companies, contacts, and generic
CSV export rows. Because it touches public network sources, exact results can
vary.
