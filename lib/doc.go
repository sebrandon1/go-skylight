// Package lib provides a Go client for the Skylight Calendar API.
//
// # Authentication
//
// Create a client using email/password (auto-login) or a pre-existing user ID
// and API token:
//
//	// Option 1: email/password — calls POST /api/sessions automatically
//	client, err := lib.NewClient("user@example.com", "password")
//
//	// Option 2: pre-existing credentials (skip the login round-trip)
//	client, err := lib.NewClientWithToken("user-id", "api-token")
//
// # Resources
//
// The client exposes methods for every Skylight resource:
//
//   - Calendar events: ListCalendarEvents, CreateCalendarEvent,
//     UpdateCalendarEvent, DeleteCalendarEvent, ListSourceCalendars
//   - Chores: ListChores, CreateChore, UpdateChore, DeleteChore
//   - Rewards: ListRewards, CreateReward, UpdateReward, DeleteReward,
//     RedeemReward, UnredeemReward, GetRewardPoints
//   - Lists and items: ListLists, GetList, CreateList, UpdateList, DeleteList,
//     AddListItem, UpdateListItem, DeleteListItem, CreateTaskBoxItem
//   - Meals: ListRecipes, GetRecipe, CreateRecipe, UpdateRecipe, DeleteRecipe,
//     ListMealSittings, CreateMealSitting, ListMealCategories,
//     AddRecipeToGroceryList
//   - Categories (family members): ListCategories
//   - Frame: GetFrame, ListDevices, GetAvatars, GetColors
//   - Dashboard (today's aggregate): GetDashboard
//   - Bounties (chore+reward pairs): CreateBounty, ListBounties
//   - Chore rotations: CreateChoreRotation
//
// # Base URL
//
// All requests target SkylightURL (defaults to https://app.ourskylight.com/api).
// Tests override it directly:
//
//	old := lib.SkylightURL
//	lib.SkylightURL = testServer.URL + "/api"
//	defer func() { lib.SkylightURL = old }()
package lib
