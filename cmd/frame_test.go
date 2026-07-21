package cmd

import (
	"fmt"
	"net/http"
	"strings"
	"testing"
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
