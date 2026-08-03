package lib

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClientWithToken(t *testing.T) {
	tests := []struct {
		name    string
		userID  string
		token   string
		wantErr bool
	}{
		{"valid credentials", "user1", "token1", false},
		{"empty user ID", "", "token1", true},
		{"empty token", "user1", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client, err := NewClientWithToken(tc.userID, tc.token)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr {
				if client.UserID != tc.userID {
					t.Errorf("UserID: want %q got %q", tc.userID, client.UserID)
				}
				if client.APIToken != tc.token {
					t.Errorf("APIToken: want %q got %q", tc.token, client.APIToken)
				}
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	tests := []struct {
		name     string
		email    string
		password string
		handler  http.HandlerFunc
		wantUID  string
		wantTok  string
		wantErr  bool
	}{
		{
			name:     "empty email",
			email:    "",
			password: "pw",
			wantErr:  true,
		},
		{
			name:     "empty password",
			email:    "e@example.com",
			password: "",
			wantErr:  true,
		},
		{
			name:     "login success",
			email:    "test@example.com",
			password: "password123",
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/sessions" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var body SessionRequest
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode: %v", err)
				}
				if body.Email != "test@example.com" {
					t.Errorf("email: want test@example.com got %q", body.Email)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`{"data":{"id":"user123","type":"authenticated_user","attributes":{"token":"tok456"}}}`)); err != nil {
					t.Errorf("write: %v", err)
				}
			},
			wantUID: "user123",
			wantTok: "tok456",
		},
		{
			name:     "login failure/unauthorized",
			email:    "bad@example.com",
			password: "wrong",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte(`{"error":"Invalid credentials"}`)); err != nil {
					t.Errorf("write: %v", err)
				}
			},
			wantErr: true,
		},
		{
			name:     "invalid JSON response",
			email:    "test@example.com",
			password: "password123",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`not valid json`)); err != nil {
					t.Errorf("write: %v", err)
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// For cases without a handler, use nil (no server needed)
			if tc.handler == nil {
				_, err := NewClient(tc.email, tc.password)
				if (err != nil) != tc.wantErr {
					t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
				}
				return
			}

			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, err := NewClient(tc.email, tc.password)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr {
				if client.UserID != tc.wantUID {
					t.Errorf("UserID: want %q got %q", tc.wantUID, client.UserID)
				}
				if client.APIToken != tc.wantTok {
					t.Errorf("APIToken: want %q got %q", tc.wantTok, client.APIToken)
				}
				// NewClient (legacy) should still produce Bearer auth cache.
				if client.UserID != tc.wantUID {
					t.Errorf("UserID: want %q got %q", tc.wantUID, client.UserID)
				}
			}
		})
	}
}

