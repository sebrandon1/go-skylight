package lib

import (
	"errors"
	"fmt"
	"time"
)

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

// DeleteMealSitting deletes a specific instance of a meal sitting by sitting ID and date (YYYY-MM-DD).
func (c *Client) DeleteMealSitting(frameID, sittingID, date string) error {
	req, err := newRequest("DELETE", fmt.Sprintf("%s/frames/%s/meals/sittings/%s/instances/%s", c.effectiveURL(), frameID, sittingID, date))
	if err != nil {
		return fmt.Errorf("failed to create delete meal sitting request: %w", err)
	}

	if err := c.doDelete(req); err != nil {
		return fmt.Errorf("failed to delete meal sitting: %w", err)
	}

	return nil
}

// GetMealSitting retrieves a single meal sitting by ID.
func (c *Client) GetMealSitting(frameID, sittingID string) (*MealSitting, error) {
	req, err := newRequest("GET", fmt.Sprintf("%s/frames/%s/meals/sittings/%s", c.effectiveURL(), frameID, sittingID))
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
func (c *Client) GetSittingRecipe(frameID, sittingID string) (*SittingWithRecipe, error) {
	sitting, err := c.GetMealSitting(frameID, sittingID)
	if err != nil {
		return nil, err
	}

	result := &SittingWithRecipe{Sitting: *sitting}

	if sitting.RecipeID == "" {
		return result, nil
	}

	recipe, err := c.GetRecipe(frameID, sitting.RecipeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get recipe for sitting: %w", err)
	}
	result.Recipe = recipe

	return result, nil
}

// ListMealPlanSittings lists scheduled meal sittings in an optional date window.
//
// The Skylight API has no separate "meal plan" resource: PlanMeals creates ordinary
// meal sittings. Listing those sittings is how callers inspect existing plans (#256).
func (c *Client) ListMealPlanSittings(frameID string, opts MealSittingListOptions) ([]MealSitting, error) {
	return c.ListMealSittings(frameID, opts)
}

// DeleteMealPlanSitting removes one planned meal sitting instance (sitting ID + date).
// Same underlying endpoint as DeleteMealSitting; named for the plan CLI surface (#256).
func (c *Client) DeleteMealPlanSitting(frameID, sittingID, date string) error {
	return c.DeleteMealSitting(frameID, sittingID, date)
}

// DeleteMealPlanRange deletes every meal sitting instance whose date falls in [dateMin, dateMax].
// Both bounds are inclusive YYYY-MM-DD strings. Deletions always run; errors are joined.
func (c *Client) DeleteMealPlanRange(frameID, dateMin, dateMax string) (deleted int, err error) {
	if dateMin == "" || dateMax == "" {
		return 0, errors.New("date_min and date_max are required")
	}
	if _, e := time.Parse(DateFormat, dateMin); e != nil {
		return 0, fmt.Errorf("invalid date_min (expected YYYY-MM-DD): %w", e)
	}
	if _, e := time.Parse(DateFormat, dateMax); e != nil {
		return 0, fmt.Errorf("invalid date_max (expected YYYY-MM-DD): %w", e)
	}

	sittings, listErr := c.ListMealSittings(frameID, MealSittingListOptions{
		DateMin: dateMin,
		DateMax: dateMax,
	})
	if listErr != nil {
		return 0, fmt.Errorf("failed to list meal plan sittings: %w", listErr)
	}

	// List responses often omit instance dates; expand the requested window so
	// DELETE /sittings/{id}/instances/{date} can still be formed.
	dates := dateRangeInclusive(dateMin, dateMax)

	var errs []error
	for _, s := range sittings {
		tryDates := dates
		if s.Date != "" {
			tryDates = []string{s.Date}
		}
		removed := false
		var lastErr error
		for _, date := range tryDates {
			if delErr := c.DeleteMealSitting(frameID, s.ID, date); delErr != nil {
				lastErr = delErr
				continue
			}
			deleted++
			removed = true
			break
		}
		if !removed {
			errs = append(errs, fmt.Errorf("sitting %s: %w", s.ID, lastErr))
		}
	}
	return deleted, errors.Join(errs...)
}

// dateRangeInclusive returns every YYYY-MM-DD from start through end inclusive.
func dateRangeInclusive(start, end string) []string {
	s, err1 := time.Parse(DateFormat, start)
	e, err2 := time.Parse(DateFormat, end)
	if err1 != nil || err2 != nil || e.Before(s) {
		return []string{start}
	}
	var out []string
	for d := s; !d.After(e); d = d.AddDate(0, 0, 1) {
		out = append(out, d.Format(DateFormat))
		if len(out) > 366 {
			break
		}
	}
	return out
}

// PlanMeals schedules a list of recipes as meal sittings starting from a given date,
// rotating through the provided meal categories across consecutive days.
func (c *Client) PlanMeals(frameID string, data MealPlanData) (*MealPlanResult, error) {
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

		sitting, err := c.CreateMealSitting(frameID, MealSittingData{
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
