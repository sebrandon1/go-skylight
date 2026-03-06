package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRecipes(t *testing.T) {
	mockRecipes := []Recipe{
		{ID: "1", Title: "Pasta"},
		{ID: "2", Title: "Salad"},
	}

	mockResponseJSON, _ := json.Marshal(mockRecipes)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/meals/recipes" {
			t.Errorf("Expected path /api/frames/frame1/meals/recipes, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	recipes, err := client.ListRecipes("frame1")
	if err != nil {
		t.Fatalf("ListRecipes failed: %v", err)
	}

	if len(recipes) != 2 {
		t.Errorf("Expected 2 recipes, got %d", len(recipes))
	}
}

func TestGetRecipe(t *testing.T) {
	mockRecipe := Recipe{ID: "1", Title: "Pasta", Ingredients: []string{"noodles", "sauce"}}

	mockResponseJSON, _ := json.Marshal(mockRecipe)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/meals/recipes/1" {
			t.Errorf("Expected path /api/frames/frame1/meals/recipes/1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	recipe, err := client.GetRecipe("frame1", "1")
	if err != nil {
		t.Fatalf("GetRecipe failed: %v", err)
	}

	if recipe.Title != "Pasta" {
		t.Errorf("Expected title 'Pasta', got '%s'", recipe.Title)
	}

	if len(recipe.Ingredients) != 2 {
		t.Errorf("Expected 2 ingredients, got %d", len(recipe.Ingredients))
	}
}

func TestCreateMealSitting(t *testing.T) {
	mockSitting := MealSitting{ID: "1", RecipeID: "recipe1", Date: "2024-01-15", MealType: "dinner"}

	mockResponseJSON, _ := json.Marshal(mockSitting)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/meals/sittings" {
			t.Errorf("Expected path /api/frames/frame1/meals/sittings, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	sitting, err := client.CreateMealSitting("frame1", MealSittingData{RecipeID: "recipe1", Date: "2024-01-15", MealType: "dinner"})
	if err != nil {
		t.Fatalf("CreateMealSitting failed: %v", err)
	}

	if sitting.MealType != "dinner" {
		t.Errorf("Expected meal type 'dinner', got '%s'", sitting.MealType)
	}
}

func TestListMealCategories(t *testing.T) {
	mockCategories := []MealCategory{
		{ID: "1", Name: "Italian"},
		{ID: "2", Name: "Mexican"},
	}

	mockResponseJSON, _ := json.Marshal(mockCategories)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/meals/categories" {
			t.Errorf("Expected path /api/frames/frame1/meals/categories, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	categories, err := client.ListMealCategories("frame1")
	if err != nil {
		t.Fatalf("ListMealCategories failed: %v", err)
	}

	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}
}

func TestCreateRecipe(t *testing.T) {
	mockRecipe := Recipe{ID: "3", Title: "Tacos"}

	mockResponseJSON, _ := json.Marshal(mockRecipe)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/meals/recipes" {
			t.Errorf("Expected path /api/frames/frame1/meals/recipes, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	recipe, err := client.CreateRecipe("frame1", RecipeData{Title: "Tacos"})
	if err != nil {
		t.Fatalf("CreateRecipe failed: %v", err)
	}

	if recipe.Title != "Tacos" {
		t.Errorf("Expected title 'Tacos', got '%s'", recipe.Title)
	}
}

func TestUpdateRecipe(t *testing.T) {
	mockRecipe := Recipe{ID: "1", Title: "Updated Pasta"}

	mockResponseJSON, _ := json.Marshal(mockRecipe)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/meals/recipes/1" {
			t.Errorf("Expected path /api/frames/frame1/meals/recipes/1, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	recipe, err := client.UpdateRecipe("frame1", "1", RecipeData{Title: "Updated Pasta"})
	if err != nil {
		t.Fatalf("UpdateRecipe failed: %v", err)
	}

	if recipe.Title != "Updated Pasta" {
		t.Errorf("Expected title 'Updated Pasta', got '%s'", recipe.Title)
	}
}

func TestDeleteRecipe(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/meals/recipes/1" {
			t.Errorf("Expected path /api/frames/frame1/meals/recipes/1, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.DeleteRecipe("frame1", "1")
	if err != nil {
		t.Fatalf("DeleteRecipe failed: %v", err)
	}
}

func TestListMealSittings(t *testing.T) {
	mockSittings := []MealSitting{
		{ID: "1", RecipeID: "r1", Date: "2024-01-15", MealType: "dinner"},
	}

	mockResponseJSON, _ := json.Marshal(mockSittings)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/meals/sittings" {
			t.Errorf("Expected path /api/frames/frame1/meals/sittings, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	sittings, err := client.ListMealSittings("frame1")
	if err != nil {
		t.Fatalf("ListMealSittings failed: %v", err)
	}

	if len(sittings) != 1 {
		t.Errorf("Expected 1 sitting, got %d", len(sittings))
	}
}

func TestAddRecipeToGroceryList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/meals/recipes/recipe1/add_to_grocery_list" {
			t.Errorf("Expected path /api/frames/frame1/meals/recipes/recipe1/add_to_grocery_list, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.AddRecipeToGroceryList("frame1", "recipe1")
	if err != nil {
		t.Fatalf("AddRecipeToGroceryList failed: %v", err)
	}
}

func TestMealErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "Not found"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListRecipes("frame1")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	_, err = client.GetRecipe("frame1", "nonexistent")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
