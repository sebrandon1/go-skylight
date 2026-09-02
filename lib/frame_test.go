package lib

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListFrames(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantName string
		wantErr  bool
	}{
		{
			name:     "returns frames",
			status:   http.StatusOK,
			response: `{"data":[{"id":"f1","type":"frame_show","attributes":{"name":"Family Frame","timezone":"America/Chicago"}},{"id":"f2","type":"frame_show","attributes":{"name":"Office","timezone":"America/New_York"}}]}`,
			wantLen:  2,
			wantName: "Family Frame",
		},
		{
			name:     "empty list",
			status:   http.StatusOK,
			response: `{"data":[]}`,
			wantLen:  0,
		},
		{
			name:    "unauthorized returns error",
			status:  http.StatusUnauthorized,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodGet {
					t.Errorf("expected GET, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames" {
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
			frames, err := client.ListFrames(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(frames) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(frames))
			}
			if tc.wantName != "" && len(frames) > 0 && frames[0].Name != tc.wantName {
				t.Errorf("Name: want %q got %q", tc.wantName, frames[0].Name)
			}
		})
	}
}

func TestGetFrame(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantName string
		wantTZ   string
		wantErr  bool
	}{
		{
			name:     "returns frame info",
			status:   http.StatusOK,
			response: `{"data":{"id":"frame1","type":"frame_show","attributes":{"name":"Family Frame","timezone":"America/Chicago"}}}`,
			wantName: "Family Frame",
			wantTZ:   "America/Chicago",
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
				if r.URL.Path != "/api/frames/frame1" {
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
			frame, err := client.GetFrame(context.Background(), "frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if frame.Name != tc.wantName {
				t.Errorf("Name: want %q got %q", tc.wantName, frame.Name)
			}
			if frame.TimeZone != tc.wantTZ {
				t.Errorf("TimeZone: want %q got %q", tc.wantTZ, frame.TimeZone)
			}
		})
	}
}

func TestListDevices(t *testing.T) {
	tests := []struct {
		name         string
		status       int
		response     string
		wantLen      int
		wantOnline   bool
		wantTimezone string
		wantErr      bool
	}{
		{
			name:   "returns devices",
			status: http.StatusOK,
			response: fmt.Sprintf(`{"data":[{"id":"dev1","type":"device","attributes":{"name":"Kitchen","activated":true,"timezone":%q,"brightness":%d,"sleeps_at":%q,"wakes_at":%q,"nightlight":true}},{"id":"dev2","type":"device","attributes":{"name":"Living Room","activated":false}}]}`,
				testDeviceTimezone, testDeviceBrightness, testDeviceSleepsAt, testDeviceWakesAt),
			wantLen:      2,
			wantOnline:   true,
			wantTimezone: testDeviceTimezone,
		},
		{
			name:    "not found returns error",
			status:  http.StatusNotFound,
			wantErr: true,
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
				if r.URL.Path != "/api/frames/frame1/devices" {
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
			devices, err := client.ListDevices(context.Background(), "frame1")
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(devices) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(devices))
			}
			if len(devices) > 0 && devices[0].Online != tc.wantOnline {
				t.Errorf("devices[0].Online: want %v got %v", tc.wantOnline, devices[0].Online)
			}
			if tc.wantTimezone != "" && devices[0].Timezone != tc.wantTimezone {
				t.Errorf("devices[0].Timezone: want %q got %q", tc.wantTimezone, devices[0].Timezone)
			}
		})
	}
}

func TestGetAvatars(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantErr  bool
	}{
		{
			name:     "returns avatars",
			status:   http.StatusOK,
			response: `{"data":[{"id":"1","type":"avatar","attributes":{"name":"Cat","image_url":"https://example.com/cat.png"}},{"id":"2","type":"avatar","attributes":{"name":"Dog","image_url":"https://example.com/dog.png"}}]}`,
			wantLen:  2,
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
				if r.URL.Path != "/api/avatars" {
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
			avatars, err := client.GetAvatars(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && len(avatars) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(avatars))
			}
		})
	}
}

