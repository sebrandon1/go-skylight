package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListLists(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns lists",
			status:   http.StatusOK,
			response: `{"data":[{"id":"1","type":"list","attributes":{"label":"Grocery","color":"#FF0000","kind":"to_do"}},{"id":"2","type":"list","attributes":{"label":"Todo","color":"#00FF00","kind":"to_do"}}]}`,
			wantLen:  2,
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
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/lists" {
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
			lists, err := client.ListLists("frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && len(lists) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(lists))
			}
		})
	}
}

func TestGetList(t *testing.T) {
	tests := []struct {
		name      string
		listID    string
		status    int
		response  string
		wantTitle string
		wantItems int
		wantErr   bool
	}{
		{
			name:      "returns list with items",
			listID:    "1",
			status:    http.StatusOK,
			response:  `{"data":{"id":"1","type":"list","attributes":{"label":"Grocery","color":"#FF0000","kind":"to_do"}},"included":[{"id":"item1","type":"list_item","attributes":{"label":"Milk","status":"pending","position":1}}]}`,
			wantTitle: "Grocery",
			wantItems: 1,
		},
		{
			name:    "server error returns error",
			listID:  "1",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
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
			list, err := client.GetList("frame1", tc.listID)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if list.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, list.Title)
			}
			if len(list.Items) != tc.wantItems {
				t.Errorf("Items: want %d got %d", tc.wantItems, len(list.Items))
			}
		})
	}
}

func TestCreateList(t *testing.T) {
	tests := []struct {
		name      string
		input     ListData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "creates list",
			input:     ListData{Title: "Shopping", Color: "#00FF00"},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"3","type":"list","attributes":{"label":"Shopping","color":"#00FF00","kind":"to_do"}}}`,
			wantTitle: "Shopping",
		},
		{
			name:      "sends label and color in request body",
			input:     ListData{Title: "Grocery List", Color: "#00FF00"},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"list1","type":"list","attributes":{"label":"Grocery List","color":"#00FF00","kind":"to_do"}}}`,
			wantTitle: "Grocery List",
		},
		{
			name:    "server error returns error",
			input:   ListData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/lists" {
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
			list, err := client.CreateList("frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && list.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, list.Title)
			}
		})
	}
}

func TestUpdateList(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "updates list",
			status:    http.StatusOK,
			response:  `{"data":{"id":"1","type":"list","attributes":{"label":"Updated List","color":"#FF0000","kind":"to_do"}}}`,
			wantTitle: "Updated List",
		},
		{
			name:      "204 no content succeeds",
			status:    http.StatusNoContent,
			wantTitle: "",
		},
		{
			name:      "201 created with body",
			status:    http.StatusCreated,
			response:  `{"data":{"id":"1","type":"list","attributes":{"label":"New List","color":"#FF0000","kind":"to_do"}}}`,
			wantTitle: "New List",
		},
		{
			name:    "server error returns error",
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
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
			list, err := client.UpdateList("frame1", "1", ListData{Title: "Test"})
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && tc.wantTitle != "" && list.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, list.Title)
			}
		})
	}
}

func TestDeleteList(t *testing.T) {
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
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.DeleteList("frame1", "1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestAddListItem(t *testing.T) {
	tests := []struct {
		name      string
		input     ListItemData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "adds item",
			input:     ListItemData{Title: "Eggs"},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"item2","type":"list_item","attributes":{"label":"Eggs","status":"pending","position":1}}}`,
			wantTitle: "Eggs",
		},
		{
			name:      "sends label and position in request body",
			input:     ListItemData{Title: "Milk", Position: 1},
			status:    http.StatusCreated,
			response:  `{"data":{"id":"item1","type":"list_item","attributes":{"label":"Milk","status":"pending","position":1}}}`,
			wantTitle: "Milk",
		},
		{
			name:    "server error returns error",
			input:   ListItemData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
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
			item, err := client.AddListItem("frame1", "1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && item.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, item.Title)
			}
		})
	}
}

func TestUpdateListItem(t *testing.T) {
	tests := []struct {
		name          string
		input         ListItemData
		status        int
		response      string
		wantCompleted bool
		wantErr       bool
	}{
		{
			name:          "marks item completed",
			input:         ListItemData{Title: "Updated Item", Completed: true},
			status:        http.StatusOK,
			response:      `{"data":{"id":"item1","type":"list_item","attributes":{"label":"Updated Item","status":"completed","position":1}}}`,
			wantCompleted: true,
		},
		{
			name:    "server error returns error",
			input:   ListItemData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPut {
					t.Errorf("expected PUT, got %s", r.Method)
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
			item, err := client.UpdateListItem("frame1", "1", "item1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && item.Completed != tc.wantCompleted {
				t.Errorf("Completed: want %v got %v", tc.wantCompleted, item.Completed)
			}
		})
	}
}

func TestDeleteListItem(t *testing.T) {
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
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.DeleteListItem("frame1", "1", "item1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestCreateTaskBoxItem(t *testing.T) {
	tests := []struct {
		name      string
		input     TaskBoxItemData
		status    int
		response  string
		wantTitle string
		wantErr   bool
	}{
		{
			name:      "creates task box item",
			input:     TaskBoxItemData{Title: "Quick task"},
			status:    http.StatusCreated,
			response:  `{"id":"tb1","title":"Quick task"}`,
			wantTitle: "Quick task",
		},
		{
			name:    "server error returns error",
			input:   TaskBoxItemData{Title: "Test"},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1/task_box_items" {
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
			item, err := client.CreateTaskBoxItem("frame1", tc.input)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && item.Title != tc.wantTitle {
				t.Errorf("Title: want %q got %q", tc.wantTitle, item.Title)
			}
		})
	}
}
