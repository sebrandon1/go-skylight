package cmd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/sebrandon1/go-skylight/lib"
	"github.com/spf13/cobra"
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

// setTestContext sets context.Background() on every cobra command in the tree
// so that cmd.Context() returns a valid (non-nil) context when RunE is called
// directly in tests (bypassing cobra.ExecuteContext).
func setTestContext(root *cobra.Command) {
	root.SetContext(context.Background())
	for _, c := range root.Commands() {
		setTestContext(c)
	}
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
	setTestContext(rootCmd)
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

// mockStdin replaces os.Stdin with a pipe pre-loaded with input for the
// duration of the test, then restores the original stdin via t.Cleanup.
func mockStdin(t *testing.T, input string) {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating stdin pipe: %v", err)
	}
	if _, err := w.WriteString(input); err != nil {
		t.Fatalf("writing to stdin pipe: %v", err)
	}
	w.Close()
	orig := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = orig
		r.Close()
	})
}

// assertCommandRegistered fails the test unless parent has a direct
// subcommand whose Use matches use.
func assertCommandRegistered(t *testing.T, parent *cobra.Command, use string) {
	t.Helper()
	for _, c := range parent.Commands() {
		if c.Use == use {
			return
		}
	}
	t.Errorf("%q command not registered on %q", use, parent.Use)
}
