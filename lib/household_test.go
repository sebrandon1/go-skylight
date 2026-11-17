package lib

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// Shapes captured from the Skylight web app (API version 2026-10-15).
const householdConfigJSON = `{"data":{"id":"1","type":"household_config","attributes":{"disney_profile_pictures":true,"disney_screensaver":true}}}`

func TestGetHouseholdConfig(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/api/frames/f1/household_config" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		_, _ = w.Write([]byte(householdConfigJSON))
	}))
	defer srv.Close()

	client, _ := NewClientWithToken("u", "t", WithBaseURL(srv.URL+"/api"))
	cfg, err := client.GetHouseholdConfig(context.Background(), "f1")
	if err != nil {
		t.Fatalf("GetHouseholdConfig: %v", err)
	}
	if !cfg.DisneyProfilePictures || !cfg.DisneyScreensaver {
		t.Errorf("decoded %+v, want both true", cfg)
	}
}

func TestUpdateHouseholdConfig_SendsOnlyChangedFields(t *testing.T) {
	var body string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch || r.URL.Path != "/api/frames/f1/household_config" {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		b, _ := io.ReadAll(r.Body)
		body = string(b)
		_, _ = w.Write([]byte(`{"data":{"id":"1","type":"household_config","attributes":{"disney_profile_pictures":true,"disney_screensaver":false}}}`))
	}))
	defer srv.Close()

	client, _ := NewClientWithToken("u", "t", WithBaseURL(srv.URL+"/api"))
	off := false
	cfg, err := client.UpdateHouseholdConfig(context.Background(), "f1", HouseholdConfigUpdate{DisneyScreensaver: &off})
	if err != nil {
		t.Fatalf("UpdateHouseholdConfig: %v", err)
	}
	if body != `{"disney_screensaver":false}` {
		t.Errorf("PATCH body = %s, want {\"disney_screensaver\":false}", body)
	}
	if cfg.DisneyScreensaver || !cfg.DisneyProfilePictures {
		t.Errorf("decoded %+v", cfg)
	}
}

func TestUpdateHouseholdConfig_NoChange(t *testing.T) {
	client, _ := NewClientWithToken("u", "t", WithBaseURL("http://127.0.0.1:0/api"))
	if _, err := client.UpdateHouseholdConfig(context.Background(), "f1", HouseholdConfigUpdate{}); err == nil {
		t.Error("expected an error for an empty update")
	}
}

func TestHouseholdConfig_ServerErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "nope", http.StatusNotFound)
	}))
	defer srv.Close()

	client, _ := NewClientWithToken("u", "t", WithBaseURL(srv.URL+"/api"))
	if _, err := client.GetHouseholdConfig(context.Background(), "f1"); err == nil {
		t.Error("GetHouseholdConfig: expected error on 404")
	}
	on := true
	if _, err := client.UpdateHouseholdConfig(context.Background(), "f1", HouseholdConfigUpdate{DisneyProfilePictures: &on}); err == nil {
		t.Error("UpdateHouseholdConfig: expected error on 404")
	}
}
