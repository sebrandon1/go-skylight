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

func TestCreateCategory(t *testing.T) {
	tests := []struct {
		name      string
		input     CategoryData
		status    int
		response  string
		wantName  string
		wantColor string
		wantErr   bool
	}{
		{
			name:      "creates category",
			input:     CategoryData{Name: "Test Profile", Color: "#FF0000"},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"99","attributes":{"label":"Test Profile","color":"#FF0000"}}}`,
			wantName:  "Test Profile",
			wantColor: "#FF0000",
		},
		{
			name:    "server error returns error",
			input:   CategoryData{Name: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:     "invalid JSON returns error",
			input:    CategoryData{Name: "Test"},
			status:   http.StatusCreated,
			response: `{invalid json`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/categories" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				var raw map[string]any
				if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
					t.Errorf("decode body: %v", err)
				}
				if raw["label"] != tc.input.Name {
					t.Errorf("label: want %q got %v", tc.input.Name, raw["label"])
				}
				if tc.input.Color != "" && raw["color"] != tc.input.Color {
					t.Errorf("color: want %q got %v", tc.input.Color, raw["color"])
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
			category, err := client.CreateCategory("frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && category.Name != tc.wantName {
				t.Errorf("Name: want %q got %q", tc.wantName, category.Name)
			}
			if !tc.wantErr && tc.wantColor != "" && category.Color != tc.wantColor {
				t.Errorf("Color: want %q got %q", tc.wantColor, category.Color)
			}
		})
	}
}

func TestDeleteCategory(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{"deletes with 204", http.StatusNoContent, false},
		{"deletes with 200", http.StatusOK, false},
		{"server error returns error", http.StatusInternalServerError, true},
		{"not found returns error", http.StatusNotFound, true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodDelete {
					t.Errorf("expected DELETE, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/categories/cat1" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.DeleteCategory("frame1", "cat1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestUpdateCategory(t *testing.T) {
	tests := []struct {
		name     string
		catID    string
		input    CategoryData
		status   int
		response string
		wantName string
		wantErr  bool
	}{
		{
			name:     "updates category name",
			catID:    "cat1",
			input:    CategoryData{Name: "Updated Name"},
			status:   http.StatusOK,
			response: `{"data":{"id":"cat1","attributes":{"label":"Updated Name","color":"#FF0000"}}}`,
			wantName: "Updated Name",
		},
		{
			name:   "204 no content succeeds",
			catID:  "cat1",
			input:  CategoryData{Color: "#00FF00"},
			status: http.StatusNoContent,
		},
		{
			name:    "bad request returns error",
			catID:   "cat1",
			input:   CategoryData{Name: "Test"},
			status:  http.StatusBadRequest,
			wantErr: true,
		},
		{
			name:    "server error returns error",
			catID:   "cat1",
			input:   CategoryData{Name: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
		{
			name:     "invalid JSON returns error",
			catID:    "cat1",
			input:    CategoryData{Name: "Test"},
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
				if r.URL.Path != "/api/frames/frame1/categories/"+tc.catID {
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
			category, err := client.UpdateCategory("frame1", tc.catID, tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && tc.wantName != "" && category != nil && category.Name != tc.wantName {
				t.Errorf("Name: want %q got %q", tc.wantName, category.Name)
			}
		})
	}
}
