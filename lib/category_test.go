package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCategories(t *testing.T) {
	mockCategories := []Category{
		{ID: "1", Name: "Mom", Color: "#FF0000"},
		{ID: "2", Name: "Dad", Color: "#0000FF"},
	}

	mockResponseJSON, _ := json.Marshal(mockCategories)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/categories" {
			t.Errorf("Expected path /api/frames/frame1/categories, got %s", r.URL.Path)
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

	categories, err := client.ListCategories("frame1")
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}

	if len(categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(categories))
	}

	if categories[0].Name != "Mom" {
		t.Errorf("Expected name 'Mom', got '%s'", categories[0].Name)
	}
}

func TestCategoryErrorHandling(t *testing.T) {
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

	_, err = client.ListCategories("frame1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
