package cmd

import (
	"fmt"
	"os"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
)

var (
	recipeID    string
	recipeTitle string
	sittingDate string
	mealType    string
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

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		categories, err := client.ListMealCategories(frameID)
		if err != nil {
			fmt.Printf("Error listing meal categories: %v\n", err)
			os.Exit(1)
		}

		printJSON(categories)
	},
}

var mealRecipesCmd = &cobra.Command{
	Use:   "recipes",
	Short: "List recipes",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		recipes, err := client.ListRecipes(frameID)
		if err != nil {
			fmt.Printf("Error listing recipes: %v\n", err)
			os.Exit(1)
		}

		printJSON(recipes)
	},
}

var mealRecipeInfoCmd = &cobra.Command{
	Use:   "recipe-info",
	Short: "Get recipe details",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		recipe, err := client.GetRecipe(frameID, recipeID)
		if err != nil {
			fmt.Printf("Error getting recipe: %v\n", err)
			os.Exit(1)
		}

		printJSON(recipe)
	},
}

var mealCreateRecipeCmd = &cobra.Command{
	Use:   "create-recipe",
	Short: "Create a recipe",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		recipe, err := client.CreateRecipe(frameID, lib.RecipeData{
			Title: recipeTitle,
		})
		if err != nil {
			fmt.Printf("Error creating recipe: %v\n", err)
			os.Exit(1)
		}

		printJSON(recipe)
	},
}

var mealDeleteRecipeCmd = &cobra.Command{
	Use:   "delete-recipe",
	Short: "Delete a recipe",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		err = client.DeleteRecipe(frameID, recipeID)
		if err != nil {
			fmt.Printf("Error deleting recipe: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Recipe deleted successfully")
	},
}

var mealSittingsCmd = &cobra.Command{
	Use:   "sittings",
	Short: "List meal sittings",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		sittings, err := client.ListMealSittings(frameID)
		if err != nil {
			fmt.Printf("Error listing meal sittings: %v\n", err)
			os.Exit(1)
		}

		printJSON(sittings)
	},
}

var mealCreateSittingCmd = &cobra.Command{
	Use:   "create-sitting",
	Short: "Create a meal sitting",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		sitting, err := client.CreateMealSitting(frameID, lib.MealSittingData{
			RecipeID: recipeID,
			Date:     sittingDate,
			MealType: mealType,
		})
		if err != nil {
			fmt.Printf("Error creating meal sitting: %v\n", err)
			os.Exit(1)
		}

		printJSON(sitting)
	},
}

var mealAddToGroceryCmd = &cobra.Command{
	Use:   "add-to-grocery",
	Short: "Add recipe ingredients to grocery list",
	Run: func(cmd *cobra.Command, args []string) {
		requireFrameID()

		client, err := lib.NewClientWithToken(userID, token)
		if err != nil {
			fmt.Printf("Error creating client: %v\n", err)
			os.Exit(1)
		}

		err = client.AddRecipeToGroceryList(frameID, recipeID)
		if err != nil {
			fmt.Printf("Error adding to grocery list: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Recipe added to grocery list successfully")
	},
}

func init() {
	mealCmd.AddCommand(mealCategoriesCmd)
	mealCmd.AddCommand(mealRecipesCmd)
	mealCmd.AddCommand(mealRecipeInfoCmd)
	mealCmd.AddCommand(mealCreateRecipeCmd)
	mealCmd.AddCommand(mealDeleteRecipeCmd)
	mealCmd.AddCommand(mealSittingsCmd)
	mealCmd.AddCommand(mealCreateSittingCmd)
	mealCmd.AddCommand(mealAddToGroceryCmd)

	mealRecipeInfoCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	mealCreateRecipeCmd.Flags().StringVar(&recipeTitle, "title", "", "Recipe title")
	mealDeleteRecipeCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")

	mealCreateSittingCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
	mealCreateSittingCmd.Flags().StringVar(&sittingDate, "date", "", "Sitting date")
	mealCreateSittingCmd.Flags().StringVar(&mealType, "meal-type", "", "Meal type (breakfast, lunch, dinner)")

	mealAddToGroceryCmd.Flags().StringVar(&recipeID, "recipe-id", "", "Recipe ID")
}
