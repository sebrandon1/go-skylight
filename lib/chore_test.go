package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListChores(t *testing.T) {
	mockResp := choreAPIResponse{
		Data: []choreAPIEntry{
			{
				ID: "1",
				Attributes: struct {
					Summary      string `json:"summary"`
					Status       string `json:"status"`
					Start        string `json:"start"`
					RewardPoints int    `json:"reward_points"`
					Recurring    bool   `json:"recurring"`
				}{Summary: "Clean room", Status: "pending"},
			},
			{
				ID: "2",
				Attributes: struct {
					Summary      string `json:"summary"`
					Status       string `json:"status"`
					Start        string `json:"start"`
					RewardPoints int    `json:"reward_points"`
					Recurring    bool   `json:"recurring"`
				}{Summary: "Do homework", Status: "completed"},
			},
		},
	}

	mockResponseJSON, _ := json.Marshal(mockResp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/chores" {
			t.Errorf("Expected path /api/frames/frame1/chores, got %s", r.URL.Path)
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

	chores, err := client.ListChores("frame1", ChoreListOptions{})
	if err != nil {
		t.Fatalf("ListChores failed: %v", err)
	}

	if len(chores) != 2 {
		t.Errorf("Expected 2 chores, got %d", len(chores))
	}

	if chores[0].Title != "Clean room" {
		t.Errorf("Expected title 'Clean room', got '%s'", chores[0].Title)
	}
}

func TestCreateChore(t *testing.T) {
	mockResp := choreAPISingleResponse{
		Data: choreAPIEntry{
			ID: "3",
			Attributes: struct {
				Summary      string `json:"summary"`
				Status       string `json:"status"`
				Start        string `json:"start"`
				RewardPoints int    `json:"reward_points"`
				Recurring    bool   `json:"recurring"`
			}{Summary: "Walk dog", RewardPoints: 5},
		},
	}

	mockResponseJSON, _ := json.Marshal(mockResp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
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

	chore, err := client.CreateChore("frame1", ChoreData{Title: "Walk dog", Points: 5})
	if err != nil {
		t.Fatalf("CreateChore failed: %v", err)
	}

	if chore.Title != "Walk dog" {
		t.Errorf("Expected title 'Walk dog', got '%s'", chore.Title)
	}
}

func TestDeleteChore(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
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

	err = client.DeleteChore("frame1", "chore1")
	if err != nil {
		t.Fatalf("DeleteChore failed: %v", err)
	}
}

func TestListChoresWithFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("date") != "2024-01-15" {
			t.Errorf("Expected date query param '2024-01-15', got '%s'", r.URL.Query().Get("date"))
		}
		if r.URL.Query().Get("status") != "pending" {
			t.Errorf("Expected status query param 'pending', got '%s'", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("assignee_id") != "cat1" {
			t.Errorf("Expected assignee_id query param 'cat1', got '%s'", r.URL.Query().Get("assignee_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListChores("frame1", ChoreListOptions{Date: "2024-01-15", Status: "pending", AssigneeID: "cat1"})
	if err != nil {
		t.Fatalf("ListChores with filters failed: %v", err)
	}
}

func TestUpdateChore(t *testing.T) {
	mockResp := choreAPISingleResponse{
		Data: choreAPIEntry{
			ID: "1",
			Attributes: struct {
				Summary      string `json:"summary"`
				Status       string `json:"status"`
				Start        string `json:"start"`
				RewardPoints int    `json:"reward_points"`
				Recurring    bool   `json:"recurring"`
			}{Summary: "Updated chore", Status: "completed"},
		},
	}

	mockResponseJSON, _ := json.Marshal(mockResp)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/chores/chore1" {
			t.Errorf("Expected path /api/frames/frame1/chores/chore1, got %s", r.URL.Path)
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

	chore, err := client.UpdateChore("frame1", "chore1", ChoreData{Title: "Updated chore", Status: "completed"})
	if err != nil {
		t.Fatalf("UpdateChore failed: %v", err)
	}

	if chore.Status != "completed" {
		t.Errorf("Expected status 'completed', got '%s'", chore.Status)
	}
}

func TestChoreErrorHandling(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Server error"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListChores("frame1", ChoreListOptions{})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateChoreError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Server error"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.CreateChore("frame1", ChoreData{Title: "Test"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUpdateChoreError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Server error"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.UpdateChore("frame1", "chore1", ChoreData{Title: "Test"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDeleteChoreError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "Server error"}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	err = client.DeleteChore("frame1", "chore1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestChoreInvalidJSONResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListChores("frame1", ChoreListOptions{})
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestCreateChoreRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var raw map[string]map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		chore := raw["chore"]
		if chore["summary"] != "Walk the dog" {
			t.Errorf("Expected summary 'Walk the dog', got '%v'", chore["summary"])
		}
		if chore["start"] != "2024-01-15" {
			t.Errorf("Expected start '2024-01-15', got '%v'", chore["start"])
		}
		if chore["reward_points"] != float64(10) {
			t.Errorf("Expected reward_points 10, got %v", chore["reward_points"])
		}
		if chore["category_id"] != "cat1" {
			t.Errorf("Expected category_id 'cat1', got '%v'", chore["category_id"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"id":"chore1","attributes":{"summary":"Walk the dog"}}}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.CreateChore("frame1", ChoreData{
		Title:      "Walk the dog",
		DueDate:    "2024-01-15",
		Points:     10,
		AssigneeID: "cat1",
	})
	if err != nil {
		t.Fatalf("CreateChore failed: %v", err)
	}
}

func TestListChoresWithDateFilterOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("date") != "2024-01-15" {
			t.Errorf("Expected date '2024-01-15', got '%s'", r.URL.Query().Get("date"))
		}
		if r.URL.Query().Get("status") != "" {
			t.Errorf("Expected no status param, got '%s'", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("assignee_id") != "" {
			t.Errorf("Expected no assignee_id param, got '%s'", r.URL.Query().Get("assignee_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListChores("frame1", ChoreListOptions{Date: "2024-01-15"})
	if err != nil {
		t.Fatalf("ListChores with date filter only failed: %v", err)
	}
}

func TestListChoresWithStatusFilterOnly(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("date") != "" {
			t.Errorf("Expected no date param, got '%s'", r.URL.Query().Get("date"))
		}
		if r.URL.Query().Get("status") != "completed" {
			t.Errorf("Expected status 'completed', got '%s'", r.URL.Query().Get("status"))
		}
		if r.URL.Query().Get("assignee_id") != "" {
			t.Errorf("Expected no assignee_id param, got '%s'", r.URL.Query().Get("assignee_id"))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data":[]}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.ListChores("frame1", ChoreListOptions{Status: "completed"})
	if err != nil {
		t.Fatalf("ListChores with status filter only failed: %v", err)
	}
}

func TestListChoresWithCategoryRelationship(t *testing.T) {
	mockJSON := `{"data":[{"id":"1","attributes":{"summary":"Clean room","status":"pending","start":"2024-01-15","reward_points":5,"recurring":false},"relationships":{"category":{"data":{"id":"cat123","type":"category"}}}}]}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockJSON))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	chores, err := client.ListChores("frame1", ChoreListOptions{})
	if err != nil {
		t.Fatalf("ListChores failed: %v", err)
	}

	if len(chores) != 1 {
		t.Fatalf("Expected 1 chore, got %d", len(chores))
	}
	if chores[0].AssigneeID != "cat123" {
		t.Errorf("Expected assignee_id 'cat123', got '%s'", chores[0].AssigneeID)
	}
	if chores[0].Title != "Clean room" {
		t.Errorf("Expected title 'Clean room', got '%s'", chores[0].Title)
	}
	if chores[0].Points != 5 {
		t.Errorf("Expected points 5, got %d", chores[0].Points)
	}
}
