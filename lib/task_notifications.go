package lib

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// TaskNotificationFilters selects whose tasks trigger a notification.
type TaskNotificationFilters struct {
	CategoryIDs []string `json:"category_ids"`
	UpForGrabs  bool     `json:"up_for_grabs"`
}

// TaskNotification is one notification rule (task due, or task completed).
type TaskNotification struct {
	Enabled bool                    `json:"enabled"`
	Filters TaskNotificationFilters `json:"filters"`
}

// MarshalJSON omits filters from a disabled rule: the API rejects filters on a
// disabled rule (422 "cannot be set when disabled") and clears them anyway.
func (n TaskNotification) MarshalJSON() ([]byte, error) {
	if !n.Enabled {
		return []byte(`{"enabled":false}`), nil
	}
	type plain TaskNotification
	return json.Marshal(plain(n))
}

// TaskNotificationSettings holds a frame's task notification settings. The
// settings are frame-wide: there is no per-task reminder field. A task due
// reminder only fires for tasks that have a due time (see ChoreData.StartTime).
type TaskNotificationSettings struct {
	TaskDue       TaskNotification `json:"task_due"`
	TaskCompleted TaskNotification `json:"task_completed"`
}

// ErrNoNotificationRecipients is returned when a rule would be enabled without
// any category IDs and without up-for-grabs tasks.
var ErrNoNotificationRecipients = errors.New("notification rule has no recipients")

// ErrFiltersOnDisabledRule is returned when filters are changed on a rule that
// stays disabled; the API only stores filters on enabled rules.
var ErrFiltersOnDisabledRule = errors.New("filters can only be set on an enabled notification rule")

// taskNotificationSettingsResponse is the API envelope: settings arrive under "data".
type taskNotificationSettingsResponse struct {
	Data TaskNotificationSettings `json:"data"`
}

// TaskNotificationChange describes a partial change to one notification rule.
// Nil fields are left as they are.
type TaskNotificationChange struct {
	Enabled     *bool
	CategoryIDs *[]string
	UpForGrabs  *bool
}

func (ch *TaskNotificationChange) apply(n TaskNotification) TaskNotification {
	if ch.Enabled != nil {
		n.Enabled = *ch.Enabled
	}
	if ch.CategoryIDs != nil {
		n.Filters.CategoryIDs = *ch.CategoryIDs
	}
	if ch.UpForGrabs != nil {
		n.Filters.UpForGrabs = *ch.UpForGrabs
	}
	if n.Filters.CategoryIDs == nil {
		n.Filters.CategoryIDs = []string{}
	}
	return n
}

// resolve applies the change to a rule and rejects combinations the API would
// reject or silently discard.
func (ch *TaskNotificationChange) resolve(key string, current TaskNotification) (TaskNotification, error) {
	n := ch.apply(current)
	if !n.Enabled && (ch.CategoryIDs != nil || ch.UpForGrabs != nil) {
		return n, fmt.Errorf("%w: turn %s on in the same command", ErrFiltersOnDisabledRule, key)
	}
	// The API silently stores an enabled rule with no recipients as
	// disabled, so reject it rather than report a change that didn't stick.
	if n.Enabled && len(n.Filters.CategoryIDs) == 0 && !n.Filters.UpForGrabs {
		return n, fmt.Errorf("%w: enabling %s needs at least one category ID or up-for-grabs", ErrNoNotificationRecipients, key)
	}
	return n, nil
}

// TaskNotificationUpdate is a partial update to the notification settings.
// Nil rules are left unchanged.
type TaskNotificationUpdate struct {
	TaskDue       *TaskNotificationChange
	TaskCompleted *TaskNotificationChange
}

func (c *Client) taskNotificationSettingsURL(frameID string) string {
	return fmt.Sprintf("%s/frames/%s/task_notification_settings", c.effectiveURL(), pathSeg(frameID))
}

// GetTaskNotificationSettings retrieves the frame's task notification settings.
func (c *Client) GetTaskNotificationSettings(ctx context.Context, frameID string) (*TaskNotificationSettings, error) {
	req, err := newRequest(ctx, "GET", c.taskNotificationSettingsURL(frameID))
	if err != nil {
		return nil, fmt.Errorf("failed to create get task notification settings request: %w", err)
	}

	var resp taskNotificationSettingsResponse
	if err := c.get(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to get task notification settings: %w", err)
	}
	return &resp.Data, nil
}

// UpdateTaskNotificationSettings applies a partial update. The current settings
// are fetched first so a changed rule is sent whole (an enabled rule keeps its
// unchanged filters; a disabled rule is sent without any); rules without a
// change are not sent at all. It returns the
// settings as stored by the API after the update.
func (c *Client) UpdateTaskNotificationSettings(ctx context.Context, frameID string, update TaskNotificationUpdate) (*TaskNotificationSettings, error) {
	if update.TaskDue == nil && update.TaskCompleted == nil {
		return nil, fmt.Errorf("no task notification settings to update")
	}

	current, err := c.GetTaskNotificationSettings(ctx, frameID)
	if err != nil {
		return nil, err
	}

	body := map[string]TaskNotification{}
	for _, r := range []struct {
		key     string
		change  *TaskNotificationChange
		current TaskNotification
	}{
		{"task_due", update.TaskDue, current.TaskDue},
		{"task_completed", update.TaskCompleted, current.TaskCompleted},
	} {
		if r.change == nil {
			continue
		}
		n, err := r.change.resolve(r.key, r.current)
		if err != nil {
			return nil, err
		}
		body[r.key] = n
	}

	req, err := newRequestWithBody(ctx, "PATCH", c.taskNotificationSettingsURL(frameID), body)
	if err != nil {
		return nil, fmt.Errorf("failed to create update task notification settings request: %w", err)
	}

	var resp taskNotificationSettingsResponse
	if err := c.patch(req, &resp); err != nil {
		return nil, fmt.Errorf("failed to update task notification settings: %w", err)
	}
	// The PATCH response only echoes the rules that were sent, so merge them
	// into the settings fetched above to return the full picture.
	if update.TaskDue != nil {
		current.TaskDue = resp.Data.TaskDue
	}
	if update.TaskCompleted != nil {
		current.TaskCompleted = resp.Data.TaskCompleted
	}
	return current, nil
}
