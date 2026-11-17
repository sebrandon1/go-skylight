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
// Recommended: OAuth2 refresh token — auto-rotates and persists to config
client, err := lib.NewClientWithRefreshToken("your-refresh-token", "your-device-fingerprint")
```

```go
// Pre-existing bearer token
client, err := lib.NewClientWithToken("user-id", "api-token")
```

```go
// With functional options: retry, rate limiting, logging, custom HTTP client,
// API version, custom base URL
client, err := lib.NewClientWithToken("user-id", "api-token",
    lib.WithRetry(3, 500*time.Millisecond, 10*time.Second),
    lib.WithRateLimit(rate.Limit(5), 10),
    lib.WithLogger(slog.Default()),
    lib.WithAPIVersion("2026-06-01"),
    lib.WithBaseURL("https://staging.example.com/api"), // test seam
)
```

`WithRetry` retries 5xx responses and transport errors for every method except POST. POST is retried only on 429, because a 5xx can follow a create the server already applied, and a retry could duplicate it.

A client built with `NewClientWithRefreshToken` renews its access token on a 401 and replays the request once. Refresh tokens are single-use, so a program that restarts must save each new one with `WithTokenRefreshHandler`. After the server rejects a refresh, `RefreshFailure` returns its error and the client needs new credentials.

## Examples

```go
ctx := context.Background()

// List chores (a missing after/before window is defaulted automatically —
// Date becomes a same-day window; otherwise the current calendar month)
chores, err := client.ListChores(ctx, "frame-id", lib.ChoreListOptions{Date: "2024-01-15"})

// All Tasks view: one entry per series, bucketed late/today/today_timed/any_day/future
all, err := client.ListAllTasks(ctx, "frame-id", lib.AllTasksOptions{Limit: 50, IncludeUpForGrabs: true})
// all.Chores.Late, all.Routines.TodayTimed, ...
// Create a chore with a due time ("HH:MM", 24-hour). Invalid values are rejected
// client-side because the API silently accepts junk (see lib.ValidateStartTime).
chore, err := client.CreateChore(ctx, "frame-id", lib.ChoreData{
    Title:     "Feed the cat",
    DueDate:   "2026-11-20",
    StartTime: "07:30",
})

// Task notifications are frame-wide, not per task. Due notifications fire for
// tasks that have a due time. Only the rules you set are sent. A disabled rule
// is sent without filters (the API rejects and clears them). Enabling a rule
// needs a category ID or up-for-grabs (lib.ErrNoNotificationRecipients), and
// changing filters on a rule that stays disabled returns
// lib.ErrFiltersOnDisabledRule.
on := true
settings, err := client.UpdateTaskNotificationSettings(ctx, "frame-id", lib.TaskNotificationUpdate{
    TaskDue: &lib.TaskNotificationChange{Enabled: &on, CategoryIDs: &[]string{"cat-1"}},
})
current, err := client.GetTaskNotificationSettings(ctx, "frame-id")

// Create a bounty (chore + matched reward)
bounty, err := client.CreateBounty(ctx, "frame-id", lib.BountyData{
    Title:       "Clean the kitchen",
    Points:      10,
    RewardTitle: "Ice cream night",
    EmojiIcon:   "🍦",
})

// Send a household message to two family members on Saturday morning, every week
msg, err := client.SendMessage(ctx, "frame-id", lib.MessageData{
    Body:        "Soccer practice at 9!",
    CategoryIDs: []string{"cat-1", "cat-2"},
    DeliverAt:   "2026-11-07T08:00:00-06:00", // empty = deliver now
    RRule:       "FREQ=WEEKLY;BYDAY=SA",
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
import "errors"

// err comes from a lib API call, e.g.:
//   _, err = client.GetChore(ctx, frameID, choreID)

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
| Chores | ✓ | ✓ | ✓ | ✓ | filter by date/assignee/status, due time, frame-wide task notification settings |
| Rewards | ✓ | ✓ | ✓ | ✓ | redeem, unredeem, points |
| Lists | ✓ | ✓ | ✓ | ✓ | items CRUD, task box |
| Recipes | ✓ | ✓ | ✓ | ✓ | sittings, grocery |
| Categories | ✓ | ✓ | ✓ | ✓ | profiles & labels |
| Messages | ✓ | ✓ | ✓ | ✓ | household text messages (`/nudges`); scheduled and recurring |
| Frame | ✓ | — | ✓ | — | info, devices, avatars, colors, per-device tab visibility |
| Photos | ✓ | ✓ | — | ✓ | paginated list, upload (S3), download, bulk delete |
| Bounties | ✓ | ✓ | ✓ | ✓ | chore + reward pairs |
| Rotations | — | ✓ | — | — | rotating chore assignments |
| Routines | ✓ | ✓ | ✓ | ✓ | get/update supported; modeled as a chore with a time-of-day slot; `ListHabits` and `Routine.Habit` expose habit streaks |
| Grocery | ✓ | ✓ | — | — | organize, order (Instacart), clear completed |
| Status / Home / Analytics / Watch | — | — | — | — | dashboards, stats, and live polling |
