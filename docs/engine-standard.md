# Engine Standard

Spyber must be judged by whether the engine does what it claims, not by whether
the UI looks like a crawler product.

## First-Principles Question

The user wants:

```text
Find businesses of type X in country Y and return usable public contacts.
```

The user should not need to know business URLs.

## Engine Outcome

For each run, Spyber should be able to report:

- candidate businesses discovered
- unique businesses after dedupe
- websites crawled
- businesses matched to the intent
- businesses rejected
- emails found
- generic business emails found
- suppressed contacts excluded
- exportable rows
- source URLs retained

## Required Algorithms

1. Intent/profile expansion
2. Candidate discovery
3. Canonicalization and dedupe
4. Crawl planning
5. Segment/ecommerce scoring
6. Contact extraction
7. Evidence retention
8. Suppression and export gating
9. Run evaluation

## Current Gaps

- No segment profile model yet.
- No phone extraction yet.
- No business-name extraction beyond source-provided names and host fallback.
- No search-job/result-set history yet.
- No precision/false-positive report from reviewed samples yet.
- Local JSON is not the production durability target.

## Acceptance Standard

A run is useful only when:

- exported rows have source URLs
- duplicate companies are not re-added
- suppressed emails do not export
- rejected businesses do not export
- contact rows are tied to matched businesses
- the operator can inspect why a business matched

For profile-based searches, target quality should be measured by reviewed
precision, duplicate rate, crawl success rate, contact yield, and exportable
rows.

## Literal Test

Run a small country scrape and inspect actual output:

```bash
rm -f /tmp/spyber-ke.json
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber init
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber scrape --country KE --limit 5
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber companies list --country KE
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber contacts list --country KE
SPYBER_STORE=/tmp/spyber-ke.json go run ./cmd/spyber export --country KE --format csv --only generic
```

The output should be judged against the acceptance standard above.
