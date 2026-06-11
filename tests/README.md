# Testing

This folder organizes shared test assets and live smoke tests. Go unit tests
remain beside the packages they test because `go test ./...` expects that shape.

## Test Types

- **Unit tests:** package-local `*_test.go` files.
- **Fixture tests:** package tests that use reusable files from
  `tests/fixtures/`.
- **Live smoke tests:** scripts under `tests/live/` that touch public network
  sources and should not run in ordinary unit-test workflows.

## Standard Checks

```bash
make test
make vet
make check-build
make lines
```

## Live Engine Check

```bash
make live-test-ke
```

This verifies that country discovery, crawling, contact extraction, and export
still work together against real public data. Results can vary because external
websites and public indexes change.

## File Rule

Every test file and fixture must stay under 700 lines. Split large fixtures by
scenario instead of creating one giant sample.
