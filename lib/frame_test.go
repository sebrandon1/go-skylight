package lib

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

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
			frame, err := client.GetFrame("frame1")
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
		name       string
		status     int
		response   string
		wantLen    int
		wantOnline bool
		wantErr    bool
	}{
		{
			name:       "returns devices",
			status:     http.StatusOK,
			response:   `{"data":[{"id":"dev1","type":"device","attributes":{"name":"Kitchen","activated":true}},{"id":"dev2","type":"device","attributes":{"name":"Living Room","activated":false}}]}`,
			wantLen:    2,
			wantOnline: true,
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
			devices, err := client.ListDevices("frame1")
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
			avatars, err := client.GetAvatars()
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if !tc.wantErr && len(avatars) != tc.wantLen {
				t.Errorf("wantLen=%d got %d", tc.wantLen, len(avatars))
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
			colors, err := client.GetColors()
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
