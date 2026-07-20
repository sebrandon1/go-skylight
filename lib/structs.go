package lib

import "encoding/json"

// apiRelationshipData is the JSON-API "data" object inside a to-one relationship.
type apiRelationshipData struct {
	ID string `json:"id"`
}

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

// OAuthTokenResponse holds the OAuth2 token response from Skylight.
type OAuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

// CalendarEvent represents a calendar event.
type CalendarEvent struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	StartAt     string `json:"starts_at,omitempty"`
	EndAt       string `json:"ends_at,omitempty"`
	AllDay      bool   `json:"all_day,omitempty"`
	Color       string `json:"color,omitempty"`
	CategoryID  string `json:"category_id,omitempty"`
	CreatedAt   string `json:"created_at,omitempty"`
	UpdatedAt   string `json:"updated_at,omitempty"`
}

const CalendarEventTypeCountdown = "countdown"

// CalendarEventData holds the event fields for create/update requests.
// Field names match the Skylight API JSON field names.
type CalendarEventData struct {
	Title       string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	StartAt     string `json:"starts_at,omitempty"`
	EndAt       string `json:"ends_at,omitempty"`
	AllDay      bool   `json:"all_day,omitempty"`
	Color       string `json:"color,omitempty"`
	CategoryID  string `json:"category_id,omitempty"`
	EventType   string `json:"event_type,omitempty"`
}

// calendarAPIResponse wraps the JSON-API envelope for calendar event list responses.
type calendarAPIResponse struct {
	Data []calendarAPIEntry `json:"data"`
}

// calendarAPISingleResponse wraps the JSON-API envelope for single calendar event responses.
type calendarAPISingleResponse struct {
	Data calendarAPIEntry `json:"data"`
}

// calendarAPIEntry represents a single calendar event in JSON-API format.
type calendarAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Summary     string `json:"summary"`
		Description string `json:"description"`
		StartsAt    string `json:"starts_at"`
		EndsAt      string `json:"ends_at"`
		AllDay      bool   `json:"all_day"`
		Color       string `json:"color"`
		CreatedAt   string `json:"created_at"`
		UpdatedAt   string `json:"updated_at"`
	} `json:"attributes"`
	Relationships struct {
		Categories struct {
			Data []apiRelationshipData `json:"data"`
		} `json:"categories"`
	} `json:"relationships"`
}

func (e *calendarAPIEntry) toCalendarEvent() CalendarEvent {
	ev := CalendarEvent{
		ID:          e.ID,
		Title:       e.Attributes.Summary,
		Description: e.Attributes.Description,
		StartAt:     e.Attributes.StartsAt,
		EndAt:       e.Attributes.EndsAt,
		AllDay:      e.Attributes.AllDay,
		Color:       e.Attributes.Color,
		CreatedAt:   e.Attributes.CreatedAt,
		UpdatedAt:   e.Attributes.UpdatedAt,
	}
	if len(e.Relationships.Categories.Data) > 0 {
		ev.CategoryID = e.Relationships.Categories.Data[0].ID
	}
	return ev
}

