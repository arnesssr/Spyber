#!/usr/bin/env sh
set -eu

STORE="${SPYBER_STORE:-/tmp/spyber-ke-live-test.json}"
LIMIT="${SPYBER_LIVE_LIMIT:-5}"

rm -f "$STORE"

SPYBER_STORE="$STORE" go run ./cmd/spyber init
SCRAPE_OUTPUT="$(SPYBER_STORE="$STORE" go run ./cmd/spyber scrape --country KE --limit "$LIMIT")"
echo "$SCRAPE_OUTPUT"
case "$SCRAPE_OUTPUT" in
  *"discovered=0"*)
    echo "live smoke failed: zero businesses discovered" >&2
    exit 1
    ;;
esac
SPYBER_STORE="$STORE" go run ./cmd/spyber companies list --country KE
CONTACTS_OUTPUT="$(SPYBER_STORE="$STORE" go run ./cmd/spyber contacts list --country KE)"
echo "$CONTACTS_OUTPUT"
if [ -z "$CONTACTS_OUTPUT" ]; then
  echo "live smoke failed: zero contacts found" >&2
  exit 1
fi
SPYBER_STORE="$STORE" go run ./cmd/spyber export --country KE --format csv --only generic
