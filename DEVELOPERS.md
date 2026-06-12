# Developers

This guide is for people changing Spyber's code, engine, tests, or public docs.

## Product Standard

Start from the fundamental user job:

```text
Find businesses of type X in country Y and return usable public contacts.
```

Do not build around the assumption that the user already knows URLs. Manual
sources are an advanced fallback, not the core workflow.

Before adding a feature, answer:

- What outcome does this produce for the user?
- What data proves the result is correct?
- What algorithm or rule is responsible?
- What can fail, duplicate, leak, or silently mislead the user?
- How will we test it locally?

## Architecture Rules

- Keep domain rules in `internal/domain`.
- Keep use cases in `internal/app`.
- Keep interfaces in `internal/ports`.
- Keep external adapters in `internal/infra`.
- Keep CLI and web handlers in `internal/interface`.
- Do not put business logic in CLI commands, HTTP handlers, or templates.
- Do not add dependencies unless they remove real risk or complexity.
- Keep every created or edited file under 700 lines.

## Engine Rules

The engine must be measurable. A search or scrape run should eventually report:

- candidates discovered
- unique businesses after dedupe
- businesses crawled
- businesses matched
- businesses rejected
- contacts found
- exportable rows
- duplicate/suppression exclusions
- source URLs retained

Exportable contacts must be tied to matched businesses and source evidence.
Rejected businesses and suppressed contacts must not export.
Provider failures and zero-candidate discovery runs must be visible failures,
not successful empty runs.

## Testing Rules

- Unit tests stay beside the Go package they test.
- Shared fixtures and live smoke scripts live under `tests/`.
- External-network tests must not run inside `make test`.
- Use `make test`, `make vet`, `make check-build`, and `make lines` before
  committing.
- Use `make live-test-ke` when changing discovery, crawling, scoring, or export.

## Versioning Rules

- Keep the current release in `VERSION`.
- Keep `internal/version/version.go` in sync with `VERSION`.
- Add user-visible changes to `CHANGELOG.md`.
- Tag releases as `vX.Y.Z` so GitHub shows the version in repository tags.

## Public Docs

There are no secret product docs in this repo. Anything committed under `docs/`
must address one of:

- users
- developers
- contributors
- product behavior
- operational/compliance expectations

Planning notes belong in ignored local paths such as `planning/` or
`docs/planning/`.

## Security And Compliance

- Never hardcode secrets.
- Never log secrets or private credentials.
- Never bypass source access controls.
- Keep source URLs for extracted contacts.
- Keep suppression and deletion behavior explicit.
- Prefer generic business emails for export.
