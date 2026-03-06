package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListChores(t *testing.T) {
	mockChores := []Chore{
		{ID: "1", Title: "Clean room", Status: "pending"},
		{ID: "2", Title: "Do homework", Status: "completed"},
	}

	mockResponseJSON, _ := json.Marshal(mockChores)

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

	chores, err := client.ListChores("frame1", "", "", "")
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
	mockChore := Chore{ID: "3", Title: "Walk dog", Points: 5}

	mockResponseJSON, _ := json.Marshal(mockChore)

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

	_, err = client.ListChores("frame1", "2024-01-15", "pending", "cat1")
	if err != nil {
		t.Fatalf("ListChores with filters failed: %v", err)
	}
}

func TestUpdateChore(t *testing.T) {
	mockChore := Chore{ID: "1", Title: "Updated chore", Status: "completed"}

	mockResponseJSON, _ := json.Marshal(mockChore)

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

	_, err = client.ListChores("frame1", "", "", "")
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

	_, err = client.ListChores("frame1", "", "", "")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestCreateChoreRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody ChoreRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.Chore.Title != "Walk the dog" {
			t.Errorf("Expected title 'Walk the dog', got '%s'", reqBody.Chore.Title)
		}
		if reqBody.Chore.DueDate != "2024-01-15" {
			t.Errorf("Expected due_date '2024-01-15', got '%s'", reqBody.Chore.DueDate)
		}
		if reqBody.Chore.Points != 10 {
			t.Errorf("Expected points 10, got %d", reqBody.Chore.Points)
		}
		if reqBody.Chore.AssigneeID != "cat1" {
			t.Errorf("Expected assignee_id 'cat1', got '%s'", reqBody.Chore.AssigneeID)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"chore1","title":"Walk the dog"}`))
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

	_, err = client.ListChores("frame1", "2024-01-15", "", "")
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

	_, err = client.ListChores("frame1", "", "completed", "")
	if err != nil {
		t.Fatalf("ListChores with status filter only failed: %v", err)
	}
}
