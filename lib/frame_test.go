package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetFrame(t *testing.T) {
	mockFrame := Frame{ID: "frame1", Name: "Family Frame", TimeZone: "America/Chicago"}

	mockResponseJSON, _ := json.Marshal(mockFrame)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1" {
			t.Errorf("Expected path /api/frames/frame1, got %s", r.URL.Path)
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

	frame, err := client.GetFrame("frame1")
	if err != nil {
		t.Fatalf("GetFrame failed: %v", err)
	}

	if frame.Name != "Family Frame" {
		t.Errorf("Expected name 'Family Frame', got '%s'", frame.Name)
	}

	if frame.TimeZone != "America/Chicago" {
		t.Errorf("Expected timezone 'America/Chicago', got '%s'", frame.TimeZone)
	}
}

func TestListDevices(t *testing.T) {
	mockDevices := []Device{
		{ID: "dev1", Name: "Kitchen", Online: true},
		{ID: "dev2", Name: "Living Room", Online: false},
	}

	mockResponseJSON, _ := json.Marshal(mockDevices)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/devices" {
			t.Errorf("Expected path /api/frames/frame1/devices, got %s", r.URL.Path)
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

	devices, err := client.ListDevices("frame1")
	if err != nil {
		t.Fatalf("ListDevices failed: %v", err)
	}

	if len(devices) != 2 {
		t.Errorf("Expected 2 devices, got %d", len(devices))
	}

	if !devices[0].Online {
		t.Error("Expected first device to be online")
	}
}

func TestGetAvatars(t *testing.T) {
	mockAvatars := []Avatar{
		{ID: "1", Name: "Cat"},
		{ID: "2", Name: "Dog"},
	}

	mockResponseJSON, _ := json.Marshal(mockAvatars)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/avatars" {
			t.Errorf("Expected path /api/avatars, got %s", r.URL.Path)
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

	avatars, err := client.GetAvatars()
	if err != nil {
		t.Fatalf("GetAvatars failed: %v", err)
	}

	if len(avatars) != 2 {
		t.Errorf("Expected 2 avatars, got %d", len(avatars))
	}
}

func TestGetColors(t *testing.T) {
	mockColors := []Color{
		{ID: "1", Name: "Red", Value: "#FF0000"},
		{ID: "2", Name: "Blue", Value: "#0000FF"},
	}

	mockResponseJSON, _ := json.Marshal(mockColors)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/colors" {
			t.Errorf("Expected path /api/colors, got %s", r.URL.Path)
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

	colors, err := client.GetColors()
	if err != nil {
		t.Fatalf("GetColors failed: %v", err)
	}

	if len(colors) != 2 {
		t.Errorf("Expected 2 colors, got %d", len(colors))
	}

	if colors[0].Value != "#FF0000" {
		t.Errorf("Expected value '#FF0000', got '%s'", colors[0].Value)
	}
}

func TestFrameErrorHandling(t *testing.T) {
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

	_, err = client.GetFrame("nonexistent")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	_, err = client.ListDevices("nonexistent")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
