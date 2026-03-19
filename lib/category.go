package lib

import "fmt"

// ListCategories retrieves categories (family members) for a frame.
func (c *Client) ListCategories(frameID string) ([]Category, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/categories", c.effectiveURL(), frameID))
	if err != nil {
		return nil, fmt.Errorf("failed to create list categories request: %w", err)
	}

	var apiResp categoryAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}

	categories := make([]Category, len(apiResp.Data))
	for i := range apiResp.Data {
		categories[i] = apiResp.Data[i].toCategory()
	}
	return categories, nil
}