// SourceCalendar represents a source calendar linked to a frame.
type SourceCalendar struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"label,omitempty"`
	Color    string `json:"color,omitempty"`
	Enabled  bool   `json:"enabled,omitempty"`
	Provider string `json:"kind,omitempty"`
}

// sourceCalendarAPIResponse wraps the JSON-API envelope for source calendar list responses.
type sourceCalendarAPIResponse struct {
	Data []sourceCalendarAPIEntry `json:"data"`
}

// sourceCalendarAPIEntry represents a single source calendar in JSON-API format.
type sourceCalendarAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Label   string `json:"label"`
		Kind    string `json:"kind"`
		Enabled bool   `json:"editable"`
	} `json:"attributes"`
}

func (e *sourceCalendarAPIEntry) toSourceCalendar() SourceCalendar {
	return SourceCalendar{
		ID:       e.ID,
		Name:     e.Attributes.Label,
		Provider: e.Attributes.Kind,
		Enabled:  e.Attributes.Enabled,
	}
}

// Chore represents a chore/task (flattened from JSON-API response).
type Chore struct {
	ID          string `json:"id,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Status      string `json:"status,omitempty"`
	DueDate     string `json:"due_date,omitempty"`
	Points      int    `json:"points,omitempty"`
	Recurring   bool   `json:"recurring"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	UpForGrabs  bool   `json:"up_for_grabs,omitempty"`
}

// choreAPIResponse wraps the JSON-API envelope for chore list responses.
type choreAPIResponse struct {
	Data []choreAPIEntry `json:"data"`
}

type choreAPIAttributes struct {
	Summary      string `json:"summary"`
	Description  string `json:"description"`
	Status       string `json:"status"`
	Start        string `json:"start"`
	RewardPoints int    `json:"reward_points"`
	Recurring    bool   `json:"recurring"`
	UpForGrabs   bool   `json:"up_for_grabs"`
}

// choreAPIEntry represents a single chore in JSON-API format.
type choreAPIEntry struct {
	ID            string             `json:"id"`
	Attributes    choreAPIAttributes `json:"attributes"`
	Relationships struct {
		Category struct {
			Data *apiRelationshipData `json:"data"`
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
		ID:          e.ID,
		Title:       e.Attributes.Summary,
		Description: e.Attributes.Description,
		Status:      e.Attributes.Status,
		DueDate:     e.Attributes.Start,
		Points:      e.Attributes.RewardPoints,
		Recurring:   e.Attributes.Recurring,
		UpForGrabs:  e.Attributes.UpForGrabs,
	}
	if e.Relationships.Category.Data != nil {
		c.AssigneeID = e.Relationships.Category.Data.ID
	}
	return c
}

// ChoreData holds the chore fields for create/update requests.
// JSON tags match the Skylight API field names.
type ChoreData struct {
	Title       string `json:"summary,omitempty"`
	Description string `json:"description,omitempty"`
	DueDate     string `json:"start,omitempty"`
	Points      int    `json:"reward_points,omitempty"`
	Status      string `json:"status,omitempty"`
	AssigneeID  string `json:"category_id,omitempty"`
	Recurring   bool   `json:"recurring,omitempty"`
	UpForGrabs  bool   `json:"up_for_grabs,omitempty"`
}

// List represents a list (e.g., grocery list, todo list).
type List struct {
	ID            string     `json:"id,omitempty"`
	Title         string     `json:"label,omitempty"`
	Color         string     `json:"color,omitempty"`
	Kind          string     `json:"kind,omitempty"`
	HideFromFrame bool       `json:"hide_from_frame"`
	Items         []ListItem `json:"list_items,omitempty"`
	CreatedAt     string     `json:"created_at,omitempty"`
	UpdatedAt     string     `json:"updated_at,omitempty"`
}

// ListItem represents an item within a list.
type ListItem struct {
	ID        string `json:"id,omitempty"`
	Title     string `json:"label,omitempty"`
	Completed bool   `json:"completed,omitempty"`
	Status    string `json:"status,omitempty"`
	Position  int    `json:"position,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// ListData holds the list fields for create/update requests.
type ListData struct {
	Title         string `json:"label,omitempty"`
	Color         string `json:"color,omitempty"`
	Kind          string `json:"kind,omitempty"`
	HideFromFrame *bool  `json:"hide_from_frame,omitempty"`
}

// ListItemData holds the list item fields for create/update requests.
type ListItemData struct {
	Title     string `json:"label,omitempty"`
	Completed bool   `json:"-"`
	Position  int    `json:"position,omitempty"`
}

// listItemSendData is the internal struct used when sending list item requests to the API.
type listItemSendData struct {
	Label    string `json:"label,omitempty"`
	Status   string `json:"status,omitempty"`
	Position int    `json:"position,omitempty"`
}

// listAPIResponse wraps the JSON-API envelope for list responses.
type listAPIResponse struct {
	Data []listAPIEntry `json:"data"`
}

// listAPISingleResponse wraps the JSON-API envelope for single list responses.
type listAPISingleResponse struct {
	Data     listAPIEntry       `json:"data"`
	Included []listItemAPIEntry `json:"included"`
}

