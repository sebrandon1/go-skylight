package lib

import "fmt"

// ListMealCategories retrieves meal categories for a frame.
func (c *Client) ListMealCategories(frameID string) ([]MealCategory, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/meals/categories", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list meal categories request: %w", err)
	}

	var categories []MealCategory
	if err := c.get(req, &categories); err != nil {
		return nil, fmt.Errorf("failed to list meal categories: %w", err)
	}

	return categories, nil
}

// ListRecipes retrieves recipes for a frame.
func (c *Client) ListRecipes(frameID string) ([]Recipe, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/meals/recipes", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list recipes request: %w", err)
	}

	var recipes []Recipe
	if err := c.get(req, &recipes); err != nil {
		return nil, fmt.Errorf("failed to list recipes: %w", err)
	}

	return recipes, nil
}

// GetRecipe retrieves a single recipe by ID.
func (c *Client) GetRecipe(frameID, recipeID string) (*Recipe, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/meals/recipes/%s", SkylightURL, frameID, recipeID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create get recipe request: %w", err)
	}

	var recipe Recipe
	if err := c.get(req, &recipe); err != nil {
		return nil, fmt.Errorf("failed to get recipe: %w", err)
	}

	return &recipe, nil
}

// CreateRecipe creates a new recipe.
func (c *Client) CreateRecipe(frameID string, recipe RecipeData) (*Recipe, error) {
	reqBody := RecipeRequest{Recipe: recipe}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/meals/recipes", SkylightURL, frameID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create recipe request: %w", err)
	}

	var created Recipe
	if err := c.post(req, &created); err != nil {
		return nil, fmt.Errorf("failed to create recipe: %w", err)
	}

	return &created, nil
}

// UpdateRecipe updates an existing recipe.
func (c *Client) UpdateRecipe(frameID, recipeID string, recipe RecipeData) (*Recipe, error) {
	reqBody := RecipeRequest{Recipe: recipe}

	req, err := newRequestWithBody("PATCH", fmt.Sprintf("%s/frames/%s/meals/recipes/%s", SkylightURL, frameID, recipeID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create update recipe request: %w", err)
	}

	var updated Recipe
	if err := c.patch(req, &updated); err != nil {
		return nil, fmt.Errorf("failed to update recipe: %w", err)
	}

	return &updated, nil
}

// DeleteRecipe deletes a recipe.
func (c *Client) DeleteRecipe(frameID, recipeID string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/meals/recipes/%s", SkylightURL, frameID, recipeID), nil)
	if err != nil {
		return fmt.Errorf("failed to create delete recipe request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}

	return nil
}

// ListMealSittings retrieves meal sittings for a frame.
func (c *Client) ListMealSittings(frameID string) ([]MealSitting, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/meals/sittings", SkylightURL, frameID), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create list meal sittings request: %w", err)
	}

	var sittings []MealSitting
	if err := c.get(req, &sittings); err != nil {
		return nil, fmt.Errorf("failed to list meal sittings: %w", err)
	}

	return sittings, nil
}

// CreateMealSitting creates a new meal sitting.
func (c *Client) CreateMealSitting(frameID string, sitting MealSittingData) (*MealSitting, error) {
	reqBody := MealSittingRequest{MealSitting: sitting}

	req, err := newRequestWithBody("POST", fmt.Sprintf("%s/frames/%s/meals/sittings", SkylightURL, frameID), reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create meal sitting request: %w", err)
	}

	var created MealSitting
	if err := c.post(req, &created); err != nil {
		return nil, fmt.Errorf("failed to create meal sitting: %w", err)
	}

	return &created, nil
}

// AddRecipeToGroceryList adds a recipe's ingredients to the grocery list.
func (c *Client) AddRecipeToGroceryList(frameID, recipeID string) error {
	req, err := newRequest("POST", fmt.Sprintf("%s/frames/%s/meals/recipes/%s/add_to_grocery_list", SkylightURL, frameID, recipeID), nil)
	if err != nil {
		return fmt.Errorf("failed to create add to grocery list request: %w", err)
	}

	if err := c.post(req, nil); err != nil {
		return fmt.Errorf("failed to add recipe to grocery list: %w", err)
	}

	return nil
}
