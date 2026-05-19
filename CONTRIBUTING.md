# Contributing to go-skylight

Thank you for your interest in contributing!

## Prerequisites

- Go 1.25+ (the module requires 1.26.1; use `go.mod` as the source of truth)
- `golangci-lint` — install via `go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`
- `make` — used for all dev workflows

## Workflow

1. Fork the repository and create a feature branch off `main`.
2. Write your code. Follow the patterns in the existing `lib/` files.
3. Run `make lint` and fix all issues before pushing.
4. Run `make test` (`go test ./... -v -race`) and ensure all tests pass.
5. Open a pull request against `main`. CI must be green before a PR can be merged.

## Commit Message Format

This project uses [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>(<scope>): <short description>

[optional body]
```

Common types: `feat`, `fix`, `docs`, `chore`, `refactor`, `test`.

Examples:
- `feat(lib): add WithTimeout functional option`
- `fix(poller): prevent duplicate events after restart`
- `docs: update Alpaca integration section in README`

## Functional Options Convention

All new client configuration must be exposed as a `ClientOption` via a `With*`
function in `lib/options.go`. Never add new exported fields to `Client` for
configuration — use functional options to keep the API additive and
backward-compatible.

## Breaking Change Policy

go-skylight follows [Semantic Versioning](https://semver.org/).

- **Patch releases** (v1.0.x): bug fixes only; no API changes.
- **Minor releases** (v1.x.0): backward-compatible additions (new methods, new `With*` options, new error types).
- **Major releases** (vX.0.0): breaking changes to exported APIs, struct fields, or error types.

A change is breaking if it requires callers to update their code. Examples:
- Removing or renaming an exported function or type
- Changing the signature of an exported function
- Removing a `json` tag field from a publicly documented struct

Add a `BREAKING CHANGE:` footer to the commit message when introducing a
breaking change:

```
feat(lib): replace SkylightURL global with WithBaseURL option

BREAKING CHANGE: direct assignment to lib.SkylightURL is no longer the
recommended way to override the API base URL. Use WithBaseURL instead.
```

## Adding a New Resource

1. Add types to `lib/structs.go`.
2. Create `lib/<resource>.go` with methods on `*Client`. Use `c.effectiveURL()` for all URL construction.
3. Create `lib/<resource>_test.go` with table-driven tests using `httptest.NewServer` and `WithBaseURL`.
4. Add CLI commands under `cmd/` using Cobra.
5. Update the API Coverage table in `README.md`.

## Running Tests

```bash
make test         # all packages, verbose, with race detector
make lint         # golangci-lint
make vet          # go vet
make build        # CLI binary
make build-trigger # alpaca-trigger binary
```

## Running Integration Tests

Integration tests call the live Skylight API. They require credentials and a frame ID.

### Setup (one-time)

Add your credentials to `~/.skylight/config` (same file used by the CLI):

```
SKYLIGHT_EMAIL=you@example.com
SKYLIGHT_PASSWORD=yourpassword
SKYLIGHT_FRAME_ID=your-frame-id
```

`SKYLIGHT_FRAME_ID` is already written by `skylight login --save`. You only need
to add `SKYLIGHT_EMAIL` and `SKYLIGHT_PASSWORD`. Environment variables override
the config file if set.

### Running

```bash
make integration-read   # read-only tests — safe to run repeatedly
make integration        # all tests including CRUD — creates and deletes real data
```

CRUD tests always clean up after themselves via `t.Cleanup`, but prefer
`integration-read` for routine local verification.
