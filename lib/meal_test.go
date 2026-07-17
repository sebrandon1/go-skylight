package lib

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
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
		wantDate string
		wantErr  bool
	}{
		{
			name:     "returns sittings",
			status:   http.StatusOK,
			response: `{"data":[{"id":"1","type":"meal_sitting","attributes":{"summary":"Dinner","instances":["2024-03-10"]},"relationships":{"meal_category":{"data":{"id":"cat1","type":"meal_category"}},"meal_recipe":{"data":{"id":"r1","type":"meal_recipe"}}}}]}`,
			wantLen:  1,
			wantDate: "2024-03-10",
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
			if tc.wantDate != "" && len(sittings) > 0 && sittings[0].Date != tc.wantDate {
				t.Errorf("Date: want %q got %q", tc.wantDate, sittings[0].Date)
			}
		})
	}
}

func TestMealSittingAPIEntry_DateFromInstances(t *testing.T) {
	tests := []struct {
		name      string
		instances []string
		wantDate  string
	}{
		{
			name:      "first instance becomes Date",
			instances: []string{"2026-05-03", "2026-05-10"},
			wantDate:  "2026-05-03",
		},
		{
			name:      "single instance",
			instances: []string{"2024-03-15"},
			wantDate:  "2024-03-15",
		},
		{
			name:      "no instances leaves Date empty",
			instances: nil,
			wantDate:  "",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := mealSittingAPIEntry{ID: "1"}
			e.Attributes.Summary = "Dinner"
			e.Attributes.Instances = tc.instances
			s := e.toMealSitting()
			if s.Date != tc.wantDate {
				t.Errorf("Date: want %q got %q", tc.wantDate, s.Date)
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
		wantDate    string
		wantErr     bool
	}{
		{
			name:        "creates sitting",
			input:       MealSittingData{RecipeID: "recipe1", Date: "2024-01-15", MealCategoryID: "cat1"},
			status:      http.StatusCreated,
			response:    `{"data":[{"id":"1","type":"meal_sitting","attributes":{"summary":"dinner","instances":["2024-01-15"]},"relationships":{"meal_category":{"data":{"id":"cat1","type":"meal_category"}},"meal_recipe":{"data":{"id":"recipe1","type":"meal_recipe"}}}}]}`,
			wantSummary: "dinner",
			wantDate:    "2024-01-15",
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
			if tc.wantDate != "" && sitting != nil && sitting.Date != tc.wantDate {
				t.Errorf("Date: want %q got %q", tc.wantDate, sitting.Date)
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

func TestCreateMealSittingFieldNames(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if _, ok := body["meal_recipe_id"]; !ok {
			t.Error(`request body must have "meal_recipe_id" key`)
		}
		if _, ok := body["recipe_id"]; ok {
			t.Error(`request body must not have "recipe_id" key`)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"data":[{"id":"s1","type":"meal_sitting","attributes":{"summary":""},"relationships":{"meal_category":{"data":null},"meal_recipe":{"data":{"id":"recipe1","type":"meal_recipe"}}}}]}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	if _, err := client.CreateMealSitting("frame1", MealSittingData{RecipeID: "recipe1", Date: "2024-01-15"}); err != nil {
		t.Fatalf("CreateMealSitting: %v", err)
	}
}

func TestCreateMealSittingSummary(t *testing.T) {
	tests := []struct {
		name        string
		input       MealSittingData
		wantSummary string
		wantAbsent  bool
	}{
		{
			name:        "sends summary when provided",
			input:       MealSittingData{Summary: "Pasta Night", Date: "2024-01-15"},
			wantSummary: "Pasta Night",
		},
		{
			name:       "omits summary when empty",
			input:      MealSittingData{Date: "2024-01-15"},
			wantAbsent: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode body: %v", err)
					w.WriteHeader(http.StatusBadRequest)
					return
				}
				val, present := body["summary"]
				if tc.wantAbsent && present {
					t.Errorf("expected summary to be absent, got %v", val)
				}
				if !tc.wantAbsent {
					if !present {
						t.Error("expected summary to be present")
					} else if val != tc.wantSummary {
						t.Errorf("summary: want %q got %v", tc.wantSummary, val)
					}
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				if _, err := w.Write([]byte(`{"data":[{"id":"s1","type":"meal_sitting","attributes":{"summary":""},"relationships":{"meal_category":{"data":null},"meal_recipe":{"data":null}}}]}`)); err != nil {
					t.Errorf("write: %v", err)
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			if _, err := client.CreateMealSitting("frame1", tc.input); err != nil {
				t.Fatalf("CreateMealSitting: %v", err)
			}
		})
	}
}

func TestGetMealSitting(t *testing.T) {
	tests := []struct {
		name       string
		sittingID  string
		status     int
		response   string
		wantID     string
		wantRecipe string
		wantErr    bool
	}{
		{
			name:       "returns sitting with recipe",
			sittingID:  "s1",
			status:     http.StatusOK,
			response:   `{"data":{"id":"s1","type":"meal_sitting","attributes":{"summary":"Dinner"},"relationships":{"meal_category":{"data":{"id":"cat1","type":"meal_category"}},"meal_recipe":{"data":{"id":"r1","type":"meal_recipe"}}}}}`,
			wantID:     "s1",
			wantRecipe: "r1",
		},
		{
			name:      "returns sitting without recipe",
			sittingID: "s2",
			status:    http.StatusOK,
			response:  `{"data":{"id":"s2","type":"meal_sitting","attributes":{"summary":"Lunch"},"relationships":{"meal_category":{"data":null},"meal_recipe":{"data":null}}}}`,
			wantID:    "s2",
		},
		{
			name:      "not found returns error",
			sittingID: "nonexistent",
			status:    http.StatusNotFound,
			wantErr:   true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
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
			sitting, err := client.GetMealSitting("frame1", tc.sittingID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if sitting.ID != tc.wantID {
				t.Errorf("ID: want %q got %q", tc.wantID, sitting.ID)
			}
			if sitting.RecipeID != tc.wantRecipe {
				t.Errorf("RecipeID: want %q got %q", tc.wantRecipe, sitting.RecipeID)
			}
		})
	}
}

func TestGetSittingRecipe(t *testing.T) {
	tests := []struct {
		name          string
		sittingResp   string
		sittingStatus int
		recipeResp    string
		recipeStatus  int
		wantRecipe    bool
		wantErr       bool
	}{
		{
			name:          "returns sitting with linked recipe",
			sittingResp:   `{"data":{"id":"s1","type":"meal_sitting","attributes":{"summary":"Dinner"},"relationships":{"meal_category":{"data":null},"meal_recipe":{"data":{"id":"r1","type":"meal_recipe"}}}}}`,
			sittingStatus: http.StatusOK,
			recipeResp:    `{"data":{"id":"r1","type":"meal_recipe","attributes":{"summary":"Pasta","description":"Classic dish"}}}`,
			recipeStatus:  http.StatusOK,
			wantRecipe:    true,
		},
		{
			name:          "returns sitting with no linked recipe",
			sittingResp:   `{"data":{"id":"s2","type":"meal_sitting","attributes":{"summary":"Lunch"},"relationships":{"meal_category":{"data":null},"meal_recipe":{"data":null}}}}`,
			sittingStatus: http.StatusOK,
			wantRecipe:    false,
		},
		{
			name:          "sitting not found returns error",
			sittingStatus: http.StatusNotFound,
			wantErr:       true,
		},
		{
			name:          "recipe fetch error returns error",
			sittingResp:   `{"data":{"id":"s1","type":"meal_sitting","attributes":{"summary":"Dinner"},"relationships":{"meal_category":{"data":null},"meal_recipe":{"data":{"id":"r1","type":"meal_recipe"}}}}}`,
			sittingStatus: http.StatusOK,
			recipeStatus:  http.StatusInternalServerError,
			wantErr:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			callCount := 0
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				callCount++
				w.Header().Set("Content-Type", "application/json")
				if callCount == 1 {
					w.WriteHeader(tc.sittingStatus)
					if tc.sittingResp != "" {
						if _, err := w.Write([]byte(tc.sittingResp)); err != nil {
							t.Errorf("write: %v", err)
						}
					}
				} else {
					w.WriteHeader(tc.recipeStatus)
					if tc.recipeResp != "" {
						if _, err := w.Write([]byte(tc.recipeResp)); err != nil {
							t.Errorf("write: %v", err)
						}
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			result, err := client.GetSittingRecipe("frame1", "s1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if (result.Recipe != nil) != tc.wantRecipe {
				t.Errorf("wantRecipe=%v got recipe=%v", tc.wantRecipe, result.Recipe)
			}
		})
	}
}

func TestDeleteMealSitting(t *testing.T) {
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
			err := client.DeleteMealSitting("frame1", "sitting1", "2026-04-28")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestPlanMeals(t *testing.T) {
	var n int

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		n++
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(mealSittingAPIResponse{
			Data: []mealSittingAPIEntry{
				{ID: fmt.Sprintf("s%d", n)},
			},
		}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	result, err := client.PlanMeals("frame1", MealPlanData{
		RecipeIDs:   []string{"r1", "r2", "r3", "r4", "r5"},
		CategoryIDs: []string{"cat-lunch", "cat-dinner"},
		StartDate:   "2026-04-21",
	})
	if err != nil {
		t.Fatalf("PlanMeals failed: %v", err)
	}

	if len(result.Sittings) != 5 {
		t.Errorf("want 5 sittings, got %d", len(result.Sittings))
	}
}

func TestPlanMealsValidation(t *testing.T) {
	client, _ := NewClientWithToken("u", "t")

	tests := []struct {
		name string
		data MealPlanData
	}{
		{"no recipes", MealPlanData{CategoryIDs: []string{"cat1"}, StartDate: "2026-04-21"}},
		{"no categories", MealPlanData{RecipeIDs: []string{"r1"}, StartDate: "2026-04-21"}},
		{"bad date", MealPlanData{RecipeIDs: []string{"r1"}, CategoryIDs: []string{"cat1"}, StartDate: "not-a-date"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := client.PlanMeals("frame1", tc.data)
			if err == nil {
				t.Error("expected validation error, got nil")
			}
		})
	}
}

func TestPlanMealsDayRotation(t *testing.T) {
	type call struct {
		recipeID   string
		categoryID string
		date       string
	}
	var calls []call

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		calls = append(calls, call{
			recipeID:   fmt.Sprint(body["meal_recipe_id"]),
			categoryID: fmt.Sprint(body["meal_category_id"]),
			date:       fmt.Sprint(body["date"]),
		})
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"data":[{"id":"s1","attributes":{},"relationships":{"meal_category":{"data":null},"meal_recipe":{"data":null}}}]}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.PlanMeals("frame1", MealPlanData{
		RecipeIDs:   []string{"r1", "r2", "r3", "r4"},
		CategoryIDs: []string{"lunch", "dinner"},
		StartDate:   "2026-04-21",
	})
	if err != nil {
		t.Fatalf("PlanMeals: %v", err)
	}

	want := []call{
		{"r1", "lunch", "2026-04-21"},
		{"r2", "dinner", "2026-04-21"},
		{"r3", "lunch", "2026-04-22"},
		{"r4", "dinner", "2026-04-22"},
	}

	if len(calls) != len(want) {
		t.Fatalf("want %d calls, got %d", len(want), len(calls))
	}
	for i, w := range want {
		if calls[i] != w {
			t.Errorf("call[%d]: want %+v, got %+v", i, w, calls[i])
		}
	}
}

func TestPlanMealsPartialFailure(t *testing.T) {
	var callCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		idx := callCount.Add(1)
		w.Header().Set("Content-Type", "application/json")
		if idx >= 3 {
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
				t.Errorf("write: %v", err)
			}
			return
		}
		w.WriteHeader(http.StatusCreated)
		if err := json.NewEncoder(w).Encode(mealSittingAPIResponse{
			Data: []mealSittingAPIEntry{{ID: fmt.Sprintf("s%d", idx)}},
		}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	result, err := client.PlanMeals("frame1", MealPlanData{
		RecipeIDs:   []string{"r1", "r2", "r3"},
		CategoryIDs: []string{"cat1"},
		StartDate:   "2026-04-21",
	})
	if err == nil {
		t.Fatal("expected error for partial failure, got nil")
	}
	if result == nil {
		t.Fatal("expected partial result, got nil")
	}
	if len(result.Sittings) != 2 {
		t.Errorf("want 2 partial sittings, got %d", len(result.Sittings))
	}
}