// listAPIEntry represents a single list in JSON-API format.
type listAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Label         string `json:"label"`
		Color         string `json:"color"`
		Kind          string `json:"kind"`
		HideFromFrame bool   `json:"hide_from_frame"`
		CreatedAt     string `json:"created_at"`
		UpdatedAt     string `json:"updated_at"`
	} `json:"attributes"`
}

func (e *listAPIEntry) toList() List {
	return List{
		ID:            e.ID,
		Title:         e.Attributes.Label,
		Color:         e.Attributes.Color,
		Kind:          e.Attributes.Kind,
		HideFromFrame: e.Attributes.HideFromFrame,
		CreatedAt:     e.Attributes.CreatedAt,
		UpdatedAt:     e.Attributes.UpdatedAt,
	}
}

// listItemAPIEntry represents a single list item in JSON-API format.
type listItemAPIEntry struct {
	ID         string `json:"id"`
	Type       string `json:"type"`
	Attributes struct {
		Label     string `json:"label"`
		Status    string `json:"status"`
		Position  int    `json:"position"`
		CreatedAt string `json:"created_at"`
	} `json:"attributes"`
}

// listItemAPISingleResponse wraps the JSON-API envelope for single list item responses.
type listItemAPISingleResponse struct {
	Data listItemAPIEntry `json:"data"`
}

func (e *listItemAPIEntry) toListItem() ListItem {
	return ListItem{
		ID:        e.ID,
		Title:     e.Attributes.Label,
		Status:    e.Attributes.Status,
		Completed: e.Attributes.Status == listItemStatusCompleted,
		Position:  e.Attributes.Position,
		CreatedAt: e.Attributes.CreatedAt,
	}
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

// rewardAPISingleResponse wraps the JSON-API envelope for single reward responses.
type rewardAPISingleResponse struct {
	Data rewardAPIEntry `json:"data"`
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
			Data *apiRelationshipData `json:"data"`
		} `json:"category"`
	} `json:"relationships"`
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

// RewardData holds the reward fields for create/update requests.
// JSON tags match the Skylight API field names.
type RewardData struct {
	Title               string `json:"name,omitempty"`
	Points              int    `json:"point_value,omitempty"`
	EmojiIcon           string `json:"emoji_icon,omitempty"`
	RespawnOnRedemption *bool  `json:"respawn_on_redemption,omitempty"`
	CategoryIDs         []int  `json:"category_ids,omitempty"`
}

// removeStarsRequest is the request body for updating reward points.
// The Skylight API uses POST /frames/{id}/reward_points with category_ids array
// and a negative points value to deduct stars.
type removeStarsRequest struct {
	CategoryIDs []int `json:"category_ids"`
	Points      int   `json:"points"`
}

// ChoreCompletionData is the request body for the chore completions endpoint.
type ChoreCompletionData struct {
	Status       string `json:"status"`
	InstanceDate string `json:"instance_date,omitempty"`
}

// ChoreListOptions holds optional filters for listing chores.
type ChoreListOptions struct {
	Date        string
	Status      string
	AssigneeID  string
	After       string
	Before      string
	IncludeLate bool
	UpForGrabs  bool
}

// RewardPointEntry represents a per-category point balance.
type RewardPointEntry struct {
	CategoryID          int `json:"category_id"`
	CurrentPointBalance int `json:"current_point_balance"`
}

// Recipe represents a meal recipe.
type Recipe struct {
	ID             string   `json:"id,omitempty"`
	Title          string   `json:"summary,omitempty"`
	Description    string   `json:"description,omitempty"`
	Ingredients    []string `json:"ingredients,omitempty"`
	URL            string   `json:"url,omitempty"`
	ImageURL       string   `json:"image_url,omitempty"`
	MealCategoryID string   `json:"meal_category_id,omitempty"`
	CreatedAt      string   `json:"created_at,omitempty"`
	UpdatedAt      string   `json:"updated_at,omitempty"`
}

// RecipeData holds the recipe fields for create/update requests.
type RecipeData struct {
	Title          string   `json:"summary,omitempty"`
	Description    string   `json:"description,omitempty"`
	Ingredients    []string `json:"ingredients,omitempty"`
	URL            string   `json:"url,omitempty"`
	MealCategoryID string   `json:"meal_category_id,omitempty"`
}