func TestAuthorizationHeader(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth != "Bearer token1" {
			t.Errorf("Authorization: want Bearer token1, got %q", auth)
		}
		apiVer := r.Header.Get("skylight-api-version")
		if apiVer != "2026-03-01" {
			t.Errorf("skylight-api-version: want 2026-03-01, got %q", apiVer)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"data":[]}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("user1", "token1")
	_, err := client.ListCategories(context.Background(), "frame1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestHTTPMethodErrorStatus(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*Client) error
	}{
		{
			name: "PUT 400 returns error",
			fn: func(c *Client) error {
				_, err := c.UpdateCalendarEvent(context.Background(), "frame1", "evt1", CalendarEventData{Title: "Test"})
				return err
			},
		},
		{
			name: "PATCH 400 returns error",
			fn: func(c *Client) error {
				_, err := c.UpdateReward(context.Background(), "frame1", "r1", RewardData{Title: "Test"})
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				if _, err := w.Write([]byte(`{"error":"Bad request"}`)); err != nil {
					t.Errorf("write: %v", err)
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			if err := tc.fn(client); err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestStatusCodeHandling(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		body    string
		fn      func(*Client) error
		wantErr bool
	}{
		{
			name:   "PUT 204 no content succeeds",
			status: http.StatusNoContent,
			fn: func(c *Client) error {
				_, err := c.UpdateList(context.Background(), "frame1", "1", ListData{Title: "Test"})
				return err
			},
		},
		{
			name:   "PUT 201 created with body succeeds",
			status: http.StatusCreated,
			body:   `{"data":{"id":"1","attributes":{"label":"New List","color":"","kind":""}}}`,
			fn: func(c *Client) error {
				list, err := c.UpdateList(context.Background(), "frame1", "1", ListData{Title: "New List"})
				if err == nil && list.Title != "New List" {
					t.Errorf("Title: want 'New List' got %q", list.Title)
				}
				return err
			},
		},
		{
			name:   "PATCH 204 no content succeeds",
			status: http.StatusNoContent,
			fn: func(c *Client) error {
				_, err := c.UpdateReward(context.Background(), "frame1", "r1", RewardData{Title: "Test"})
				return err
			},
		},
		{
			name:   "DELETE 200 OK succeeds",
			status: http.StatusOK,
			fn: func(c *Client) error {
				return c.DeleteCalendarEvent(context.Background(), "frame1", "evt1")
			},
		},
		{
			name:   "POST with nil response target succeeds",
			status: http.StatusOK,
			fn: func(c *Client) error {
				return c.RedeemReward(context.Background(), "frame1", "r1")
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.body != "" {
					if _, err := w.Write([]byte(tc.body)); err != nil {
						t.Errorf("write: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := tc.fn(client)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestInvalidJSONResponse(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*Client) error
	}{
		{
			name: "GET invalid JSON",
			fn: func(c *Client) error {
				_, err := c.GetRewardPoints(context.Background(), "frame1")
				return err
			},
		},
		{
			name: "POST invalid JSON",
			fn: func(c *Client) error {
				_, err := c.CreateList(context.Background(), "frame1", ListData{Title: "Test"})
				return err
			},
		},
		{
			name: "PUT invalid JSON",
			fn: func(c *Client) error {
				_, err := c.UpdateCalendarEvent(context.Background(), "frame1", "evt1", CalendarEventData{Title: "Test"})
				return err
			},
		},
		{
			name: "PATCH invalid JSON",
			fn: func(c *Client) error {
				_, err := c.UpdateReward(context.Background(), "frame1", "r1", RewardData{Title: "Test"})
				return err
			},
		},
		{
			name: "GET not-JSON string",
			fn: func(c *Client) error {
				_, err := c.ListCategories(context.Background(), "frame1")
				return err
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				if _, err := w.Write([]byte(`{invalid json`)); err != nil {
					t.Errorf("write: %v", err)
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			if err := tc.fn(client); err == nil {
				t.Error("expected error for invalid JSON, got nil")
			}
		})
	}
}

func TestConnectionRefusedErrors(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*Client) error
	}{
		{"GET", func(c *Client) error { _, err := c.GetFrame(context.Background(), "frame1"); return err }},
		{"POST", func(c *Client) error {
			_, err := c.CreateReward(context.Background(), "frame1", RewardData{Title: "T"})
			return err
		}},
		{"PUT", func(c *Client) error {
			_, err := c.UpdateCalendarEvent(context.Background(), "frame1", "e1", CalendarEventData{Title: "T"})
			return err
		}},
		{"PATCH", func(c *Client) error {
			_, err := c.UpdateReward(context.Background(), "frame1", "r1", RewardData{Title: "T"})
			return err
		}},
		{"DELETE", func(c *Client) error { return c.DeleteCalendarEvent(context.Background(), "frame1", "evt1") }},
		{"Login", func(*Client) error { _, err := NewClient("test@example.com", "pw"); return err }},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			old := SkylightURL
			SkylightURL = "http://localhost:1/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			if err := tc.fn(client); err == nil {
				t.Error("expected error for connection refused, got nil")
			}
		})
	}
}

func TestBadURLNewRequestErrors(t *testing.T) {
	tests := []struct {
		name string
		fn   func(*Client) error
	}{
		// GET-based
		{"ListCalendarEvents", func(c *Client) error {
			_, err := c.ListCalendarEvents(context.Background(), "f", "", "", "")
			return err
		}},
		{"ListSourceCalendars", func(c *Client) error { _, err := c.ListSourceCalendars(context.Background(), "f"); return err }},
		{"ListCategories", func(c *Client) error { _, err := c.ListCategories(context.Background(), "f"); return err }},
		{"ListChores", func(c *Client) error {
			_, err := c.ListChores(context.Background(), "f", ChoreListOptions{})
			return err
		}},
		{"GetFrame", func(c *Client) error { _, err := c.GetFrame(context.Background(), "f"); return err }},
		{"ListDevices", func(c *Client) error { _, err := c.ListDevices(context.Background(), "f"); return err }},
		{"GetAvatars", func(c *Client) error { _, err := c.GetAvatars(context.Background()); return err }},
		{"GetColors", func(c *Client) error { _, err := c.GetColors(context.Background()); return err }},
		{"ListLists", func(c *Client) error { _, err := c.ListLists(context.Background(), "f"); return err }},
		{"GetList", func(c *Client) error { _, err := c.GetList(context.Background(), "f", "1"); return err }},
		{"ListRewards", func(c *Client) error { _, err := c.ListRewards(context.Background(), "f"); return err }},
		{"GetRewardPoints", func(c *Client) error { _, err := c.GetRewardPoints(context.Background(), "f"); return err }},
		{"ListRecipes", func(c *Client) error { _, err := c.ListRecipes(context.Background(), "f"); return err }},
		{"GetRecipe", func(c *Client) error { _, err := c.GetRecipe(context.Background(), "f", "1"); return err }},
		{"ListMealCategories", func(c *Client) error { _, err := c.ListMealCategories(context.Background(), "f"); return err }},
		{"ListMealSittings", func(c *Client) error {
			_, err := c.ListMealSittings(context.Background(), "f", MealSittingListOptions{})
			return err
		}},
		// DELETE-based
		{"DeleteCalendarEvent", func(c *Client) error { return c.DeleteCalendarEvent(context.Background(), "f", "e1") }},
		{"DeleteChore", func(c *Client) error { return c.DeleteChore(context.Background(), "f", "c1") }},
		{"DeleteList", func(c *Client) error { return c.DeleteList(context.Background(), "f", "1") }},
		{"DeleteListItem", func(c *Client) error { return c.DeleteListItem(context.Background(), "f", "1", "i1") }},
		{"DeleteReward", func(c *Client) error { return c.DeleteReward(context.Background(), "f", "r1") }},
		{"DeleteRecipe", func(c *Client) error { return c.DeleteRecipe(context.Background(), "f", "1") }},
		// POST-based (nil response)
		{"RedeemReward", func(c *Client) error { return c.RedeemReward(context.Background(), "f", "r1") }},
		{"UnredeemReward", func(c *Client) error { return c.UnredeemReward(context.Background(), "f", "r1") }},
		{"AddRecipeToGroceryList", func(c *Client) error { return c.AddRecipeToGroceryList(context.Background(), "f", "r1") }},
		// POST/PUT/PATCH with body
		{"CreateCalendarEvent", func(c *Client) error {
			_, err := c.CreateCalendarEvent(context.Background(), "f", CalendarEventData{Title: "T"})
			return err
		}},
		{"UpdateCalendarEvent", func(c *Client) error {
			_, err := c.UpdateCalendarEvent(context.Background(), "f", "e1", CalendarEventData{Title: "T"})
			return err
		}},
		{"CreateChore", func(c *Client) error {
			_, err := c.CreateChore(context.Background(), "f", ChoreData{Title: "T"})
			return err
		}},
		{"UpdateChore", func(c *Client) error {
			_, err := c.UpdateChore(context.Background(), "f", "c1", ChoreData{Title: "T"})
			return err
		}},
		{"CreateList", func(c *Client) error {
			_, err := c.CreateList(context.Background(), "f", ListData{Title: "T"})
			return err
		}},
		{"UpdateList", func(c *Client) error {
			_, err := c.UpdateList(context.Background(), "f", "1", ListData{Title: "T"})
			return err
		}},
		{"AddListItem", func(c *Client) error {
			_, err := c.AddListItem(context.Background(), "f", "1", ListItemData{Title: "T"})
			return err
		}},
		{"UpdateListItem", func(c *Client) error {
			_, err := c.UpdateListItem(context.Background(), "f", "1", "i1", ListItemData{Title: "T"})
			return err
		}},
		{"CreateTaskBoxItem", func(c *Client) error {
			_, err := c.CreateTaskBoxItem(context.Background(), "f", TaskBoxItemData{Title: "T"})
			return err
		}},
		{"CreateReward", func(c *Client) error {
			_, err := c.CreateReward(context.Background(), "f", RewardData{Title: "T"})
			return err
		}},
		{"UpdateReward", func(c *Client) error {
			_, err := c.UpdateReward(context.Background(), "f", "r1", RewardData{Title: "T"})
			return err
		}},
		{"CreateRecipe", func(c *Client) error {
			_, err := c.CreateRecipe(context.Background(), "f", RecipeData{Title: "T"})
			return err
		}},
		{"UpdateRecipe", func(c *Client) error {
			_, err := c.UpdateRecipe(context.Background(), "f", "1", RecipeData{Title: "T"})
			return err
		}},
		{"CreateMealSitting", func(c *Client) error {
			_, err := c.CreateMealSitting(context.Background(), "f", MealSittingData{RecipeID: "r1"})
			return err
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			old := SkylightURL
			SkylightURL = "://bad"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			if err := tc.fn(client); err == nil {
				t.Errorf("expected error with bad URL, got nil")
			}
		})
	}
}

func TestEffectiveURLFallback(t *testing.T) {
	client, _ := NewClientWithToken("u", "t")
	// Client created without WithBaseURL — baseURL is set to default
	// Clear it to test the fallback branch
	client.baseURL = ""
	got := client.effectiveURL()
	if got != SkylightURL {
		t.Errorf("effectiveURL = %q, want %q", got, SkylightURL)
	}
}

func TestDoWithLogger(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(categoryAPIResponse{}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}))
	defer srv.Close()

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	client, _ := NewClientWithToken("u", "t", WithBaseURL(srv.URL+"/api"), WithLogger(logger))

	// Make a request that exercises both logger.Debug branches in do()
	_, err := client.ListCategories(context.Background(), "frame1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDoDeleteErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		if _, err := w.Write([]byte(`{"error":"forbidden"}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	err := client.DeleteChore(context.Background(), "frame1", "ch1")
	if err == nil {
		t.Error("expected error for 403 response, got nil")
	}
}

func TestNewClientWithRefreshToken(t *testing.T) {
	tests := []struct {
		name         string
		refreshToken string
		fingerprint  string
		handler      http.HandlerFunc
		wantErr      bool
		wantToken    string
		wantRefresh  string
	}{
		{
			name:         "empty refresh token returns error",
			refreshToken: "",
			fingerprint:  "fp1",
			wantErr:      true,
		},
		{
			name:         "empty fingerprint returns error",
			refreshToken: "rt1",
			fingerprint:  "",
			wantErr:      true,
		},
		{
			name:         "refresh token auth success",
			refreshToken: "rt1",
			fingerprint:  "fp1",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				if _, err := w.Write([]byte(`{"access_token":"at1","refresh_token":"rt2","expires_in":3600}`)); err != nil {
					t.Errorf("write: %v", err)
				}
			},
			wantToken:   "at1",
			wantRefresh: "rt2",
		},
		{
			name:         "refresh token auth failure",
			refreshToken: "rt1",
			fingerprint:  "fp1",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusUnauthorized)
				if _, err := w.Write([]byte(`{"error":"invalid_grant"}`)); err != nil {
					t.Errorf("write: %v", err)
				}
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.handler == nil {
				client, err := NewClientWithRefreshToken(tc.refreshToken, tc.fingerprint)
				if (err != nil) != tc.wantErr {
					t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
				}
				if !tc.wantErr && client == nil {
					t.Fatal("expected non-nil client")
				}
				return
			}

			srv := httptest.NewServer(tc.handler)
			defer srv.Close()

			old := OAuthURL
			OAuthURL = srv.URL
			defer func() { OAuthURL = old }()

			client, err := NewClientWithRefreshToken(tc.refreshToken, tc.fingerprint)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if client.APIToken != tc.wantToken {
				t.Errorf("APIToken: want %q got %q", tc.wantToken, client.APIToken)
			}
			if client.RefreshToken != tc.wantRefresh {
				t.Errorf("RefreshToken: want %q got %q", tc.wantRefresh, client.RefreshToken)
			}
		})
	}
}

func TestLoginBadURL(t *testing.T) {
	old := SkylightURL
	SkylightURL = "://bad"
	defer func() { SkylightURL = old }()

	_, err := NewClient("test@example.com", "password123")
	if err == nil {
		t.Error("expected error for Login with bad URL, got nil")
	}
}

func TestLoginRequestBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var body SessionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		if body.Email != "user@test.com" {
			t.Errorf("email: want user@test.com got %q", body.Email)
		}
		if body.Password != "mypassword" {
			t.Errorf("password: want mypassword got %q", body.Password)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"data":{"id":"u1","type":"authenticated_user","attributes":{"token":"t1"}}}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, err := NewClient("user@test.com", "mypassword")
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}
	if client.UserID != "u1" {
		t.Errorf("UserID: want u1 got %q", client.UserID)
	}
}

