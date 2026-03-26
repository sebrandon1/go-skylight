package lib

import "fmt"

// ListMealCategories retrieves meal categories for a frame.
func (c *Client) ListMealCategories(frameID string) ([]MealCategory, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/meals/categories", c.effectiveURL(), frameID))
	if err != nil {
		return nil, fmt.Errorf("failed to create list meal categories request: %w", err)
	}

	var apiResp mealCategoryAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list meal categories: %w", err)
	}

	categories := make([]MealCategory, len(apiResp.Data))
	for i := range apiResp.Data {
		categories[i] = apiResp.Data[i].toMealCategory()
	}
	return categories, nil
}

// ListRecipes retrieves recipes for a frame.
func (c *Client) ListRecipes(frameID string) ([]Recipe, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/meals/recipes", c.effectiveURL(), frameID))
	if err != nil {
		return nil, fmt.Errorf("failed to create list recipes request: %w", err)
	}

	var apiResp recipeAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}

	recipes := make([]Recipe, len(apiResp.Data))
	for i := range apiResp.Data {
		recipes[i] = apiResp.Data[i].toRecipe()
	}
	return recipes, nil
}

// GetRecipe retrieves a single recipe by ID.
func (c *Client) GetRecipe(frameID, recipeID string) (*Recipe, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/meals/recipes/%s", c.effectiveURL(), frameID, recipeID))
	if err != nil {
		return nil, fmt.Errorf("failed to create get recipe request: %w", err)
	}

	var apiResp recipeAPISingleResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}

	result := apiResp.Data.toRecipe()
	return &result, nil
}

// CreateRecipe creates a new recipe.
// The API expects a flat JSON body (no recipe wrapper).
func (c *Client) CreateRecipe(frameID string, recipe RecipeData) (*Recipe, error) {
	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/meals/recipes", c.effectiveURL(), frameID), recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe request: %w", err)
	}

	var apiResp recipeAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create recipe: %w", err)
	}

	result := apiResp.Data.toRecipe()
	return &result, nil
}

// UpdateRecipe updates an existing recipe.
// The API expects a flat JSON body (no recipe wrapper).
func (c *Client) UpdateRecipe(frameID, recipeID string, recipe RecipeData) (*Recipe, error) {
	req, err := newRequestWithBody("PATCH", fmt.Sprintf("%s/frames/%s/meals/recipes/%s", c.effectiveURL(), frameID, recipeID), recipe)
	if err != nil {
		return nil, fmt.Errorf("failed to create update recipe request: %w", err)
	}

	var apiResp recipeAPISingleResponse
	if err := c.patch(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}

	result := apiResp.Data.toRecipe()
	return &result, nil
}

// DeleteRecipe deletes a recipe.
func (c *Client) DeleteRecipe(frameID, recipeID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/meals/recipes/%s", c.effectiveURL(), frameID, recipeID))
	if err != nil {
		return fmt.Errorf("failed to create delete recipe request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}

	return nil
}

// ListMealSittings retrieves meal sittings for a frame within an optional date range.
func (c *Client) ListMealSittings(frameID string, opts MealSittingListOptions) ([]MealSitting, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/meals/sittings", c.effectiveURL(), frameID))
	if err != nil {
		return nil, fmt.Errorf("failed to create list meal sittings request: %w", err)
	}

	params := map[string]string{}
	if opts.DateMin != "" {
		params["date_min"] = opts.DateMin
	}
	if opts.DateMax != "" {
		params["date_max"] = opts.DateMax
	}
	if len(params) > 0 {
		addQueryParams(req, params)
	}

	var apiResp mealSittingAPIResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to list meal sittings: %w", err)
	}

	sittings := make([]MealSitting, len(apiResp.Data))
	for i := range apiResp.Data {
		sittings[i] = apiResp.Data[i].toMealSitting()
	}
	return sittings, nil
}

// CreateMealSitting creates a new meal sitting.
// The API expects a flat JSON body (no meal_sitting wrapper).
func (c *Client) CreateMealSitting(frameID string, sitting MealSittingData) (*MealSitting, error) {
	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/meals/sittings", c.effectiveURL(), frameID), sitting)
	if err != nil {
		return nil, fmt.Errorf("failed to create meal sitting request: %w", err)
	}

	var apiResp mealSittingAPIResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create meal sitting: %w", err)
	}

	if len(apiResp.Data) == 0 {
		return nil, fmt.Errorf("failed to create meal sitting: empty response")
	}

	result := apiResp.Data[0].toMealSitting()
	return &result, nil
}

// DeleteMealSitting deletes a meal sitting by ID.
func (c *Client) DeleteMealSitting(frameID, sittingID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/meals/sittings/%s", c.effectiveURL(), frameID, sittingID))
	if err != nil {
		return fmt.Errorf("failed to create delete meal sitting request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete meal sitting: %w", err)
	}

	return nil
}

// AddRecipeToGroceryList adds a recipe's ingredients to the grocery list.
func (c *Client) AddRecipeToGroceryList(frameID, recipeID string) error {
	req, err := newRequest("POST", fmt.Sprintf("%s/frames/%s/meals/recipes/%s/add_to_grocery_list", c.effectiveURL(), frameID, recipeID))
	if err != nil {
		return fmt.Errorf("failed to create add to grocery list request: %w", err)
	}

	if err := c.post(req, nil); err != nil {
		return fmt.Errorf("failed to add recipe to grocery list: %w", err)
	}

	return nil
}