// recipeAPIResponse wraps the JSON-API envelope for recipe list responses.
type recipeAPIResponse struct {
	Data []recipeAPIEntry `json:"data"`
}

// recipeAPISingleResponse wraps the JSON-API envelope for single recipe responses.
type recipeAPISingleResponse struct {
	Data recipeAPIEntry `json:"data"`
}

// recipeAPIEntry represents a single recipe in JSON-API format.
type recipeAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Summary     string   `json:"summary"`
		Description string   `json:"description"`
		Ingredients []string `json:"ingredients"`
		URL         string   `json:"url"`
		ImageURL    string   `json:"image_url"`
		CreatedAt   string   `json:"created_at"`
		UpdatedAt   string   `json:"updated_at"`
	} `json:"attributes"`
	Relationships struct {
		MealCategory struct {
			Data *apiRelationshipData `json:"data"`
		} `json:"meal_category"`
	} `json:"relationships"`
}

func (e *recipeAPIEntry) toRecipe() Recipe {
	r := Recipe{
		ID:          e.ID,
		Title:       e.Attributes.Summary,
		Description: e.Attributes.Description,
		Ingredients: e.Attributes.Ingredients,
		URL:         e.Attributes.URL,
		ImageURL:    e.Attributes.ImageURL,
		CreatedAt:   e.Attributes.CreatedAt,
		UpdatedAt:   e.Attributes.UpdatedAt,
	}
	if e.Relationships.MealCategory.Data != nil {
		r.MealCategoryID = e.Relationships.MealCategory.Data.ID
	}
	return r
}

// MealSitting represents a meal sitting (scheduled meal).
type MealSitting struct {
	ID             string `json:"id,omitempty"`
	Summary        string `json:"summary,omitempty"`
	RecipeID       string `json:"recipe_id,omitempty"`
	MealCategoryID string `json:"meal_category_id,omitempty"`
	Date           string `json:"date,omitempty"`
	CreatedAt      string `json:"created_at,omitempty"`
	UpdatedAt      string `json:"updated_at,omitempty"`
}

// MealSittingData holds the meal sitting fields for create requests.
type MealSittingData struct {
	Summary        string `json:"summary,omitempty"`
	RecipeID       string `json:"meal_recipe_id,omitempty"`
	MealCategoryID string `json:"meal_category_id,omitempty"`
	Date           string `json:"date,omitempty"`
}

// MealSittingListOptions holds filters for listing meal sittings.
type MealSittingListOptions struct {
	DateMin string
	DateMax string
}

// mealSittingAPIResponse wraps the JSON-API envelope for meal sitting list/create responses.
type mealSittingAPIResponse struct {
	Data []mealSittingAPIEntry `json:"data"`
}

// mealSittingAPISingleResponse wraps the JSON-API envelope for single meal sitting responses.
type mealSittingAPISingleResponse struct {
	Data mealSittingAPIEntry `json:"data"`
}

// SittingWithRecipe holds a meal sitting and its linked recipe details.
type SittingWithRecipe struct {
	Sitting MealSitting `json:"sitting"`
	Recipe  *Recipe     `json:"recipe,omitempty"`
}

// mealSittingAPIEntry represents a single meal sitting in JSON-API format.
type mealSittingAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Summary   string   `json:"summary"`
		Instances []string `json:"instances"`
	} `json:"attributes"`
	Relationships struct {
		MealCategory struct {
			Data *apiRelationshipData `json:"data"`
		} `json:"meal_category"`
		MealRecipe struct {
			Data *apiRelationshipData `json:"data"`
		} `json:"meal_recipe"`
	} `json:"relationships"`
}

func (e *mealSittingAPIEntry) toMealSitting() MealSitting {
	s := MealSitting{
		ID:      e.ID,
		Summary: e.Attributes.Summary,
	}
	if len(e.Attributes.Instances) > 0 {
		s.Date = e.Attributes.Instances[0]
	}
	if e.Relationships.MealCategory.Data != nil {
		s.MealCategoryID = e.Relationships.MealCategory.Data.ID
	}
	if e.Relationships.MealRecipe.Data != nil {
		s.RecipeID = e.Relationships.MealRecipe.Data.ID
	}
	return s
}

