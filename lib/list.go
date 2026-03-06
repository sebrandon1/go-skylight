package lib

import "fmt"

// ListLists retrieves all lists for a frame.
func (c *Client) ListLists(frameID string) ([]List, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/lists", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list lists request: %w", err)
	}

	var lists []List
	if err := c.get(req, &lists); err != nil {
		return nil, fmt.Errorf("failed to list lists: %w", err)
	}

	return lists, nil
}

// GetList retrieves a single list by ID.
func (c *Client) GetList(frameID, listID string) (*List, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/lists/%s", SkylightURL, frameID, listID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get list request: %w", err)
	}

	var list List
	if err := c.get(req, &list); err != nil {
		return nil, fmt.Errorf("failed to get list: %w", err)
	}

	return &list, nil
}

// CreateList creates a new list on a frame.
func (c *Client) CreateList(frameID string, list ListData) (*List, error) {
	reqBody := ListRequest{List: list}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/lists", SkylightURL, frameID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create list request: %w", err)
	}

	var created List
	if err := c.post(req, &created); err != nil {
		return nil, fmt.Errorf("failed to create list: %w", err)
	}

	return &created, nil
}

// UpdateList updates an existing list.
func (c *Client) UpdateList(frameID, listID string, list ListData) (*List, error) {
	reqBody := ListRequest{List: list}

	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/lists/%s", SkylightURL, frameID, listID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create update list request: %w", err)
	}

	var updated List
	if err := c.put(req, &updated); err != nil {
		return nil, fmt.Errorf("failed to update list: %w", err)
	}

	return &updated, nil
}

// DeleteList deletes a list.
func (c *Client) DeleteList(frameID, listID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/lists/%s", SkylightURL, frameID, listID), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete list request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete list: %w", err)
	}

	return nil
}

// AddListItem adds an item to a list.
func (c *Client) AddListItem(frameID, listID string, item ListItemData) (*ListItem, error) {
	reqBody := ListItemRequest{ListItem: item}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/lists/%s/list_items", SkylightURL, frameID, listID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create add list item request: %w", err)
	}

	var created ListItem
	if err := c.post(req, &created); err != nil {
		return nil, fmt.Errorf("failed to add list item: %w", err)
	}

	return &created, nil
}

// UpdateListItem updates an item in a list.
func (c *Client) UpdateListItem(frameID, listID, itemID string, item ListItemData) (*ListItem, error) {
	reqBody := ListItemRequest{ListItem: item}

	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/lists/%s/list_items/%s", SkylightURL, frameID, listID, itemID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create update list item request: %w", err)
	}

	var updated ListItem
	if err := c.put(req, &updated); err != nil {
		return nil, fmt.Errorf("failed to update list item: %w", err)
	}

	return &updated, nil
}

// DeleteListItem deletes an item from a list.
func (c *Client) DeleteListItem(frameID, listID, itemID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/lists/%s/list_items/%s", SkylightURL, frameID, listID, itemID), nil)
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

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/task_box_items", SkylightURL, frameID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create task box item request: %w", err)
	}

	var created TaskBoxItem
	if err := c.post(req, &created); err != nil {
		return nil, fmt.Errorf("failed to create task box item: %w", err)
	}

	return &created, nil
}
