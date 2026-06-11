# Compliance

Spyber is designed for public ecommerce business intelligence, not email spam.

## Collection Rules

- Keep the source URL for every contact.
- Prefer role-based addresses such as `sales@`, `support@`, `hello@`, and
  `wholesale@`.
- Classify named work addresses separately.
- Do not guess personal emails in v1.
- Do not bypass access controls or login walls.
- Respect source rules and rate limits.

## Export Rules

- Suppressed contacts must never export.
- Exports should default to generic role-based emails.
- Every export should record filters, timestamp, and row count.
- Operators should review contacts before commercial use.

## Deletion And Suppression

Suppression is a first-class workflow. If an address opts out, requests deletion,
or is legally unsuitable, it should be added to the suppression list with a
reason and timestamp.

## Legal Notes

AGPL controls software licensing. It does not grant permission to collect or use
personal data. Operators remain responsible for local privacy, anti-spam, data
broker, and marketing rules in every country they target.
