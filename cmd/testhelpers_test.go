package cmd

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
)

// captureStderr redirects os.Stderr for the duration of fn and returns
// whatever was written to it.
func captureStderr(fn func()) string {
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w
	fn()
	w.Close()
	os.Stderr = old
	var buf bytes.Buffer
	buf.ReadFrom(r)
	return buf.String()
}

// clientAtURL builds a *lib.Client pointed at baseURL.
func clientAtURL(t *testing.T, baseURL string) *lib.Client {
	t.Helper()
	client, err := lib.NewClientWithToken("u", "t", lib.WithBaseURL(baseURL))
	if err != nil {
		t.Fatalf("creating test client: %v", err)
	}
	return client
}

// newMockClient starts an httptest server driven by handler and returns a
// *lib.Client pointed at it.
func newMockClient(t *testing.T, handler http.HandlerFunc) *lib.Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return clientAtURL(t, srv.URL)
}

// pointClientAt wires client up as the active client for a cobra command's
// Run closure: getClient() returns it via autoClient, and requireFrameID()
// passes via frameID. Restores prior state via t.Cleanup so package-level
// state doesn't leak between tests.
func pointClientAt(t *testing.T, client *lib.Client) {
	t.Helper()
	origAutoClient, origFrameID := autoClient, frameID
	autoClient = client
	frameID = "test-frame"
	t.Cleanup(func() {
		autoClient = origAutoClient
		frameID = origFrameID
	})
}

// newCmdTestClient builds a mock client and wires it up as the active client
// for a cobra command's Run closure. Use clientAtURL + pointClientAt
// directly when the mock server's own URL needs to be embedded in a
// response (e.g. a presigned upload URL).
func newCmdTestClient(t *testing.T, handler http.HandlerFunc) {
	t.Helper()
	pointClientAt(t, newMockClient(t, handler))
}
