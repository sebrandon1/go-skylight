package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewClientWithToken(t *testing.T) {
	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("NewClientWithToken failed: %v", err)
	}

	if client.UserID != "user1" {
		t.Errorf("Expected UserID 'user1', got '%s'", client.UserID)
	}

	if client.APIToken != "token1" {
		t.Errorf("Expected APIToken 'token1', got '%s'", client.APIToken)
	}
}

func TestNewClientWithTokenEmptyUserID(t *testing.T) {
	_, err := NewClientWithToken("", "token1")
	if err == nil {
		t.Error("Expected error for empty user ID, got nil")
	}
}

func TestNewClientWithTokenEmptyToken(t *testing.T) {
	_, err := NewClientWithToken("user1", "")
	if err == nil {
		t.Error("Expected error for empty token, got nil")
	}
}

func TestNewClientEmptyEmail(t *testing.T) {
	_, err := NewClient("", "password")
	if err == nil {
		t.Error("Expected error for empty email, got nil")
	}
}

func TestNewClientEmptyPassword(t *testing.T) {
	_, err := NewClient("email@example.com", "")
	if err == nil {
		t.Error("Expected error for empty password, got nil")
	}
}

func TestNewClientLoginSuccess(t *testing.T) {
	mockSession := Session{UserID: "user123", APIToken: "tok456"}
	mockResponseJSON, _ := json.Marshal(mockSession)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/sessions" {
			t.Errorf("Expected path /api/sessions, got %s", r.URL.Path)
		}

		var reqBody SessionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.Email != "test@example.com" {
			t.Errorf("Expected email 'test@example.com', got '%s'", reqBody.Email)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(mockResponseJSON)
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClient("test@example.com", "password123")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.UserID != "user123" {
		t.Errorf("Expected UserID 'user123', got '%s'", client.UserID)
	}

	if client.APIToken != "tok456" {
		t.Errorf("Expected APIToken 'tok456', got '%s'", client.APIToken)
	}
}

func TestNewClientLoginFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "Invalid credentials"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	_, err := NewClient("bad@example.com", "wrongpassword")
	if err == nil {
		t.Error("Expected error for failed login, got nil")
	}
}

func TestAuthHeader(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			t.Error("Expected Authorization header to be set")
		}

		// Basic base64("user1:token1") = "dXNlcjE6dG9rZW4x"
		expected := "Basic dXNlcjE6dG9rZW4x"
		if authHeader != expected {
			t.Errorf("Expected Authorization '%s', got '%s'", expected, authHeader)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListCategories("frame1")
	if err != nil {
		t.Fatalf("Request with auth header failed: %v", err)
	}
}

func TestPutMethodErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad request"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.UpdateCalendarEvent("frame1", "evt1", CalendarEventData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
}

func TestPatchMethodErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "Bad request"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.UpdateReward("frame1", "r1", RewardData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for bad request, got nil")
	}
}

func TestPutNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// UpdateList with 204 No Content - should succeed without trying to decode
	_, err = client.UpdateList("frame1", "1", ListData{Title: "Test"})
	if err != nil {
		t.Fatalf("UpdateList with 204 failed: %v", err)
	}
}

func TestGetInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetRewardPoints("frame1")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestPostInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.CreateReward("frame1", RewardData{Title: "Test", Points: 10})
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestPutInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.UpdateCalendarEvent("frame1", "evt1", CalendarEventData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestPatchInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.UpdateReward("frame1", "r1", RewardData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestDeleteWithOKStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	err = client.DeleteCalendarEvent("frame1", "evt1")
	if err != nil {
		t.Fatalf("DeleteCalendarEvent with 200 OK failed: %v", err)
	}
}

func TestPatchNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	_, err = client.UpdateReward("frame1", "r1", RewardData{Title: "Test"})
	if err != nil {
		t.Fatalf("UpdateReward with 204 failed: %v", err)
	}
}

func TestPutWithCreatedStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"1","title":"New List"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	list, err := client.UpdateList("frame1", "1", ListData{Title: "New List"})
	if err != nil {
		t.Fatalf("UpdateList with 201 Created failed: %v", err)
	}

	if list.Title != "New List" {
		t.Errorf("Expected title 'New List', got '%s'", list.Title)
	}
}

