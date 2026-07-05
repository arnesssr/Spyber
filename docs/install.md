# Install

Spyber builds two Go binaries:

- `spyber`: CLI for discovery, review, verification, and export.
- `spyberd`: local operator UI server.

## From GitHub

```bash
go install github.com/arnesssr/Spyber/cmd/spyber@latest
go install github.com/arnesssr/Spyber/cmd/spyberd@latest
```

Make sure Go's binary directory is on your `PATH`:

```bash
export PATH="$(go env GOPATH)/bin:$PATH"
```

Check the install:

```bash
spyber version
spyberd --addr 127.0.0.1:8091
```

## From A Local Clone

```bash
make install
export PATH="$(go env GOPATH)/bin:$PATH"
make install-check
```

Run the UI:

```bash
spyberd --addr 127.0.0.1:8091
```

Then open:

```text
http://127.0.0.1:8091
```

## Storage

For durable use, set PostgreSQL:

```bash
export SPYBER_DATABASE_URL='postgres://user:pass@localhost:5432/spyber?sslmode=disable'
spyber init
spyberd --addr 127.0.0.1:8091
```

Without `SPYBER_DATABASE_URL`, normal CLI and UI commands fail fast. Set
`SPYBER_STORE=/tmp/spyber-dev.json` only for explicit development JSON runs.
