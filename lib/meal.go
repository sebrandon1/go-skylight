package lib

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ListMealCategories retrieves meal categories for a frame.
func (c *Client) ListMealCategories(ctx context.Context, frameID string) ([]MealCategory, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/meals/categories", c.effectiveURL(), pathSeg(frameID)))
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

// CreateMealCategory creates a new meal category.
func (c *Client) CreateMealCategory(ctx context.Context, frameID string, data MealCategoryData) (*MealCategory, error) {
	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/meals/categories", c.effectiveURL(), pathSeg(frameID)), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create meal category request: %w", err)
	}

	var apiResp mealCategoryAPISingleResponse
	if err := c.post(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to create meal category: %w", err)
	}

	result := apiResp.Data.toMealCategory()
	return &result, nil
}

// UpdateMealCategory updates an existing meal category.
func (c *Client) UpdateMealCategory(ctx context.Context, frameID, categoryID string, data MealCategoryData) (*MealCategory, error) {
	req, err := newRequestWithBody(ctx, "PATCH", fmt.Sprintf("%s/frames/%s/meals/categories/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(categoryID)), data)
	if err != nil {
		return nil, fmt.Errorf("failed to create update meal category request: %w", err)
	}

	var apiResp mealCategoryAPISingleResponse
	if err := c.patch(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to update meal category: %w", err)
	}

	result := apiResp.Data.toMealCategory()
	return &result, nil
}

// DeleteMealCategory deletes a meal category.
func (c *Client) DeleteMealCategory(ctx context.Context, frameID, categoryID string) error {
	req, err := newRequest(ctx, "DELETE", fmt.Sprintf("%s/frames/%s/meals/categories/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(categoryID)))
	if err != nil {
		return fmt.Errorf("failed to create delete meal category request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete meal category: %w", err)
	}

	return nil
}

// ListRecipes retrieves recipes for a frame.
func (c *Client) ListRecipes(ctx context.Context, frameID string) ([]Recipe, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/meals/recipes", c.effectiveURL(), pathSeg(frameID)))
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
func (c *Client) GetRecipe(ctx context.Context, frameID, recipeID string) (*Recipe, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/meals/recipes/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(recipeID)))
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
func (c *Client) CreateRecipe(ctx context.Context, frameID string, recipe RecipeData) (*Recipe, error) {
	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/meals/recipes", c.effectiveURL(), pathSeg(frameID)), recipe)
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
func (c *Client) UpdateRecipe(ctx context.Context, frameID, recipeID string, recipe RecipeData) (*Recipe, error) {
	req, err := newRequestWithBody(ctx, "PATCH", fmt.Sprintf("%s/frames/%s/meals/recipes/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(recipeID)), recipe)
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
func (c *Client) DeleteRecipe(ctx context.Context, frameID, recipeID string) error {
	req, err := newRequest(ctx, "DELETE", fmt.Sprintf("%s/frames/%s/meals/recipes/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(recipeID)))
	if err != nil {
		return fmt.Errorf("failed to create delete recipe request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete recipe: %w", err)
	}

	return nil
}

// ListMealSittings retrieves meal sittings for a frame within an optional date range.
func (c *Client) ListMealSittings(ctx context.Context, frameID string, opts MealSittingListOptions) ([]MealSitting, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/meals/sittings", c.effectiveURL(), pathSeg(frameID)))
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
func (c *Client) CreateMealSitting(ctx context.Context, frameID string, sitting MealSittingData) (*MealSitting, error) {
	req, err := newRequestWithBody(ctx, "POST", fmt.Sprintf("%s/frames/%s/meals/sittings", c.effectiveURL(), pathSeg(frameID)), sitting)
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

// DeleteMealSitting deletes a specific instance of a meal sitting by sitting ID and date (YYYY-MM-DD).
func (c *Client) DeleteMealSitting(ctx context.Context, frameID, sittingID, date string) error {
	req, err := newRequest(ctx, "DELETE", fmt.Sprintf("%s/frames/%s/meals/sittings/%s/instances/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(sittingID), pathSeg(date)))
	if err != nil {
		return fmt.Errorf("failed to create delete meal sitting request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete meal sitting: %w", err)
	}

	return nil
}

// GetMealSitting retrieves a single meal sitting by ID.
func (c *Client) GetMealSitting(ctx context.Context, frameID, sittingID string) (*MealSitting, error) {
	req, err := newRequest(ctx, "GET", fmt.Sprintf("%s/frames/%s/meals/sittings/%s", c.effectiveURL(), pathSeg(frameID), pathSeg(sittingID)))
	if err != nil {
		return nil, fmt.Errorf("failed to create get meal sitting request: %w", err)
	}

	var apiResp mealSittingAPISingleResponse
	if err := c.get(req, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to get meal sitting: %w", err)
	}

	result := apiResp.Data.toMealSitting()
	return &result, nil
}

// GetSittingRecipe fetches a meal sitting and its linked recipe in one call.
// Returns a result with a nil Recipe field if the sitting has no linked recipe.
func (c *Client) GetSittingRecipe(ctx context.Context, frameID, sittingID string) (*SittingWithRecipe, error) {
	sitting, err := c.GetMealSitting(ctx, frameID, sittingID)
	if err != nil {
		return nil, err
	}

	result := &SittingWithRecipe{Sitting: *sitting}

	if sitting.RecipeID == "" {
		return result, nil
	}

	recipe, err := c.GetRecipe(ctx, frameID, sitting.RecipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe for sitting: %w", err)
	}
	result.Recipe = recipe

	return result, nil
}

// PlanMeals schedules a list of recipes as meal sittings starting from a given date,
// rotating through the provided meal categories across consecutive days.
func (c *Client) PlanMeals(ctx context.Context, frameID string, data MealPlanData) (*MealPlanResult, error) {
	if len(data.RecipeIDs) == 0 {
		return nil, errors.New("at least one recipe is required")
	}
	if len(data.CategoryIDs) == 0 {
		return nil, errors.New("at least one category is required")
	}

	startDate, err := time.Parse(DateFormat, data.StartDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start_date format (expected YYYY-MM-DD): %w", err)
	}

	numCategories := len(data.CategoryIDs)
	var created []MealSitting

	for i, recipeID := range data.RecipeIDs {
		dayOffset, catIdx := i/numCategories, i%numCategories
		categoryID := data.CategoryIDs[catIdx]
		date := startDate.AddDate(0, 0, dayOffset).Format(DateFormat)

		sitting, err := c.CreateMealSitting(ctx, frameID, MealSittingData{
			RecipeID:       recipeID,
			Date:           date,
			MealCategoryID: categoryID,
		})
		if err != nil {
			return &MealPlanResult{Sittings: created}, fmt.Errorf("failed scheduling recipe %q on %s: %w", recipeID, date, err)
		}

		created = append(created, *sitting)
	}

	return &MealPlanResult{Sittings: created}, nil
}

// AddRecipeToGroceryList adds a recipe's ingredients to the grocery list.
func (c *Client) AddRecipeToGroceryList(ctx context.Context, frameID, recipeID string) error {
	req, err := newRequest(ctx, "POST", fmt.Sprintf("%s/frames/%s/meals/recipes/%s/add_to_grocery_list", c.effectiveURL(), pathSeg(frameID), pathSeg(recipeID)))
	if err != nil {
		return fmt.Errorf("failed to create add to grocery list request: %w", err)
	}

	if err := c.post(req, nil); err != nil {
		return fmt.Errorf("failed to add recipe to grocery list: %w", err)
	}

	return nil
}
