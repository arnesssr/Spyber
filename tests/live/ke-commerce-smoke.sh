#!/usr/bin/env sh
set -eu

STORE="${SPYBER_STORE:-/tmp/spyber-ke-live-test.json}"
LIMIT="${SPYBER_LIVE_LIMIT:-5}"

rm -f "$STORE"

SPYBER_STORE="$STORE" go run ./cmd/spyber init
SPYBER_STORE="$STORE" go run ./cmd/spyber scrape --country KE --limit "$LIMIT"
SPYBER_STORE="$STORE" go run ./cmd/spyber companies list --country KE
SPYBER_STORE="$STORE" go run ./cmd/spyber contacts list --country KE
SPYBER_STORE="$STORE" go run ./cmd/spyber export --country KE --format csv --only generic
