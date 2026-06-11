# Contributing

Spyber welcomes contributions that improve the business discovery engine,
crawler safety, result quality, documentation, and tests.

## Before You Start

Read:

- [Developers](DEVELOPERS.md)
- [Architecture](docs/architecture.md)
- [Product Engine](docs/product-engine.md)
- [Testing](tests/README.md)
- [Compliance](docs/compliance.md)

## Contribution Rules

- Keep changes scoped.
- Keep every created or edited file under 700 lines.
- Add or update tests for behavior changes.
- Keep docs public-facing and useful to users, developers, or contributors.
- Do not commit planning notes, local data stores, generated binaries, secrets,
  credentials, or scraped datasets.
- Do not introduce a dependency unless it clearly reduces risk or complexity.

## Verification

Run:

```bash
make test
make vet
make check-build
make lines
```

For discovery, crawling, scoring, or export changes, also run:

```bash
make live-test-ke
```

The live test uses public network sources, so results can vary. It is useful for
catching whether the engine still produces source-backed, exportable contacts.

## Pull Request Expectations

Describe:

- what changed
- why it matters
- how it was tested
- known tradeoffs
- whether the change affects compliance, crawling, exports, or suppression
