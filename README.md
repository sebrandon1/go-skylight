# go-skylight

A Go wrapper for the [Skylight Calendar](https://app.ourskylight.com) API. Provides both a CLI tool and a Go library for managing Skylight frames, calendar events, chores, lists, rewards, meals, and more.

## Installation

```bash
go install github.com/sebrandon1/go-skylight@latest
```

## Authentication

Skylight uses Basic auth with a `user-id` and `api-token` pair. You can obtain these by logging in with your email and password:

```bash
go-skylight login --email user@example.com --password yourpassword
```

Output:

```
Login successful!
User ID: abc123
API Token: xyz789
```

Use the returned `user-id` and `token` in all subsequent commands.

## CLI Usage

All commands require `--user-id` and `--token` flags. For brevity, the examples below use shell variables:

```bash
export SKYLIGHT_UID="your-user-id"
export SKYLIGHT_TOKEN="your-api-token"
export SKYLIGHT_FRAME="your-frame-id"
```

### Calendar Events

```bash
# List events in a date range
go-skylight get calendar list \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --start-date 2024-01-01 --end-date 2024-01-31

# Create an event
go-skylight get calendar create \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --title "Family Dinner" --start-at "2024-01-15T18:00:00Z" --end-at "2024-01-15T19:30:00Z"

# Create an all-day event
go-skylight get calendar create \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --title "Vacation" --start-at "2024-06-01" --end-at "2024-06-07" --all-day

# Delete an event
go-skylight get calendar delete \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --event-id EVENT_ID

# List linked source calendars (Google, iCal, etc.)
go-skylight get calendar sources \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME
```

### Chores

```bash
# List all chores
go-skylight get chore list \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME

# List chores with filters
go-skylight get chore list \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --date 2024-01-15 --status pending --assignee-id CATEGORY_ID

# Create a chore
go-skylight get chore create \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --title "Clean room" --points 5

# Delete a chore
go-skylight get chore delete \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --chore-id CHORE_ID
```

### Lists & List Items

```bash
# List all lists
go-skylight get list all \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME

# Get a specific list with its items
go-skylight get list info \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --list-id LIST_ID

# Create a new list
go-skylight get list create \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --title "Grocery List"

# Add an item to a list
go-skylight get list add-item \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --list-id LIST_ID --title "Milk"

# Delete an item from a list
go-skylight get list delete-item \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --list-id LIST_ID --item-id ITEM_ID

# Delete a list
go-skylight get list delete \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --list-id LIST_ID
```

### Rewards

```bash
# List rewards
go-skylight get reward list \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME

# Create a reward
go-skylight get reward create \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --title "Ice Cream" --points 10

# Redeem a reward
go-skylight get reward redeem \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --reward-id REWARD_ID

# Unredeem a reward
go-skylight get reward unredeem \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --reward-id REWARD_ID

# Check points balance
go-skylight get reward points \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME

# Delete a reward
go-skylight get reward delete \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --reward-id REWARD_ID
```

### Meals & Recipes

```bash
# List meal categories
go-skylight get meal categories \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME

# List all recipes
go-skylight get meal recipes \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME

# Get a specific recipe
go-skylight get meal recipe-info \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --recipe-id RECIPE_ID

# Create a recipe
go-skylight get meal create-recipe \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --title "Spaghetti"

# Delete a recipe
go-skylight get meal delete-recipe \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --recipe-id RECIPE_ID

# List meal sittings (scheduled meals)
go-skylight get meal sittings \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME

# Schedule a meal
go-skylight get meal create-sitting \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --recipe-id RECIPE_ID --date 2024-01-15 --meal-type dinner

# Add a recipe's ingredients to the grocery list
go-skylight get meal add-to-grocery \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME \
  --recipe-id RECIPE_ID
```

### Family Members (Categories)

```bash
go-skylight get category \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME
```

### Frame & Device Info

```bash
# Get frame details
go-skylight get frame info \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME

# List connected devices
go-skylight get frame devices \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN --frame-id $SKYLIGHT_FRAME

# List available avatars
go-skylight get frame avatars \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN

# List available colors
go-skylight get frame colors \
  --user-id $SKYLIGHT_UID --token $SKYLIGHT_TOKEN
```

## Library Usage

### Authentication

```go
package main

import (
    "fmt"
    "log"

    "github.com/sebrandon1/go-skylight/lib"
)

func main() {
    // Option 1: Login with email/password (returns user ID and token)
    client, err := lib.NewClient("user@example.com", "password")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Authenticated as %s\n", client.UserID)

    // Option 2: Use an existing user ID and API token
    client, err = lib.NewClientWithToken("your-user-id", "your-api-token")
    if err != nil {
        log.Fatal(err)
    }
}
```

### Calendar Events

```go
// List events in a date range
events, err := client.ListCalendarEvents("frame-id", "2024-01-01", "2024-01-31")
if err != nil {
    log.Fatal(err)
}
for _, e := range events {
    fmt.Printf("%s - %s: %s\n", e.StartAt, e.EndAt, e.Title)
}

// Create an event
event, err := client.CreateCalendarEvent("frame-id", lib.CalendarEventData{
    Title:   "Team Meeting",
    StartAt: "2024-01-15T10:00:00Z",
    EndAt:   "2024-01-15T11:00:00Z",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Created event: %s\n", event.ID)

// Update an event
updated, err := client.UpdateCalendarEvent("frame-id", event.ID, lib.CalendarEventData{
    Title: "Updated Meeting",
})
if err != nil {
    log.Fatal(err)
}
fmt.Printf("Updated: %s\n", updated.Title)

// Delete an event
err = client.DeleteCalendarEvent("frame-id", event.ID)
if err != nil {
    log.Fatal(err)
}

// List linked source calendars
calendars, err := client.ListSourceCalendars("frame-id")
if err != nil {
    log.Fatal(err)
}
for _, c := range calendars {
    fmt.Printf("%s (%s) - enabled: %t\n", c.Name, c.Provider, c.Enabled)
}
```

### Chores

```go
// List chores (with optional filters — pass empty strings to skip)
chores, err := client.ListChores("frame-id", "2024-01-15", "pending", "assignee-id")
if err != nil {
    log.Fatal(err)
}
for _, c := range chores {
    fmt.Printf("[%s] %s (%d pts)\n", c.Status, c.Title, c.Points)
}

// Create a chore
chore, err := client.CreateChore("frame-id", lib.ChoreData{
    Title:  "Walk the dog",
    Points: 5,
})
if err != nil {
    log.Fatal(err)
}

// Update a chore
updated, err := client.UpdateChore("frame-id", chore.ID, lib.ChoreData{
    Status: "completed",
})
if err != nil {
    log.Fatal(err)
}

// Delete a chore
err = client.DeleteChore("frame-id", chore.ID)
```

### Lists & List Items

```go
// Create a list
list, err := client.CreateList("frame-id", lib.ListData{
    Title: "Grocery List",
    Color: "#4CAF50",
})
if err != nil {
    log.Fatal(err)
}

// Add items to the list
item, err := client.AddListItem("frame-id", list.ID, lib.ListItemData{
    Title: "Eggs",
})
if err != nil {
    log.Fatal(err)
}

// Mark item as completed
_, err = client.UpdateListItem("frame-id", list.ID, item.ID, lib.ListItemData{
    Completed: true,
})

// Get a list with all its items
fullList, err := client.GetList("frame-id", list.ID)
if err != nil {
    log.Fatal(err)
}
for _, item := range fullList.Items {
    status := "[ ]"
    if item.Completed {
        status = "[x]"
    }
    fmt.Printf("%s %s\n", status, item.Title)
}

// Delete an item, then the list
_ = client.DeleteListItem("frame-id", list.ID, item.ID)
_ = client.DeleteList("frame-id", list.ID)

// Create a task box item (quick task)
_, err = client.CreateTaskBoxItem("frame-id", lib.TaskBoxItemData{
    Title: "Call dentist",
})
```

### Rewards

```go
// Create a reward
reward, err := client.CreateReward("frame-id", lib.RewardData{
    Title:  "Movie Night",
    Points: 20,
})
if err != nil {
    log.Fatal(err)
}

// List all rewards
rewards, err := client.ListRewards("frame-id")
for _, r := range rewards {
    fmt.Printf("%s - %d pts (redeemed: %t)\n", r.Title, r.Points, r.Redeemed)
}

// Check points balance
points, err := client.GetRewardPoints("frame-id")
fmt.Printf("Total points: %d\n", points.Points)

// Redeem and unredeem
_ = client.RedeemReward("frame-id", reward.ID)
_ = client.UnredeemReward("frame-id", reward.ID)

// Update and delete
_, _ = client.UpdateReward("frame-id", reward.ID, lib.RewardData{Points: 30})
_ = client.DeleteReward("frame-id", reward.ID)
```

### Meals & Recipes

```go
// List meal categories
categories, err := client.ListMealCategories("frame-id")

// Create a recipe
recipe, err := client.CreateRecipe("frame-id", lib.RecipeData{
    Title:       "Spaghetti Bolognese",
    Description: "Classic Italian pasta dish",
    Ingredients: []string{"spaghetti", "ground beef", "tomato sauce", "onion", "garlic"},
})
if err != nil {
    log.Fatal(err)
}

// Get a recipe by ID
r, err := client.GetRecipe("frame-id", recipe.ID)
fmt.Printf("%s: %s\n", r.Title, r.Description)
for _, ing := range r.Ingredients {
    fmt.Printf("  - %s\n", ing)
}

// Schedule a meal
sitting, err := client.CreateMealSitting("frame-id", lib.MealSittingData{
    RecipeID: recipe.ID,
    Date:     "2024-01-15",
    MealType: "dinner",
})

// Add recipe ingredients to the grocery list
err = client.AddRecipeToGroceryList("frame-id", recipe.ID)

// Update and delete recipes
_, _ = client.UpdateRecipe("frame-id", recipe.ID, lib.RecipeData{Title: "Updated Spaghetti"})
_ = client.DeleteRecipe("frame-id", recipe.ID)

// List meal sittings and recipes
sittings, _ := client.ListMealSittings("frame-id")
recipes, _ := client.ListRecipes("frame-id")
```

### Family Members, Frame & Utilities

```go
// List family members (categories)
members, err := client.ListCategories("frame-id")
for _, m := range members {
    fmt.Printf("%s (color: %s)\n", m.Name, m.Color)
}

// Get frame info
frame, err := client.GetFrame("frame-id")
fmt.Printf("Frame: %s (timezone: %s)\n", frame.Name, frame.TimeZone)

// List connected devices
devices, err := client.ListDevices("frame-id")
for _, d := range devices {
    fmt.Printf("%s - online: %t\n", d.Name, d.Online)
}

// List available avatars and colors
avatars, _ := client.GetAvatars()
colors, _ := client.GetColors()
```

## API Coverage

| Resource | Operations |
|----------|-----------|
| Session | Login (email/password) |
| Calendar Events | List, Create, Update, Delete |
| Source Calendars | List |
| Chores | List (with filters), Create, Update, Delete |
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
