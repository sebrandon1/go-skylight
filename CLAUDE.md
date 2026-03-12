# go-skylight

Go CLI and client library for the [Skylight Calendar](https://www.ourskylight.com/) API. Provides command-line access to manage frames, calendars, chores, rewards, lists, meals, and family member categories on Skylight devices.

## Go Version

Go 1.26.1 (see `go.mod`)

## Dependencies

- `github.com/spf13/cobra` -- CLI framework

## Build / Test / Lint

```bash
make build          # go build with version injection via ldflags
make test           # go test ./... -v
make lint           # golangci-lint run ./...
make vet            # go vet ./...
make clean          # rm -f go-skylight
```

Run `make lint` before committing and fix any issues.

The build target injects version via ldflags: `-X main.Version=$(git describe --tags --always)`. Use `./go-skylight --version` to verify.

## Project Structure

```
main.go                 # Entrypoint, sets Version, calls cmd.Execute()
cmd/                    # Cobra command definitions
  root.go               # Root command, persistent flags (--email, --password, --token, --user-id, --frame-id, --config), version
  session.go            # login command (with --save flag for config file)
  config.go             # Config file loading/saving (~/.skylight/config)
  config_test.go        # Config file tests
  frame.go              # frame info, devices, avatars, colors
  calendar.go           # calendar list, create, update, delete, sources
  chore.go              # chore list, create, update, delete
  reward.go             # reward list, create, update, delete, redeem, unredeem, points
  list.go               # list all, info, create, update, delete, add-item, update-item, delete-item
  meal.go               # meal categories, recipes (create, update, delete), sittings, grocery list
  category.go           # list family member categories
  dashboard.go          # today command (aggregates events, chores, points, meals, lists)
  bounty.go             # bounty create/list (chore + reward pairs)
  rotation.go           # chore rotation create (rotating assignments across members)
  helpers.go            # printJSON utility
lib/                    # API client library
  client.go             # HTTP client, auth, request helpers (get/post/put/patch/delete)
  session.go            # Login (POST /api/sessions)
  structs.go            # All API types and request/response structs
  calendar.go           # Calendar event CRUD, source calendars
  category.go           # List categories
  chore.go              # Chore CRUD (JSON-API format)
  frame.go              # Frame info, devices, avatars, colors
  list.go               # List CRUD, list item CRUD, task box items
  meal.go               # Recipes, meal sittings, meal categories, grocery list
  reward.go             # Reward CRUD, redeem/unredeem, points (JSON-API format)
  bounty.go             # Bounty (chore + reward pair) create and list
  dashboard.go          # Dashboard aggregator (today's events, chores, points, meals, lists)
  rotation.go           # Chore rotation generator (rotating assignments across members)
  *_test.go             # Unit tests using httptest mock servers
```

## Authentication

Three modes:
1. **Email/password flags:** `--email` + `--password` (auto-login via `POST /api/sessions`)
2. **Token flags:** `--user-id` + `--token` (skip login, use pre-existing credentials)
3. **Config file:** `~/.skylight/config` (loaded automatically, CLI flags take precedence)

Most commands require `--frame-id` to identify the target Skylight frame.

### Config File

Location: `~/.skylight/config` (override with `--config` flag)

Format (KEY=VALUE, one per line):
```
SKYLIGHT_EMAIL=user@example.com
SKYLIGHT_PASSWORD=secret
SKYLIGHT_TOKEN=abc123
SKYLIGHT_USER_ID=uid456
SKYLIGHT_FRAME_ID=fid789
```

- CLI flags take precedence over config file values
- `login --save` writes credentials to config file after successful login
- Comments (`#`) and blank lines are supported

## CLI Commands

### Top-level
- `login` -- Authenticate and print credentials (with `--save` to write config)
- `dashboard` / `today` -- Today's aggregated dashboard
- `bounty create|list` -- Chore + reward pairs
- `rotation create` -- Rotating chore assignments

### Under `get`
- `get calendar list|create|update|delete|sources`
- `get chore list|create|update|delete`
- `get reward list|create|update|delete|redeem|unredeem|points`
- `get list all|info|create|update|delete|add-item|update-item|delete-item`
- `get meal categories|recipes|recipe-info|create-recipe|update-recipe|delete-recipe|sittings|create-sitting|add-to-grocery`
- `get category`
- `get frame info|devices|avatars|colors`

### Update Commands

All update commands use `cmd.Flags().Changed()` to only send fields that were explicitly set:

- `get chore update --chore-id ID [--title] [--status] [--points] [--assignee-id] [--date]`
- `get calendar update --event-id ID [--title] [--start-at] [--end-at] [--all-day]`
- `get list update --list-id ID [--title] [--color]`
- `get list update-item --list-id ID --item-id ID [--title] [--completed]`
- `get reward update --reward-id ID [--title] [--points] [--emoji-icon]`
- `get meal update-recipe --recipe-id ID [--title] [--description] [--ingredients] [--url]`

### Recipe Create Flags

`get meal create-recipe --title TITLE [--description] [--ingredients item1,item2] [--url]`

## API Base URL

`https://app.ourskylight.com/api` (set in `lib/client.go` as `SkylightURL`; overridden in tests)

## CI/CD

- Pre-main workflow at `.github/workflows/pre-main.yaml` -- runs lint, test, build on matrix (ubuntu + macos), plus govulncheck
- Release workflow at `.github/workflows/release-binaries.yaml` -- cross-platform binaries (linux/darwin, amd64/arm64) with SHA256 checksums
- Version injected via ldflags at build time
- Linters: goconst, gocritic, gocyclo, misspell, unparam, errcheck, gosec, ineffassign, revive, staticcheck

## Notes

- The Skylight API uses JSON-API format for chores and rewards; the library flattens these into simpler structs.
- Tests use `httptest.NewServer` with a swapped `SkylightURL` for isolation.
- Do not add `Co-Authored-By` lines to commit messages.
