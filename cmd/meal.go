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

	mealPlanRecipeIDs   []string
	mealPlanCategoryIDs []string
	mealPlanStartDate   string
)

var mealCmd = &cobra.Command{
	Use:   "meal",
	Short: "Meal and recipe management commands",
}

var mealCategoriesCmd = &cobra.Command{
	Use:   "categories",
	Short: "List meal categories",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		categories, err := client.ListMealCategories(frameID)
		if err != nil {
			fatal("listing meal categories", err)
		}

		printOutput(categories)
	},
}

var mealRecipesCmd = &cobra.Command{
	Use:   "recipes",
	Short: "List recipes",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		recipes, err := client.ListRecipes(frameID)
		if err != nil {
			fatal("listing recipes", err)
		}

		printOutput(recipes)
	},
}

var mealRecipeInfoCmd = &cobra.Command{
	Use:   "recipe-info",
	Short: "Get recipe details",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		recipe, err := client.GetRecipe(frameID, recipeID)
		if err != nil {
			fatal("getting recipe", err)
		}

		printJSON(recipe)
	},
}

var mealCreateRecipeCmd = &cobra.Command{
	Use:   "create-recipe",
	Short: "Create a recipe",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		recipe, err := client.CreateRecipe(frameID, lib.RecipeData{
			Title:          recipeTitle,
			Description:    recipeDescription,
			Ingredients:    recipeIngredients,
			URL:            recipeURL,
			MealCategoryID: recipeCategoryID,
		})
		if err != nil {
			fatal("creating recipe", err)
		}

		printJSON(recipe)
	},
}

var mealDeleteRecipeCmd = &cobra.Command{
	Use:   "delete-recipe",
	Short: "Delete a recipe",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.DeleteRecipe(frameID, recipeID)
		if err != nil {
			fatal("deleting recipe", err)
		}

		fmt.Println("Recipe deleted successfully")
	},
}

var mealSittingsCmd = &cobra.Command{
	Use:   "sittings",
	Short: "List meal sittings",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		for _, f := range []struct {
			name string
			val  string
		}{{"date-min", sittingDateMin}, {"date-max", sittingDateMax}} {
			if cmd.Flags().Changed(f.name) {
				if err := validateDate(f.val); err != nil {
					fmt.Fprintln(os.Stderr, err)
					os.Exit(1)
				}
			}
		}

		client := getClient()

		sittings, err := client.ListMealSittings(frameID, lib.MealSittingListOptions{
			DateMin: sittingDateMin,
			DateMax: sittingDateMax,
		})
		if err != nil {
			fatal("listing meal sittings", err)
		}

		printOutput(sittings)
	},
}

var mealCreateSittingCmd = &cobra.Command{
	Use:   "create-sitting",
	Short: "Create a meal sitting",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if err := validateDate(sittingDate); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		client := getClient()

		sitting, err := client.CreateMealSitting(frameID, lib.MealSittingData{
			Summary:        sittingSummary,
			RecipeID:       recipeID,
			Date:           sittingDate,
			MealCategoryID: mealCategoryID,
		})
		if err != nil {
			fatal("creating meal sitting", err)
		}

		printJSON(sitting)
	},
}

var mealDeleteSittingCmd = &cobra.Command{
	Use:   "delete-sitting",
	Short: "Delete a meal sitting instance",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if err := validateDate(sittingDate); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		client := getClient()

		err := client.DeleteMealSitting(frameID, sittingID, sittingDate)
		if err != nil {
			fatal("deleting meal sitting", err)
		}

		fmt.Println("Meal sitting deleted successfully")
	},
}

var mealAddToGroceryCmd = &cobra.Command{
	Use:   "add-to-grocery",
	Short: "Add recipe ingredients to grocery list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		err := client.AddRecipeToGroceryList(frameID, recipeID)
		if err != nil {
			fatal("adding to grocery list", err)
		}

		fmt.Println("Recipe added to grocery list successfully")
	},
}

var mealGetSittingCmd = &cobra.Command{
	Use:   "get-sitting",
	Short: "Get meal sitting details",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		sitting, err := client.GetMealSitting(frameID, sittingID)
		if err != nil {
			fatal("getting meal sitting", err)
		}

		printJSON(sitting)
	},
}

var mealSittingRecipeCmd = &cobra.Command{
	Use:   "sitting-recipe",
	Short: "View recipe details from a meal sitting",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		result, err := client.GetSittingRecipe(frameID, sittingID)
		if err != nil {
			fatal("getting sitting recipe", err)
		}

		if result.Recipe == nil {
			fmt.Println("No recipe linked to this meal sitting")
			return
		}

		printJSON(result)
	},
}

