package cmd

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

func TestPhotoListCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"p1","attributes":{"status":"ready","asset_type":"image/jpeg","asset_url":"http://cdn/p1.jpg"}}],"meta":{"next_page_token":"next123"}}`)
	})

	stdout := captureStdout(func() {
		_ = captureStderr(func() {
			if err := photoListCmd.RunE(photoListCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	})
	if !strings.Contains(stdout, "p1") {
		t.Errorf("expected photo id in output, got: %s", stdout)
	}
}

func TestPhotoListCmd_PrintsNextPageToken(t *testing.T) {
	// A non-empty page with a token: the token must reach stdout so scripts
	// can paginate by passing it back as --page-token.
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"p1","attributes":{}}],"meta":{"next_page_token":"next123"}}`)
	})

	// Table mode prints token on stdout so scripts can parse it
	orig := outputFormat
	outputFormat = outputTable
	t.Cleanup(func() { outputFormat = orig })

	out := captureStdout(func() {
		if err := photoListCmd.RunE(photoListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "next123") {
		t.Errorf("expected next page token on stdout, got: %s", out)
	}
}

func TestPhotoListCmd_EmptyPageSuppressesToken(t *testing.T) {
	// An empty page with a token: the token must be suppressed (the API overshoot
	// fix in lib.ListPhotos). Scripts stop paging when they see next_page_token "".
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[],"meta":{"next_page_token":"should-not-appear"}}`)
	})

	out := captureStdout(func() {
		if err := photoListCmd.RunE(photoListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "should-not-appear") {
		t.Errorf("token must be suppressed on an empty page, got: %s", out)
	}
}

func TestPhotoListCmd_LimitTruncates(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"p1","attributes":{}},{"id":"p2","attributes":{}},{"id":"p3","attributes":{}}],"meta":{"next_page_token":"CURSOR-XYZ"}}`)
	})

	origLimit := photoLimit
	photoLimit = 2
	t.Cleanup(func() { photoLimit = origLimit })

	out := captureStdout(func() {
		if err := photoListCmd.RunE(photoListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	// token must be suppressed when --limit truncates
	if strings.Contains(out, "CURSOR-XYZ") {
		t.Errorf("expected next_page_token suppressed after --limit truncation, got: %s", out)
	}
	if strings.Contains(out, "p3") {
		t.Errorf("expected only 2 photos, but p3 appeared in output: %s", out)
	}
	if !strings.Contains(out, "p1") {
		t.Errorf("expected p1 in output, got: %s", out)
	}
}

func TestPhotoListCmd_LimitTruncatesTableMode(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"p1","attributes":{}},{"id":"p2","attributes":{}},{"id":"p3","attributes":{}}],"meta":{"next_page_token":"CURSOR-XYZ"}}`)
	})

	origLimit, origFmt := photoLimit, outputFormat
	photoLimit = 2
	outputFormat = outputTable
	t.Cleanup(func() { photoLimit, outputFormat = origLimit, origFmt })

	out := captureStdout(func() {
		if err := photoListCmd.RunE(photoListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	// token must be suppressed in table mode too when --limit truncates
	if strings.Contains(out, "CURSOR-XYZ") {
		t.Errorf("expected next_page_token suppressed in table mode after --limit truncation, got: %s", out)
	}
}

func TestPhotoListCmd_LimitNoTruncate(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":[{"id":"p1","attributes":{}}],"meta":{"next_page_token":"CURSOR-XYZ"}}`)
	})

	origLimit, origFmt := photoLimit, outputFormat
	photoLimit = 5
	outputFormat = outputTable
	t.Cleanup(func() { photoLimit, outputFormat = origLimit, origFmt })

	out := captureStdout(func() {
		if err := photoListCmd.RunE(photoListCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	// token must still appear when --limit is not reached
	if !strings.Contains(out, "CURSOR-XYZ") {
		t.Errorf("expected next_page_token in output when limit not reached, got: %s", out)
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
		stderr = captureStderr(func() {
			if err := photoListCmd.RunE(photoListCmd, nil); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
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

	out := captureStdout(func() {
		if err := photoUploadCmd.RunE(photoUploadCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "k1") {
		t.Errorf("expected upload result with key in output, got: %s", out)
	}
}

func TestPhotoUploadCmd_FileTooLarge(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "big.jpg")
	if err := os.WriteFile(imgPath, []byte("x"), 0o600); err != nil {
		t.Fatalf("setup: %v", err)
	}

	origMax, origFile, origFrameID := maxPhotoBytes, photoFile, frameID
	maxPhotoBytes = 0 // any file will exceed the limit
	photoFile = imgPath
	frameID = "f1"
	t.Cleanup(func() { maxPhotoBytes, photoFile, frameID = origMax, origFile, origFrameID })

	err := photoUploadCmd.RunE(photoUploadCmd, nil)
	if err == nil {
		t.Fatal("expected error for oversized file, got nil")
	}
	if !strings.Contains(err.Error(), "file too large") {
		t.Errorf("expected 'file too large' in error, got: %v", err)
	}
}

func TestPhotoUploadCmd_FileNotFound(t *testing.T) {
	origFile, origFrameID := photoFile, frameID
	photoFile = filepath.Join(t.TempDir(), "does-not-exist.jpg")
	frameID = "f1"
	t.Cleanup(func() { photoFile, frameID = origFile, origFrameID })

	err := photoUploadCmd.RunE(photoUploadCmd, nil)
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
	if !strings.Contains(err.Error(), "accessing file") {
		t.Errorf("expected 'accessing file' in error, got: %v", err)
	}
}

func TestPhotoDeleteCmd(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	origIDs, origYes := photoID, yes
	photoID, yes = []string{"1", "2"}, true
	t.Cleanup(func() { photoID, yes = origIDs, origYes })

	out := captureStdout(func() {
		if err := photoDeleteCmd.RunE(photoDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Photos deleted successfully") {
		t.Errorf("expected success message, got: %s", out)
	}
}

func TestPhotoDeleteCmd_DryRun(t *testing.T) {
	origIDs, origDryRun := photoID, dryRun
	photoID, dryRun = []string{"1", "2"}, true
	t.Cleanup(func() { photoID, dryRun = origIDs, origDryRun })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	out := captureStdout(func() {
		if err := photoDeleteCmd.RunE(photoDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Dry run") {
		t.Errorf("expected dry run output, got: %s", out)
	}
}

func TestPhotoDeleteCmd_ConfirmationDeclined(t *testing.T) {
	origIDs, origYes := photoID, yes
	photoID, yes = []string{"1", "2"}, false
	t.Cleanup(func() { photoID, yes = origIDs, origYes })

	origFrameID := frameID
	frameID = "test-frame"
	t.Cleanup(func() { frameID = origFrameID })

	mockStdin(t, "n\n")

	out := captureStdout(func() {
		if err := photoDeleteCmd.RunE(photoDeleteCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if strings.Contains(out, "deleted successfully") {
		t.Errorf("expected no deletion when confirmation declined, got: %s", out)
	}
}

// photoDownloadTestServer starts a mock HTTP server that lists the given photo
// IDs from its root endpoint and serves fake bytes from /asset.jpg. It returns
// the output directory callers should use for downloads.
func photoDownloadTestServer(t *testing.T, ids []string) string {
	t.Helper()
	var assetURL string
	mux := http.NewServeMux()
	mux.HandleFunc("/asset.jpg", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("fake-image-bytes"))
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		parts := make([]string, len(ids))
		for i, id := range ids {
			parts[i] = fmt.Sprintf(`{"id":%q,"attributes":{"status":"ready","asset_type":"image/jpeg","asset_url":%q}}`, id, assetURL)
		}
		fmt.Fprintf(w, `{"data":[%s]}`, strings.Join(parts, ","))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	assetURL = srv.URL + "/asset.jpg"
	pointClientAt(t, clientAtURL(t, srv.URL))
	return t.TempDir()
}

func TestPhotoDownloadCmd(t *testing.T) {
	dir := photoDownloadTestServer(t, []string{"p1"})

	origAll, origIDs, origDir := photoDownloadAll, photoID, photoOutputDir
	photoDownloadAll = true
	photoID = nil
	photoOutputDir = dir
	t.Cleanup(func() {
		photoDownloadAll, photoID, photoOutputDir = origAll, origIDs, origDir
	})

	out := captureStdout(func() {
		if err := photoDownloadCmd.RunE(photoDownloadCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
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

	origAll, origIDs := photoDownloadAll, photoID
	photoDownloadAll = true
	photoID = nil
	t.Cleanup(func() { photoDownloadAll, photoID = origAll, origIDs })

	out := captureStdout(func() {
		if err := photoDownloadCmd.RunE(photoDownloadCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "No matching photos found") {
		t.Errorf("expected no-matches message, got: %s", out)
	}
}

func TestPhotoDownloadCmd_ByPhotoID(t *testing.T) {
	dir := photoDownloadTestServer(t, []string{"p1", "p2"})

	origAll, origIDs, origDir := photoDownloadAll, photoID, photoOutputDir
	photoDownloadAll = false
	photoID = []string{"p1"}
	photoOutputDir = dir
	t.Cleanup(func() { photoDownloadAll, photoID, photoOutputDir = origAll, origIDs, origDir })

	out := captureStdout(func() {
		if err := photoDownloadCmd.RunE(photoDownloadCmd, nil); err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
	if !strings.Contains(out, "Saved") {
		t.Errorf("expected save confirmation, got: %s", out)
	}

	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 {
		t.Errorf("expected exactly one downloaded file (p1 only), got %v (err=%v)", entries, err)
	}
}

func TestPhotoDownloadCmd_NoFlagsError(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	origAll, origIDs := photoDownloadAll, photoID
	photoDownloadAll = false
	photoID = nil
	t.Cleanup(func() { photoDownloadAll, photoID = origAll, origIDs })

	err := photoDownloadCmd.RunE(photoDownloadCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--photo-id or --all") {
		t.Errorf("expected '--photo-id or --all' error, got: %v", err)
	}
}

func TestPhotoDeleteCmd_InvalidID(t *testing.T) {
	newCmdTestClient(t, func(w http.ResponseWriter, r *http.Request) {})

	origIDs, origYes := photoID, yes
	photoID, yes = []string{"not-a-number"}, true
	t.Cleanup(func() { photoID, yes = origIDs, origYes })

	err := photoDeleteCmd.RunE(photoDeleteCmd, nil)
	if err == nil || !strings.Contains(err.Error(), "invalid photo ID") {
		t.Errorf("expected 'invalid photo ID' error, got: %v", err)
	}
}

func photoDownloadMixedServer(t *testing.T) string {
	t.Helper()
	var base string
	mux := http.NewServeMux()
	mux.HandleFunc("/good.jpg", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("image-bytes"))
	})
	mux.HandleFunc("/bad.jpg", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":[
			{"id":"good","attributes":{"status":"ready","asset_type":"image/jpeg","asset_url":%q}},
			{"id":"bad","attributes":{"status":"ready","asset_type":"image/jpeg","asset_url":%q}},
			{"id":"..","attributes":{"status":"ready","asset_type":"image/jpeg","asset_url":%q}},
			{"id":"pending","attributes":{"status":"processing","asset_type":"image/jpeg","asset_url":""}}
		]}`, base+"/good.jpg", base+"/bad.jpg", base+"/good.jpg")
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	base = srv.URL
	pointClientAt(t, clientAtURL(t, srv.URL))
	return t.TempDir()
}

func TestPhotoDownloadCmd_PartialFailure(t *testing.T) {
	dir := photoDownloadMixedServer(t)

	origAll, origIDs, origDir := photoDownloadAll, photoID, photoOutputDir
	photoDownloadAll, photoID, photoOutputDir = true, nil, dir
	t.Cleanup(func() { photoDownloadAll, photoID, photoOutputDir = origAll, origIDs, origDir })

	var runErr error
	captureStdout(func() { runErr = photoDownloadCmd.RunE(photoDownloadCmd, nil) })

	// Three photos have an asset URL: one saves, one 500s, one has an unsafe ID.
	if runErr == nil || !strings.Contains(runErr.Error(), "2 of 3 photo downloads failed") {
		t.Fatalf("error = %v, want 2 of 3 failed", runErr)
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Name() != "good.jpg" {
		t.Errorf("want only good.jpg and no .part files, got %v", entries)
	}
	info, err := entries[0].Info()
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("mode = %04o, want 0600", perm)
	}
}

func TestListAllPhotos_StopsOnEmptyPageWithToken(t *testing.T) {
	requests := 0
	client := newMockClient(t, func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get("page_token") != "more" {
			fmt.Fprint(w, `{"data":[{"id":"p1","attributes":{"status":"ready"}}],"meta":{"next_page_token":"more"}}`)
			return
		}
		fmt.Fprint(w, `{"data":[],"meta":{"next_page_token":"more"}}`)
	})
	photos, err := listAllPhotos(context.Background(), client, "f1")
	if err != nil {
		t.Fatalf("listAllPhotos: %v", err)
	}
	if len(photos) != 1 || requests != 2 {
		t.Errorf("got %d photos in %d requests, want 1 in 2", len(photos), requests)
	}
}

func TestDownloadPhotoFile_Errors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("image-bytes"))
	}))
	defer srv.Close()
	client := clientAtURL(t, srv.URL)
	photo := lib.Photo{ID: "p1", AssetURL: srv.URL + "/a.jpg", AssetType: "image/jpeg"}

	t.Run("missing directory", func(t *testing.T) {
		missing := filepath.Join(t.TempDir(), "missing")
		if _, err := downloadPhotoFile(context.Background(), client, photo, missing); err == nil {
			t.Error("expected a temp file error")
		}
	})

	t.Run("rename failure removes the part file", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.MkdirAll(filepath.Join(dir, "p1.jpg", "child"), 0o700); err != nil {
			t.Fatal(err)
		}
		if _, err := downloadPhotoFile(context.Background(), client, photo, dir); err == nil {
			t.Fatal("expected a rename error")
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 1 {
			t.Errorf("part file left behind: %v", entries)
		}
	})
}
