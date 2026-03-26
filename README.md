# go-skylight

[![Build Status](https://github.com/sebrandon1/go-skylight/actions/workflows/pre-main.yaml/badge.svg)](https://github.com/sebrandon1/go-skylight/actions/workflows/pre-main.yaml)
[![codecov](https://codecov.io/gh/sebrandon1/go-skylight/branch/main/graph/badge.svg)](https://codecov.io/gh/sebrandon1/go-skylight)
[![Go Reference](https://pkg.go.dev/badge/github.com/sebrandon1/go-skylight/lib.svg)](https://pkg.go.dev/github.com/sebrandon1/go-skylight/lib)
[![Go Version](https://img.shields.io/github/go-mod/go-version/sebrandon1/go-skylight)](https://go.dev/)
[![License](https://img.shields.io/github/license/sebrandon1/go-skylight)](LICENSE)

Go CLI and client library for the [Skylight Calendar](https://app.ourskylight.com) API. Manage frames, calendars, chores, rewards, lists, meals, and family member categories from the terminal or from Go code.

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
  get chore create \
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
  get chore create --title "Take out the trash" --points 5
```

Images are published to [Docker Hub](https://hub.docker.com/r/sebrandon1/go-skylight) on every release for `linux/amd64` and `linux/arm64`.

### Authenticate

```bash
# Option 1: interactive login (saves credentials to ~/.skylight/config)
go-skylight login --email user@example.com --password yourpassword --save

# Option 2: supply credentials directly on every command
go-skylight get chore list --user-id YOUR_UID --token YOUR_TOKEN --frame-id FRAME_ID
```

After `login --save`, credentials are stored in `~/.skylight/config` and loaded automatically.

## Authentication Modes

| Mode | Flags / Config Keys |
|------|---------------------|
| Email + password | `--email` / `--password` |
| Pre-existing token | `--user-id` / `--token` |
| Config file | `SKYLIGHT_EMAIL`, `SKYLIGHT_PASSWORD`, `SKYLIGHT_TOKEN`, `SKYLIGHT_USER_ID`, `SKYLIGHT_FRAME_ID` |

Config file location: `~/.skylight/config` (override with `--config`). CLI flags take precedence.

## CLI Reference

All commands accept `--user-id`, `--token`, `--frame-id`, and `--config` as persistent flags.

### Calendar

```bash
go-skylight get calendar list [--start-date DATE] [--end-date DATE]
go-skylight get calendar create --title TITLE --start-at DATETIME [--end-at DATETIME] [--all-day]
go-skylight get calendar update --event-id ID [--title TITLE] [--start-at DATETIME] [--end-at DATETIME]
go-skylight get calendar delete --event-id ID
go-skylight get calendar sources
```

### Chores

```bash
go-skylight get chore list [--date DATE] [--assignee-id ID] [--status STATUS]
go-skylight get chore create --title TITLE [--points N] [--assignee-id ID] [--date DATE]
go-skylight get chore update --chore-id ID [--title T] [--status S] [--points N]
go-skylight get chore delete --chore-id ID
```

### Rewards

```bash
go-skylight get reward list
go-skylight get reward create --title TITLE --points N [--emoji-icon EMOJI]
go-skylight get reward update --reward-id ID [--title T] [--points N] [--emoji-icon EMOJI]
go-skylight get reward delete --reward-id ID
go-skylight get reward redeem   --reward-id ID
go-skylight get reward unredeem --reward-id ID
go-skylight get reward points
```

### Lists

```bash
go-skylight get list all
go-skylight get list info       --list-id ID
go-skylight get list create     --title TITLE [--color COLOR]
go-skylight get list update     --list-id ID [--title T] [--color C]
go-skylight get list delete     --list-id ID
go-skylight get list add-item   --list-id ID --title TITLE
go-skylight get list update-item --list-id ID --item-id ITEM_ID [--title T] [--completed]
go-skylight get list delete-item --list-id ID --item-id ITEM_ID
```

### Meals

```bash
go-skylight get meal categories
go-skylight get meal recipes
go-skylight get meal recipe-info --recipe-id ID
go-skylight get meal create-recipe --title TITLE [--description D] [--ingredients a,b] [--url URL]
go-skylight get meal update-recipe --recipe-id ID [--title T] [--description D]
go-skylight get meal delete-recipe --recipe-id ID
go-skylight get meal sittings
go-skylight get meal create-sitting --recipe-id ID --date DATE
go-skylight get meal add-to-grocery --recipe-id ID
```

### Bounties & Rotations

```bash
go-skylight bounty create --title TITLE --points N --reward-title R [--emoji-icon EMOJI]
go-skylight bounty list
go-skylight rotation create --chores "Dishes,Vacuum" --assignees "id1,id2" \
    --start-date DATE --weeks N --points N
```

### Dashboard

```bash
go-skylight today   # or: go-skylight dashboard
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
| Categories | ✓ | — | — | — | family members |
| Frame | — | — | — | — | info, devices, avatars, colors |
| Bounties | ✓ | ✓ | — | — | chore + reward pairs |
| Rotations | — | ✓ | — | — | rotating assignments |
| Dashboard | — | — | — | — | today aggregate |

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
make build          # build go-skylight CLI
make build-trigger  # build alpaca-trigger
make test           # go test ./... -v
make lint           # golangci-lint run ./...
make vet            # go vet ./...
make clean          # remove built binaries
```

CI runs `lint`, `test` (with `-race`), and `build` on ubuntu + macos across Go 1.25.x and 1.26.x.
