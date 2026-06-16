# go-skylight

[![Build Status](https://github.com/sebrandon1/go-skylight/actions/workflows/pre-main.yaml/badge.svg)](https://github.com/sebrandon1/go-skylight/actions/workflows/pre-main.yaml)
[![codecov](https://codecov.io/gh/sebrandon1/go-skylight/branch/main/graph/badge.svg)](https://codecov.io/gh/sebrandon1/go-skylight)
[![Go Reference](https://pkg.go.dev/badge/github.com/sebrandon1/go-skylight/lib.svg)](https://pkg.go.dev/github.com/sebrandon1/go-skylight/lib)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sebrandon1/go-skylight)](https://go.dev/)
[![License](https://img.shields.io/github/license/sebrandon1/go-skylight)](LICENSE)

Go CLI and client library for the [Skylight Calendar](https://app.ourskylight.com) API. Manage frames, calendars, chores, rewards, lists, meals, and family member categories from the terminal or from Go code.

> **Disclaimer:** This is an unofficial, community-built tool and is not affiliated with or endorsed by Skylight. It interacts with Skylight's undocumented API — behavior may change without notice. Use at your own risk.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│  go-skylight CLI  (cmd/)                                     │
│  Cobra commands: calendar · chore · reward · list · meal … │
└──────────────────────┬───────────────────────────────────────┘
                       │ calls
┌──────────────────────▼───────────────────────────────────────┐
│  lib.Client  (lib/)                                          │
│  retry · rate-limit · slog logging · typed errors           │
└──────────────────────┬───────────────────────────────────────┘
                       │ HTTPS / Basic auth
             ┌─────────▼──────────┐
             │  Skylight REST API │
             └────────────────────┘

┌──────────────────────────────────────────────────────────────┐
│  alpaca-trigger  (cmd/alpaca-trigger/)                       │
│  lib.RewardsPoller ──► lib.Client ──► Skylight API          │
│        │                                                     │
│        └──► AlpacaClient ──► Alpaca v2 REST API             │
│             POST /v2/orders  (VOO notional buy)             │
└──────────────────────────────────────────────────────────────┘
```

## Quick Start

### Install

```bash
go install github.com/sebrandon1/go-skylight@latest
```

### Docker

No Go toolchain required — pull the image and run any command directly:

```bash
docker run --rm \
  sebrandon1/go-skylight:latest \
  chore create \
  --user-id YOUR_UID \
  --token YOUR_TOKEN \
  --frame-id FRAME_ID \
  --title "Take out the trash" \
  --points 5
```

Pass credentials via environment variables to keep the command tidy:

```bash
docker run --rm \
  -e SKYLIGHT_USER_ID=YOUR_UID \
  -e SKYLIGHT_TOKEN=YOUR_TOKEN \
  -e SKYLIGHT_FRAME_ID=FRAME_ID \
  sebrandon1/go-skylight:latest \
  chore create --title "Take out the trash" --points 5
```

Images are published to [Docker Hub](https://hub.docker.com/r/sebrandon1/go-skylight) on every release for `linux/amd64` and `linux/arm64`.

### Authenticate

```bash
# Option 1: headless OAuth2 login (saves refresh token + device fingerprint to ~/.skylight/config)
skylight login --email user@example.com --password yourpassword --save

# Option 2: supply a refresh token directly on every command
skylight chore list --refresh-token YOUR_REFRESH_TOKEN --frame-id FRAME_ID

# Option 3 (legacy): pre-existing bearer token
skylight chore list --user-id YOUR_UID --token YOUR_TOKEN --frame-id FRAME_ID
```

After `login --save`, the refresh token is stored in `~/.skylight/config` and loaded automatically; it rotates on use and the new token is persisted back to the config file.

## Authentication Modes

| Mode | Flags / Config Keys |
|------|---------------------|
| OAuth2 refresh token (recommended) | `--refresh-token` / `SKYLIGHT_REFRESH_TOKEN`, `--device-fingerprint` / `SKYLIGHT_DEVICE_FINGERPRINT` |
| Pre-existing bearer token | `--user-id` / `--token` / `SKYLIGHT_USER_ID` / `SKYLIGHT_TOKEN` |
| Email + password (deprecated) | `--email` / `--password` / `SKYLIGHT_EMAIL` / `SKYLIGHT_PASSWORD` |
| Frame ID | `--frame-id` / `SKYLIGHT_FRAME_ID` |

Config file location: `~/.skylight/config` (override with `--config`). CLI flags take precedence.

## CLI Reference

All commands accept `--user-id`/`--token`, `--refresh-token`, `--frame-id`, `--config`, and `--output (json|table)` as persistent flags. Resource commands are top-level (e.g. `skylight chore list`) — the legacy `skylight get <resource> ...` nesting was flattened and is no longer used.

### Calendar

```bash
skylight calendar list [--start-date DATE] [--end-date DATE]
skylight calendar create --title TITLE --start-at DATETIME [--end-at DATETIME] [--all-day]
skylight calendar update --event-id ID [--title TITLE] [--start-at DATETIME] [--end-at DATETIME]
skylight calendar delete --event-id ID
skylight calendar sources
skylight calendar create-countdown --title TITLE --date DATE
skylight calendar week [--date DATE]
```

### Chores

```bash
skylight chore list [--date DATE] [--assignee-id ID] [--status S] [--after DATE] [--before DATE] [--include-late] [--up-for-grabs] [--week [DATE]]
skylight chore create --title TITLE [--description D] [--points N] [--assignee-id ID] [--date DATE] [--recurring] [--up-for-grabs]
skylight chore update --chore-id ID [--title T] [--description D] [--status S] [--points N] [--assignee-id ID] [--date DATE]
skylight chore delete --chore-id ID
skylight chore complete --chore-id ID
skylight chore skip --chore-id ID
skylight chore claim --chore-id ID --assignee-id ID
skylight chore streak [--days N]
```

### Rewards

```bash
skylight reward list
skylight reward create --title TITLE --points N [--emoji-icon EMOJI] [--no-respawn] [--category-ids 1,2]
skylight reward update --reward-id ID [--title T] [--points N] [--emoji-icon EMOJI]
skylight reward delete --reward-id ID
skylight reward redeem   --reward-id ID
skylight reward unredeem --reward-id ID
skylight reward points
```

### Lists

```bash
skylight list all
skylight list info       --list-id ID
skylight list create     --title TITLE [--color COLOR] [--hide-from-frame]
skylight list update     --list-id ID [--title T] [--color C] [--hide-from-frame]
skylight list delete     --list-id ID
skylight list add-item   --list-id ID --title TITLE
skylight list update-item --list-id ID --item-id ITEM_ID [--title T] [--completed]
skylight list delete-item --list-id ID --item-id ITEM_ID
skylight list clear-completed --list-id ID
```

### Meals

```bash
skylight meal categories
skylight meal recipes
skylight meal recipe-info --recipe-id ID
skylight meal create-recipe --title TITLE [--description D] [--ingredients a,b] [--url URL] [--meal-category-id ID]
skylight meal update-recipe --recipe-id ID [--title T] [--description D] [--ingredients a,b] [--url URL]
skylight meal delete-recipe --recipe-id ID
skylight meal sittings [--date-min DATE] [--date-max DATE]
skylight meal create-sitting --recipe-id ID --date DATE [--summary S] [--meal-category-id ID]
skylight meal delete-sitting --sitting-id ID [--date DATE]
skylight meal sitting-recipe --sitting-id ID
skylight meal add-to-grocery --recipe-id ID
skylight meal plan --recipes ID,ID --start-date DATE [--categories ID,ID]
```

### Photos

```bash
skylight photo list [--page-token TOKEN]
skylight photo upload --file PATH [--caption TEXT]
skylight photo delete --message-id ID [--message-id ID ...]
skylight photo download [--message-id ID ...] [--all] [--output-dir DIR]
```

### Categories (Profiles & Labels)

```bash
skylight category list
skylight category create --name NAME [--color COLOR]
skylight category update --category-id ID [--name NAME] [--color COLOR]
skylight category delete --category-id ID
```

### Frame

```bash
skylight frame list
skylight frame info
skylight frame devices
skylight frame avatars
skylight frame colors
```

### Bounties & Rotations

```bash
skylight bounty create --title TITLE --points N --reward-title R [--assignee-id ID] [--due-date DATE] [--emoji-icon EMOJI] [--recurring]
skylight bounty list
skylight bounty update --chore-id ID --reward-id ID [--title T] [--reward-title R] [--points N] [--due-date DATE] [--emoji-icon EMOJI]
skylight bounty delete --chore-id ID --reward-id ID

skylight rotation create --chores "Dishes,Vacuum" --assignee-ids "id1,id2" \
    --start-date DATE --weeks N --points N
```

### Routines

```bash
skylight routine list
skylight routine create --title TITLE [--assignee-id ID] [--steps "Step 1,Step 2"]
skylight routine update --routine-id ID [--title T] [--assignee-id ID] [--steps "..."]
skylight routine delete --routine-id ID
skylight routine reorder --routine-ids "id1,id2,id3"
```

### Grocery

```bash
skylight grocery list
skylight grocery show     --list-id ID
skylight grocery create   --title TITLE
skylight grocery add      --list-id ID --items "Milk,Eggs"
skylight grocery add-recipe --recipe-id ID
skylight grocery organize --list-id ID
skylight grocery order    --list-id ID --retailer costco
skylight grocery clear    --list-id ID
```

### Status, Home, Analytics & Watch

```bash
skylight status                     # quick overview of the connected frame
skylight home [--no-tasks] [--no-lists]   # weekly combined view of events, tasks, and lists
skylight analytics [--days N]       # family activity statistics over a time period
skylight watch [--interval SECONDS] [--resources rewards,chores,calendar] [--persist]
```

### Export & Import

```bash
skylight export [--output-file PATH] [--resources chores,rewards,lists,recipes,sittings,calendar] [--days N]
skylight import --file PATH [--dry-run] [--resources all]
```

## Library Usage

```go
import "github.com/sebrandon1/go-skylight/lib"

// Basic client
client, err := lib.NewClientWithToken("user-id", "api-token")

// With functional options: retry, rate limiting, logging, custom base URL
client, err := lib.NewClientWithToken("user-id", "api-token",
    lib.WithRetry(3, 500*time.Millisecond, 10*time.Second),
    lib.WithRateLimit(rate.Limit(5), 10),
    lib.WithLogger(slog.Default()),
    lib.WithBaseURL("https://staging.example.com/api"), // test seam
)

// List chores
chores, err := client.ListChores("frame-id", lib.ChoreListOptions{Date: "2024-01-15"})

// Create a bounty (chore + matched reward)
bounty, err := client.CreateBounty("frame-id", lib.BountyData{
    Title:       "Clean the kitchen",
    Points:      10,
    RewardTitle: "Ice cream night",
    EmojiIcon:   "🍦",
})

// Stream reward redemptions
poller := lib.NewRewardsPoller(client, "frame-id", 60*time.Second, "")
poller.Start(ctx)
for event := range poller.Events() {
    fmt.Printf("%s redeemed %s (%d pts)\n",
        event.ChildName, event.RewardName, event.Points)
}
```

### Typed Errors

```go
var authErr *lib.AuthError
var notFound *lib.NotFoundError
var rateLimit *lib.RateLimitError
var netErr *lib.NetworkError

if errors.As(err, &authErr) {
    // re-authenticate
} else if errors.As(err, &rateLimit) {
    time.Sleep(rateLimit.RetryAfter)
}
```

## API Coverage

| Resource | List | Create | Update | Delete | Extra |
|----------|------|--------|--------|--------|-------|
| Calendar events | ✓ | ✓ | ✓ | ✓ | sources |
| Chores | ✓ | ✓ | ✓ | ✓ | filter by date/assignee/status |
| Rewards | ✓ | ✓ | ✓ | ✓ | redeem, unredeem, points |
| Lists | ✓ | ✓ | ✓ | ✓ | items CRUD, task box |
| Recipes | ✓ | ✓ | ✓ | ✓ | sittings, grocery |
| Categories | ✓ | ✓ | ✓ | ✓ | profiles & labels |
| Frame | ✓ | — | — | — | info, devices, avatars, colors |
| Photos | ✓ | ✓ | — | ✓ | paginated list, upload (S3), download, bulk delete |
| Bounties | ✓ | ✓ | ✓ | ✓ | chore + reward pairs |
| Rotations | — | ✓ | — | — | rotating chore assignments |
| Routines | ✓ | ✓ | ✓ | ✓ | reorder |
| Grocery | ✓ | ✓ | — | — | organize, order (Instacart), clear completed |
| Status / Home / Analytics / Watch | — | — | — | — | dashboards, stats, and live polling |

## Alpaca Integration

`alpaca-trigger` watches for Skylight reward redemptions and places a notional
VOO market buy on Alpaca Markets every time a reward is redeemed.

### Setup

```bash
make build-trigger
```

### Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ALPACA_API_KEY` | ✓ | — | Alpaca API key ID |
| `ALPACA_API_SECRET` | ✓ | — | Alpaca API secret key |
| `ALPACA_BASE_URL` | — | `https://paper-api.alpaca.markets` | **Paper trading by default** |
| `SKYLIGHT_USER_ID` | ✓ | — | Skylight user ID |
| `SKYLIGHT_TOKEN` | ✓ | — | Skylight API token |
| `SKYLIGHT_FRAME_ID` | ✓ | — | Skylight frame ID to watch |
| `POLLER_INTERVAL` | — | `60s` | Poll frequency (Go duration string) |
| `POLLER_STATE_FILE` | — | `~/.skylight/poller-state.json` | Deduplication state |
| `VOO_NOTIONAL` | — | `1.00` | Dollar amount per redemption |

> **Warning:** Set `ALPACA_BASE_URL=https://api.alpaca.markets` only when you intend to place real orders. The default is the paper-trading endpoint.

### Run

```bash
export ALPACA_API_KEY=your_key
export ALPACA_API_SECRET=your_secret
export SKYLIGHT_USER_ID=uid
export SKYLIGHT_TOKEN=tok
export SKYLIGHT_FRAME_ID=fid

./alpaca-trigger
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
