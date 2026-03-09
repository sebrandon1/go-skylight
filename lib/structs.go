package lib

// Session holds the authenticated user credentials extracted from login.
type Session struct {
	UserID   string
	APIToken string
}

// sessionResponse represents the JSON-API response from POST /api/sessions.
type sessionResponse struct {
	Data struct {
		ID         string `json:"id"`
		Attributes struct {
			Token string `json:"token"`
		} `json:"attributes"`
	} `json:"data"`
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

// Chore represents a chore/task (flattened from JSON-API response).
type Chore struct {
	ID         string `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Status     string `json:"status,omitempty"`
	DueDate    string `json:"due_date,omitempty"`
	Points     int    `json:"points,omitempty"`
	Recurring  bool   `json:"recurring,omitempty"`
	AssigneeID string `json:"assignee_id,omitempty"`
}

// choreAPIResponse wraps the JSON-API envelope for chore list responses.
type choreAPIResponse struct {
	Data []choreAPIEntry `json:"data"`
}

// choreAPIEntry represents a single chore in JSON-API format.
type choreAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Summary      string `json:"summary"`
		Status       string `json:"status"`
		Start        string `json:"start"`
		RewardPoints int    `json:"reward_points"`
		Recurring    bool   `json:"recurring"`
	} `json:"attributes"`
	Relationships struct {
		Category struct {
			Data *struct {
				ID string `json:"id"`
			} `json:"data"`
		} `json:"category"`
	} `json:"relationships"`
}

// choreAPISingleResponse wraps the JSON-API envelope for single chore responses.
type choreAPISingleResponse struct {
	Data choreAPIEntry `json:"data"`
}

// toChore converts a JSON-API chore entry to a flat Chore struct.
func (e *choreAPIEntry) toChore() Chore {
	c := Chore{
		ID:        e.ID,
		Title:     e.Attributes.Summary,
		Status:    e.Attributes.Status,
		DueDate:   e.Attributes.Start,
		Points:    e.Attributes.RewardPoints,
		Recurring: e.Attributes.Recurring,
	}
	if e.Relationships.Category.Data != nil {
		c.AssigneeID = e.Relationships.Category.Data.ID
	}
	return c
}

// ChoreRequest represents the API request body for creating/updating a chore.
type ChoreRequest struct {
	Chore ChoreData `json:"chore"`
}

// ChoreData holds the chore fields for create/update requests.
// JSON tags match the Skylight API field names.
type ChoreData struct {
	Title      string `json:"summary,omitempty"`
	DueDate    string `json:"start,omitempty"`
	Points     int    `json:"reward_points,omitempty"`
	Status     string `json:"status,omitempty"`
	AssigneeID string `json:"category_id,omitempty"`
	Recurring  bool   `json:"recurring,omitempty"`
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

// Reward represents a reward (flattened from JSON-API response).
type Reward struct {
	ID         string `json:"id,omitempty"`
	Title      string `json:"title,omitempty"`
	Points     int    `json:"points,omitempty"`
	EmojiIcon  string `json:"emoji_icon,omitempty"`
	CategoryID string `json:"category_id,omitempty"`
	Redeemed   bool   `json:"redeemed,omitempty"`
}

// rewardAPIResponse wraps the JSON-API envelope for reward list responses.
type rewardAPIResponse struct {
	Data []rewardAPIEntry `json:"data"`
}

// rewardAPIEntry represents a single reward in JSON-API format.
type rewardAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Name                string  `json:"name"`
		EmojiIcon           string  `json:"emoji_icon"`
		PointValue          int     `json:"point_value"`
		RespawnOnRedemption bool    `json:"respawn_on_redemption"`
		RedeemedAt          *string `json:"redeemed_at"`
	} `json:"attributes"`
	Relationships struct {
		Category struct {
			Data *struct {
				ID string `json:"id"`
			} `json:"data"`
		} `json:"category"`
	} `json:"relationships"`
}

// rewardAPISingleResponse wraps the JSON-API envelope for single reward responses.
type rewardAPISingleResponse struct {
	Data rewardAPIEntry `json:"data"`
}

// toReward converts a JSON-API reward entry to a flat Reward struct.
func (e *rewardAPIEntry) toReward() Reward {
	r := Reward{
		ID:        e.ID,
		Title:     e.Attributes.Name,
		Points:    e.Attributes.PointValue,
		EmojiIcon: e.Attributes.EmojiIcon,
		Redeemed:  e.Attributes.RedeemedAt != nil,
	}
	if e.Relationships.Category.Data != nil {
		r.CategoryID = e.Relationships.Category.Data.ID
	}
	return r
}

// RewardRequest represents the request body for creating/updating a reward.
type RewardRequest struct {
	Reward RewardData `json:"reward"`
}

// RewardData holds the reward fields for create/update requests.
// JSON tags match the Skylight API field names.
type RewardData struct {
	Title               string `json:"name,omitempty"`
	Points              int    `json:"point_value,omitempty"`
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
