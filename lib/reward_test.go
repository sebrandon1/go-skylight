package lib

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListRewards(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		response   string
		wantLen    int
		wantPoints int
		wantCatID  string
		wantEmoji  string
		wantErr    bool
	}{
		{
			name:       "returns rewards",
			status:     http.StatusOK,
			response:   `{"data":[{"id":"1","attributes":{"name":"Ice cream","point_value":10}},{"id":"2","attributes":{"name":"Movie night","point_value":20}}]}`,
			wantLen:    2,
			wantPoints: 10,
		},
		{
			name:      "flattens category relationship",
			status:    http.StatusOK,
			response:  `{"data":[{"id":"1","attributes":{"name":"Ice cream","point_value":10,"emoji_icon":"🍦"},"relationships":{"category":{"data":{"id":"cat456","type":"category"}}}}]}`,
			wantLen:   1,
			wantCatID: "cat456",
			wantEmoji: "🍦",
		},
		{
			name:    "not found returns error",
			status:  http.StatusNotFound,
			wantErr: true,
		},
		{
			name:     "invalid JSON returns error",
			status:   http.StatusOK,
			response: `not valid json`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/rewards" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			rewards, err := client.ListRewards(context.Background(), "frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(rewards) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(rewards))
			}
			if len(rewards) > 0 {
				if tc.wantPoints != 0 && rewards[0].Points != tc.wantPoints {
					t.Errorf("Points: want %d got %d", tc.wantPoints, rewards[0].Points)
				}
				if tc.wantCatID != "" && rewards[0].CategoryID != tc.wantCatID {
					t.Errorf("CategoryID: want %q got %q", tc.wantCatID, rewards[0].CategoryID)
				}
				if tc.wantEmoji != "" && rewards[0].EmojiIcon != tc.wantEmoji {
					t.Errorf("EmojiIcon: want %q got %q", tc.wantEmoji, rewards[0].EmojiIcon)
				}
			}
		})
	}
}

func TestCreateReward(t *testing.T) {
	tests := []struct {
		name      string
		input     RewardData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "creates reward",
			input:     RewardData{Title: "Game time", Points: 15},
			status:    http.StatusCreated,
			response:  `{"data":[{"id":"3","attributes":{"name":"Game time","point_value":15}}]}`,
			wantTitle: "Game time",
		},
		{
			name:      "sends all fields in request body",
			input:     RewardData{Title: "Pizza Night", Points: 50},
			status:    http.StatusCreated,
			response:  `{"data":[{"id":"r1","attributes":{"name":"Pizza Night","point_value":50}}]}`,
			wantTitle: "Pizza Night",
		},
		{
			name:    "server error returns error",
			input:   RewardData{Title: "Test", Points: 10},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:     "empty data array returns error",
			input:    RewardData{Title: "Test", Points: 5},
			status:   http.StatusCreated,
			response: `{"data":[]}`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				var raw map[string]any
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if raw["name"] != tc.input.Title {
					t.Errorf("name: want %q got %v", tc.input.Title, raw["name"])
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			reward, err := client.CreateReward(context.Background(), "frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && reward.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, reward.Title)
			}
		})
	}
}

func TestUpdateReward(t *testing.T) {
	tests := []struct {
		name       string
		rewardID   string
		input      RewardData
		status     int
		response   string
		wantPoints int
		wantErr    bool
	}{
		{
			name:       "updates reward",
			rewardID:   "reward1",
			input:      RewardData{Title: "Updated reward", Points: 25},
			status:     http.StatusOK,
			response:   `{"data":{"id":"1","attributes":{"name":"Updated reward","point_value":25}}}`,
			wantPoints: 25,
		},
		{
			name:     "204 no content succeeds",
			rewardID: "r1",
			input:    RewardData{Title: "Test"},
			status:   http.StatusNoContent,
		},
		{
			name:     "bad request returns error",
			rewardID: "r1",
			input:    RewardData{Title: "Test"},
			status:   http.StatusBadRequest,
			wantErr:  true,
		},
		{
			name:     "server error returns error",
			rewardID: "r1",
			input:    RewardData{Title: "Test"},
			status:   http.StatusInternalServerError,
			wantErr:  true,
		},
		{
			name:     "invalid JSON returns error",
			rewardID: "r1",
			input:    RewardData{Title: "Test"},
			status:   http.StatusOK,
			response: `{invalid json`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			reward, err := client.UpdateReward(context.Background(), "frame1", tc.rewardID, tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && tc.wantPoints != 0 && reward != nil && reward.Points != tc.wantPoints {
				t.Errorf("Points: want %d got %d", tc.wantPoints, reward.Points)
			}
		})
	}
}

