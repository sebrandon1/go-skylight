package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListCategories(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantName string
		wantErr  bool
	}{
		{
			name:     "returns family members",
			status:   http.StatusOK,
			response: `{"data":[{"id":"1","attributes":{"label":"Mom","color":"#FF0000"}},{"id":"2","attributes":{"label":"Dad","color":"#0000FF"}}]}`,
			wantLen:  2,
			wantName: "Mom",
		},
		{
			name:     "empty list",
			status:   http.StatusOK,
			response: `{"data":[]}`,
			wantLen:  0,
		},
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
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
				if r.URL.Path != "/api/frames/frame1/categories" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.status)
				if tc.response != "" {
					if _, err := w.Write([]byte(tc.response)); err != nil {
						t.Errorf("write response: %v", err)
					}
				}
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			categories, err := client.ListCategories("frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(categories) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(categories))
			}
			if tc.wantName != "" && len(categories) > 0 && categories[0].Name != tc.wantName {
				t.Errorf("want categories[0].Name=%q got %q", tc.wantName, categories[0].Name)
			}
		})
	}
}

func TestListCategoriesRequestBody(t *testing.T) {
	data, _ := json.Marshal(categoryAPIResponse{
		Data: []categoryAPIEntry{
			{ID: "1", Attributes: struct {
				Label         string  `json:"label"`
				Color         string  `json:"color"`
				ProfilePicURL *string `json:"profile_pic_url"`
			}{Label: "Mom", Color: "#FF0000"}},
			{ID: "2", Attributes: struct {
				Label         string  `json:"label"`
				Color         string  `json:"color"`
				ProfilePicURL *string `json:"profile_pic_url"`
			}{Label: "Dad", Color: "#0000FF"}},
		},
	})

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(data); err != nil {
			t.Errorf("write response: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	categories, err := client.ListCategories("frame1")
	if err != nil {
		t.Fatalf("ListCategories failed: %v", err)
	}
	if len(categories) != 2 {
		t.Errorf("expected 2 categories, got %d", len(categories))
	}
	if categories[0].Color != "#FF0000" {
		t.Errorf("expected color #FF0000, got %s", categories[0].Color)
	}
}
