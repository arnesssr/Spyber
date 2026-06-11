# Changelog

## 0.2.0

- Add PostgreSQL as the reliable store when `SPYBER_DATABASE_URL` is set.
- Keep JSON as a lightweight fallback when no database URL is configured.
- Add persisted fetch tasks and controlled parallel company fetching.
- Add intent expansion for custom business searches.
- Add background find jobs and a Jobs UI.
- Add `spyber version`.
- Add `VERSION` and release tagging convention.
- Enforce the 700-line file rule through `make lines`.

## 0.1.0

- Initial Go CLI and server-rendered UI.
- Country discovery, crawling, contact extraction, review, suppression, and CSV export.