// MealCategory represents a meal category.
type MealCategory struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"label,omitempty"`
	Color string `json:"color,omitempty"`
}

// mealCategoryAPIResponse wraps the JSON-API envelope for meal category list responses.
type mealCategoryAPIResponse struct {
	Data []mealCategoryAPIEntry `json:"data"`
}

// mealCategoryAPIEntry represents a single meal category in JSON-API format.
type mealCategoryAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Label string `json:"label"`
		Color string `json:"color"`
	} `json:"attributes"`
}

func (e *mealCategoryAPIEntry) toMealCategory() MealCategory {
	return MealCategory{
		ID:    e.ID,
		Name:  e.Attributes.Label,
		Color: e.Attributes.Color,
	}
}

// Category represents a family member/category.
type Category struct {
	ID        string `json:"id,omitempty"`
	Name      string `json:"label,omitempty"`
	Color     string `json:"color,omitempty"`
	AvatarURL string `json:"profile_pic_url,omitempty"`
	CreatedAt string `json:"created_at,omitempty"`
	UpdatedAt string `json:"updated_at,omitempty"`
}

// categoryAPIResponse wraps the JSON-API envelope for category list responses.
type categoryAPIResponse struct {
	Data []categoryAPIEntry `json:"data"`
}

// categoryAPIEntry represents a single category in JSON-API format.
type categoryAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Label         string  `json:"label"`
		Color         string  `json:"color"`
		ProfilePicURL *string `json:"profile_pic_url"`
	} `json:"attributes"`
}

// categoryAPISingleResponse wraps the JSON-API envelope for single category responses.
type categoryAPISingleResponse struct {
	Data categoryAPIEntry `json:"data"`
}

// CategoryData holds the category fields for update requests.
type CategoryData struct {
	Name  string `json:"label,omitempty"`
	Color string `json:"color,omitempty"`
}

func (e *categoryAPIEntry) toCategory() Category {
	c := Category{
		ID:    e.ID,
		Name:  e.Attributes.Label,
		Color: e.Attributes.Color,
	}
	if e.Attributes.ProfilePicURL != nil {
		c.AvatarURL = *e.Attributes.ProfilePicURL
	}
	return c
}

// Frame represents a Skylight frame/device group.
// FeatureState represents the enabled/disabled state of a single frame feature.
type FeatureState struct {
	Enabled bool `json:"enabled"`
}

type Frame struct {
	ID            string                  `json:"id,omitempty"`
	Name          string                  `json:"name,omitempty"`
	TimeZone      string                  `json:"timezone,omitempty"`
	Plus          bool                    `json:"plus,omitempty"`
	FeatureBundle map[string]FeatureState `json:"feature_bundle,omitempty"`
	CreatedAt     string                  `json:"created_at,omitempty"`
	UpdatedAt     string                  `json:"updated_at,omitempty"`
}

// frameAPIResponse wraps the JSON-API envelope for single frame responses.
type frameAPIResponse struct {
	Data frameAPIEntry `json:"data"`
}

// framesAPIResponse wraps the JSON-API envelope for frame list responses.
type framesAPIResponse struct {
	Data []frameAPIEntry `json:"data"`
}

// frameAPIEntry represents a frame in JSON-API format.
type frameAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Name          string                     `json:"name"`
		TimeZone      string                     `json:"timezone"`
		Plus          bool                       `json:"plus"`
		FeatureBundle map[string]json.RawMessage `json:"feature_bundle"`
	} `json:"attributes"`
}

func (e *frameAPIEntry) toFrame() Frame {
	bundle := make(map[string]FeatureState, len(e.Attributes.FeatureBundle))
	for k, raw := range e.Attributes.FeatureBundle {
		var fs FeatureState
		if err := json.Unmarshal(raw, &fs); err == nil {
			bundle[k] = fs
		}
	}
	return Frame{
		ID:            e.ID,
		Name:          e.Attributes.Name,
		TimeZone:      e.Attributes.TimeZone,
		Plus:          e.Attributes.Plus,
		FeatureBundle: bundle,
	}
}

// Device represents a physical Skylight device.
type Device struct {
	ID                   string `json:"id,omitempty"`
	Name                 string `json:"name,omitempty"`
	Online               bool   `json:"online,omitempty"`
	Timezone             string `json:"timezone,omitempty"`
	Role                 string `json:"role,omitempty"`
	CategoryID           string `json:"category_id,omitempty"`
	Brightness           int    `json:"brightness,omitempty"`
	SleepsAt             string `json:"sleeps_at,omitempty"`
	WakesAt              string `json:"wakes_at,omitempty"`
	CurrentlySleeping    bool   `json:"currently_sleeping,omitempty"`
	SleepModeOn          bool   `json:"sleep_mode_on,omitempty"`
	SleepMode            string `json:"sleep_mode,omitempty"`
	Nightlight           bool   `json:"nightlight,omitempty"`
	NightlightBrightness int    `json:"nightlight_brightness,omitempty"`
}

// deviceAPIResponse wraps the JSON-API envelope for device list responses.
type deviceAPIResponse struct {
	Data []deviceAPIEntry `json:"data"`
}

// deviceAPIEntry represents a single device in JSON-API format.
type deviceAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Name                 string `json:"name"`
		Activated            bool   `json:"activated"`
		Timezone             string `json:"timezone"`
		Role                 string `json:"role"`
		CategoryID           string `json:"category_id"`
		Brightness           int    `json:"brightness"`
		SleepsAt             string `json:"sleeps_at"`
		WakesAt              string `json:"wakes_at"`
		CurrentlySleeping    bool   `json:"currently_sleeping"`
		SleepModeOn          bool   `json:"sleep_mode_on"`
		SleepMode            string `json:"sleep_mode"`
		Nightlight           bool   `json:"nightlight"`
		NightlightBrightness int    `json:"nightlight_brightness"`
	} `json:"attributes"`
}

func (e *deviceAPIEntry) toDevice() Device {
	return Device{
		ID:                   e.ID,
		Name:                 e.Attributes.Name,
		Online:               e.Attributes.Activated,
		Timezone:             e.Attributes.Timezone,
		Role:                 e.Attributes.Role,
		CategoryID:           e.Attributes.CategoryID,
		Brightness:           e.Attributes.Brightness,
		SleepsAt:             e.Attributes.SleepsAt,
		WakesAt:              e.Attributes.WakesAt,
		CurrentlySleeping:    e.Attributes.CurrentlySleeping,
		SleepModeOn:          e.Attributes.SleepModeOn,
		SleepMode:            e.Attributes.SleepMode,
		Nightlight:           e.Attributes.Nightlight,
		NightlightBrightness: e.Attributes.NightlightBrightness,
	}
}

// Avatar represents an available avatar option.
type Avatar struct {
	ID       string `json:"id,omitempty"`
	Name     string `json:"name,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// avatarAPIResponse wraps the JSON-API envelope for avatar list responses.
