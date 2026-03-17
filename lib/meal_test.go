package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRecipes(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns recipes",
			status:   http.StatusOK,
			response: `{"data":[{"id":"1","type":"meal_recipe","attributes":{"summary":"Pasta","description":""}},{"id":"2","type":"meal_recipe","attributes":{"summary":"Salad","description":""}}]}`,
			wantLen:  2,
		},
		{
			name:    "not found returns error",
			status:  http.StatusNotFound,
			wantErr: true,
		},
		{
			name:     "invalid JSON returns error",
			status:   http.StatusOK,
			response: `not valid json`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/meals/recipes" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			recipes, err := client.ListRecipes("frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && len(recipes) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(recipes))
			}
		})
	}
}

func TestGetRecipe(t *testing.T) {
	tests := []struct {
		name        string
		recipeID    string
		status      int
		response    string
		wantTitle   string
		wantIngreds int
		wantErr     bool
	}{
		{
			name:        "returns recipe with ingredients",
			recipeID:    "1",
			status:      http.StatusOK,
			response:    `{"data":{"id":"1","type":"meal_recipe","attributes":{"summary":"Pasta","description":"","ingredients":["noodles","sauce"]}}}`,
			wantTitle:   "Pasta",
			wantIngreds: 2,
		},
		{
			name:     "not found returns error",
			recipeID: "nonexistent",
			status:   http.StatusNotFound,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			recipe, err := client.GetRecipe("frame1", tc.recipeID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if recipe.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, recipe.Title)
			}
			if len(recipe.Ingredients) != tc.wantIngreds {
				t.Errorf("Ingredients: want %d got %d", tc.wantIngreds, len(recipe.Ingredients))
			}
		})
	}
}

func TestCreateRecipe(t *testing.T) {
	tests := []struct {
		name      string
		input     RecipeData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "creates recipe",
			input:     RecipeData{Title: "Tacos"},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"3","type":"meal_recipe","attributes":{"summary":"Tacos","description":""}}}`,
			wantTitle: "Tacos",
		},
		{
			name:      "sends all fields in request body",
			input:     RecipeData{Title: "Spaghetti", Description: "Classic Italian", Ingredients: []string{"pasta", "sauce"}},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"r1","type":"meal_recipe","attributes":{"summary":"Spaghetti","description":"Classic Italian"}}}`,
			wantTitle: "Spaghetti",
		},
		{
			name:    "server error returns error",
			input:   RecipeData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/meals/recipes" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			recipe, err := client.CreateRecipe("frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && recipe.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, recipe.Title)
			}
		})
	}
}

func TestUpdateRecipe(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "updates recipe",
			status:    http.StatusOK,
			response:  `{"data":{"id":"1","type":"meal_recipe","attributes":{"summary":"Updated Pasta","description":""}}}`,
			wantTitle: "Updated Pasta",
		},
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			recipe, err := client.UpdateRecipe("frame1", "1", RecipeData{Title: "Updated Pasta"})
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && recipe.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, recipe.Title)
			}
		})
	}
}

func TestDeleteRecipe(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"deletes with 204", http.StatusNoContent, false},
		{"server error returns error", http.StatusInternalServerError, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.DeleteRecipe("frame1", "1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestListMealCategories(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns categories",
			status:   http.StatusOK,
			response: `{"data":[{"id":"1","type":"meal_category","attributes":{"label":"Italian","color":"#FF0000"}},{"id":"2","type":"meal_category","attributes":{"label":"Mexican","color":"#00FF00"}}]}`,
			wantLen:  2,
		},
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/frames/frame1/meals/categories" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			categories, err := client.ListMealCategories("frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && len(categories) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(categories))
			}
		})
	}
}

func TestListMealSittings(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns sittings",
			status:   http.StatusOK,
			response: `{"data":[{"id":"1","type":"meal_sitting","attributes":{"summary":"Dinner"},"relationships":{"meal_category":{"data":{"id":"cat1","type":"meal_category"}},"meal_recipe":{"data":{"id":"r1","type":"meal_recipe"}}}}]}`,
			wantLen:  1,
		},
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/frames/frame1/meals/sittings" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			sittings, err := client.ListMealSittings("frame1", MealSittingListOptions{})
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && len(sittings) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(sittings))
			}
		})
	}
}

func TestCreateMealSitting(t *testing.T) {
	tests := []struct {
		name        string
		input       MealSittingData
		status      int
		response    string
		wantSummary string
		wantErr     bool
	}{
		{
			name:        "creates sitting",
			input:       MealSittingData{RecipeID: "recipe1", Date: "2024-01-15", MealCategoryID: "cat1"},
			status:      http.StatusCreated,
			response:    `{"data":[{"id":"1","type":"meal_sitting","attributes":{"summary":"dinner"},"relationships":{"meal_category":{"data":{"id":"cat1","type":"meal_category"}},"meal_recipe":{"data":{"id":"recipe1","type":"meal_recipe"}}}}]}`,
			wantSummary: "dinner",
		},
		{
			name:    "server error returns error",
			input:   MealSittingData{RecipeID: "r1", Date: "2024-01-15"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/meals/sittings" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			sitting, err := client.CreateMealSitting("frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && sitting.Summary != tc.wantSummary {
				t.Errorf("Summary: want %q got %q", tc.wantSummary, sitting.Summary)
			}
		})
	}
}

func TestAddRecipeToGroceryList(t *testing.T) {
	tests := []struct {
		name     string
		recipeID string
		status   int
		wantErr  bool
	}{
		{
			name:     "adds recipe to grocery list",
			recipeID: "recipe1",
			status:   http.StatusOK,
		},
		{
			name:     "server error returns error",
			recipeID: "recipe1",
			status:   http.StatusInternalServerError,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/meals/recipes/recipe1/add_to_grocery_list" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.AddRecipeToGroceryList("frame1", tc.recipeID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}
