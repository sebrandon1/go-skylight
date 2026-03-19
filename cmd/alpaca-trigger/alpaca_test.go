package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuyVOOSuccess(t *testing.T) {
	tests := []struct {
		name       string
		notional   string
		respStatus int
		respBody   alpacaOrderResponse
		wantErr    bool
	}{
		{
			name:       "places market buy",
			notional:   "1.00",
			respStatus: http.StatusOK,
			respBody:   alpacaOrderResponse{ID: "ord1", Status: "accepted", Symbol: "VOO"},
		},
		{
			name:       "accepted 201",
			notional:   "5.00",
			respStatus: http.StatusCreated,
			respBody:   alpacaOrderResponse{ID: "ord2", Status: "pending_new", Symbol: "VOO"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost || r.URL.Path != "/v2/orders" {
					t.Errorf("unexpected: %s %s", r.Method, r.URL.Path)
					w.WriteHeader(http.StatusNotFound)
					return
				}
				if r.Header.Get("APCA-API-KEY-ID") == "" {
					t.Error("missing APCA-API-KEY-ID header")
				}
				if r.Header.Get("APCA-API-SECRET-KEY") == "" {
					t.Error("missing APCA-API-SECRET-KEY header")
				}

				var body alpacaOrder
				if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
					t.Errorf("decode: %v", err)
				}
				if body.Symbol != "VOO" {
					t.Errorf("want VOO, got %s", body.Symbol)
				}
				if body.Side != "buy" {
					t.Errorf("want buy, got %s", body.Side)
				}
				if body.Notional != tc.notional {
					t.Errorf("want notional %s, got %s", tc.notional, body.Notional)
				}

				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(tc.respStatus)
				if err := json.NewEncoder(w).Encode(tc.respBody); err != nil {
					t.Errorf("encode: %v", err)
				}
			}))
			defer srv.Close()

			client := NewAlpacaClient(srv.URL, "key", "secret")
			order, err := client.BuyVOO(context.Background(), tc.notional)
			if (err != nil) != tc.wantErr {
				t.Fatalf("wantErr=%v got %v", tc.wantErr, err)
			}
			if tc.wantErr {
				return
			}
			if order.ID != tc.respBody.ID {
				t.Errorf("order.ID = %q, want %q", order.ID, tc.respBody.ID)
			}
		})
	}
}

func TestBuyVOOServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		if _, err := w.Write([]byte(`{"message":"insufficient buying power"}`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	client := NewAlpacaClient(srv.URL, "key", "secret")
	_, err := client.BuyVOO(context.Background(), "1.00")
	if err == nil {
		t.Fatal("expected error for 422 response")
	}
}

func TestBuyVOOConnectionRefused(t *testing.T) {
	client := NewAlpacaClient("http://localhost:1", "key", "secret")
	_, err := client.BuyVOO(context.Background(), "1.00")
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}

func TestBuyVOOInvalidJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`not json`)); err != nil {
			t.Errorf("write: %v", err)
		}
	}))
	defer srv.Close()

	client := NewAlpacaClient(srv.URL, "key", "secret")
	_, err := client.BuyVOO(context.Background(), "1.00")
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestConfigFromEnvMissingAPISecret(t *testing.T) {
	t.Setenv("ALPACA_API_KEY", "key")
	t.Setenv("ALPACA_API_SECRET", "")
	t.Setenv("SKYLIGHT_USER_ID", "uid")
	t.Setenv("SKYLIGHT_TOKEN", "tok")
	t.Setenv("SKYLIGHT_FRAME_ID", "fid")

	_, err := configFromEnv()
	if err == nil {
		t.Error("expected error for missing ALPACA_API_SECRET")
	}
}

func TestConfigFromEnvMissingSkylight(t *testing.T) {
	tests := []struct {
		name    string
		setEnvs map[string]string
	}{
		{
			name: "missing SKYLIGHT_USER_ID",
			setEnvs: map[string]string{
				"ALPACA_API_KEY":    "key",
				"ALPACA_API_SECRET": "secret",
				"SKYLIGHT_USER_ID":  "",
				"SKYLIGHT_TOKEN":    "tok",
				"SKYLIGHT_FRAME_ID": "fid",
			},
		},
		{
			name: "missing SKYLIGHT_TOKEN",
			setEnvs: map[string]string{
				"ALPACA_API_KEY":    "key",
				"ALPACA_API_SECRET": "secret",
				"SKYLIGHT_USER_ID":  "uid",
				"SKYLIGHT_TOKEN":    "",
				"SKYLIGHT_FRAME_ID": "fid",
			},
		},
		{
			name: "missing SKYLIGHT_FRAME_ID",
			setEnvs: map[string]string{
				"ALPACA_API_KEY":    "key",
				"ALPACA_API_SECRET": "secret",
				"SKYLIGHT_USER_ID":  "uid",
				"SKYLIGHT_TOKEN":    "tok",
				"SKYLIGHT_FRAME_ID": "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.setEnvs {
				t.Setenv(k, v)
			}
			_, err := configFromEnv()
			if err == nil {
				t.Error("expected error")
			}
		})
	}
}

func TestConfigFromEnv(t *testing.T) {
	t.Run("missing required vars", func(t *testing.T) {
		t.Setenv("ALPACA_API_KEY", "")
		_, err := configFromEnv()
		if err == nil {
			t.Error("expected error for missing ALPACA_API_KEY")
		}
	})

	t.Run("all vars set", func(t *testing.T) {
		t.Setenv("ALPACA_API_KEY", "key")
		t.Setenv("ALPACA_API_SECRET", "secret")
		t.Setenv("SKYLIGHT_USER_ID", "uid")
		t.Setenv("SKYLIGHT_TOKEN", "tok")
		t.Setenv("SKYLIGHT_FRAME_ID", "fid")
		t.Setenv("POLLER_INTERVAL", "30s")
		t.Setenv("VOO_NOTIONAL", "2.50")

		cfg, err := configFromEnv()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.vooNotional != "2.50" {
			t.Errorf("want notional 2.50, got %s", cfg.vooNotional)
		}
	})

	t.Run("invalid interval", func(t *testing.T) {
		t.Setenv("ALPACA_API_KEY", "key")
		t.Setenv("ALPACA_API_SECRET", "secret")
		t.Setenv("SKYLIGHT_USER_ID", "uid")
		t.Setenv("SKYLIGHT_TOKEN", "tok")
		t.Setenv("SKYLIGHT_FRAME_ID", "fid")
		t.Setenv("POLLER_INTERVAL", "notaduration")

		_, err := configFromEnv()
		if err == nil {
			t.Error("expected error for bad POLLER_INTERVAL")
		}
	})

	t.Run("negative notional", func(t *testing.T) {
		t.Setenv("ALPACA_API_KEY", "key")
		t.Setenv("ALPACA_API_SECRET", "secret")
		t.Setenv("SKYLIGHT_USER_ID", "uid")
		t.Setenv("SKYLIGHT_TOKEN", "tok")
		t.Setenv("SKYLIGHT_FRAME_ID", "fid")
		t.Setenv("POLLER_INTERVAL", "60s")
		t.Setenv("VOO_NOTIONAL", "-1.00")

		_, err := configFromEnv()
		if err == nil {
			t.Error("expected error for negative VOO_NOTIONAL")
		}
	})
}