type avatarAPIResponse struct {
	Data []avatarAPIEntry `json:"data"`
}

// avatarAPIEntry represents a single avatar in JSON-API format.
type avatarAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Name     string `json:"name"`
		ImageURL string `json:"image_url"`
	} `json:"attributes"`
}

func (e *avatarAPIEntry) toAvatar() Avatar {
	return Avatar{
		ID:       e.ID,
		Name:     e.Attributes.Name,
		ImageURL: e.Attributes.ImageURL,
	}
}

// Color represents an available color option.
type Color struct {
	Name string `json:"name,omitempty"`
	Hex  string `json:"hex,omitempty"`
}

// colorAPIResponse wraps the data envelope for color list responses.
type colorAPIResponse struct {
	Data []Color `json:"data"`
}

// BountyData holds the input for creating a bounty (chore + paired reward).
type BountyData struct {
	Title       string `json:"title"`
	Points      int    `json:"points"`
	DueDate     string `json:"due_date,omitempty"`
	AssigneeID  string `json:"assignee_id,omitempty"`
	CategoryIDs []int  `json:"category_ids,omitempty"`
	Recurring   bool   `json:"recurring,omitempty"`
	RewardTitle string `json:"reward_title"`
	EmojiIcon   string `json:"emoji_icon,omitempty"`
}