func TestNewRequestAttachesContext(t *testing.T) {
	type ctxKey struct{}
	ctx := context.WithValue(context.Background(), ctxKey{}, "sentinel")

	req, err := newRequest(ctx, http.MethodGet, "http://example.com")
	if err != nil {
		t.Fatalf("newRequest: %v", err)
	}
	if req.Context().Value(ctxKey{}) != "sentinel" {
		t.Error("newRequest: context not attached to request")
	}

	req2, err := newRequestWithBody(ctx, http.MethodPost, "http://example.com", map[string]string{"k": "v"})
	if err != nil {
		t.Fatalf("newRequestWithBody: %v", err)
	}
	if req2.Context().Value(ctxKey{}) != "sentinel" {
		t.Error("newRequestWithBody: context not attached to request")
	}
}

func TestClientDoCancelledContextPropagates(t *testing.T) {
	ready := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		close(ready)
		time.Sleep(500 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	ctx, cancel := context.WithCancel(context.Background())
	client, _ := NewClientWithToken("u", "t")
	go func() { <-ready; cancel() }()

	_, err := client.GetFrame(ctx, "frame1")
	if err == nil {
		t.Fatal("expected error from canceled context")
	}
	if !errors.Is(err, context.Canceled) {
		var ne *NetworkError
		if !errors.As(err, &ne) {
			t.Errorf("expected context.Canceled or *NetworkError wrapping it, got %T: %v", err, err)
		}
	}
}
