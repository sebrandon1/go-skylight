package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"
	"sync"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestFrameListCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"f1","attributes":{"name":"Kitchen"}}]}`)
	})

	out := captureStdout(func() {
		if err := frameListCmd.RunE(frameListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Kitchen") {
		t.Errorf("expected frame name in output, got: %s", out)
	}
}

func TestFrameInfoCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"id":"test-frame","attributes":{"name":"Kitchen"}}}`)
	})

	out := captureStdout(func() {
		if err := frameInfoCmd.RunE(frameInfoCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Kitchen") {
		t.Errorf("expected frame name in output, got: %s", out)
	}
}

func TestFrameDevicesCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"d1","attributes":{"name":"Living Room","activated":true}}]}`)
	})

	out := captureStdout(func() {
		if err := frameDevicesCmd.RunE(frameDevicesCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Living Room") {
		t.Errorf("expected device name in output, got: %s", out)
	}
}

func TestFrameAvatarsCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"a1","attributes":{"name":"Robot","image_url":"http://example.com/a.png"}}]}`)
	})

	out := captureStdout(func() {
		if err := frameAvatarsCmd.RunE(frameAvatarsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Robot") {
		t.Errorf("expected avatar name in output, got: %s", out)
	}
}

func TestFrameColorsCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"c1","name":"Blue","hex":"#0000FF"}]}`)
	})

	out := captureStdout(func() {
		if err := frameColorsCmd.RunE(frameColorsCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Blue") {
		t.Errorf("expected color name in output, got: %s", out)
	}
}

func TestFrameSetAlbumCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	origAlbumID := currentAlbumID
	currentAlbumID = 42
	t.Cleanup(func() { currentAlbumID = origAlbumID })

	out := captureStdout(func() {
		if err := frameSetAlbumCmd.RunE(frameSetAlbumCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Current album set to 42") {
		t.Errorf("expected confirmation message, got: %s", out)
	}
}

func TestFrameSetAlbumCmdQuiet(t *testing.T) {
	// #265: --quiet must suppress set-album success output
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	origAlbumID := currentAlbumID
	origQuiet := quiet
	currentAlbumID = 7
	quiet = true
	t.Cleanup(func() {
		currentAlbumID = origAlbumID
		quiet = origQuiet
	})

	out := captureStdout(func() {
		if err := frameSetAlbumCmd.RunE(frameSetAlbumCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if out != "" {
		t.Errorf("expected no output with --quiet, got: %q", out)
	}
}

func TestFrameUpdateCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	// pflag.Set() marks the flag as permanently "changed" on the shared
	// command singleton (no unset API), so this only runs once per process.
	if err := frameUpdateCmd.Flags().Set(flagScreensaverShowWeather, "true"); err != nil {
		t.Fatalf("setting screensaver-show-weather flag: %v", err)
	}

	out := captureStdout(func() {
		if err := frameUpdateCmd.RunE(frameUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Frame settings updated") {
		t.Errorf("expected confirmation message, got: %s", out)
	}
}

func TestFrameUpdateCmd_APIError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	// Ensure something is sent even when this test runs alone.
	setTabFlags(t, "d1", []string{"meals"}, nil)

	if err := frameUpdateCmd.RunE(frameUpdateCmd, nil); err == nil {
		t.Error("expected error from API failure, got nil")
	}
}

func TestFrameUpdateCmd_NothingToUpdate(t *testing.T) {
	if frameUpdateCmd.Flags().Changed(flagScreensaverShowWeather) || frameUpdateCmd.Flags().Changed(flagScreensaverShowEvents) {
		t.Skip("a screensaver flag was set by an earlier test in this process")
	}
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		t.Errorf("unexpected request: %s", r.URL.Path)
	})
	setTabFlags(t, "", nil, nil)

	err := frameUpdateCmd.RunE(frameUpdateCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "nothing to update") {
		t.Errorf("want nothing-to-update error, got %v", err)
	}
}

func TestFrameUpdateCmd_MissingFrameID(t *testing.T) {
	origFrameID := frameID
	frameID = ""
	t.Cleanup(func() { frameID = origFrameID })

	if err := frameUpdateCmd.RunE(frameUpdateCmd, nil); err == nil {
		t.Error("expected error for missing frame ID, got nil")
	}
}

const testDeviceConfigJSON = `{"data":{"id":"d1","type":"device_config","attributes":{"home_tab":true,"meals_tab":false}}}`

// setTabFlags sets the frame tab flag vars for one test and restores them.
func setTabFlags(t *testing.T, deviceID string, hide, show []string) {
	t.Helper()
	origDevice, origHide, origShow := frameDeviceID, hideTabs, showTabs
	frameDeviceID, hideTabs, showTabs = deviceID, hide, show
	t.Cleanup(func() { frameDeviceID, hideTabs, showTabs = origDevice, origHide, origShow })
}

func TestFrameTabsCmd(t *testing.T) {
	tests := []struct {
		name      string
		deviceID  string
		devices   string
		wantPaths []string
	}{
		{
			name:      "all devices when --device-id is unset",
			devices:   `{"data":[{"id":"d1","attributes":{}},{"id":"d2","attributes":{}}]}`,
			wantPaths: []string{"/frames/test-frame/devices", "/frames/test-frame/devices/d1/device_config", "/frames/test-frame/devices/d2/device_config"},
		},
		{
			name:      "only the given device",
			deviceID:  "d9",
			wantPaths: []string{"/frames/test-frame/devices/d9/device_config"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var (
				mu    sync.Mutex
				paths []string
			)
			newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				mu.Lock()
				paths = append(paths, r.URL.Path)
				mu.Unlock()
				w.Header().Set("Content-Type", "application/json")
				if strings.HasSuffix(r.URL.Path, "/devices") {
					fmt.Fprint(w, tc.devices)
					return
				}
				fmt.Fprint(w, testDeviceConfigJSON)
			})
			setTabFlags(t, tc.deviceID, nil, nil)

			out := captureStdout(func() {
				if err := frameTabsCmd.RunE(frameTabsCmd, nil); err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			})
			slices.Sort(paths) // device configs are fetched concurrently
			if !slices.Equal(paths, tc.wantPaths) {
				t.Errorf("paths: want %v got %v", tc.wantPaths, paths)
			}
			if !strings.Contains(out, `"meals": false`) {
				t.Errorf("expected meals hidden in output, got: %s", out)
			}
		})
	}
}

func TestFrameTabsCmd_Errors(t *testing.T) {
	tests := []struct {
		name     string
		deviceID string
		noFrame  bool
		wantErr  string
	}{
		{name: "device config API error", deviceID: "d1", wantErr: "getting tabs for device d1"},
		{name: "device list API error", wantErr: "listing devices"},
		{name: "missing frame ID", noFrame: true, wantErr: "frame"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			newCmdTestClient(t, http500Handler)
			setTabFlags(t, tc.deviceID, nil, nil)
			if tc.noFrame {
				frameID = ""
			}

			err := frameTabsCmd.RunE(frameTabsCmd, nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestFrameUpdateCmd_Tabs(t *testing.T) {
	var tabBody map[string]bool
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/frames/test-frame/devices":
			fmt.Fprint(w, `{"data":[{"id":"d1","attributes":{}}]}`)
		case "/frames/test-frame/devices/d1/device_config":
			if err := json.NewDecoder(r.Body).Decode(&tabBody); err != nil {
				t.Errorf("decode body: %v", err)
			}
			fmt.Fprint(w, testDeviceConfigJSON)
		default:
			// Screensaver PATCH, if an earlier test left a screensaver flag changed.
			w.WriteHeader(http.StatusNoContent)
		}
	})
	setTabFlags(t, "", []string{"meals", " Recipes "}, []string{"photos"})

	out := captureStdout(func() {
		if err := frameUpdateCmd.RunE(frameUpdateCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	want := map[string]bool{"meals_tab": false, "recipes_tab": false, "photos_tab": true}
	if !maps.Equal(tabBody, want) {
		t.Errorf("tab body: want %v got %v", want, tabBody)
	}
	if !strings.Contains(out, "Tabs updated on device d1") {
		t.Errorf("expected confirmation message, got: %s", out)
	}
}

func TestFrameUpdateCmd_TabErrors(t *testing.T) {
	tests := []struct {
		name         string
		deviceID     string
		hide         []string
		devices      string
		configStatus int
		noRequests   bool
		wantErr      string
	}{
		{
			name:    "several devices need --device-id",
			hide:    []string{"meals"},
			devices: `{"data":[{"id":"d1","attributes":{}},{"id":"d2","attributes":{}}]}`,
			wantErr: "pass --device-id",
		},
		{
			name:         "device config API error",
			deviceID:     "d1",
			hide:         []string{"meals"},
			configStatus: http.StatusInternalServerError,
			wantErr:      "updating device tabs",
		},
		{
			name:       "unknown tab rejected before any request",
			hide:       []string{"chores"},
			noRequests: true,
			wantErr:    `unknown tab "chores"`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
				if tc.noRequests {
					t.Errorf("unexpected request: %s", r.URL.Path)
				}
				switch {
				case strings.HasSuffix(r.URL.Path, "/device_config"):
					w.WriteHeader(tc.configStatus)
				case strings.HasSuffix(r.URL.Path, "/devices"):
					fmt.Fprint(w, tc.devices)
				default:
					// Screensaver PATCH, if an earlier test left a screensaver flag changed.
					w.WriteHeader(http.StatusNoContent)
				}
			})
			setTabFlags(t, tc.deviceID, tc.hide, nil)

			err := frameUpdateCmd.RunE(frameUpdateCmd, nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
				t.Errorf("want error containing %q, got %v", tc.wantErr, err)
			}
		})
	}
}

func TestTabChanges(t *testing.T) {
	tests := []struct {
		name    string
		hide    []string
		show    []string
		want    map[string]bool
		wantErr string
	}{
		{name: "none", want: map[string]bool{}},
		{name: "hide and show", hide: []string{"meals"}, show: []string{"photos"}, want: map[string]bool{"meals": false, "photos": true}},
		{name: "normalizes case and space, skips empty", hide: []string{" Meals ", ""}, want: map[string]bool{"meals": false}},
		{name: "duplicate in same flag is fine", hide: []string{"meals", "meals"}, want: map[string]bool{"meals": false}},
		{name: "unknown tab", hide: []string{"chores"}, wantErr: `unknown tab "chores"`},
		{name: "tab in both flags", hide: []string{"meals"}, show: []string{"meals"}, wantErr: "in both"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tabChanges(tc.hide, tc.show)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !maps.Equal(got, tc.want) {
				t.Errorf("want %v got %v", tc.want, got)
			}
		})
	}
}

func TestResolveDeviceID(t *testing.T) {
	tests := []struct {
		name     string
		deviceID string
		devices  string
		status   int
		want     string
		wantErr  string
	}{
		{name: "explicit ID skips lookup", deviceID: "d9", want: "d9"},
		{name: "device list API error", devices: `{}`, status: http.StatusInternalServerError, wantErr: "listing devices"},
		{name: "single device", devices: `{"data":[{"id":"d1","attributes":{}}]}`, want: "d1"},
		{name: "no devices", devices: `{"data":[]}`, wantErr: "no devices"},
		{name: "several devices", devices: `{"data":[{"id":"d1","attributes":{}},{"id":"d2","attributes":{}}]}`, wantErr: "d1, d2"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
				if tc.devices == "" {
					t.Errorf("unexpected request: %s", r.URL.Path)
				}
				if tc.status != 0 {
					w.WriteHeader(tc.status)
				}
				fmt.Fprint(w, tc.devices)
			})
			got, err := resolveDeviceID(context.Background(), client, tc.deviceID)
			if tc.wantErr != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErr) {
					t.Fatalf("want error containing %q, got %v", tc.wantErr, err)
				}
				return
			}
			if err != nil || got != tc.want {
				t.Errorf("want %q, got %q (err %v)", tc.want, got, err)
			}
		})
	}
}

func TestPrintDeviceTabsTable(t *testing.T) {
	out := captureStdout(func() {
		printDeviceTabsTable([]lib.DeviceConfig{{DeviceID: "d1", Tabs: map[string]bool{"home": true, "meals": false}}})
	})
	for _, want := range []string{"DEVICE", "MEALS", "d1", "shown", "hidden", "-"} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got: %s", want, out)
		}
	}
}
