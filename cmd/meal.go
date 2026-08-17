package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	recipeID          string
	recipeTitle       string
	recipeDescription string
	recipeIngredients []string
	recipeURL         string
	recipeCategoryID  string
	sittingID         string
	sittingDate       string
	sittingSummary    string
	sittingDateMin    string
	sittingDateMax    string
	mealCategoryID    string
	mealCategoryName  string
	mealCategoryColor string

	mealPlanRecipeIDs   []string
	mealPlanCategoryIDs []string
	mealPlanStartDate   string
)

var mealCmd = &cobra.Command{
	Use:   subMeal,
	Short: "Meal and recipe management commands",
	Long: `Manage recipes and scheduled meal sittings on a Skylight frame.

A recipe holds the title/ingredients; a "sitting" schedules that
recipe on a specific date and meal category (breakfast, dinner,
etc.). Use "meal plan" to schedule several recipes across a date
range in one call.

  # Create a recipe, then schedule it for a specific date
  skylight meal create-recipe --title "Tacos" --meal-category-id 12345678
  skylight meal create-sitting --recipe-id 87654321 --date 2026-06-05 --meal-category-id 12345678`,
}

var mealCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List meal categories",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		categories, err := client.ListMealCategories(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing meal categories: %w", err)
		}

		printOutput(categories)
		return nil
	},
}

var mealCreateCategoryCmd = &cobra.Command{
	Use:   "create-category",
	Short: "Create a meal category",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		category, err := client.CreateMealCategory(cmd.Context(), frameID, lib.MealCategoryData{
			Name:  mealCategoryName,
			Color: mealCategoryColor,
		})
		if err != nil {
			return fmt.Errorf("creating meal category: %w", err)
		}

		printOutput([]lib.MealCategory{*category})
		return nil
	},
}

var mealUpdateCategoryCmd = &cobra.Command{
	Use:   "update-category",
	Short: "Update a meal category",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.MealCategoryData{}
		if cmd.Flags().Changed("name") {
			data.Name = mealCategoryName
		}
		if cmd.Flags().Changed("color") {
			data.Color = mealCategoryColor
		}

		category, err := client.UpdateMealCategory(cmd.Context(), frameID, mealCategoryID, data)
		if err != nil {
			return fmt.Errorf("updating meal category: %w", err)
		}

		printOutput([]lib.MealCategory{*category})
		return nil
	},
}

var mealDeleteCategoryCmd = &cobra.Command{
	Use:   "delete-category",
	Short: "Delete a meal category",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete meal category %s", mealCategoryID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete meal category %s?", mealCategoryID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteMealCategory(cmd.Context(), frameID, mealCategoryID); err != nil {
			return fmt.Errorf("deleting meal category: %w", err)
		}

		printSuccess("Meal category deleted successfully")
		return nil
	},
}

var mealRecipesCmd = &cobra.Command{
	Use:   "recipes",
	Short: "List recipes",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		recipes, err := client.ListRecipes(cmd.Context(), frameID)
		if err != nil {
			return fmt.Errorf("listing recipes: %w", err)
		}

		printOutput(recipes)
		return nil
	},
}

var mealRecipeInfoCmd = &cobra.Command{
	Use:   "recipe-info",
	Short: "Get recipe details",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		recipe, err := client.GetRecipe(cmd.Context(), frameID, recipeID)
		if err != nil {
			return fmt.Errorf("getting recipe: %w", err)
		}

		printOutput([]lib.Recipe{*recipe})
		return nil
	},
}

var mealCreateRecipeCmd = &cobra.Command{
	Use:   "create-recipe",
	Short: "Create a recipe",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		recipe, err := client.CreateRecipe(cmd.Context(), frameID, lib.RecipeData{
			Title:          recipeTitle,
			Description:    recipeDescription,
			Ingredients:    recipeIngredients,
			URL:            recipeURL,
			MealCategoryID: recipeCategoryID,
		})
		if err != nil {
			return fmt.Errorf("creating recipe: %w", err)
		}

		printOutput([]lib.Recipe{*recipe})
		return nil
	},
}

var mealDeleteRecipeCmd = &cobra.Command{
	Use:   "delete-recipe",
	Short: "Delete a recipe",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete recipe %s", recipeID)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete recipe %s?", recipeID)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteRecipe(cmd.Context(), frameID, recipeID); err != nil {
			return fmt.Errorf("deleting recipe: %w", err)
		}

		printSuccess("Recipe deleted successfully")
		return nil
	},
}