func TestSetCurrentAlbum(t *testing.T) {
	tests := []struct {
		name    string
		status  int
		wantErr bool
	}{
		{
			name:   "sets album successfully",
			status: http.StatusNoContent,
		},
		{
			name:   "200 ok also succeeds",
			status: http.StatusOK,
		},
		{
			name:    "not found returns error",
			status:  http.StatusNotFound,
			wantErr: true,
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
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/frame1" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.SetCurrentAlbum(context.Background(), "frame1", 42)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestUpdateFrameSettings(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name    string
		opts    UpdateFrameSettingsOptions
		status  int
		wantErr bool
		verify  func(t *testing.T, r *http.Request)
	}{
		{
			name: "sets both fields",
			opts: UpdateFrameSettingsOptions{
				ScreensaverShowWeather: &trueVal,
				ScreensaverShowEvents:  &falseVal,
			},
			status: http.StatusNoContent,
			verify: func(t *testing.T, r *http.Request) {
				var body struct {
					Frame struct {
						Weather *bool `json:"screensaver_show_weather"`
						Events  *bool `json:"screensaver_show_events"`
					} `json:"frame"`
				}
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				if body.Frame.Weather == nil || *body.Frame.Weather != true {
					t.Errorf("weather: want true, got %v", body.Frame.Weather)
				}
				if body.Frame.Events == nil || *body.Frame.Events != false {
					t.Errorf("events: want false, got %v", body.Frame.Events)
				}
			},
		},
		{
			name: "sets only weather",
			opts: UpdateFrameSettingsOptions{
				ScreensaverShowWeather: &trueVal,
			},
			status: http.StatusNoContent,
			verify: func(t *testing.T, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				frame := body["frame"].(map[string]any)
				if len(frame) != 1 {
					t.Errorf("expected 1 field, got %d", len(frame))
				}
				if frame["screensaver_show_weather"] != true {
					t.Errorf("weather: want true, got %v", frame["screensaver_show_weather"])
				}
			},
		},
		{
			name: "sets only events",
			opts: UpdateFrameSettingsOptions{
				ScreensaverShowEvents: &falseVal,
			},
			status: http.StatusNoContent,
			verify: func(t *testing.T, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				frame := body["frame"].(map[string]any)
				if len(frame) != 1 {
					t.Errorf("expected 1 field, got %d", len(frame))
				}
				if frame["screensaver_show_events"] != false {
					t.Errorf("events: want false, got %v", frame["screensaver_show_events"])
				}
			},
		},
		{
			name:   "no fields set",
			opts:   UpdateFrameSettingsOptions{},
			status: http.StatusNoContent,
			verify: func(t *testing.T, r *http.Request) {
				var body map[string]any
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Fatalf("decode body: %v", err)
				}
				frame := body["frame"].(map[string]any)
				if len(frame) != 0 {
					t.Errorf("expected 0 fields, got %d", len(frame))
				}
			},
		},
		{
			name:    "server error returns error",
			opts:    UpdateFrameSettingsOptions{},
			status:  http.StatusInternalServerError,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPatch {
					t.Errorf("expected PATCH, got %s", r.Method)
				}
				if r.URL.Path != "/api/frames/f1" {
					t.Errorf("unexpected path: %s", r.URL.Path)
				}
				if tc.verify != nil {
					tc.verify(t, r)
				}
				w.WriteHeader(tc.status)
			}))
			defer srv.Close()

			old := SkylightURL
			SkylightURL = srv.URL + "/api"
			defer func() { SkylightURL = old }()

			client, _ := NewClientWithToken("u", "t")
			err := client.UpdateFrameSettings(context.Background(), "f1", tc.opts)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
		})
	}
}

func TestGetColors(t *testing.T) {
	tests := []struct {
		name     string
		status   int
		response string
		wantLen  int
		wantHex  string
		wantErr  bool
	}{
		{
			name:     "returns colors",
			status:   http.StatusOK,
			response: `{"data":[{"name":"Red","hex":"#FF0000"},{"name":"Blue","hex":"#0000FF"}]}`,
			wantLen:  2,
			wantHex:  "#FF0000",
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
				if r.URL.Path != "/api/colors" {
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
			colors, err := client.GetColors(context.Background())
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if len(colors) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(colors))
			}
			if tc.wantHex != "" && len(colors) > 0 && colors[0].Hex != tc.wantHex {
				t.Errorf("colors[0].Hex: want %q got %q", tc.wantHex, colors[0].Hex)
			}
		})
	}
}
