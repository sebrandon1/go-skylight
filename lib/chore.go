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

	var apiResp choreAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list chores: %w", err)
	}

	chores := make([]Chore, len(apiResp.Data))
	for i := range apiResp.Data {
		chores[i] = apiResp.Data[i].toChore()
	}

	return chores, nil
}

// CreateChore creates a new chore on a frame.
func (c *Client) CreateChore(frameID string, chore ChoreData) (*Chore, error) {
	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/chores", SkylightURL, frameID), chore)
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

// UpdateChore updates an existing chore.
func (c *Client) UpdateChore(frameID, choreID string, chore ChoreData) (*Chore, error) {
	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/chores/%s", SkylightURL, frameID, choreID), chore)
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
