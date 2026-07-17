package lib

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExtToContentType(t *testing.T) {
	cases := []struct {
		ext  string
		want string
	}{
		{"jpg", "image/jpeg"},
		{"JPG", "image/jpeg"},
		{"jpeg", "image/jpeg"},
		{"png", "image/png"},
		{"PNG", "image/png"},
		{"gif", "image/gif"},
		{"webp", "image/webp"},
		{"mp4", "video/mp4"},
		{"MP4", "video/mp4"},
		{"mov", "video/quicktime"},
		{"MOV", "video/quicktime"},
		{"bmp", "image/jpeg"},  // unknown falls back to jpeg
		{"tiff", "image/jpeg"}, // unknown falls back to jpeg
		{"", "image/jpeg"},     // empty falls back to jpeg
	}
	for _, c := range cases {
		got := extToContentType(c.ext)
		if got != c.want {
			t.Errorf("extToContentType(%q) = %q, want %q", c.ext, got, c.want)
		}
	}
}

func TestListPhotos_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/frames/frame1/messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(photoAPIResponse{
			Data: []photoAPIEntry{
				{ID: "p1", Attributes: struct {
					Status       string `json:"status"`
					AssetType    string `json:"asset_type"`
					CreatedAt    string `json:"created_at"`
					UpdatedAt    string `json:"updated_at"`
					ThumbnailURL string `json:"thumbnail_url"`
					AssetURL     string `json:"asset_url"`
					SenderID     int    `json:"sender_id"`
				}{Status: "active", AssetType: "image/jpeg"}},
			},
			Meta: &photoMeta{NextPageToken: "tok2"},
		}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	photos, nextToken, err := client.ListPhotos("frame1", PhotoListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(photos) != 1 {
		t.Fatalf("want 1 photo, got %d", len(photos))
	}
	if photos[0].ID != "p1" {
		t.Errorf("ID: want p1, got %s", photos[0].ID)
	}
	if nextToken != "tok2" {
		t.Errorf("nextToken: want tok2, got %s", nextToken)
	}
}

func TestListPhotos_WithPageToken(t *testing.T) {
	var gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.URL.Query().Get("page_token")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(photoAPIResponse{}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, _, err := client.ListPhotos("frame1", PhotoListOptions{PageToken: "mytoken"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotToken != "mytoken" {
		t.Errorf("page_token: want mytoken, got %q", gotToken)
	}
}

func TestListPhotos_DefaultPageToken(t *testing.T) {
	var gotToken string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotToken = r.URL.Query().Get("page_token")
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(photoAPIResponse{}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, _, err := client.ListPhotos("frame1", PhotoListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotToken != "__START__" {
		t.Errorf("page_token: want __START__, got %q", gotToken)
	}
}

func TestListPhotos_NoNextToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(photoAPIResponse{Data: []photoAPIEntry{}}); err != nil {
			t.Errorf("encode: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, nextToken, err := client.ListPhotos("frame1", PhotoListOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if nextToken != "" {
		t.Errorf("nextToken: want empty, got %q", nextToken)
	}
}

func TestDeletePhotos_Success(t *testing.T) {
	var gotIDs []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("method: want DELETE, got %s", r.Method)
		}
		var body photoDeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode: %v", err)
		}
		gotIDs = body.MessageIDs
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	err := client.DeletePhotos("frame1", []int{1, 2, 3})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(gotIDs) != 3 || gotIDs[0] != 1 || gotIDs[1] != 2 || gotIDs[2] != 3 {
		t.Errorf("message IDs: want [1 2 3], got %v", gotIDs)
	}
}

func TestDeletePhotos_Error(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	err := client.DeletePhotos("frame1", []int{1})
	if err == nil {
		t.Error("expected error for 403 response, got nil")
	}
}

func TestDownloadPhoto_Success(t *testing.T) {
	imageData := []byte("fake-image-bytes")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write(imageData); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	client, _ := NewClientWithToken("u", "t")
	data, err := client.DownloadPhoto(srv.URL + "/photo.jpg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(data) != string(imageData) {
		t.Errorf("data: want %q, got %q", imageData, data)
	}
}

func TestDownloadPhoto_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.DownloadPhoto(srv.URL + "/photo.jpg")
	if err == nil {
		t.Error("expected error for 404, got nil")
	}
}

func TestDownloadPhoto_BadURL(t *testing.T) {
	client, _ := NewClientWithToken("u", "t")
	_, err := client.DownloadPhoto("://bad-url")
	if err == nil {
		t.Error("expected error for bad URL, got nil")
	}
}

func TestListPhotos_BadURL(t *testing.T) {
	old := SkylightURL
	SkylightURL = "://bad"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, _, err := client.ListPhotos("frame1", PhotoListOptions{})
	if err == nil {
		t.Error("expected error for bad URL, got nil")
	}
}

func TestDeletePhotos_BadURL(t *testing.T) {
	old := SkylightURL
	SkylightURL = "://bad"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	err := client.DeletePhotos("frame1", []int{1})
	if err == nil {
		t.Error("expected error for bad URL, got nil")
	}
}

func TestUploadPhoto(t *testing.T) {
	s3Srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("expected PUT on S3 server, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer s3Srv.Close()

	skylightSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/api/upload_url" {
			t.Errorf("expected POST /api/upload_url, got %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		body := `{"data":{"url":"` + s3Srv.URL + `","key":"k1","get_url":"","message_ids":[1],"frame_names":["frame1"]}}`
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer skylightSrv.Close()

	old := SkylightURL
	SkylightURL = skylightSrv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	result, err := client.UploadPhoto("frame1", "jpg", []byte("test data"), "")
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Key != "k1" {
		t.Errorf("Key: want k1, got %q", result.Key)
	}
}

func TestUploadPhoto_PresignError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		if _, err := w.Write([]byte(`{"error":"fail"}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.UploadPhoto("frame1", "jpg", []byte("test data"), "")
	if err == nil {
		t.Fatal("expected error for presign failure, got nil")
	}
}

func TestUploadPhoto_EmptyUploadURL(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write([]byte(`{"data":{"url":"","key":"","get_url":"","message_ids":[],"frame_names":[]}}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	old := SkylightURL
	SkylightURL = srv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.UploadPhoto("frame1", "jpg", []byte("test data"), "")
	if err == nil {
		t.Fatal("expected error for empty upload URL, got nil")
	}
	if !strings.Contains(err.Error(), "empty upload URL") {
		t.Errorf("expected error to contain 'empty upload URL', got: %v", err)
	}
}

func TestUploadPhoto_S3Error(t *testing.T) {
	s3Srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer s3Srv.Close()

	skylightSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		body := `{"data":{"url":"` + s3Srv.URL + `","key":"k1","get_url":"","message_ids":[1],"frame_names":["frame1"]}}`
		if _, err := w.Write([]byte(body)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer skylightSrv.Close()

	old := SkylightURL
	SkylightURL = skylightSrv.URL + "/api"
	defer func() { SkylightURL = old }()

	client, _ := NewClientWithToken("u", "t")
	_, err := client.UploadPhoto("frame1", "jpg", []byte("test data"), "")
	if err == nil {
		t.Fatal("expected error for S3 failure, got nil")
	}
}

