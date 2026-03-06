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
