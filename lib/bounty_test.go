package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCreateBounty(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/api/frames/frame1/chores":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(choreAPISingleResponse{
				Data: choreAPIEntry{
					ID: "ch1",
					Attributes: struct {
						Summary      string `json:"summary"`
						Status       string `json:"status"`
						Start        string `json:"start"`
						RewardPoints int    `json:"reward_points"`
						Recurring    bool   `json:"recurring"`
					}{Summary: "Do dishes", RewardPoints: 10},
				},
			})
		case r.Method == "POST" && r.URL.Path == "/api/frames/frame1/rewards":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(rewardAPISingleResponse{
				Data: rewardAPIEntry{
					ID: "rw1",
					Attributes: struct {
						Name                string  `json:"name"`
						EmojiIcon           string  `json:"emoji_icon"`
						PointValue          int     `json:"point_value"`
						RespawnOnRedemption bool    `json:"respawn_on_redemption"`
						RedeemedAt          *string `json:"redeemed_at"`
					}{Name: "Ice cream", PointValue: 10, EmojiIcon: "🍦"},
				},
			})
		default:
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	bounty, err := client.CreateBounty("frame1", BountyData{
		Title:       "Do dishes",
		Points:      10,
		RewardTitle: "Ice cream",
		EmojiIcon:   "🍦",
	})
	if err != nil {
		t.Fatalf("CreateBounty failed: %v", err)
	}

	if bounty.Chore.Title != "Do dishes" {
		t.Errorf("Expected chore title 'Do dishes', got '%s'", bounty.Chore.Title)
	}
	if bounty.Reward.Title != "Ice cream" {
		t.Errorf("Expected reward title 'Ice cream', got '%s'", bounty.Reward.Title)
	}
	if bounty.Chore.Points != 10 || bounty.Reward.Points != 10 {
		t.Errorf("Expected points 10, got chore=%d reward=%d", bounty.Chore.Points, bounty.Reward.Points)
	}
}

func TestCreateBountyCleanupOnRewardFailure(t *testing.T) {
	var deleteChoreCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == "POST" && r.URL.Path == "/api/frames/frame1/chores":
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(choreAPISingleResponse{
				Data: choreAPIEntry{
					ID: "ch1",
					Attributes: struct {
						Summary      string `json:"summary"`
						Status       string `json:"status"`
						Start        string `json:"start"`
						RewardPoints int    `json:"reward_points"`
						Recurring    bool   `json:"recurring"`
					}{Summary: "Test"},
				},
			})
		case r.Method == "POST" && r.URL.Path == "/api/frames/frame1/rewards":
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":"fail"}`))
		case r.Method == "DELETE" && r.URL.Path == "/api/frames/frame1/chores/ch1":
			deleteChoreCount.Add(1)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("Unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	_, err = client.CreateBounty("frame1", BountyData{
		Title:       "Test",
		Points:      5,
		RewardTitle: "Prize",
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if deleteChoreCount.Load() != 1 {
		t.Errorf("Expected 1 delete chore call for cleanup, got %d", deleteChoreCount.Load())
	}
}

func TestListBounties(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		switch r.URL.Path {
		case "/api/frames/frame1/chores":
			json.NewEncoder(w).Encode(choreAPIResponse{
				Data: []choreAPIEntry{
					{ID: "ch1", Attributes: struct {
						Summary      string `json:"summary"`
						Status       string `json:"status"`
						Start        string `json:"start"`
						RewardPoints int    `json:"reward_points"`
						Recurring    bool   `json:"recurring"`
					}{Summary: "Task A", Status: "pending", RewardPoints: 10}},
					{ID: "ch2", Attributes: struct {
						Summary      string `json:"summary"`
						Status       string `json:"status"`
						Start        string `json:"start"`
						RewardPoints int    `json:"reward_points"`
						Recurring    bool   `json:"recurring"`
					}{Summary: "Task B", Status: "pending", RewardPoints: 0}},
				},
			})
		case "/api/frames/frame1/rewards":
			json.NewEncoder(w).Encode(rewardAPIResponse{
				Data: []rewardAPIEntry{
					{ID: "rw1", Attributes: struct {
						Name                string  `json:"name"`
						EmojiIcon           string  `json:"emoji_icon"`
						PointValue          int     `json:"point_value"`
						RespawnOnRedemption bool    `json:"respawn_on_redemption"`
						RedeemedAt          *string `json:"redeemed_at"`
					}{Name: "Prize", PointValue: 10}},
				},
			})
		default:
			t.Errorf("Unexpected path: %s", r.URL.Path)
		}
	}))
	defer server.Close()

	originalURL := SkylightURL
	SkylightURL = server.URL + "/api"
	defer func() { SkylightURL = originalURL }()

	client, err := NewClientWithToken("user1", "token1")
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	bounties, err := client.ListBounties("frame1")
	if err != nil {
		t.Fatalf("ListBounties failed: %v", err)
	}

	if len(bounties) != 1 {
		t.Fatalf("Expected 1 bounty (matched by points), got %d", len(bounties))
	}
	if bounties[0].Chore.Title != "Task A" {
		t.Errorf("Expected chore 'Task A', got '%s'", bounties[0].Chore.Title)
	}
	if bounties[0].Reward.Title != "Prize" {
		t.Errorf("Expected reward 'Prize', got '%s'", bounties[0].Reward.Title)
	}
}