func TestDeleteReward(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"deletes with 204", http.StatusNoContent, false},
		{"server error returns error", http.StatusInternalServerError, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/rewards/reward1" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.DeleteReward(context.Background(), "frame1", "reward1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestRedeemReward(t *testing.T) {
	tests := []struct {
		name     string
		rewardID string
		status   int
		wantErr  bool
	}{
		{
			name:     "redeems reward",
			rewardID: "reward1",
			status:   http.StatusOK,
		},
		{
			name:     "not found returns error",
			rewardID: "nonexistent",
			status:   http.StatusNotFound,
			wantErr:  true,
		},
		{
			name:     "server error returns error",
			rewardID: "r1",
			status:   http.StatusInternalServerError,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.RedeemReward(context.Background(), "frame1", tc.rewardID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestUnredeemReward(t *testing.T) {
	tests := []struct {
		name     string
		rewardID string
		status   int
		wantErr  bool
	}{
		{"unredeems reward", "reward1", http.StatusOK, false},
		{"server error returns error", "r1", http.StatusInternalServerError, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.UnredeemReward(context.Background(), "frame1", tc.rewardID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestRemoveStars(t *testing.T) {
	tests := []struct {
		name       string
		categoryID int
		points     int
		status     int
		wantErr    bool
	}{
		{
			name:       "removes stars",
			categoryID: 123,
			points:     5,
			status:     http.StatusOK,
		},
		{
			name:       "not found returns error",
			categoryID: 999,
			points:     1,
			status:     http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "server error returns error",
			categoryID: 1,
			points:     3,
			status:     http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var gotBody removeStarsRequest
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/reward_points" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
					t.Errorf("decode body: %v", err)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.RemoveStars(context.Background(), "frame1", tc.categoryID, tc.points)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr {
				if len(gotBody.CategoryIDs) != 1 || gotBody.CategoryIDs[0] != tc.categoryID {
					t.Errorf("CategoryIDs: want [%d] got %v", tc.categoryID, gotBody.CategoryIDs)
				}
				if gotBody.Points != -tc.points {
					t.Errorf("Points: want %d (negative) got %d", -tc.points, gotBody.Points)
				}
			}
		})
	}
}

func TestGetRewardPoints(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantCat  int
		wantBal  int
		wantErr  bool
	}{
		{
			name:     "returns point entries",
			status:   http.StatusOK,
			response: `[{"category_id":123,"current_point_balance":42}]`,
			wantLen:  1,
			wantCat:  123,
			wantBal:  42,
		},
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:     "invalid JSON returns error",
			status:   http.StatusOK,
			response: `{invalid json`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/frames/frame1/reward_points" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			points, err := client.GetRewardPoints(context.Background(), "frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(points) != tc.wantLen {
				t.Fatalf("wantLen=%d got %d", tc.wantLen, len(points))
			}
			if tc.wantCat != 0 && points[0].CategoryID != tc.wantCat {
				t.Errorf("CategoryID: want %d got %d", tc.wantCat, points[0].CategoryID)
			}
			if tc.wantBal != 0 && points[0].CurrentPointBalance != tc.wantBal {
				t.Errorf("CurrentPointBalance: want %d got %d", tc.wantBal, points[0].CurrentPointBalance)
			}
		})
	}
}
