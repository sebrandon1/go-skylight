package lib

// Session represents the response from POST /api/sessions.
type Session struct {
	UserID   string `json:"user_id,omitempty"`
	APIToken string `json:"api_token,omitempty"`
}

// SessionRequest represents the login request body.
type SessionRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CalendarEvent represents a calendar event.
type CalendarEvent struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	StartAt     string `json:"start_at,omitempty"`
	EndAt       string `json:"end_at,omitempty"`
	AllDay      bool   `json:"all_day,omitempty"`
	Color       string `json:"color,omitempty"`
	CategoryID  string `json:"category_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// CalendarEventRequest represents the request body for creating/updating a calendar event.
type CalendarEventRequest struct {
	CalendarEvent CalendarEventData `json:"calendar_event"`
}

// CalendarEventData holds the event fields for create/update requests.
type CalendarEventData struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	StartAt     string `json:"start_at,omitempty"`
	EndAt       string `json:"end_at,omitempty"`
	AllDay      bool   `json:"all_day,omitempty"`
	Color       string `json:"color,omitempty"`
	CategoryID  string `json:"category_id,omitempty"`
}

// SourceCalendar represents a source calendar linked to a frame.
type SourceCalendar struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	Color    string `json:"color,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	Provider string `json:"provider,omitempty"`
}

// Chore represents a chore/task.
type Chore struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Status      string `json:"status,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	Points      int    `json:"points,omitempty"`
	Recurring   bool   `json:"recurring,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

// ChoreRequest represents the request body for creating/updating a chore.
type ChoreRequest struct {
	Chore ChoreData `json:"chore"`
}

// ChoreData holds the chore fields for create/update requests.
type ChoreData struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Status      string `json:"status,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	Points      int    `json:"points,omitempty"`
	Recurring   bool   `json:"recurring,omitempty"`
}

// List represents a list (e.g., grocery list, todo list).
type List struct {
	ID        string     `json:"id,omitempty"`
	Title     string     `json:"title,omitempty"`
	Color     string     `json:"color,omitempty"`
	Items     []ListItem `json:"list_items,omitempty"`
	CreatedAt string     `json:"created_at,omitempty"`
	UpdatedAt string     `json:"updated_at,omitempty"`
}

// ListItem represents an item within a list.
type ListItem struct {
	ID        string `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	Completed bool   `json:"completed,omitempty"`
	Position  int    `json:"position,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ListRequest represents the request body for creating/updating a list.
type ListRequest struct {
	List ListData `json:"list"`
}

// ListData holds the list fields for create/update requests.
type ListData struct {
	Title string `json:"title,omitempty"`
	Color string `json:"color,omitempty"`
}

// ListItemRequest represents the request body for creating/updating a list item.
type ListItemRequest struct {
	ListItem ListItemData `json:"list_item"`
}

// ListItemData holds the list item fields for create/update requests.
type ListItemData struct {
	Title     string `json:"title,omitempty"`
	Completed bool   `json:"completed,omitempty"`
	Position  int    `json:"position,omitempty"`
}

// TaskBoxItem represents a task box item.
type TaskBoxItem struct {
	ID        string `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
}

// TaskBoxItemRequest represents the request body for creating a task box item.
type TaskBoxItemRequest struct {
	TaskBoxItem TaskBoxItemData `json:"task_box_item"`
}

// TaskBoxItemData holds the task box item fields for create requests.
type TaskBoxItemData struct {
	Title string `json:"title,omitempty"`
}

// Reward represents a reward.
type Reward struct {
	ID        string `json:"id,omitempty"`
	Title     string `json:"title,omitempty"`
	Points    int    `json:"points,omitempty"`
	Redeemed  bool   `json:"redeemed,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// RewardRequest represents the request body for creating/updating a reward.
type RewardRequest struct {
	Reward RewardData `json:"reward"`
}

// RewardData holds the reward fields for create/update requests.
type RewardData struct {
	Title               string `json:"title,omitempty"`
	Points              int    `json:"points,omitempty"`
	EmojiIcon           string `json:"emoji_icon,omitempty"`
	RespawnOnRedemption *bool  `json:"respawn_on_redemption,omitempty"`
	CategoryIDs         []int  `json:"category_ids,omitempty"`
}

// ChoreListOptions holds optional filters for listing chores.
type ChoreListOptions struct {
	Date        string
	Status      string
	AssigneeID  string
	After       string
	Before      string
	IncludeLate bool
}

// RewardPointEntry represents a per-category point balance.
type RewardPointEntry struct {
	CategoryID          int `json:"category_id"`
	CurrentPointBalance int `json:"current_point_balance"`
}

// Recipe represents a meal recipe.
type Recipe struct {
	ID          string   `json:"id,omitempty"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Ingredients []string `json:"ingredients,omitempty"`
	URL         string   `json:"url,omitempty"`
	ImageURL    string   `json:"image_url,omitempty"`
	CategoryID  string   `json:"category_id,omitempty"`
	CreatedAt   string   `json:"created_at,omitempty"`
	UpdatedAt   string   `json:"updated_at,omitempty"`
}

// RecipeRequest represents the request body for creating/updating a recipe.
type RecipeRequest struct {
	Recipe RecipeData `json:"recipe"`
}

// RecipeData holds the recipe fields for create/update requests.
type RecipeData struct {
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Ingredients []string `json:"ingredients,omitempty"`
	URL         string   `json:"url,omitempty"`
	ImageURL    string   `json:"image_url,omitempty"`
	CategoryID  string   `json:"category_id,omitempty"`
}

// MealSitting represents a meal sitting (scheduled meal).
type MealSitting struct {
	ID        string `json:"id,omitempty"`
	RecipeID  string `json:"recipe_id,omitempty"`
	Date      string `json:"date,omitempty"`
	MealType  string `json:"meal_type,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// MealSittingRequest represents the request body for creating a meal sitting.
type MealSittingRequest struct {
	MealSitting MealSittingData `json:"meal_sitting"`
}

// MealSittingData holds the meal sitting fields for create requests.
type MealSittingData struct {
	RecipeID string `json:"recipe_id,omitempty"`
	Date     string `json:"date,omitempty"`
	MealType string `json:"meal_type,omitempty"`
}

// MealCategory represents a meal category.
type MealCategory struct {
	ID   string `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

// Category represents a family member/category.
type Category struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	Color     string `json:"color,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// Frame represents a Skylight frame/device group.
type Frame struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	TimeZone  string `json:"time_zone,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// Device represents a physical Skylight device.
type Device struct {
	ID           string `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Model        string `json:"model,omitempty"`
	FirmwareVer  string `json:"firmware_version,omitempty"`
	LastOnlineAt string `json:"last_online_at,omitempty"`
	Online       bool   `json:"online,omitempty"`
	CreatedAt    string `json:"created_at,omitempty"`
	UpdatedAt    string `json:"updated_at,omitempty"`
}

// Avatar represents an available avatar option.
type Avatar struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// Color represents an available color option.
type Color struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}
