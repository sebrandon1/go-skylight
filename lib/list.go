package lib

import "fmt"

const listItemStatusCompleted = "completed"
const listItemStatusPending = "pending"

// ListLists retrieves all lists for a frame.
func (c *Client) ListLists(frameID string) ([]List, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/lists", c.effectiveURL(), frameID))
	if err != nil {
		return nil, fmt.Errorf("failed to create list lists request: %w", err)
	}

	var apiResp listAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list lists: %w", err)
	}

	lists := make([]List, len(apiResp.Data))
	for i := range apiResp.Data {
		lists[i] = apiResp.Data[i].toList()
	}
	return lists, nil
}

// GetList retrieves a single list by ID (includes items from the included array).
func (c *Client) GetList(frameID, listID string) (*List, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/lists/%s", c.effectiveURL(), frameID, listID))
	if err != nil {
		return nil, fmt.Errorf("failed to create get list request: %w", err)
	}

	var apiResp listAPISingleResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	list := apiResp.Data.toList()
	for _, entry := range apiResp.Included {
		if entry.Type == "list_item" {
			list.Items = append(list.Items, entry.toListItem())
		}
	}
	return &list, nil
}

// CreateList creates a new list on a frame.
// The API expects a flat JSON body (no wrapper) with label, kind, and color.
func (c *Client) CreateList(frameID string, list ListData) (*List, error) {
	if list.Kind == "" {
		list.Kind = "to_do"
	}
	if list.Color == "" {
		list.Color = "#2178AF"
	}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/lists", c.effectiveURL(), frameID), list)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}

	var apiResp listAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create list: %w", err)
	}

	result := apiResp.Data.toList()
	return &result, nil
}

// UpdateList updates an existing list.
// The API expects a flat JSON body (no {"list":{...}} wrapper).
func (c *Client) UpdateList(frameID, listID string, list ListData) (*List, error) {
	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/lists/%s", c.effectiveURL(), frameID, listID), list)
	if err != nil {
		return nil, fmt.Errorf("failed to create update list request: %w", err)
	}

	var apiResp listAPISingleResponse
	if err := c.put(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update list: %w", err)
	}

	result := apiResp.Data.toList()
	return &result, nil
}

// DeleteList deletes a list.
func (c *Client) DeleteList(frameID, listID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/lists/%s", c.effectiveURL(), frameID, listID))
	if err != nil {
		return fmt.Errorf("failed to create delete list request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete list: %w", err)
	}

	return nil
}

// AddListItem adds an item to a list.
// The API expects a flat JSON body with label (no list_item wrapper).
func (c *Client) AddListItem(frameID, listID string, item ListItemData) (*ListItem, error) {
	send := listItemSendData{
		Label:    item.Title,
		Position: item.Position,
	}
	if item.Completed {
		send.Status = listItemStatusCompleted
	}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/lists/%s/list_items", c.effectiveURL(), frameID, listID), send)
	if err != nil {
		return nil, fmt.Errorf("failed to create add list item request: %w", err)
	}

	var apiResp listItemAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to add list item: %w", err)
	}

	result := apiResp.Data.toListItem()
	return &result, nil
}

// UpdateListItem updates an item in a list.
// The API expects a flat JSON body (no wrapper).
func (c *Client) UpdateListItem(frameID, listID, itemID string, item ListItemData) (*ListItem, error) {
	send := listItemSendData{
		Label:    item.Title,
		Position: item.Position,
	}
	if item.Completed {
		send.Status = listItemStatusCompleted
	} else if item.Title == "" && item.Position == 0 {
		// Explicit incomplete when only status is being changed
		send.Status = listItemStatusPending
	}

	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/lists/%s/list_items/%s", c.effectiveURL(), frameID, listID, itemID), send)
	if err != nil {
		return nil, fmt.Errorf("failed to create update list item request: %w", err)
	}

	var apiResp listItemAPISingleResponse
	if err := c.put(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update list item: %w", err)
	}

	result := apiResp.Data.toListItem()
	return &result, nil
}

// DeleteListItem deletes an item from a list.
func (c *Client) DeleteListItem(frameID, listID, itemID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/lists/%s/list_items/%s", c.effectiveURL(), frameID, listID, itemID))
	if err != nil {
		return fmt.Errorf("failed to create delete list item request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete list item: %w", err)
	}

	return nil
}

// CreateTaskBoxItem creates a new task box item on a frame.
func (c *Client) CreateTaskBoxItem(frameID string, item TaskBoxItemData) (*TaskBoxItem, error) {
	reqBody := TaskBoxItemRequest{TaskBoxItem: item}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/task_box_items", c.effectiveURL(), frameID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create task box item request: %w", err)
	}

	var created TaskBoxItem
	if err := c.post(req, &created); err != nil {
		return nil, fmt.Errorf("failed to create task box item: %w", err)
	}

	return &created, nil
}
