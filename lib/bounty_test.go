package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

func TestCreateBounty(t *testing.T) {
	tests := []struct {
		name        string
		input       BountyData
		handler     http.HandlerFunc
		wantChore   string
		wantReward  string
		wantPoints  int
		wantDeleteN int32
		wantErr     bool
	}{
		{
			name:  "creates chore and reward",
			input: BountyData{Title: "Do dishes", Points: 10, RewardTitle: "Ice cream", EmojiIcon: "🍦"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				switch {
				case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/chores":
					w.WriteHeader(http.StatusCreated)
					if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
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
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/rewards":
					w.WriteHeader(http.StatusCreated)
					if err := json.NewEncoder(w).Encode(rewardAPIResponse{
						Data: []rewardAPIEntry{{
							ID: "rw1",
							Attributes: struct {
								Name                string  `json:"name"`
								EmojiIcon           string  `json:"emoji_icon"`
								PointValue          int     `json:"point_value"`
								RespawnOnRedemption bool    `json:"respawn_on_redemption"`
								RedeemedAt          *string `json:"redeemed_at"`
							}{Name: "Ice cream", PointValue: 10, EmojiIcon: "🍦"},
						}},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
			},
			wantChore:  "Do dishes",
			wantReward: "Ice cream",
			wantPoints: 10,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			bounty, err := client.CreateBounty("frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if bounty.Chore.Title != tc.wantChore {
				t.Errorf("Chore.Title: want %q got %q", tc.wantChore, bounty.Chore.Title)
			}
			if bounty.Reward.Title != tc.wantReward {
				t.Errorf("Reward.Title: want %q got %q", tc.wantReward, bounty.Reward.Title)
			}
			if bounty.Chore.Points != tc.wantPoints || bounty.Reward.Points != tc.wantPoints {
				t.Errorf("Points: want %d, chore=%d reward=%d", tc.wantPoints, bounty.Chore.Points, bounty.Reward.Points)
			}
		})
	}
}

func TestCreateBountyCleanupOnRewardFailure(t *testing.T) {
	var deleteChoreCount atomic.Int32

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/chores":
			w.WriteHeader(http.StatusCreated)
			if err := json.NewEncoder(w).Encode(choreAPISingleResponse{
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
			}); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/frames/frame1/rewards":
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
				t.Errorf("write: %v", err)
			}
		case r.Method == http.MethodDelete && r.URL.Path == "/api/frames/frame1/chores/ch1":
			deleteChoreCount.Add(1)
			w.WriteHeader(http.StatusNoContent)
		default:
			t.Errorf("unexpected request: %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.CreateBounty("frame1", BountyData{Title: "Test", Points: 5, RewardTitle: "Prize"})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if deleteChoreCount.Load() != 1 {
		t.Errorf("expected 1 cleanup DELETE, got %d", deleteChoreCount.Load())
	}
}

func TestListBounties(t *testing.T) {
	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantLen    int
		wantChore  string
		wantReward string
		wantErr    bool
	}{
		{
			name: "matches chores to rewards by points",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				switch r.URL.Path {
				case "/api/frames/frame1/chores":
					if err := json.NewEncoder(w).Encode(choreAPIResponse{
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
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				case "/api/frames/frame1/rewards":
					if err := json.NewEncoder(w).Encode(rewardAPIResponse{
						Data: []rewardAPIEntry{{ID: "rw1", Attributes: struct {
							Name                string  `json:"name"`
							EmojiIcon           string  `json:"emoji_icon"`
							PointValue          int     `json:"point_value"`
							RespawnOnRedemption bool    `json:"respawn_on_redemption"`
							RedeemedAt          *string `json:"redeemed_at"`
						}{Name: "Prize", PointValue: 10}}},
					}); err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
			},
			wantLen:    1,
			wantChore:  "Task A",
			wantReward: "Prize",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			bounties, err := client.ListBounties("frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(bounties) != tc.wantLen {
				t.Fatalf("wantLen=%d got %d", tc.wantLen, len(bounties))
			}
			if tc.wantLen > 0 {
				if bounties[0].Chore.Title != tc.wantChore {
					t.Errorf("Chore.Title: want %q got %q", tc.wantChore, bounties[0].Chore.Title)
				}
				if bounties[0].Reward.Title != tc.wantReward {
					t.Errorf("Reward.Title: want %q got %q", tc.wantReward, bounties[0].Reward.Title)
				}
			}
		})
	}
}
