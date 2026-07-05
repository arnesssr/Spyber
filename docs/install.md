# Install

Spyber builds two Go binaries:

- `spyber`: CLI for discovery, review, verification, and export.
- `spyberd`: local operator UI server.

## From GitHub

This does not require cloning the repository. Go downloads the module and
builds native binaries for your current OS and CPU.

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
spyberd --help
```

For a fixed release:

```bash
go install github.com/arnesssr/Spyber/cmd/spyber@v0.2.3
go install github.com/arnesssr/Spyber/cmd/spyberd@v0.2.3
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

For a quick local PostgreSQL with Docker:

```bash
docker run --name spyber-postgres \
  -e POSTGRES_USER=spyber \
  -e POSTGRES_PASSWORD=spyber \
  -e POSTGRES_DB=spyber \
  -p 5432:5432 \
  -d postgres:16

export SPYBER_DATABASE_URL='postgres://spyber:spyber@127.0.0.1:5432/spyber?sslmode=disable'
spyber init
```

Without `SPYBER_DATABASE_URL`, normal CLI and UI commands fail fast. Set
`SPYBER_STORE=/tmp/spyber-dev.json` only for explicit development JSON runs.