// Bounty represents a chore with a paired reward.
type Bounty struct {
	Chore  Chore  `json:"chore"`
	Reward Reward `json:"reward"`
}

// RotationData holds the input for creating a chore rotation schedule.
type RotationData struct {
	Chores      []string `json:"chores"`
	AssigneeIDs []string `json:"assignee_ids"`
	StartDate   string   `json:"start_date"`
	Weeks       int      `json:"weeks"`
	Points      int      `json:"points,omitempty"`
}

// RotationResult holds the chores created by a rotation schedule.
type RotationResult struct {
	Chores []Chore `json:"chores"`
}

// MealPlanData holds the input for scheduling meals from a recipe list.
type MealPlanData struct {
	RecipeIDs   []string
	CategoryIDs []string
	StartDate   string
}

// MealPlanResult holds the sittings created by a meal plan.
type MealPlanResult struct {
	Sittings []MealSitting `json:"sittings"`
}

// Photo represents a photo (or video) message on a Skylight frame.
type Photo struct {
	ID           string `json:"id"`
	Status       string `json:"status"`
	AssetType    string `json:"asset_type"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	ThumbnailURL string `json:"thumbnail_url"`
	AssetURL     string `json:"asset_url"`
	SenderID     int    `json:"sender_id"`
}

// PhotoListOptions holds optional filters for listing photos.
type PhotoListOptions struct {
	// PageToken controls pagination. Use "__START__" for the first page.
	PageToken string
}

// PhotoUploadResponse holds the result of a successful upload URL request.
type PhotoUploadResponse struct {
	// UploadURL is the presigned S3 PUT URL to upload the image bytes to.
	UploadURL  string   `json:"url"`
	Key        string   `json:"key"`
	GetURL     string   `json:"get_url"`
	MessageIDs []int    `json:"message_ids"`
	FrameNames []string `json:"frame_names"`
}

// photoAPIResponse is the JSON-API envelope for messages list responses.
type photoAPIResponse struct {
	Data []photoAPIEntry `json:"data"`
	Meta *photoMeta      `json:"meta"`
}

// photoMeta holds pagination tokens returned by the messages endpoint.
type photoMeta struct {
	NextPageToken string `json:"next_page_token"`
	SyncToken     string `json:"sync_token"`
}

// photoAPIEntry is a single message_status item in JSON-API format.
type photoAPIEntry struct {
	ID         string `json:"id"`
	Attributes struct {
		Status       string `json:"status"`
		AssetType    string `json:"asset_type"`
		CreatedAt    string `json:"created_at"`
		UpdatedAt    string `json:"updated_at"`
		ThumbnailURL string `json:"thumbnail_url"`
		AssetURL     string `json:"asset_url"`
		SenderID     int    `json:"sender_id"`
	} `json:"attributes"`
}

func (e photoAPIEntry) toPhoto() Photo {
	return Photo{
		ID:           e.ID,
		Status:       e.Attributes.Status,
		AssetType:    e.Attributes.AssetType,
		CreatedAt:    e.Attributes.CreatedAt,
		UpdatedAt:    e.Attributes.UpdatedAt,
		ThumbnailURL: e.Attributes.ThumbnailURL,
		AssetURL:     e.Attributes.AssetURL,
		SenderID:     e.Attributes.SenderID,
	}
}

// photoUploadURLRequest is the request body for POST /api/upload_url.
type photoUploadURLRequest struct {
	Ext      string   `json:"ext"`
	FrameIDs []string `json:"frame_ids"`
	Caption  string   `json:"caption,omitempty"`
}

// photoUploadURLResponse is the API response for POST /api/upload_url.
type photoUploadURLResponse struct {
	Data PhotoUploadResponse `json:"data"`
}

// photoDeleteRequest is the body for DELETE /messages/destroy_multiple.
type photoDeleteRequest struct {
	MessageIDs []int `json:"message_ids"`
}
