.PHONY: test fmt vet build install install-cli install-ui install-check check-build run-ui smoke live-test-ke lines

test:
	go test ./...

fmt:
	gofmt -w $$(find . -name '*.go' -not -path './.git/*')

vet:
	go vet ./...

build:
	go build -o bin/spyber ./cmd/spyber
	go build -o bin/spyberd ./cmd/spyberd

install: install-cli install-ui

install-cli:
	go install ./cmd/spyber

install-ui:
	go install ./cmd/spyberd

install-check:
	command -v spyber
	command -v spyberd
	spyber version

check-build:
	go build -o /tmp/spyber-check ./cmd/spyber
	go build -o /tmp/spyberd-check ./cmd/spyberd

run-ui:
	go run ./cmd/spyberd --addr 127.0.0.1:8091

smoke:
	SPYBER_STORE=/tmp/spyber-smoke.json go run ./cmd/spyber init
	SPYBER_STORE=/tmp/spyber-smoke.json go run ./cmd/spyber source add --country GB --type seed --url https://example.com
	SPYBER_STORE=/tmp/spyber-smoke.json go run ./cmd/spyber discover --country GB --domain https://shop.example
	SPYBER_STORE=/tmp/spyber-smoke.json go run ./cmd/spyber companies list --country GB

live-test-ke:
	sh tests/live/ke-commerce-smoke.sh

lines:
	find . -type f -not -path './.git/*' -not -path './.spyber/*' -not -path './bin/*' -exec wc -l {} + | awk '$$2 != "total" && $$1 > 700 { print; failed = 1 } END { exit failed }'