var mealPlanCmd = &cobra.Command{
	Use:   "plan",
	Short: "Schedule a batch of meals from a recipe list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		if err := validateDate(mealPlanStartDate); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}

		client := getClient()

		result, err := client.PlanMeals(frameID, lib.MealPlanData{
			RecipeIDs:   mealPlanRecipeIDs,
			CategoryIDs: mealPlanCategoryIDs,
			StartDate:   mealPlanStartDate,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: planning meals: %v\n", err)
			if result != nil && len(result.Sittings) > 0 {
				fmt.Fprintf(os.Stderr, "Partial result (%d sittings created):\n", len(result.Sittings))
				printJSON(result)
			}
			os.Exit(1)
		}

		printJSON(result)
	},
}

var mealUpdateRecipeCmd = &cobra.Command{
	Use:   "update-recipe",
	Short: "Update a recipe",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client := getClient()

		data := lib.RecipeData{}
		if cmd.Flags().Changed("title") {
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

		recipe, err := client.UpdateRecipe(frameID, recipeID, data)
		if err != nil {
			fatal("updating recipe", err)
		}

		printJSON(recipe)
	},
}

func init() {
	mealCmd.AddCommand(mealCategoriesCmd)
	mealCmd.AddCommand(mealRecipesCmd)
	mealCmd.AddCommand(mealRecipeInfoCmd)
	mealCmd.AddCommand(mealCreateRecipeCmd)
	mealCmd.AddCommand(mealUpdateRecipeCmd)
	mealCmd.AddCommand(mealDeleteRecipeCmd)
	mealCmd.AddCommand(mealSittingsCmd)
	mealCmd.AddCommand(mealCreateSittingCmd)
	mealCmd.AddCommand(mealDeleteSittingCmd)
	mealCmd.AddCommand(mealGetSittingCmd)
	mealCmd.AddCommand(mealAddToGroceryCmd)
	mealCmd.AddCommand(mealSittingRecipeCmd)
	mealCmd.AddCommand(mealPlanCmd)

	mealRecipeInfoCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	markFlagRequired(mealRecipeInfoCmd, "recipe-id")

	mealCreateRecipeCmd.Flags().StringVar(&recipeTitle, "title", "", "Recipe title")
	mealCreateRecipeCmd.Flags().StringVar(&recipeDescription, "description", "", "Recipe description")
	mealCreateRecipeCmd.Flags().StringSliceVar(&recipeIngredients, "ingredients", nil, "Ingredients (comma-separated)")
	mealCreateRecipeCmd.Flags().StringVar(&recipeURL, "url", "", "Recipe URL")
	mealCreateRecipeCmd.Flags().StringVar(&recipeCategoryID, "meal-category-id", "", "Meal category ID")
	markFlagRequired(mealCreateRecipeCmd, "title")

	mealUpdateRecipeCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID to update")
	mealUpdateRecipeCmd.Flags().StringVar(&recipeTitle, "title", "", "Recipe title")
	mealUpdateRecipeCmd.Flags().StringVar(&recipeDescription, "description", "", "Recipe description")
	mealUpdateRecipeCmd.Flags().StringSliceVar(&recipeIngredients, "ingredients", nil, "Ingredients (comma-separated)")
	mealUpdateRecipeCmd.Flags().StringVar(&recipeURL, "url", "", "Recipe URL")
	markFlagRequired(mealUpdateRecipeCmd, "recipe-id")

	mealDeleteRecipeCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	markFlagRequired(mealDeleteRecipeCmd, "recipe-id")

	mealSittingsCmd.Flags().StringVar(&sittingDateMin, "date-min", "", "Minimum date filter (YYYY-MM-DD)")
	mealSittingsCmd.Flags().StringVar(&sittingDateMax, "date-max", "", "Maximum date filter (YYYY-MM-DD)")

	mealCreateSittingCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	mealCreateSittingCmd.Flags().StringVar(&sittingSummary, "summary", "", "Meal sitting summary/title")
	mealCreateSittingCmd.Flags().StringVar(&sittingDate, "date", "", "Sitting date")
	mealCreateSittingCmd.Flags().StringVar(&mealCategoryID, "meal-category-id", "", "Meal category ID")
	markFlagRequired(mealCreateSittingCmd, "recipe-id")
	markFlagRequired(mealCreateSittingCmd, "date")
	markFlagRequired(mealCreateSittingCmd, "meal-category-id")

	mealDeleteSittingCmd.Flags().StringVar(&sittingID, "sitting-id", "", "Meal sitting ID")
	mealDeleteSittingCmd.Flags().StringVar(&sittingDate, "date", "", "Instance date to delete (YYYY-MM-DD)")
	markFlagRequired(mealDeleteSittingCmd, "sitting-id")
	markFlagRequired(mealDeleteSittingCmd, "date")

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
