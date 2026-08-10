package lib

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const (
	ChoreStatusComplete = "complete"
	ChoreStatusPending  = "pending"
	ChoreStatusSkipped  = "skipped"

	choreStatusPending = ChoreStatusPending
	paramTrue          = "true"
	applyToAll         = "all"
)

// choreIDRe matches composite chore IDs like "18731133-2026-04-28" or "70190003-2026-04-28-0600".
var choreIDRe = regexp.MustCompile(`^(\d+)-(\d{4}-\d{2}-\d{2})`)

// parseChoreID splits a composite chore ID into its base numeric ID and instance date.
// For non-recurring chores with plain numeric IDs, instanceDate is empty.
func parseChoreID(choreID string) (baseID, instanceDate string) {
	if m := choreIDRe.FindStringSubmatch(choreID); m != nil {
		return m[1], m[2]
	}
	return choreID, ""
}

// setCompletion calls the chore completions endpoint with the given data.
// It fills in InstanceDate from the choreID before sending.
func (c *Client) setCompletion(ctx context.Context, frameID, choreID string, data ChoreCompletionData) error {
	baseID, instanceDate := parseChoreID(choreID)
	data.InstanceDate = instanceDate
	req, err := newRequestWithBody(ctx, "PUT",
		fmt.Sprintf("%s/frames/%s/chores/%s/completions", c.effectiveURL(), pathSeg(frameID), pathSeg(baseID)), data)
	if err != nil {
		return fmt.Errorf("failed to create completion request: %w", err)
	}
	var result choreAPISingleResponse
	return c.put(req, &result)
}

func (opts ChoreListOptions) queryParams() map[string]string {
	params := map[string]string{}
	if opts.Date != "" {
		params["date"] = opts.Date
	}
	if opts.Status != "" {
		params["status"] = opts.Status
	}
	if opts.AssigneeID != "" {
		params["assignee_id"] = opts.AssigneeID
	}
	if opts.After != "" {
		params["after"] = opts.After
	}
	if opts.Before != "" {
		params["before"] = opts.Before
	}
	if opts.IncludeLate {
		params["include_late"] = paramTrue
	}
	if opts.UpForGrabs {
		params["include_up_for_grabs"] = paramTrue
		params["filter"] = "linked_to_profile"
		if opts.After == "" {
			params["after"] = time.Now().Format("2006-01-02")
		}
		if opts.Before == "" {
			params["before"] = time.Now().AddDate(0, 0, 7).Format("2006-01-02")
		}
	}
	if opts.Search != "" {
		params["q"] = opts.Search
	}
	return params
}

// ListChores retrieves chores for a frame with optional filters.
func (c *Client) ListChores(ctx context.Context, frameID string, opts ChoreListOptions) ([]Chore, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/chores", c.effectiveURL(), pathSeg(frameID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create list chores request: %w", err)
	}

	if params := opts.queryParams(); len(params) > 0 {
		addQueryParams(req, params)
	}

	var apiResp choreAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list chores: %w", err)
	}

	searchLower := strings.ToLower(opts.Search)
	chores := make([]Chore, 0, len(apiResp.Data))
	for i := range apiResp.Data {
		c := apiResp.Data[i].toChore()
		if opts.Status != "" && c.Status != opts.Status {
			continue
		}
		if opts.UpForGrabs && !c.UpForGrabs {
			continue
		}
		if searchLower != "" && !strings.Contains(strings.ToLower(c.Title), searchLower) && !strings.Contains(strings.ToLower(c.Description), searchLower) {
			continue
		}
		chores = append(chores, c)
	}

	return chores, nil
}

// CreateChore creates a new chore on a frame.
func (c *Client) CreateChore(ctx context.Context, frameID string, chore ChoreData) (*Chore, error) {
	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/chores", c.effectiveURL(), pathSeg(frameID)), chore)
	if err != nil {
		return nil, fmt.Errorf("failed to create chore request: %w", err)
	}

	var apiResp choreAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create chore: %w", err)
	}

	result := apiResp.Data.toChore()
	return &result, nil
}

// CreateUpForGrabsChore creates a shared chore that anyone can claim, using the
// create_multiple endpoint which accepts up_for_grabs without a category_id.
func (c *Client) CreateUpForGrabsChore(ctx context.Context, frameID string, chore ChoreData) (*Chore, error) {
	chore.UpForGrabs = true
	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/chores/create_multiple", c.effectiveURL(), pathSeg(frameID)), chore)
	if err != nil {
		return nil, fmt.Errorf("failed to create up-for-grabs chore request: %w", err)
	}

	var apiResp choreAPIResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create up-for-grabs chore: %w", err)
	}
	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("no chore returned from create_multiple")
	}

	result := apiResp.Data[0].toChore()
	return &result, nil
}

// UpdateChore updates an existing chore.
func (c *Client) UpdateChore(ctx context.Context, frameID, choreID string, chore ChoreData) (*Chore, error) {
	req, err := newRequestWithBody(ctx, "PUT", fmt.Sprintf("%s/frames/%s/chores/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(choreID)), chore)
	if err != nil {
		return nil, fmt.Errorf("failed to create update chore request: %w", err)
	}

	var apiResp choreAPISingleResponse
	if err := c.put(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update chore: %w", err)
	}

	result := apiResp.Data.toChore()
	return &result, nil
}

// SkipChore skips a single instance of a recurring chore. Pass deferUntil as
// a YYYY-MM-DD date to reschedule the instance, or "" to skip without deferring.
func (c *Client) SkipChore(ctx context.Context, frameID, choreID, deferUntil string) error {
	return c.setCompletion(ctx, frameID, choreID, ChoreCompletionData{Status: ChoreStatusSkipped, DeferUntil: deferUntil})
}

// CompleteChore marks a chore instance as completed via the completions endpoint.
func (c *Client) CompleteChore(ctx context.Context, frameID, choreID string) error {
	return c.setCompletion(ctx, frameID, choreID, ChoreCompletionData{Status: ChoreStatusComplete})
}

// ClaimChore assigns an up-for-grabs chore to the given assignee.
func (c *Client) ClaimChore(ctx context.Context, frameID, choreID, assigneeID string) (*Chore, error) {
	return c.UpdateChore(ctx, frameID, choreID, ChoreData{AssigneeID: assigneeID})
}

// DeleteChore deletes a chore. apply_to=all is sent unconditionally because
// the Skylight API requires it for recurring chores (400 otherwise) and ignores
// it for non-recurring ones.
func (c *Client) DeleteChore(ctx context.Context, frameID, choreID string) error {
	req, err := newRequest(ctx, "DELETE", fmt.Sprintf("%s/frames/%s/chores/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(choreID)))
	if err != nil {
		return fmt.Errorf("failed to create delete chore request: %w", err)
	}

	addQueryParams(req, map[string]string{"apply_to": applyToAll})

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete chore: %w", err)
	}

	return nil
}
