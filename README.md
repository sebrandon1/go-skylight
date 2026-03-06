# go-skylight

A Go wrapper for the [Skylight Calendar](https://app.ourskylight.com) API. Provides both a CLI and a library for interacting with Skylight frames.

## Installation

```bash
go install github.com/sebrandon1/go-skylight@latest
```

## Authentication

Skylight supports two authentication methods:

### Email/Password Login

```bash
go-skylight login --email user@example.com --password yourpassword
```

This returns a `user-id` and `token` for subsequent requests.

### Direct Token Auth

```bash
go-skylight get calendar list --user-id YOUR_USER_ID --token YOUR_TOKEN --frame-id YOUR_FRAME_ID
```

## CLI Usage

### Calendar Events

```bash
# List events
go-skylight get calendar list --frame-id FRAME_ID --start-date 2024-01-01 --end-date 2024-01-31

# Create event
go-skylight get calendar create --frame-id FRAME_ID --title "Meeting" --start-at "2024-01-15T10:00:00Z" --end-at "2024-01-15T11:00:00Z"

# Delete event
go-skylight get calendar delete --frame-id FRAME_ID --event-id EVENT_ID

# List source calendars
go-skylight get calendar sources --frame-id FRAME_ID
```

### Chores

```bash
# List chores
go-skylight get chore list --frame-id FRAME_ID

# Create chore
go-skylight get chore create --frame-id FRAME_ID --title "Clean room" --points 5
```

### Lists

```bash
# List all lists
go-skylight get list all --frame-id FRAME_ID

# Get a specific list
go-skylight get list info --frame-id FRAME_ID --list-id LIST_ID

# Add item to list
go-skylight get list add-item --frame-id FRAME_ID --list-id LIST_ID --title "Milk"
```

### Rewards

```bash
# List rewards
go-skylight get reward list --frame-id FRAME_ID

# Redeem a reward
go-skylight get reward redeem --frame-id FRAME_ID --reward-id REWARD_ID

# Get points
go-skylight get reward points --frame-id FRAME_ID
```

### Meals

```bash
# List recipes
go-skylight get meal recipes --frame-id FRAME_ID

# List meal sittings
go-skylight get meal sittings --frame-id FRAME_ID
```

### Frame & Device Info

```bash
# Get frame info
go-skylight get frame info --frame-id FRAME_ID

# List devices
go-skylight get frame devices --frame-id FRAME_ID

# List avatars
go-skylight get frame avatars

# List colors
go-skylight get frame colors
```

## Library Usage

```go
package main

import (
    "fmt"
    "log"

    "github.com/sebrandon1/go-skylight/lib"
)

func main() {
    // Authenticate with email/password
    client, err := lib.NewClient("user@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }

    // Or use existing token
    // client, err := lib.NewClientWithToken("user-id", "api-token")

    // List calendar events
    events, err := client.ListCalendarEvents("frame-id", "2024-01-01", "2024-01-31")
    if err != nil {
        log.Fatal(err)
    }

    for _, event := range events {
        fmt.Printf("%s: %s\n", event.StartAt, event.Title)
    }
}
```

## API Coverage

| Resource | Operations |
|----------|-----------|
| Session | Login (email/password) |
| Calendar Events | List, Create, Update, Delete |
| Source Calendars | List |
| Chores | List, Create, Update, Delete |
| Lists | List, Get, Create, Update, Delete |
| List Items | Add, Update, Delete |
| Task Box | Create |
| Rewards | List, Create, Update, Delete, Redeem, Unredeem, Points |
| Recipes | List, Get, Create, Update, Delete, Add to Grocery |
| Meal Sittings | List, Create |
| Meal Categories | List |
| Categories | List |
| Frame | Get |
| Devices | List |
| Avatars | List |
| Colors | List |

## Development

```bash
make build    # Build binary
make test     # Run tests
make lint     # Run linter
make vet      # Run go vet
make clean    # Remove binary
```

## License

Apache License 2.0
