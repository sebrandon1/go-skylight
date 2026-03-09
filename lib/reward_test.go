package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRewards(t *testing.T) {
	mockResp := rewardAPIResponse{
		Data: []rewardAPIEntry{
			{
				ID: "1",
				Attributes: struct {
					Name                string  `json:"name"`
					EmojiIcon           string  `json:"emoji_icon"`
					PointValue          int     `json:"point_value"`
					RespawnOnRedemption bool    `json:"respawn_on_redemption"`
					RedeemedAt          *string `json:"redeemed_at"`
				}{Name: "Ice cream", PointValue: 10},
			},
			{
				ID: "2",
				Attributes: struct {
					Name                string  `json:"name"`
					EmojiIcon           string  `json:"emoji_icon"`
					PointValue          int     `json:"point_value"`
					RespawnOnRedemption bool    `json:"respawn_on_redemption"`
					RedeemedAt          *string `json:"redeemed_at"`
				}{Name: "Movie night", PointValue: 20},
			},
		},
	}

	mockResponseJSON, _ := json.Marshal(mockResp)

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
	mockResp := rewardAPISingleResponse{
		Data: rewardAPIEntry{
			ID: "3",
			Attributes: struct {
				Name                string  `json:"name"`
				EmojiIcon           string  `json:"emoji_icon"`
				PointValue          int     `json:"point_value"`
				RespawnOnRedemption bool    `json:"respawn_on_redemption"`
				RedeemedAt          *string `json:"redeemed_at"`
			}{Name: "Game time", PointValue: 15},
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
	mockPoints := []RewardPointEntry{{CategoryID: 123, CurrentPointBalance: 42}}

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

	if len(points) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(points))
	}
	if points[0].CategoryID != 123 {
		t.Errorf("Expected category_id 123, got %d", points[0].CategoryID)
	}
	if points[0].CurrentPointBalance != 42 {
		t.Errorf("Expected 42 points, got %d", points[0].CurrentPointBalance)
	}
}

func TestUpdateReward(t *testing.T) {
	mockResp := rewardAPISingleResponse{
		Data: rewardAPIEntry{
			ID: "1",
			Attributes: struct {
				Name                string  `json:"name"`
				EmojiIcon           string  `json:"emoji_icon"`
				PointValue          int     `json:"point_value"`
				RespawnOnRedemption bool    `json:"respawn_on_redemption"`
				RedeemedAt          *string `json:"redeemed_at"`
			}{Name: "Updated reward", PointValue: 25},
		},
	}

	mockResponseJSON, _ := json.Marshal(mockResp)

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
		var raw map[string]any
		if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if raw["name"] != "Pizza Night" {
			t.Errorf("Expected name 'Pizza Night', got '%v'", raw["name"])
		}
		if raw["point_value"] != float64(50) {
			t.Errorf("Expected point_value 50, got %v", raw["point_value"])
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"data":{"id":"r1","attributes":{"name":"Pizza Night","point_value":50}}}`))
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

func TestListRewardsWithCategoryRelationship(t *testing.T) {
	mockJSON := `{"data":[{"id":"1","attributes":{"name":"Ice cream","point_value":10,"emoji_icon":"🍦"},"relationships":{"category":{"data":{"id":"cat456","type":"category"}}}}]}`

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

	rewards, err := client.ListRewards("frame1")
	if err != nil {
		t.Fatalf("ListRewards failed: %v", err)
	}

	if len(rewards) != 1 {
		t.Fatalf("Expected 1 reward, got %d", len(rewards))
	}
	if rewards[0].CategoryID != "cat456" {
		t.Errorf("Expected category_id 'cat456', got '%s'", rewards[0].CategoryID)
	}
	if rewards[0].EmojiIcon != "🍦" {
		t.Errorf("Expected emoji_icon '🍦', got '%s'", rewards[0].EmojiIcon)
	}
}
