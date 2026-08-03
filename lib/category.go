package lib

import (
	"context"
	"fmt"
)

// ListCategories retrieves categories (family members) for a frame.
func (c *Client) ListCategories(ctx context.Context, frameID string) ([]Category, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/categories", c.effectiveURL(), pathSeg(frameID)))
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

// CreateCategory creates a new category (profile/label) on a frame.
func (c *Client) CreateCategory(ctx context.Context, frameID string, data CategoryData) (*Category, error) {
	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/categories", c.effectiveURL(), pathSeg(frameID)), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create category request: %w", err)
	}

	var apiResp categoryAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	result := apiResp.Data.toCategory()
	return &result, nil
}

// UpdateCategory updates an existing category (profile/label).
func (c *Client) UpdateCategory(ctx context.Context, frameID, categoryID string, data CategoryData) (*Category, error) {
	req, err := newRequestWithBody(ctx, "PATCH", fmt.Sprintf("%s/frames/%s/categories/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(categoryID)), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create update category request: %w", err)
	}

	var apiResp categoryAPISingleResponse
	if err := c.patch(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	result := apiResp.Data.toCategory()
	return &result, nil
}

// DeleteCategory deletes a category (profile/label).
func (c *Client) DeleteCategory(ctx context.Context, frameID, categoryID string) error {
	req, err := newRequest(ctx, "DELETE", fmt.Sprintf("%s/frames/%s/categories/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(categoryID)))
	if err != nil {
		return fmt.Errorf("failed to create delete category request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	return nil
}
