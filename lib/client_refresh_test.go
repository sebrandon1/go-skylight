package lib

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// refreshFixture is a mock API that accepts only the newest access token, and
// a mock OAuth endpoint that issues access-N and refresh-N.
type refreshFixture struct {
	exchanges atomic.Int32
	oauthFail atomic.Int32 // non-zero: OAuth responds with this status
	api       *httptest.Server
	mu        sync.Mutex
	valid     string
	bodies    []string
}

func newRefreshFixture(t *testing.T) *refreshFixture {
	t.Helper()
	f := &refreshFixture{valid: "access-1"}
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if status := f.oauthFail.Load(); status != 0 {
			w.WriteHeader(int(status))
			fmt.Fprint(w, `{"error":"invalid_grant"}`)
			return
		}
		n := f.exchanges.Add(1)
		f.mu.Lock()
		f.valid = fmt.Sprintf("access-%d", n)
		f.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"access_token":"access-%d","refresh_token":"refresh-%d","expires_in":3600}`, n, n)
	}))
	t.Cleanup(oauth.Close)
	orig := OAuthURL
	OAuthURL = oauth.URL
	t.Cleanup(func() { OAuthURL = orig })

	f.api = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Header.Get("Authorization") != "Bearer "+f.valid {
			w.WriteHeader(http.StatusUnauthorized)
			fmt.Fprint(w, `{"errors":["Invalid token"]}`)
			return
		}
		f.bodies = append(f.bodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, `{"data":[{"id":"r1","attributes":{"name":"x"}}]}`)
	}))
	t.Cleanup(f.api.Close)
	return f
}

// expire makes the API reject every access token issued so far.
func (f *refreshFixture) expire() {
	f.mu.Lock()
	f.valid = "none-yet"
	f.mu.Unlock()
}

func (f *refreshFixture) client(t *testing.T, opts ...ClientOption) *Client {
	t.Helper()
	c, err := NewClientWithRefreshToken("refresh-0", "fp", append(opts, WithBaseURL(f.api.URL))...)
	if err != nil {
		t.Fatalf("NewClientWithRefreshToken: %v", err)
	}
	return c
}

func TestRefreshOn401_ReplaysPostBody(t *testing.T) {
	f := newRefreshFixture(t)
	c := f.client(t)
	f.expire()

	if _, err := c.CreateReward(context.Background(), "f1", RewardData{Title: "Ice cream", Points: 5, CategoryIDs: []int{1}}); err != nil {
		t.Fatalf("CreateReward: %v", err)
	}
	if len(f.bodies) != 1 || !strings.Contains(f.bodies[0], "Ice cream") {
		t.Errorf("replayed bodies = %q, want the original body", f.bodies)
	}
	if c.APIToken != "access-2" || c.RefreshToken != "refresh-2" {
		t.Errorf("tokens = %q, %q", c.APIToken, c.RefreshToken)
	}
}

func TestRefreshOn401_ConcurrentRequestsShareOneExchange(t *testing.T) {
	f := newRefreshFixture(t)
	c := f.client(t)
	f.expire()

	var wg sync.WaitGroup
	errs := make(chan error, 20)
	for range 20 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.ListRewards(context.Background(), "f1")
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Errorf("request failed: %v", err)
		}
	}
	// One exchange at construction, one for all 20 expired requests.
	if got := f.exchanges.Load(); got != 2 {
		t.Errorf("exchanges = %d, want 2", got)
	}
}

func TestRefreshOn401_HandlerReceivesEachToken(t *testing.T) {
	f := newRefreshFixture(t)
	var got []string
	c := f.client(t, WithTokenRefreshHandler(func(rt string) error {
		got = append(got, rt)
		return errors.New("handler errors do not fail requests")
	}))
	f.expire()

	if _, err := c.ListRewards(context.Background(), "f1"); err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	if want := "refresh-1,refresh-2"; strings.Join(got, ",") != want {
		t.Errorf("handler got %v, want %s", got, want)
	}
}

func TestRefreshOn401_ReplaysOnlyOnce(t *testing.T) {
	var exchanges atomic.Int32
	oauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := exchanges.Add(1)
		fmt.Fprintf(w, `{"access_token":"a%d","refresh_token":"r%d"}`, n, n)
	}))
	defer oauth.Close()
	orig := OAuthURL
	OAuthURL = oauth.URL
	t.Cleanup(func() { OAuthURL = orig })

	var calls atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer api.Close()

	c, err := NewClientWithRefreshToken("r0", "fp", WithBaseURL(api.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.ListRewards(context.Background(), "f1"); !IsAuthError(err) {
		t.Errorf("err = %v, want an AuthError", err)
	}
	if calls.Load() != 2 || exchanges.Load() != 2 {
		t.Errorf("API calls = %d, exchanges = %d; want 2, 2", calls.Load(), exchanges.Load())
	}
}

func TestRefreshOn401_ExchangeFailures(t *testing.T) {
	tests := []struct {
		name        string
		status      int32
		wantAuthErr bool
	}{
		{"rejected token", http.StatusBadRequest, true},
		{"unauthorized client", http.StatusUnauthorized, true},
		{"server outage", http.StatusInternalServerError, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := newRefreshFixture(t)
			c := f.client(t)
			f.expire()
			f.oauthFail.Store(tt.status)

			_, err := c.ListRewards(context.Background(), "f1")
			if err == nil {
				t.Fatal("expected an error")
			}
			if IsAuthError(err) != tt.wantAuthErr {
				t.Errorf("IsAuthError = %v, want %v (err: %v)", IsAuthError(err), tt.wantAuthErr, err)
			}
			if (c.RefreshFailure() != nil) != tt.wantAuthErr {
				t.Errorf("RefreshFailure = %v", c.RefreshFailure())
			}
		})
	}
}

func TestRefreshOn401_SuccessClearsFailure(t *testing.T) {
	f := newRefreshFixture(t)
	c := f.client(t)
	f.expire()
	f.oauthFail.Store(http.StatusBadRequest)
	if _, err := c.ListRewards(context.Background(), "f1"); err == nil {
		t.Fatal("expected an error")
	}

	f.oauthFail.Store(0)
	if _, err := c.ListRewards(context.Background(), "f1"); err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	if err := c.RefreshFailure(); err != nil {
		t.Errorf("RefreshFailure = %v, want nil after a successful refresh", err)
	}
}

func TestWithTokenExchange_UsedForConstructionAndRefresh(t *testing.T) {
	f := newRefreshFixture(t)
	var seen []string
	c := f.client(t, WithTokenExchange(func(rt, fp string) (*OAuthTokenResponse, error) {
		seen = append(seen, rt+"/"+fp)
		f.mu.Lock()
		f.valid = "custom"
		f.mu.Unlock()
		return &OAuthTokenResponse{AccessToken: "custom", RefreshToken: "next"}, nil
	}))
	f.expire()

	if _, err := c.ListRewards(context.Background(), "f1"); err != nil {
		t.Fatalf("ListRewards: %v", err)
	}
	if want := "refresh-0/fp,next/fp"; strings.Join(seen, ",") != want {
		t.Errorf("exchange calls = %v, want %s", seen, want)
	}
	if f.exchanges.Load() != 0 {
		t.Errorf("default exchange ran %d times, want 0", f.exchanges.Load())
	}
}

func TestStaticTokenClient_DoesNotRefreshOn401(t *testing.T) {
	var calls atomic.Int32
	api := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer api.Close()

	c, err := NewClientWithToken("u", "t", WithBaseURL(api.URL))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := c.ListRewards(context.Background(), "f1"); !IsAuthError(err) {
		t.Errorf("err = %v, want an AuthError", err)
	}
	if calls.Load() != 1 {
		t.Errorf("calls = %d, want 1", calls.Load())
	}
}

func TestRefreshOn401_HandlerErrorIsLogged(t *testing.T) {
	f := newRefreshFixture(t)
	var logs bytes.Buffer
	f.client(t,
		WithLogger(slog.New(slog.NewTextHandler(&logs, nil))),
		WithTokenRefreshHandler(func(string) error { return errors.New("save failed") }),
	)
	if !strings.Contains(logs.String(), "save failed") {
		t.Errorf("logs = %q, want the handler error", logs.String())
	}
}