var mealSittingsCmd = &cobra.Command{
	Use:   "sittings",
	Short: "List meal sittings",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		for _, f := range []struct {
			name string
			val  string
		}{{"date-min", sittingDateMin}, {"date-max", sittingDateMax}} {
			if cmd.Flags().Changed(f.name) {
				if err := validateDate(f.val); err != nil {
					return err
				}
			}
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		sittings, err := client.ListMealSittings(cmd.Context(), frameID, lib.MealSittingListOptions{
			DateMin: sittingDateMin,
			DateMax: sittingDateMax,
		})
		if err != nil {
			return fmt.Errorf("listing meal sittings: %w", err)
		}

		printOutput(sittings)
		return nil
	},
}

var mealCreateSittingCmd = &cobra.Command{
	Use:   "create-sitting",
	Short: "Create a meal sitting",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if err := validateDate(sittingDate); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		sitting, err := client.CreateMealSitting(cmd.Context(), frameID, lib.MealSittingData{
			Summary:        sittingSummary,
			RecipeID:       recipeID,
			Date:           sittingDate,
			MealCategoryID: mealCategoryID,
		})
		if err != nil {
			return fmt.Errorf("creating meal sitting: %w", err)
		}

		printOutput([]lib.MealSitting{*sitting})
		return nil
	},
}

var mealUpdateSittingCmd = &cobra.Command{
	Use:   "update-sitting",
	Short: "Update a meal sitting",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.MealSittingData{}
		if cmd.Flags().Changed("summary") {
			data.Summary = sittingSummary
		}
		if cmd.Flags().Changed(subDate) {
			if err := validateDate(sittingDate); err != nil {
				return err
			}
			data.Date = sittingDate
		}
		if cmd.Flags().Changed("recipe-id") {
			data.RecipeID = recipeID
		}
		if cmd.Flags().Changed("meal-category-id") {
			data.MealCategoryID = mealCategoryID
		}

		sitting, err := client.UpdateMealSitting(cmd.Context(), frameID, sittingID, data)
		if err != nil {
			return fmt.Errorf("updating meal sitting: %w", err)
		}

		printOutput([]lib.MealSitting{*sitting})
		return nil
	},
}

var mealDeleteSittingCmd = &cobra.Command{
	Use:   "delete-sitting",
	Short: "Delete a meal sitting instance",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if err := validateDate(sittingDate); err != nil {
			return err
		}

		if dryRun {
			printDryRun("delete meal sitting %s on %s", sittingID, sittingDate)
			return nil
		}

		if !confirmAction(fmt.Sprintf("Delete meal sitting %s on %s?", sittingID, sittingDate)) {
			return nil
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.DeleteMealSitting(cmd.Context(), frameID, sittingID, sittingDate); err != nil {
			return fmt.Errorf("deleting meal sitting: %w", err)
		}

		printSuccess("Meal sitting deleted successfully")
		return nil
	},
}

var mealAddToGroceryCmd = &cobra.Command{
	Use:   "add-to-grocery",
	Short: "Add recipe ingredients to grocery list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		if err := client.AddRecipeToGroceryList(cmd.Context(), frameID, recipeID); err != nil {
			return fmt.Errorf("adding to grocery list: %w", err)
		}

		printSuccess("Recipe added to grocery list successfully")
		return nil
	},
}

var mealGetSittingCmd = &cobra.Command{
	Use:   "get-sitting",
	Short: "Get meal sitting details",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		sitting, err := client.GetMealSitting(cmd.Context(), frameID, sittingID)
		if err != nil {
			return fmt.Errorf("getting meal sitting: %w", err)
		}

		printOutput([]lib.MealSitting{*sitting})
		return nil
	},
}

var mealSittingRecipeCmd = &cobra.Command{
	Use:   "sitting-recipe",
	Short: "View recipe details from a meal sitting",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		result, err := client.GetSittingRecipe(cmd.Context(), frameID, sittingID)
		if err != nil {
			return fmt.Errorf("getting sitting recipe: %w", err)
		}

		if result.Recipe == nil {
			fmt.Println("No recipe linked to this meal sitting")
			return nil
		}

		printJSON(result)
		return nil
	},
}

var mealPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Schedule a batch of meals from a recipe list",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		if err := validateDate(mealPlanStartDate); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		result, err := client.PlanMeals(cmd.Context(), frameID, lib.MealPlanData{
			RecipeIDs:   mealPlanRecipeIDs,
			CategoryIDs: mealPlanCategoryIDs,
			StartDate:   mealPlanStartDate,
		})
		if err != nil {
			if result != nil && len(result.Sittings) > 0 {
				fmt.Fprintf(os.Stderr, "Partial result (%d sittings created):\n", len(result.Sittings))
				printJSON(result)
			}
			return fmt.Errorf("planning meals: %w", err)
		}

		printJSON(result)
		return nil
	},
}

var mealUpdateRecipeCmd = &cobra.Command{
	Use:   "update-recipe",
	Short: "Update a recipe",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireFrameID(); err != nil {
			return err
		}

		client, err := getClient()
		if err != nil {
			return err
		}

		data := lib.RecipeData{}
		if cmd.Flags().Changed(subTitle) {
			data.Title = recipeTitle
		}
		if cmd.Flags().Changed("description") {
			data.Description = recipeDescription
		}
		if cmd.Flags().Changed("ingredients") {
			data.Ingredients = recipeIngredients
		}
		if cmd.Flags().Changed("url") {
			data.URL = recipeURL
		}

		recipe, err := client.UpdateRecipe(cmd.Context(), frameID, recipeID, data)
		if err != nil {
			return fmt.Errorf("updating recipe: %w", err)
		}

		printOutput([]lib.Recipe{*recipe})
		return nil
	},
}

