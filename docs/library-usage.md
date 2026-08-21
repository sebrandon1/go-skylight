# Library Usage

Import the `lib` package to use go-skylight as a Go library.

```go
import (
    "context"
    "fmt"
    "log/slog"
    "time"

    "github.com/sebrandon1/go-skylight/lib"
    "golang.org/x/time/rate"
)
```

## Creating a Client

```go
// Basic client
client, err := lib.NewClientWithToken("user-id", "api-token")

// With functional options: retry, rate limiting, logging, custom HTTP client,
// API version, custom base URL
client, err := lib.NewClientWithToken("user-id", "api-token",
    lib.WithRetry(3, 500*time.Millisecond, 10*time.Second),
    lib.WithRateLimit(rate.Limit(5), 10),
    lib.WithLogger(slog.Default()),
    lib.WithBaseURL("https://staging.example.com/api"), // test seam
)
```

## Examples

```go
ctx := context.Background()

// List chores
chores, err := client.ListChores(ctx, "frame-id", lib.ChoreListOptions{Date: "2024-01-15"})

// Create a bounty (chore + matched reward)
bounty, err := client.CreateBounty(ctx, "frame-id", lib.BountyData{
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

## Typed Errors

```go
var authErr *lib.AuthError
var notFound *lib.NotFoundError
var rateLimit *lib.RateLimitError
var netErr *lib.NetworkError
var valErr *lib.ValidationError

if errors.As(err, &authErr) {
    // re-authenticate
} else if errors.As(err, &rateLimit) {
    time.Sleep(rateLimit.RetryAfter)
} else if errors.As(err, &valErr) {
    fmt.Println("validation failed:", valErr.Fields)
}
```

Helper predicates are available when you don't need the typed value:

```go
if lib.IsAuthError(err) { /* re-authenticate */ }
if lib.IsRateLimited(err) { /* back off */ }
if lib.IsValidation(err) { /* bad request */ }
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
| Routines | ✓ | ✓ | ✓ | ✓ | get/update supported; modeled as a chore with a time-of-day slot |
| Grocery | ✓ | ✓ | — | — | organize, order (Instacart), clear completed |
| Status / Home / Analytics / Watch | — | — | — | — | dashboards, stats, and live polling |
