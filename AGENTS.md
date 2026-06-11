# Global Codex Working Rules

## Core Operating Principle

Build durable, secure, useful software. Functionality beats decoration.
Correctness beats cleverness. Security is not an optional pass at the end.

## Engineering Standards

- Prefer simple, durable architecture over fashionable complexity.
- Design around clear boundaries: domain, application, infrastructure, interface.
- Avoid god files, god services, hidden coupling, and magic behavior.
- Make state explicit.
- Make failure modes explicit.
- Add validation at system boundaries.
- Prefer boring, proven technology unless there is a strong reason not to.
- Do not introduce dependencies unless they clearly reduce risk or complexity.
- Write code that another serious engineer can maintain under pressure.
- Keep every created or edited file under 700 lines.

## Durable Engine Rules

- Identify the source of truth.
- Define invariants before implementation.
- Handle retries, idempotency, race conditions, timeouts, and partial failure.
- Treat persistence, migrations, and recovery as first-class concerns.
- Add observability where behavior can fail silently.
- Prefer deterministic behavior over clever abstractions.
- Do not fake durability with optimistic assumptions.

## Security Rules

Always consider authentication, authorization, input validation, secrets
handling, data exposure, injection risks, rate limiting, auditability,
privilege boundaries, and supply-chain risk.

Never hardcode secrets. Never log secrets. Never expose internal errors to
users. Never trust client-side checks.

## UI Rules

- Functionality first.
- Clarity first.
- Minimal wording.
- No decorative bloat.
- No fake marketing fluff.
- No huge paragraphs in the UI.
- Every section must earn its place.
- Prefer direct labels, obvious actions, strong hierarchy, and clean spacing.

## Completion Standard

Before saying a task is done:

- Explain what changed.
- Explain why the structure is durable.
- Mention security considerations.
- Mention any tradeoffs.
- Run or suggest the most relevant verification command.