func init() {
	mealCmd.AddCommand(mealCategoriesCmd)
	mealCmd.AddCommand(mealCreateCategoryCmd)
	mealCmd.AddCommand(mealUpdateCategoryCmd)
	mealCmd.AddCommand(mealDeleteCategoryCmd)
	mealCmd.AddCommand(mealRecipesCmd)
	mealCmd.AddCommand(mealRecipeInfoCmd)
	mealCmd.AddCommand(mealCreateRecipeCmd)
	mealCmd.AddCommand(mealUpdateRecipeCmd)
	mealCmd.AddCommand(mealDeleteRecipeCmd)
	mealCmd.AddCommand(mealSittingsCmd)
	mealCmd.AddCommand(mealCreateSittingCmd)
	mealCmd.AddCommand(mealUpdateSittingCmd)
	mealCmd.AddCommand(mealDeleteSittingCmd)
	mealCmd.AddCommand(mealGetSittingCmd)
	mealCmd.AddCommand(mealAddToGroceryCmd)
	mealCmd.AddCommand(mealSittingRecipeCmd)
	mealCmd.AddCommand(mealPlanCmd)

	mealCreateCategoryCmd.Flags().StringVar(&mealCategoryName, "name", "", "Category name")
	mealCreateCategoryCmd.Flags().StringVar(&mealCategoryColor, "color", "", "Category color")
	markFlagRequired(mealCreateCategoryCmd, "name")

	mealUpdateCategoryCmd.Flags().StringVar(&mealCategoryID, "category-id", "", "Meal category ID")
	mealUpdateCategoryCmd.Flags().StringVar(&mealCategoryName, "name", "", "Category name")
	mealUpdateCategoryCmd.Flags().StringVar(&mealCategoryColor, "color", "", "Category color")
	markFlagRequired(mealUpdateCategoryCmd, "category-id")

	mealDeleteCategoryCmd.Flags().StringVar(&mealCategoryID, "category-id", "", "Meal category ID")
	mealDeleteCategoryCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	mealDeleteCategoryCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(mealDeleteCategoryCmd, "category-id")

	mealRecipeInfoCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	markFlagRequired(mealRecipeInfoCmd, "recipe-id")

	mealCreateRecipeCmd.Flags().StringVar(&recipeTitle, subTitle, "", "Recipe title")
	mealCreateRecipeCmd.Flags().StringVar(&recipeDescription, "description", "", "Recipe description")
	mealCreateRecipeCmd.Flags().StringSliceVar(&recipeIngredients, "ingredients", nil, "Ingredients (comma-separated)")
	mealCreateRecipeCmd.Flags().StringVar(&recipeURL, "url", "", "Recipe URL")
	mealCreateRecipeCmd.Flags().StringVar(&recipeCategoryID, "meal-category-id", "", "Meal category ID")
	markFlagRequired(mealCreateRecipeCmd, subTitle)

	mealUpdateRecipeCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID to update")
	mealUpdateRecipeCmd.Flags().StringVar(&recipeTitle, subTitle, "", "Recipe title")
	mealUpdateRecipeCmd.Flags().StringVar(&recipeDescription, "description", "", "Recipe description")
	mealUpdateRecipeCmd.Flags().StringSliceVar(&recipeIngredients, "ingredients", nil, "Ingredients (comma-separated)")
	mealUpdateRecipeCmd.Flags().StringVar(&recipeURL, "url", "", "Recipe URL")
	markFlagRequired(mealUpdateRecipeCmd, "recipe-id")

	mealDeleteRecipeCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	mealDeleteRecipeCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	mealDeleteRecipeCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(mealDeleteRecipeCmd, "recipe-id")

	mealSittingsCmd.Flags().StringVar(&sittingDateMin, "date-min", "", "Minimum date filter (YYYY-MM-DD)")
	mealSittingsCmd.Flags().StringVar(&sittingDateMax, "date-max", "", "Maximum date filter (YYYY-MM-DD)")

	mealCreateSittingCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	mealCreateSittingCmd.Flags().StringVar(&sittingSummary, "summary", "", "Meal sitting summary/title")
	mealCreateSittingCmd.Flags().StringVar(&sittingDate, subDate, "", "Sitting date")
	mealCreateSittingCmd.Flags().StringVar(&mealCategoryID, "meal-category-id", "", "Meal category ID")
	markFlagRequired(mealCreateSittingCmd, "recipe-id")
	markFlagRequired(mealCreateSittingCmd, subDate)
	markFlagRequired(mealCreateSittingCmd, "meal-category-id")

	mealUpdateSittingCmd.Flags().StringVar(&sittingID, "sitting-id", "", "Meal sitting ID")
	mealUpdateSittingCmd.Flags().StringVar(&sittingSummary, "summary", "", "Meal sitting summary/title")
	mealUpdateSittingCmd.Flags().StringVar(&sittingDate, subDate, "", "Sitting date (YYYY-MM-DD)")
	mealUpdateSittingCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	mealUpdateSittingCmd.Flags().StringVar(&mealCategoryID, "meal-category-id", "", "Meal category ID")
	markFlagRequired(mealUpdateSittingCmd, "sitting-id")

	mealDeleteSittingCmd.Flags().StringVar(&sittingID, "sitting-id", "", "Meal sitting ID")
	mealDeleteSittingCmd.Flags().StringVar(&sittingDate, subDate, "", "Instance date to delete (YYYY-MM-DD)")
	mealDeleteSittingCmd.Flags().BoolVar(&dryRun, "dry-run", false, "Preview without making API calls")
	mealDeleteSittingCmd.Flags().BoolVar(&yes, "yes", false, "Skip confirmation prompt")
	markFlagRequired(mealDeleteSittingCmd, "sitting-id")
	markFlagRequired(mealDeleteSittingCmd, subDate)

	mealGetSittingCmd.Flags().StringVar(&sittingID, "sitting-id", "", "Meal sitting ID")
	markFlagRequired(mealGetSittingCmd, "sitting-id")

	mealAddToGroceryCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	markFlagRequired(mealAddToGroceryCmd, "recipe-id")

	mealSittingRecipeCmd.Flags().StringVar(&sittingID, "sitting-id", "", "Meal sitting ID")
	markFlagRequired(mealSittingRecipeCmd, "sitting-id")

	mealPlanCmd.Flags().StringSliceVar(&mealPlanRecipeIDs, "recipes", nil, "Recipe IDs (comma-separated)")
	mealPlanCmd.Flags().StringSliceVar(&mealPlanCategoryIDs, "categories", nil, "Meal category IDs (comma-separated)")
	mealPlanCmd.Flags().StringVar(&mealPlanStartDate, "start-date", "", "Start date (YYYY-MM-DD)")
	markFlagRequired(mealPlanCmd, "recipes")
	markFlagRequired(mealPlanCmd, "categories")
	markFlagRequired(mealPlanCmd, "start-date")
}
