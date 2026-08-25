# Library Usage

This guide provides examples of how to use the `go-skylight` client library.

## Setup

First, initialize your client:

```go
import (
	"context"
	"github.com/sebrandon1/go-skylight"
)

func main() {
	ctx := context.Background()
	client, err := lib.NewClient("frame-id", "token")
	if err!= nil {
		panic(err)
	}
	_ = client
	_ = ctx
}
```

## Examples

### Listing Chores

```go
chores, err := client.ListChores(ctx, "frame-id", lib.ChoreListOptions{Date: "2024-01-15"})
```

### Creating a Bounty

```go
// Note: Replace... with valid BountyData
// bounty, err := client.CreateBounty(ctx, "frame-id", lib.BountyData{...})
```

### Managing Lists

```go
// list, err := client.GetList(ctx, "list-id")
```

## Error Handling

All methods return a standard Go error. Use `errors.As` to check for specific library errors if needed.