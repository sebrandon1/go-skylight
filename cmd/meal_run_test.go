package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
)

func mealMockHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/meals/categories"):
			fmt.Fprint(w, `{"data":[{"id":"mc1","attributes":{"label":"Dinner","color":"#FF0000"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/recipe1/add_to_grocery_list"):
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/recipe1") && r.Method == http.MethodPatch:
			fmt.Fprint(w, `{"data":{"id":"recipe1","type":"meal_recipe","attributes":{"summary":"Updated","description":""}}}`)
		case strings.HasSuffix(r.URL.Path, "/recipe1") && r.Method == http.MethodDelete:
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/recipe1"):
			fmt.Fprint(w, `{"data":{"id":"recipe1","type":"meal_recipe","attributes":{"summary":"Tacos","description":""}}}`)
		case strings.HasSuffix(r.URL.Path, "/meals/recipes") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":{"id":"recipe1","type":"meal_recipe","attributes":{"summary":"Tacos","description":""}}}`)
		case strings.HasSuffix(r.URL.Path, "/meals/recipes"):
			fmt.Fprint(w, `{"data":[{"id":"recipe1","type":"meal_recipe","attributes":{"summary":"Tacos","description":""}}]}`)
		case strings.HasSuffix(r.URL.Path, "/sitting1/instances/2026-01-01"):
			w.WriteHeader(http.StatusOK)
		case strings.HasSuffix(r.URL.Path, "/sitting1"):
			fmt.Fprint(w, `{"data":{"id":"sitting1","type":"meal_sitting","attributes":{"summary":"dinner"},"relationships":{"meal_recipe":{"data":{"id":"recipe1","type":"meal_recipe"}}}}}`)
		case strings.HasSuffix(r.URL.Path, "/meals/sittings") && r.Method == http.MethodPost:
			fmt.Fprint(w, `{"data":[{"id":"sitting1","type":"meal_sitting","attributes":{"summary":"dinner"}}]}`)
		case strings.HasSuffix(r.URL.Path, "/meals/sittings"):
			fmt.Fprint(w, `{"data":[{"id":"sitting1","type":"meal_sitting","attributes":{"summary":"dinner"}}]}`)
		default:
			fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen"}}}`)
		}
	}
}

func TestMealCategoriesCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())

	out := captureStdout(func() {
		if err := mealCategoriesCmd.RunE(mealCategoriesCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dinner") {
		t.Errorf("expected category in output, got: %s", out)
	}
}

func TestMealRecipesCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())

	out := captureStdout(func() {
		if err := mealRecipesCmd.RunE(mealRecipesCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Tacos") {
		t.Errorf("expected recipe in output, got: %s", out)
	}
}

func TestMealRecipeInfoCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origID := recipeID
	recipeID = "recipe1"
	t.Cleanup(func() { recipeID = origID })

	out := captureStdout(func() {
		if err := mealRecipeInfoCmd.RunE(mealRecipeInfoCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Tacos") {
		t.Errorf("expected recipe in output, got: %s", out)
	}
}

func TestMealCreateRecipeCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origTitle := recipeTitle
	recipeTitle = "Tacos"
	t.Cleanup(func() { recipeTitle = origTitle })

	out := captureStdout(func() {
		if err := mealCreateRecipeCmd.RunE(mealCreateRecipeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Tacos") {
		t.Errorf("expected created recipe in output, got: %s", out)
	}
}

func TestMealDeleteRecipeCmd_DryRun(t *testing.T) {
	origID, origDryRun := recipeID, dryRun
	recipeID, dryRun = "recipe1", true
	t.Cleanup(func() { recipeID, dryRun = origID, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := mealDeleteRecipeCmd.RunE(mealDeleteRecipeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestMealDeleteRecipeCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origID, origYes := recipeID, yes
	recipeID, yes = "recipe1", true
	t.Cleanup(func() { recipeID, yes = origID, origYes })

	out := captureStdout(func() {
		if err := mealDeleteRecipeCmd.RunE(mealDeleteRecipeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestMealSittingsCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())

	out := captureStdout(func() {
		if err := mealSittingsCmd.RunE(mealSittingsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "sitting1") {
		t.Errorf("expected sitting in output, got: %s", out)
	}
}

func TestMealCreateSittingCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origRecipe, origDate, origCat := recipeID, sittingDate, mealCategoryID
	recipeID, sittingDate, mealCategoryID = "recipe1", "2026-01-01", "mc1"
	t.Cleanup(func() { recipeID, sittingDate, mealCategoryID = origRecipe, origDate, origCat })

	out := captureStdout(func() {
		if err := mealCreateSittingCmd.RunE(mealCreateSittingCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "sitting1") {
		t.Errorf("expected created sitting in output, got: %s", out)
	}
}

func TestMealDeleteSittingCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origID, origDate, origYes := sittingID, sittingDate, yes
	sittingID, sittingDate, yes = "sitting1", "2026-01-01", true
	t.Cleanup(func() { sittingID, sittingDate, yes = origID, origDate, origYes })

	out := captureStdout(func() {
		if err := mealDeleteSittingCmd.RunE(mealDeleteSittingCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "deleted successfully") {
		t.Errorf("expected deletion confirmation, got: %s", out)
	}
}

func TestMealDeleteSittingCmd_DryRun(t *testing.T) {
	origID, origDate, origDryRun := sittingID, sittingDate, dryRun
	sittingID, sittingDate, dryRun = "sitting1", "2026-01-01", true
	t.Cleanup(func() { sittingID, sittingDate, dryRun = origID, origDate, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := mealDeleteSittingCmd.RunE(mealDeleteSittingCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestMealAddToGroceryCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origID := recipeID
	recipeID = "recipe1"
	t.Cleanup(func() { recipeID = origID })

	out := captureStdout(func() {
		if err := mealAddToGroceryCmd.RunE(mealAddToGroceryCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "added to grocery list successfully") {
		t.Errorf("expected confirmation, got: %s", out)
	}
}

func TestMealGetSittingCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origID := sittingID
	sittingID = "sitting1"
	t.Cleanup(func() { sittingID = origID })

	out := captureStdout(func() {
		if err := mealGetSittingCmd.RunE(mealGetSittingCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "sitting1") {
		t.Errorf("expected sitting ID in output, got: %s", out)
	}
	if !strings.Contains(out, "dinner") {
		t.Errorf("expected sitting summary in output, got: %s", out)
	}
}

func TestMealSittingRecipeCmd_WithRecipe(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origID := sittingID
	sittingID = "sitting1"
	t.Cleanup(func() { sittingID = origID })

	out := captureStdout(func() {
		if err := mealSittingRecipeCmd.RunE(mealSittingRecipeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Tacos") {
		t.Errorf("expected linked recipe in output, got: %s", out)
	}
}

func TestMealSittingRecipeCmd_NoRecipe(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"sitting2","type":"meal_sitting","attributes":{"summary":"dinner"}}}`)
	})
	origID := sittingID
	sittingID = "sitting2"
	t.Cleanup(func() { sittingID = origID })

	out := captureStdout(func() {
		if err := mealSittingRecipeCmd.RunE(mealSittingRecipeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No recipe linked") {
		t.Errorf("expected no-recipe message, got: %s", out)
	}
}

func TestMealPlanCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origRecipes, origCats, origStart := mealPlanRecipeIDs, mealPlanCategoryIDs, mealPlanStartDate
	mealPlanRecipeIDs, mealPlanCategoryIDs, mealPlanStartDate = []string{"recipe1"}, []string{"mc1"}, "2026-01-01"
	t.Cleanup(func() {
		mealPlanRecipeIDs, mealPlanCategoryIDs, mealPlanStartDate = origRecipes, origCats, origStart
	})

	out := captureStdout(func() {
		if err := mealPlanCmd.RunE(mealPlanCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "sitting1") {
		t.Errorf("expected planned sitting in output, got: %s", out)
	}
}

func TestMealUpdateRecipeCmd(t *testing.T) {
	newCmdTestClient(t, mealMockHandler())
	origID, origTitle := recipeID, recipeTitle
	recipeID, recipeTitle = "recipe1", "Updated"
	t.Cleanup(func() { recipeID, recipeTitle = origID, origTitle })

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := mealUpdateRecipeCmd.Flags().Set("title", "Updated"); err != nil {
		t.Fatalf("setting title flag: %v", err)
	}

	out := captureStdout(func() {
		if err := mealUpdateRecipeCmd.RunE(mealUpdateRecipeCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Updated") {
		t.Errorf("expected updated recipe in output, got: %s", out)
	}
}

func TestMealCmdExists(t *testing.T) {
	assertCommandRegistered(t, rootCmd, "meal")
}
