package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListLists(t *testing.T) {
	mockLists := []List{
		{ID: "1", Title: "Grocery"},
		{ID: "2", Title: "Todo"},
	}

	mockResponseJSON, _ := json.Marshal(mockLists)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/lists" {
			t.Errorf("Expected path /api/frames/frame1/lists, got %s", r.URL.Path)
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

	lists, err := client.ListLists("frame1")
	if err != nil {
		t.Fatalf("ListLists failed: %v", err)
	}

	if len(lists) != 2 {
		t.Errorf("Expected 2 lists, got %d", len(lists))
	}
}

func TestGetList(t *testing.T) {
	mockList := List{ID: "1", Title: "Grocery", Items: []ListItem{
		{ID: "item1", Title: "Milk", Completed: false},
	}}

	mockResponseJSON, _ := json.Marshal(mockList)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/lists/1" {
			t.Errorf("Expected path /api/frames/frame1/lists/1, got %s", r.URL.Path)
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

	list, err := client.GetList("frame1", "1")
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}

	if list.Title != "Grocery" {
		t.Errorf("Expected title 'Grocery', got '%s'", list.Title)
	}

	if len(list.Items) != 1 {
		t.Errorf("Expected 1 item, got %d", len(list.Items))
	}
}

func TestAddListItem(t *testing.T) {
	mockItem := ListItem{ID: "item2", Title: "Eggs", Completed: false}

	mockResponseJSON, _ := json.Marshal(mockItem)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/lists/1/list_items" {
			t.Errorf("Expected path /api/frames/frame1/lists/1/list_items, got %s", r.URL.Path)
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

	item, err := client.AddListItem("frame1", "1", ListItemData{Title: "Eggs"})
	if err != nil {
		t.Fatalf("AddListItem failed: %v", err)
	}

	if item.Title != "Eggs" {
		t.Errorf("Expected title 'Eggs', got '%s'", item.Title)
	}
}

func TestDeleteListItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/lists/1/list_items/item1" {
			t.Errorf("Expected path /api/frames/frame1/lists/1/list_items/item1, got %s", r.URL.Path)
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

	err = client.DeleteListItem("frame1", "1", "item1")
	if err != nil {
		t.Fatalf("DeleteListItem failed: %v", err)
	}
}

func TestCreateList(t *testing.T) {
	mockList := List{ID: "3", Title: "Shopping", Color: "#00FF00"}

	mockResponseJSON, _ := json.Marshal(mockList)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/lists" {
			t.Errorf("Expected path /api/frames/frame1/lists, got %s", r.URL.Path)
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

	list, err := client.CreateList("frame1", ListData{Title: "Shopping", Color: "#00FF00"})
	if err != nil {
		t.Fatalf("CreateList failed: %v", err)
	}

	if list.Title != "Shopping" {
		t.Errorf("Expected title 'Shopping', got '%s'", list.Title)
	}
}

func TestUpdateList(t *testing.T) {
	mockList := List{ID: "1", Title: "Updated List"}

	mockResponseJSON, _ := json.Marshal(mockList)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/lists/1" {
			t.Errorf("Expected path /api/frames/frame1/lists/1, got %s", r.URL.Path)
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

	list, err := client.UpdateList("frame1", "1", ListData{Title: "Updated List"})
	if err != nil {
		t.Fatalf("UpdateList failed: %v", err)
	}

	if list.Title != "Updated List" {
		t.Errorf("Expected title 'Updated List', got '%s'", list.Title)
	}
}

func TestDeleteList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/lists/1" {
			t.Errorf("Expected path /api/frames/frame1/lists/1, got %s", r.URL.Path)
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

	err = client.DeleteList("frame1", "1")
	if err != nil {
		t.Fatalf("DeleteList failed: %v", err)
	}
}

func TestUpdateListItem(t *testing.T) {
	mockItem := ListItem{ID: "item1", Title: "Updated Item", Completed: true}

	mockResponseJSON, _ := json.Marshal(mockItem)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/lists/1/list_items/item1" {
			t.Errorf("Expected path /api/frames/frame1/lists/1/list_items/item1, got %s", r.URL.Path)
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

	item, err := client.UpdateListItem("frame1", "1", "item1", ListItemData{Title: "Updated Item", Completed: true})
	if err != nil {
		t.Fatalf("UpdateListItem failed: %v", err)
	}

	if !item.Completed {
		t.Error("Expected item to be completed")
	}
}

func TestCreateTaskBoxItem(t *testing.T) {
	mockItem := TaskBoxItem{ID: "tb1", Title: "Quick task"}

	mockResponseJSON, _ := json.Marshal(mockItem)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/api/frames/frame1/task_box_items" {
			t.Errorf("Expected path /api/frames/frame1/task_box_items, got %s", r.URL.Path)
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

	item, err := client.CreateTaskBoxItem("frame1", TaskBoxItemData{Title: "Quick task"})
	if err != nil {
		t.Fatalf("CreateTaskBoxItem failed: %v", err)
	}

	if item.Title != "Quick task" {
		t.Errorf("Expected title 'Quick task', got '%s'", item.Title)
	}
}

func TestListErrorHandling(t *testing.T) {
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

	_, err = client.ListLists("frame1")
	if err == nil {
		t.Error("Expected error, got nil")
	}

	_, err = client.GetList("frame1", "nonexistent")
	if err == nil {
		t.Error("Expected error, got nil")
	}
}
