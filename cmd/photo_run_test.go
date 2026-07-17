package cmd

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPhotoListCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"p1","attributes":{"status":"ready","asset_type":"image/jpeg","asset_url":"http://cdn/p1.jpg"}}],"meta":{"next_page_token":"next123"}}`)
	})

	stdout := captureStdout(func() {
		_ = captureStderr(func() {
			photoListCmd.Run(photoListCmd, nil)
		})
	})
	if !strings.Contains(stdout, "p1") {
		t.Errorf("expected photo id in output, got: %s", stdout)
	}
}

func TestPhotoListCmd_PrintsNextPageToken(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[],"meta":{"next_page_token":"next123"}}`)
	})

	// Table mode still prints token on stderr for humans
	orig := outputFormat
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = orig })

	stderr := captureStderr(func() { photoListCmd.Run(photoListCmd, nil) })
	if !strings.Contains(stderr, "next123") {
		t.Errorf("expected next page token on stderr, got: %s", stderr)
	}
}

func TestPhotoListCmd_JSONIncludesNextPageToken(t *testing.T) {
	// #264: JSON mode must expose next_page_token in stdout body
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"p1","attributes":{"status":"ready","asset_type":"image/jpeg","asset_url":"http://cdn/p1.jpg"}}],"meta":{"next_page_token":"tok-json"}}`)
	})

	orig := outputFormat
	outputFormat = outputJSON
	t.Cleanup(func() { outputFormat = orig })

	var stderr string
	stdout := captureStdout(func() {
		stderr = captureStderr(func() { photoListCmd.Run(photoListCmd, nil) })
	})
	if !strings.Contains(stdout, "next_page_token") || !strings.Contains(stdout, "tok-json") {
		t.Errorf("expected next_page_token in JSON stdout, got: %s", stdout)
	}
	if strings.Contains(stderr, "tok-json") {
		t.Errorf("JSON mode should not write token to stderr, got: %s", stderr)
	}
}

func TestPhotoUploadCmd(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "photo.jpg")
	if err := os.WriteFile(imgPath, []byte("fake-image-bytes"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	var s3URL string
	mux := http.NewServeMux()
	mux.HandleFunc("/upload_url", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"url":%q,"key":"k1","get_url":"http://cdn/k1.jpg"}}`, s3URL)
	})
	mux.HandleFunc("/s3-put", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	s3URL = srv.URL + "/s3-put"
	pointClientAt(t, clientAtURL(t, srv.URL))

	origFile, origCaption := photoFile, photoCaption
	photoFile, photoCaption = imgPath, "test caption"
	t.Cleanup(func() { photoFile, photoCaption = origFile, origCaption })

	out := captureStdout(func() { photoUploadCmd.Run(photoUploadCmd, nil) })
	if !strings.Contains(out, "k1") {
		t.Errorf("expected upload result with key in output, got: %s", out)
	}
}

func TestPhotoDeleteCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	origIDs := photoMessageID
	photoMessageID = []string{"1", "2"}
	t.Cleanup(func() { photoMessageID = origIDs })

	out := captureStdout(func() { photoDeleteCmd.Run(photoDeleteCmd, nil) })
	if !strings.Contains(out, "Photos deleted successfully") {
		t.Errorf("expected success message, got: %s", out)
	}
}

func TestPhotoDownloadCmd(t *testing.T) {
	var assetURL string
	mux := http.NewServeMux()
	mux.HandleFunc("/asset.jpg", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fake-image-bytes"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":[{"id":"p1","attributes":{"status":"ready","asset_type":"image/jpeg","asset_url":%q}}]}`, assetURL)
	})

	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	assetURL = srv.URL + "/asset.jpg"
	pointClientAt(t, clientAtURL(t, srv.URL))

	dir := t.TempDir()
	origAll, origIDs, origDir := photoDownloadAll, photoMessageID, photoOutputDir
	photoDownloadAll = true
	photoMessageID = nil
	photoOutputDir = dir
	t.Cleanup(func() {
		photoDownloadAll, photoMessageID, photoOutputDir = origAll, origIDs, origDir
	})

	out := captureStdout(func() { photoDownloadCmd.Run(photoDownloadCmd, nil) })
	if !strings.Contains(out, "Saved") {
		t.Errorf("expected save confirmation, got: %s", out)
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Errorf("expected one downloaded file, got %v (err=%v)", entries, err)
	}
}

func TestPhotoDownloadCmd_NoMatches(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[]}`)
	})

	origAll, origIDs := photoDownloadAll, photoMessageID
	photoDownloadAll = true
	photoMessageID = nil
	t.Cleanup(func() { photoDownloadAll, photoMessageID = origAll, origIDs })

	out := captureStdout(func() { photoDownloadCmd.Run(photoDownloadCmd, nil) })
	if !strings.Contains(out, "No matching photos found") {
		t.Errorf("expected no-matches message, got: %s", out)
	}
}
