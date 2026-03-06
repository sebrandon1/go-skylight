package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRewards(t *testing.T) {
	mockRewards := []Reward{
		{ID: "1", Title: "Ice cream", Points: 10},
		{ID: "2", Title: "Movie night", Points: 20},
	}

	mockResponseJSON, _ := json.Marshal(mockRewards)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/rewards" {
			t.Errorf("Expected path /api/frames/frame1/rewards, got %s", r.URL.Path)
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

	rewards, err := client.ListRewards("frame1")
	if err != nil {
		t.Fatalf("ListRewards failed: %v", err)
	}

	if len(rewards) != 2 {
		t.Errorf("Expected 2 rewards, got %d", len(rewards))
	}

	if rewards[0].Points != 10 {
		t.Errorf("Expected 10 points, got %d", rewards[0].Points)
	}
}

func TestCreateReward(t *testing.T) {
	mockReward := Reward{ID: "3", Title: "Game time", Points: 15}

	mockResponseJSON, _ := json.Marshal(mockReward)

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

	reward, err := client.CreateReward("frame1", RewardData{Title: "Game time", Points: 15})
	if err != nil {
		t.Fatalf("CreateReward failed: %v", err)
	}

	if reward.Title != "Game time" {
		t.Errorf("Expected title 'Game time', got '%s'", reward.Title)
	}
}

func TestRedeemReward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/rewards/reward1/redeem" {
			t.Errorf("Expected path /api/frames/frame1/rewards/reward1/redeem, got %s", r.URL.Path)
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

	err = client.RedeemReward("frame1", "reward1")
	if err != nil {
		t.Fatalf("RedeemReward failed: %v", err)
	}
}

func TestGetRewardPoints(t *testing.T) {
	mockPoints := RewardPoints{Points: 42}

	mockResponseJSON, _ := json.Marshal(mockPoints)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/reward_points" {
			t.Errorf("Expected path /api/frames/frame1/reward_points, got %s", r.URL.Path)
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

	points, err := client.GetRewardPoints("frame1")
	if err != nil {
		t.Fatalf("GetRewardPoints failed: %v", err)
	}

	if points.Points != 42 {
		t.Errorf("Expected 42 points, got %d", points.Points)
	}
}

func TestUpdateReward(t *testing.T) {
	mockReward := Reward{ID: "1", Title: "Updated reward", Points: 25}

	mockResponseJSON, _ := json.Marshal(mockReward)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/rewards/reward1" {
			t.Errorf("Expected path /api/frames/frame1/rewards/reward1, got %s", r.URL.Path)
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

	reward, err := client.UpdateReward("frame1", "reward1", RewardData{Title: "Updated reward", Points: 25})
	if err != nil {
		t.Fatalf("UpdateReward failed: %v", err)
	}

	if reward.Points != 25 {
		t.Errorf("Expected 25 points, got %d", reward.Points)
	}
}

func TestDeleteReward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/rewards/reward1" {
			t.Errorf("Expected path /api/frames/frame1/rewards/reward1, got %s", r.URL.Path)
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

	err = client.DeleteReward("frame1", "reward1")
	if err != nil {
		t.Fatalf("DeleteReward failed: %v", err)
	}
}

func TestUnredeemReward(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/rewards/reward1/unredeem" {
			t.Errorf("Expected path /api/frames/frame1/rewards/reward1/unredeem, got %s", r.URL.Path)
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

	err = client.UnredeemReward("frame1", "reward1")
	if err != nil {
		t.Fatalf("UnredeemReward failed: %v", err)
	}
}

func TestRewardErrorHandling(t *testing.T) {
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

	_, err = client.ListRewards("frame1")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	err = client.RedeemReward("frame1", "nonexistent")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestCreateRewardError(t *testing.T) {
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

	_, err = client.CreateReward("frame1", RewardData{Title: "Test", Points: 10})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUpdateRewardError(t *testing.T) {
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

	_, err = client.UpdateReward("frame1", "r1", RewardData{Title: "Test"})
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestDeleteRewardError(t *testing.T) {
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

	err = client.DeleteReward("frame1", "r1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestUnredeemRewardError(t *testing.T) {
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

	err = client.UnredeemReward("frame1", "r1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestGetRewardPointsError(t *testing.T) {
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

	_, err = client.GetRewardPoints("frame1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}

func TestRewardInvalidJSONResponse(t *testing.T) {
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

	_, err = client.ListRewards("frame1")
	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
}

func TestCreateRewardRequestBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var reqBody RewardRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.Reward.Title != "Pizza Night" {
			t.Errorf("Expected title 'Pizza Night', got '%s'", reqBody.Reward.Title)
		}
		if reqBody.Reward.Points != 50 {
			t.Errorf("Expected points 50, got %d", reqBody.Reward.Points)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":"r1","title":"Pizza Night","points":50}`))
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.CreateReward("frame1", RewardData{Title: "Pizza Night", Points: 50})
	if err != nil {
		t.Fatalf("CreateReward failed: %v", err)
	}
}

func TestRedeemRewardError(t *testing.T) {
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

	err = client.RedeemReward("frame1", "r1")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
