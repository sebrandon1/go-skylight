package lib

import "fmt"

const listItemStatusCompleted = "completed"
const listItemStatusPending = "pending"

// ListKindGrocery is the list kind value for grocery lists.
const ListKindGrocery = "grocery"

// ListLists retrieves all lists for a frame.
func (c *Client) ListLists(frameID string) ([]List, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/lists", c.effectiveURL(), pathSeg(frameID)))
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
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/lists/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(listID)))
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

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/lists", c.effectiveURL(), pathSeg(frameID)), list)
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
	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/lists/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(listID)), list)
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
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/lists/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(listID)))
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

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/lists/%s/list_items", c.effectiveURL(), pathSeg(frameID), pathSeg(listID)), send)
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

	req, err := newRequestWithBody("PUT", fmt.Sprintf("%s/frames/%s/lists/%s/list_items/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(listID), pathSeg(itemID)), send)
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
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/lists/%s/list_items/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(listID), pathSeg(itemID)))
	if err != nil {
		return fmt.Errorf("failed to create delete list item request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete list item: %w", err)
	}

	return nil
}

// ClearCompletedListItems deletes all completed items from a list.
func (c *Client) ClearCompletedListItems(frameID, listID string) (int, error) {
	list, err := c.GetList(frameID, listID)
	if err != nil {
		return 0, fmt.Errorf("failed to get list: %w", err)
	}

	var deleted int
	for _, item := range list.Items {
		if item.Completed {
			if err := c.DeleteListItem(frameID, listID, item.ID); err != nil {
				return deleted, fmt.Errorf("failed to delete item %s: %w", item.ID, err)
			}
			deleted++
		}
	}
	return deleted, nil
}

// OrganizeGroceryList deduplicates and sorts a grocery list by aisle.
func (c *Client) OrganizeGroceryList(frameID, listID string) error {
	req, err := newRequest("POST", fmt.Sprintf("%s/frames/%s/lists/%s/organize", c.effectiveURL(), pathSeg(frameID), pathSeg(listID)))
	if err != nil {
		return fmt.Errorf("failed to create organize list request: %w", err)
	}

	var result any
	if err := c.post(req, &result); err != nil {
		return fmt.Errorf("failed to organize list: %w", err)
	}

	return nil
}

// OrderGroceryList sends the grocery list to an Instacart order.
// retailer may be empty (default) or a specific retailer slug (e.g. "costco").
func (c *Client) OrderGroceryList(frameID, listID, retailer string) (string, error) {
	body := struct {
		Retailer string `json:"retailer,omitempty"`
	}{Retailer: retailer}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/lists/%s/order", c.effectiveURL(), pathSeg(frameID), pathSeg(listID)), body)
	if err != nil {
		return "", fmt.Errorf("failed to create order list request: %w", err)
	}

	var result struct {
		RedirectURL string `json:"redirect_url"`
	}
	if err := c.post(req, &result); err != nil {
		return "", fmt.Errorf("failed to order list: %w", err)
	}

	return result.RedirectURL, nil
}

// CreateTaskBoxItem creates a new task box item on a frame.
func (c *Client) CreateTaskBoxItem(frameID string, item TaskBoxItemData) (*TaskBoxItem, error) {
	reqBody := TaskBoxItemRequest{TaskBoxItem: item}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/task_box_items", c.effectiveURL(), pathSeg(frameID)), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create task box item request: %w", err)
	}

	var created TaskBoxItem
	if err := c.post(req, &created); err != nil {
		return nil, fmt.Errorf("failed to create task box item: %w", err)
	}

	return &created, nil
}
