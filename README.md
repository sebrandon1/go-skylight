# go-skylight

[![Build Status](https://github.com/sebrandon1/go-skylight/actions/workflows/pre-main.yaml/badge.svg)](https://github.com/sebrandon1/go-skylight/actions/workflows/pre-main.yaml)
[![codecov](https://codecov.io/gh/sebrandon1/go-skylight/branch/main/graph/badge.svg)](https://codecov.io/gh/sebrandon1/go-skylight)
[![Go Reference](https://pkg.go.dev/badge/github.com/sebrandon1/go-skylight/lib.svg)](https://pkg.go.dev/github.com/sebrandon1/go-skylight/lib)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sebrandon1/go-skylight)](https://go.dev/)
[![License](https://img.shields.io/github/license/sebrandon1/go-skylight)](LICENSE)

Go CLI and client library for the [Skylight Calendar](https://app.ourskylight.com) API. Manage frames, calendars, chores, rewards, lists, meals, and family member categories from the terminal or from Go code.

> **Disclaimer:** This is an unofficial, community-built tool and is not affiliated with or endorsed by Skylight. It interacts with Skylight's undocumented API — behavior may change without notice. Use at your own risk.

## Key Features

- Full CRUD for calendar events, chores, rewards, lists, recipes, categories, photos, routines, and grocery lists
- OAuth2 login with automatic token rotation and config file persistence
- Table and JSON output formats
- Go client library with retry, rate limiting, and typed errors
- Alpaca Markets integration for reward-triggered stock purchases
- Docker images for `linux/amd64` and `linux/arm64`

## Quick Start

```bash
# Install
go install github.com/sebrandon1/go-skylight@latest

# Login (saves credentials to ~/.skylight/config)
skylight login --email user@example.com --password yourpassword --save

# Use any command
skylight chore list --frame-id FRAME_ID
skylight calendar list --start-date 2024-01-15
skylight reward points
```

Or run via Docker:

```bash
docker run --rm \
  -e SKYLIGHT_USER_ID=YOUR_UID \
  -e SKYLIGHT_TOKEN=YOUR_TOKEN \
  -e SKYLIGHT_FRAME_ID=FRAME_ID \
  sebrandon1/go-skylight:latest chore list
```

## Authentication

| Mode | Flags / Env Vars |
|------|------------------|
| OAuth2 refresh token (recommended) | `--refresh-token` / `SKYLIGHT_REFRESH_TOKEN` |
| Pre-existing bearer token | `--user-id` + `--token` / `SKYLIGHT_USER_ID` + `SKYLIGHT_TOKEN` |
| Email + password (deprecated) | `--email` + `--password` / `SKYLIGHT_EMAIL` + `SKYLIGHT_PASSWORD` |

Config file: `~/.skylight/config` (override with `--config`). CLI flags take precedence.

## Guides

| Document | Description |
|----------|-------------|
| [Examples](docs/examples/) | Common scenarios: deleting profiles/labels, managing chores, scripting |
| [CLI Reference](docs/cli-reference.md) | Full command listing for all resources |
| [Library Usage](docs/library-usage.md) | Go client API, examples, typed errors, and coverage matrix |
| [Alpaca Integration](docs/alpaca-trigger.md) | Reward-triggered stock purchases via Alpaca Markets |

## Architecture

```
CLI (cmd/)  -->  lib.Client (lib/)  -->  Skylight REST API
                                    -->  Alpaca v2 REST API (alpaca-trigger)
```

## Development

```bash
make build          # build skylight CLI
make build-trigger  # build alpaca-trigger
make test           # go test ./... -v
make lint           # golangci-lint run ./...
make vet            # go vet ./...
make clean          # remove built binaries
```

CI runs `lint`, `test` (with `-race`), and `build` on ubuntu + macos across Go 1.25.x and 1.26.x.
