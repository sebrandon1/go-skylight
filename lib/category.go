package lib

import "fmt"

// ListCategories retrieves categories (family members) for a frame.
func (c *Client) ListCategories(frameID string) ([]Category, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/categories", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list categories request: %w", err)
	}

	var categories []Category
	if err := c.get(req, &categories); err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	return categories, nil
}
