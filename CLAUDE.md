# go-skylight

Go CLI and client library for the [Skylight Calendar](https://www.ourskylight.com/) API. Provides command-line access to manage frames, calendars, chores, rewards, lists, meals, photos, routines, groceries, and family member categories on Skylight devices.

## Go Version

Go 1.26 (see `go.mod`)

## Dependencies

- `github.com/spf13/cobra` -- CLI framework
- `golang.org/x/time` -- Rate limiter for API client

## Build / Test / Lint

```bash
make build              # go build with version injection via ldflags → ./skylight
make build-trigger      # go build → ./alpaca-trigger (Alpaca Markets integration)
make test               # go test ./... -v
make lint               # golangci-lint run ./...
make vet                # go vet ./...
make clean              # rm -f skylight alpaca-trigger
make integration        # go test with -tags integration (full API, requires credentials)
make integration-read   # integration tests limited to read-only operations
```

Run `make lint` before committing and fix any issues.

The build target injects version via ldflags: `-X main.Version=$(git describe --tags --always)`. Use `./skylight --version` to verify.

## Project Structure

```
main.go                       # Entrypoint, sets Version, calls cmd.Execute()
cmd/                           # Cobra command definitions
  root.go                      # Root command, persistent flags (--email, --password, --token, --user-id, --frame-id, --config, --refresh-token, --device-fingerprint, --output, --quiet), version
  session.go                   # login command (with --save flag for config file)
  config.go                    # Config file loading/saving (~/.skylight/config)
  configcmd.go                 # config show|get|set|unset|edit subcommands
  frame.go                     # frame info, devices, avatars, colors
  calendar.go                  # calendar list, create, create-countdown, update, delete, sources, week
  calendar_week.go             # Weekly calendar view builder (Mon-Sun slots)
  chore.go                     # chore list (with --week), create, update, delete, complete, skip, claim
  chore_streak.go              # chore streak — per-assignee completion streak stats
  chore_week.go                # Weekly chore view builder (Mon-Sun slots)
  reward.go                    # reward list, create, update, delete, redeem, unredeem, points, remove-stars
  reward_remove_stars.go       # reward remove-stars — deduct points from a user balance
  list.go                      # list all, info, create, update, delete, add-item, update-item, delete-item
  meal.go                      # meal categories, recipes (create, update, delete), sittings, grocery list
  category.go                  # list family member categories
  photo.go                     # photo list, upload, delete, download
  routine.go                   # routine list, create, update, delete, reorder
  grocery.go                   # grocery list, create, show, add, add-recipe, clear, organize, order
  addon.go                     # addon list — show frame add-ons and enabled state
  analytics.go                 # analytics — family activity stats over a time period
  export.go                    # export — dump frame data to JSON file
  import.go                    # import — restore frame data from export JSON file
  watch.go                     # watch — poll for changes and print events as they happen
  home.go                      # home — weekly combined view of events, tasks, lists, meals
  status.go                    # status — quick overview of the connected frame
  week.go                      # Generic weekly slot builder (used by chore_week and calendar_week)
  table_output.go              # Table renderers for all resource types (--output table)
  helpers.go                   # printJSON, printOutput, printSuccess, printDryRun, table writer, utilities
  *_test.go                    # Unit tests
cmd/alpaca-trigger/            # Standalone binary: reward-triggered stock purchases
  main.go                      # Entrypoint, env-based config, polls for reward redemptions
  alpaca.go                    # Alpaca Markets v2 REST API client (BuyVOO)
  alpaca_test.go               # Unit tests
lib/                           # API client library
  client.go                    # HTTP client, auth, request helpers (get/post/put/patch/delete)
  session.go                   # Login (POST /api/sessions), OAuth2 refresh token flow
  structs.go                   # All API types and request/response structs
  options.go                   # ClientOption functional options (WithBaseURL, WithRetry, WithRateLimit, WithLogger)
  errors.go                    # Typed errors: AuthError, NotFoundError, RateLimitError, ValidationError
  retry.go                     # Retry with exponential backoff, jitter, and rate limiting
  poller.go                    # RewardsPoller — background poll loop with persistent dedup state
  doc.go                       # Package-level godoc
  calendar.go                  # Calendar event CRUD, source calendars
  category.go                  # List categories
  chore.go                     # Chore CRUD (JSON-API format)
  frame.go                     # Frame info, devices, avatars, colors
  list.go                      # List CRUD, list item CRUD, task box items
  meal.go                      # Recipes, meal sittings, meal categories, grocery list
  reward.go                    # Reward CRUD, redeem/unredeem, points, remove-stars (JSON-API format)
  bounty.go                    # Bounty (chore + reward pair) create and list
  rotation.go                  # Chore rotation generator (rotating assignments across members)
  photo.go                     # Photo list, upload, delete, download
  routine.go                   # Routine CRUD, reorder
  *_test.go                    # Unit tests using httptest mock servers
  integration_test.go          # Integration tests (build tag: integration)
  integration_crud_test.go     # CRUD integration tests (build tag: integration)
  example_test.go              # Testable examples for godoc
  config_loader_test.go        # Config loader tests
```

## Authentication

Three modes (in order of preference):
1. **OAuth2 refresh token (recommended):** `--refresh-token` / `SKYLIGHT_REFRESH_TOKEN` — auto-rotates tokens, persists to config
2. **Pre-existing bearer token:** `--user-id` + `--token` / `SKYLIGHT_USER_ID` + `SKYLIGHT_TOKEN` (skip login)
3. **Email/password (deprecated):** `--email` + `--password` / `SKYLIGHT_EMAIL` + `SKYLIGHT_PASSWORD` (auto-login via `POST /api/sessions`)

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
SKYLIGHT_REFRESH_TOKEN=rt_abc123
SKYLIGHT_DEVICE_FINGERPRINT=00000000-0000-4000-8000-000000000001
```

- CLI flags take precedence over config file values
- `login --save` writes credentials to config file after successful login
- Rotated refresh tokens are automatically persisted back to the config file
- Comments (`#`) and blank lines are supported

## CLI Commands

### Global Flags

- `--output, -o` -- Output format: `json` (default) or `table`
- `--quiet, -q` -- Suppress non-essential success messages
- `--config` -- Config file path (default `~/.skylight/config`)
- `--frame-id` -- Target Skylight frame ID (required by most commands)

### Top-level Commands

- `login` -- Authenticate and print credentials (with `--save` to write config)
- `status` -- Quick overview of the connected frame (chores, events, meals, lists, points)
- `home` -- Weekly combined view of events, tasks, lists, and meals (with `--no-tasks`, `--no-lists`, `--no-meals`)
- `analytics` -- Family activity statistics over a time period (with `--days`)
- `watch` -- Poll for changes and print events as they happen (with `--interval`, `--resources`, `--persist`)
- `export` -- Dump frame data to JSON file (with `--output-file`, `--resources`, `--days`)
- `import` -- Restore frame data from export JSON file (with `--file`, `--dry-run`, `--resources`)
- `bounty create|list` -- Chore + reward pairs
- `rotation create` -- Rotating chore assignments
- `config show|get|set|unset|edit` -- View and modify the local configuration file

### Resource Commands

- `calendar list|create|create-countdown|update|delete|sources|week` -- Calendar events and source calendars
- `chore list|create|update|delete|complete|skip|claim|streak` -- Chore management (list supports `--week` for weekly view)
- `reward list|create|update|delete|redeem|unredeem|points|remove-stars` -- Rewards and point management
- `list all|info|create|update|delete|add-item|update-item|delete-item` -- List management
- `meal categories|recipes|recipe-info|create-recipe|update-recipe|delete-recipe|sittings|create-sitting|add-to-grocery` -- Meal planning
- `category` -- List family member categories
- `frame info|devices|avatars|colors` -- Frame info and settings
- `photo list|upload|delete|download` -- Photo and video management
- `routine list|create|update|delete|reorder` -- Routine (ordered recurring task list) management
- `grocery list|create|show|add|add-recipe|clear|organize|order` -- Grocery list management (Instacart ordering)
- `addon list` -- Frame add-ons and enabled state

### Legacy `get` Command

The `get` prefix (`get calendar list`, etc.) is still supported but hidden. Prefer the top-level resource commands (`calendar list`, etc.).

### Update Commands

All update commands use `cmd.Flags().Changed()` to only send fields that were explicitly set:

- `chore update --chore-id ID [--title] [--status] [--points] [--assignee-id] [--date]`
- `calendar update --event-id ID [--title] [--start-at] [--end-at] [--all-day]`
- `list update --list-id ID [--title] [--color]`
- `list update-item --list-id ID --item-id ID [--title] [--completed]`
- `reward update --reward-id ID [--title] [--points] [--emoji-icon]`
- `meal update-recipe --recipe-id ID [--title] [--description] [--ingredients] [--url]`
- `routine update --routine-id ID [--title] [--assignee-id] [--steps]`

### Delete/Mutate Commands

Several destructive commands support `--dry-run` to preview the action without making API calls:

- `import --dry-run` -- Preview what would be imported
- Delete and redeem commands with `--dry-run` flag

## alpaca-trigger Binary

Standalone daemon (`cmd/alpaca-trigger/`) that watches for Skylight reward redemptions and places a notional VOO market buy order on Alpaca Markets (paper trading by default).

Build: `make build-trigger`

Configuration (all via environment variables):
- `ALPACA_API_KEY`, `ALPACA_API_SECRET` -- Alpaca Markets API credentials (required)
- `ALPACA_BASE_URL` -- Alpaca API base URL (default: paper trading)
- `SKYLIGHT_USER_ID`, `SKYLIGHT_TOKEN`, `SKYLIGHT_FRAME_ID` -- Skylight credentials (required)
- `POLLER_INTERVAL` -- Poll interval (default: `60s`)
- `POLLER_STATE_FILE` -- Path to persist dedup state across restarts
- `VOO_NOTIONAL` -- Dollar amount per buy (default: `1.00`)

## API Base URL

`https://app.ourskylight.com/api` (set in `lib/client.go` as `SkylightURL`; overridden in tests and via `WithBaseURL` option)

## CI/CD

- Pre-main workflow at `.github/workflows/pre-main.yaml` -- runs lint, test, build on matrix (ubuntu + macos), plus govulncheck
- Release workflow at `.github/workflows/release-binaries.yaml` -- cross-platform binaries (linux/darwin, amd64/arm64) with SHA256 checksums
- Docker images for `linux/amd64` and `linux/arm64`
- Version injected via ldflags at build time
- Linters: goconst, gocritic, gocyclo, misspell, unparam, errcheck, gosec, ineffassign, revive, staticcheck

## Notes

- The Skylight API uses JSON-API format for chores and rewards; the library flattens these into simpler structs.
- Tests use `httptest.NewServer` with a swapped `SkylightURL` for isolation.
- The client library supports functional options (WithRetry, WithRateLimit, WithLogger, WithBaseURL) and typed errors (AuthError, NotFoundError, RateLimitError, ValidationError).
- Do not add `Co-Authored-By` lines to commit messages.