func TestDecodeJSONError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not json at all`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListCategories("frame1")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestAddQueryParamsEmpty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != "" {
			t.Errorf("Expected no query params, got '%s'", r.URL.RawQuery)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`[]`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// ListCalendarEvents with no params - addQueryParams should not be called
	_, err = client.ListCalendarEvents("frame1", "", "")
	if err != nil {
		t.Fatalf("ListCalendarEvents with empty params failed: %v", err)
	}
}

func TestLoginInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	_, err := NewClient("test@example.com", "password123")
	if err == nil {
		t.Error("Expected error for invalid JSON login response, got nil")
	}
}

func TestGetWithHTTPClientError(t *testing.T) {
	originalURL := SkylightURL
	SkylightURL = "http://localhost:1/api" // port 1 will refuse connection
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.GetFrame("frame1")
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestPostWithHTTPClientError(t *testing.T) {
	originalURL := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.CreateReward("frame1", RewardData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestPutWithHTTPClientError(t *testing.T) {
	originalURL := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.UpdateCalendarEvent("frame1", "evt1", CalendarEventData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestPatchWithHTTPClientError(t *testing.T) {
	originalURL := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.UpdateReward("frame1", "r1", RewardData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestDeleteWithHTTPClientError(t *testing.T) {
	originalURL := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.DeleteCalendarEvent("frame1", "evt1")
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestLoginWithHTTPClientError(t *testing.T) {
	originalURL := SkylightURL
	SkylightURL = "http://localhost:1/api"
	defer func() { SkylightURL = originalURL }()

	_, err := NewClient("test@example.com", "password123")
	if err == nil {
		t.Error("Expected error for connection refused, got nil")
	}
}

func TestPostWithNilResponseTarget(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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

	// RedeemReward passes nil as v to post()
	err = client.RedeemReward("frame1", "r1")
	if err != nil {
		t.Fatalf("RedeemReward failed: %v", err)
	}
}

func TestNewRequestErrorPaths(t *testing.T) {
	originalURL := SkylightURL
	SkylightURL = "://bad"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// GET-based functions
	_, err = client.ListCalendarEvents("frame1", "", "")
	if err == nil {
		t.Error("Expected error for ListCalendarEvents with bad URL")
	}

	_, err = client.ListSourceCalendars("frame1")
	if err == nil {
		t.Error("Expected error for ListSourceCalendars with bad URL")
	}

	_, err = client.ListCategories("frame1")
	if err == nil {
		t.Error("Expected error for ListCategories with bad URL")
	}

	_, err = client.ListChores("frame1", "", "", "")
	if err == nil {
		t.Error("Expected error for ListChores with bad URL")
	}

	_, err = client.GetFrame("frame1")
	if err == nil {
		t.Error("Expected error for GetFrame with bad URL")
	}

	_, err = client.ListDevices("frame1")
	if err == nil {
		t.Error("Expected error for ListDevices with bad URL")
	}

	_, err = client.GetAvatars()
	if err == nil {
		t.Error("Expected error for GetAvatars with bad URL")
	}

	_, err = client.GetColors()
	if err == nil {
		t.Error("Expected error for GetColors with bad URL")
	}

	_, err = client.ListLists("frame1")
	if err == nil {
		t.Error("Expected error for ListLists with bad URL")
	}

	_, err = client.GetList("frame1", "1")
	if err == nil {
		t.Error("Expected error for GetList with bad URL")
	}

	_, err = client.ListRewards("frame1")
	if err == nil {
		t.Error("Expected error for ListRewards with bad URL")
	}

	_, err = client.GetRewardPoints("frame1")
	if err == nil {
		t.Error("Expected error for GetRewardPoints with bad URL")
	}

	_, err = client.ListRecipes("frame1")
	if err == nil {
		t.Error("Expected error for ListRecipes with bad URL")
	}

	_, err = client.GetRecipe("frame1", "1")
	if err == nil {
		t.Error("Expected error for GetRecipe with bad URL")
	}

	_, err = client.ListMealCategories("frame1")
	if err == nil {
		t.Error("Expected error for ListMealCategories with bad URL")
	}

	_, err = client.ListMealSittings("frame1")
	if err == nil {
		t.Error("Expected error for ListMealSittings with bad URL")
	}

	// DELETE-based functions
	err = client.DeleteCalendarEvent("frame1", "evt1")
	if err == nil {
		t.Error("Expected error for DeleteCalendarEvent with bad URL")
	}

	err = client.DeleteChore("frame1", "chore1")
	if err == nil {
		t.Error("Expected error for DeleteChore with bad URL")
	}

	err = client.DeleteList("frame1", "1")
	if err == nil {
		t.Error("Expected error for DeleteList with bad URL")
	}

	err = client.DeleteListItem("frame1", "1", "item1")
	if err == nil {
		t.Error("Expected error for DeleteListItem with bad URL")
	}

	err = client.DeleteReward("frame1", "r1")
	if err == nil {
		t.Error("Expected error for DeleteReward with bad URL")
	}

	err = client.DeleteRecipe("frame1", "1")
	if err == nil {
		t.Error("Expected error for DeleteRecipe with bad URL")
	}

	// POST-based functions with nil response (using newRequest, not newRequestWithBody)
	err = client.RedeemReward("frame1", "r1")
	if err == nil {
		t.Error("Expected error for RedeemReward with bad URL")
	}

	err = client.UnredeemReward("frame1", "r1")
	if err == nil {
		t.Error("Expected error for UnredeemReward with bad URL")
	}

	err = client.AddRecipeToGroceryList("frame1", "recipe1")
	if err == nil {
		t.Error("Expected error for AddRecipeToGroceryList with bad URL")
	}

	// POST/PUT/PATCH functions using newRequestWithBody
	_, err = client.CreateCalendarEvent("frame1", CalendarEventData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for CreateCalendarEvent with bad URL")
	}

	_, err = client.UpdateCalendarEvent("frame1", "evt1", CalendarEventData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for UpdateCalendarEvent with bad URL")
	}

	_, err = client.CreateChore("frame1", ChoreData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for CreateChore with bad URL")
	}

	_, err = client.UpdateChore("frame1", "chore1", ChoreData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for UpdateChore with bad URL")
	}

	_, err = client.CreateList("frame1", ListData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for CreateList with bad URL")
	}

	_, err = client.UpdateList("frame1", "1", ListData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for UpdateList with bad URL")
	}

	_, err = client.AddListItem("frame1", "1", ListItemData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for AddListItem with bad URL")
	}

	_, err = client.UpdateListItem("frame1", "1", "item1", ListItemData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for UpdateListItem with bad URL")
	}

	_, err = client.CreateTaskBoxItem("frame1", TaskBoxItemData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for CreateTaskBoxItem with bad URL")
	}

	_, err = client.CreateReward("frame1", RewardData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for CreateReward with bad URL")
	}

	_, err = client.UpdateReward("frame1", "r1", RewardData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for UpdateReward with bad URL")
	}

	_, err = client.CreateRecipe("frame1", RecipeData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for CreateRecipe with bad URL")
	}

	_, err = client.UpdateRecipe("frame1", "1", RecipeData{Title: "Test"})
	if err == nil {
		t.Error("Expected error for UpdateRecipe with bad URL")
	}

	_, err = client.CreateMealSitting("frame1", MealSittingData{RecipeID: "r1"})
	if err == nil {
		t.Error("Expected error for CreateMealSitting with bad URL")
	}
}

func TestLoginNewRequestError(t *testing.T) {
	originalURL := SkylightURL
	SkylightURL = "://bad"
	defer func() { SkylightURL = originalURL }()

	_, err := NewClient("test@example.com", "password123")
	if err == nil {
		t.Error("Expected error for Login with bad URL")
	}
}

func TestLoginRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		var reqBody SessionRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.Email != "user@test.com" {
			t.Errorf("Expected email 'user@test.com', got '%s'", reqBody.Email)
		}
		if reqBody.Password != "mypassword" {
			t.Errorf("Expected password 'mypassword', got '%s'", reqBody.Password)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"user_id":"u1","api_token":"t1"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClient("user@test.com", "mypassword")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.UserID != "u1" {
		t.Errorf("Expected UserID 'u1', got '%s'", client.UserID)
	}
}
