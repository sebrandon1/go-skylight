// Package main provides the alpaca-trigger binary, which watches for Skylight
// reward redemptions and places a notional VOO market buy on Alpaca Markets.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// alpacaOrder is the request body for POST /v2/orders.
type alpacaOrder struct {
	Symbol      string `json:"symbol"`
	Notional    string `json:"notional"`
	Side        string `json:"side"`
	Type        string `json:"type"`
	TimeInForce string `json:"time_in_force"`
}

// alpacaOrderResponse is the subset of fields we care about from Alpaca.
type alpacaOrderResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
	Symbol string `json:"symbol"`
}

// AlpacaClient places orders via the Alpaca v2 REST API using key-based auth.
type AlpacaClient struct {
	BaseURL    string
	APIKey     string
	APISecret  string
	HTTPClient *http.Client
}

// NewAlpacaClient returns a client pointed at baseURL (use the paper-trading
// URL by default to prevent accidental live orders).
func NewAlpacaClient(baseURL, apiKey, apiSecret string) *AlpacaClient {
	return &AlpacaClient{
		BaseURL:    baseURL,
		APIKey:     apiKey,
		APISecret:  apiSecret,
		HTTPClient: &http.Client{Timeout: 15 * time.Second},
	}
}

// BuyVOO places a notional market buy order for VOO.
// notional is a dollar amount string, e.g. "1.00".
func (a *AlpacaClient) BuyVOO(ctx context.Context, notional string) (*alpacaOrderResponse, error) {
	order := alpacaOrder{
		Symbol:      "VOO",
		Notional:    notional,
		Side:        "buy",
		Type:        "market",
		TimeInForce: "day",
	}

	body, err := json.Marshal(order)
	if err != nil {
		return nil, fmt.Errorf("marshal order: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.BaseURL+"/v2/orders", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("APCA-API-KEY-ID", a.APIKey)
	req.Header.Set("APCA-API-SECRET-KEY", a.APISecret)
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("alpaca API error %d: %s", resp.StatusCode, string(respBody))
	}

	var orderResp alpacaOrderResponse
	if err := json.Unmarshal(respBody, &orderResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return &orderResp, nil
}
