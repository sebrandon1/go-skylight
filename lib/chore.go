package lib

import "fmt"

// ListChores retrieves chores for a frame with optional filters.
func (c *Client) ListChores(frameID string, opts ChoreListOptions) ([]Chore, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/chores", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list chores request: %w", err)
	}

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
		params["include_late"] = "true"
	}
	if len(params) > 0 {
		addQueryParams(req, params)
	}

	var chores []Chore
	if err := c.get(req, &chores); err != nil {
		return nil, fmt.Errorf("failed to list chores: %w", err)
	}

	return chores, nil
}

// CreateChore creates a new chore on a frame.
func (c *Client) CreateChore(frameID string, chore ChoreData) (*Chore, error) {
	reqBody := ChoreRequest{Chore: chore}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/chores", SkylightURL, frameID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create chore request: %w", err)
	}

	var created Chore
	if err := c.post(req, &created); err != nil {
		return nil, fmt.Errorf("failed to create chore: %w", err)
	}

	return &created, nil
}

// UpdateChore updates an existing chore.
func (c *Client) UpdateChore(frameID, choreID string, chore ChoreData) (*Chore, error) {
	reqBody := ChoreRequest{Chore: chore}

	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/chores/%s", SkylightURL, frameID, choreID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create update chore request: %w", err)
	}

	var updated Chore
	if err := c.put(req, &updated); err != nil {
		return nil, fmt.Errorf("failed to update chore: %w", err)
	}

	return &updated, nil
}

// DeleteChore deletes a chore.
func (c *Client) DeleteChore(frameID, choreID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/chores/%s", SkylightURL, frameID, choreID), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete chore request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete chore: %w", err)
	}

	return nil
}
