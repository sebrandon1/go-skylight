package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/sebrandon1/go-skylight/lib"
)

const (
	defaultAlpacaBaseURL = "https://paper-api.alpaca.markets"
	defaultVOONotional   = "1.00"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	cfg, err := configFromEnv()
	if err != nil {
		logger.Error("configuration error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	alpaca := NewAlpacaClient(cfg.alpacaBaseURL, cfg.alpacaAPIKey, cfg.alpacaAPISecret)
	skylightClient, err := lib.NewClientWithToken(cfg.skylightUserID, cfg.skylightToken,
		lib.WithLogger(logger),
		lib.WithRetry(3, 500*time.Millisecond, 10*time.Second),
	)
	if err != nil {
		logger.Error("skylight client error", slog.String("error", err.Error()))
		os.Exit(1)
	}

	poller := lib.NewRewardsPoller(skylightClient, cfg.skylightFrameID, cfg.pollerInterval, cfg.stateFile)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logger.Info("alpaca-trigger starting",
		slog.String("frame_id", cfg.skylightFrameID),
		slog.Duration("interval", cfg.pollerInterval),
		slog.String("alpaca_base_url", cfg.alpacaBaseURL),
		slog.String("voo_notional", cfg.vooNotional),
	)

	poller.Start(ctx)
	defer poller.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("alpaca-trigger shutting down")
			return
		case event := <-poller.Events():
			logger.Info("redemption detected",
				slog.String("reward_id", event.RewardID),
				slog.String("reward_name", event.RewardName),
				slog.String("child_name", event.ChildName),
				slog.Int("points", event.Points),
			)
			order, err := alpaca.BuyVOO(ctx, cfg.vooNotional)
			if err != nil {
				logger.Error("VOO buy failed",
					slog.String("reward_id", event.RewardID),
					slog.String("error", err.Error()),
				)
				continue
			}
			logger.Info("VOO buy placed",
				slog.String("order_id", order.ID),
				slog.String("status", order.Status),
				slog.String("notional", cfg.vooNotional),
			)
		}
	}
}

type triggerConfig struct {
	alpacaBaseURL   string
	alpacaAPIKey    string
	alpacaAPISecret string

	skylightUserID  string
	skylightToken   string
	skylightFrameID string

	pollerInterval time.Duration
	stateFile      string
	vooNotional    string
}

func configFromEnv() (*triggerConfig, error) {
	cfg := &triggerConfig{
		alpacaBaseURL:   getEnvDefault("ALPACA_BASE_URL", defaultAlpacaBaseURL),
		alpacaAPIKey:    os.Getenv("ALPACA_API_KEY"),
		alpacaAPISecret: os.Getenv("ALPACA_API_SECRET"),

		skylightUserID:  os.Getenv("SKYLIGHT_USER_ID"),
		skylightToken:   os.Getenv("SKYLIGHT_TOKEN"),
		skylightFrameID: os.Getenv("SKYLIGHT_FRAME_ID"),

		stateFile:   getEnvDefault("POLLER_STATE_FILE", ""),
		vooNotional: getEnvDefault("VOO_NOTIONAL", defaultVOONotional),
	}

	intervalStr := getEnvDefault("POLLER_INTERVAL", "60s")
	d, err := time.ParseDuration(intervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid POLLER_INTERVAL %q: %w", intervalStr, err)
	}
	cfg.pollerInterval = d

	// Validate required fields.
	if cfg.alpacaAPIKey == "" {
		return nil, fmt.Errorf("ALPACA_API_KEY is required")
	}
	if cfg.alpacaAPISecret == "" {
		return nil, fmt.Errorf("ALPACA_API_SECRET is required")
	}
	if cfg.skylightUserID == "" {
		return nil, fmt.Errorf("SKYLIGHT_USER_ID is required")
	}
	if cfg.skylightToken == "" {
		return nil, fmt.Errorf("SKYLIGHT_TOKEN is required")
	}
	if cfg.skylightFrameID == "" {
		return nil, fmt.Errorf("SKYLIGHT_FRAME_ID is required")
	}

	// Validate notional is a positive number.
	if n, err := strconv.ParseFloat(cfg.vooNotional, 64); err != nil || n <= 0 {
		return nil, fmt.Errorf("VOO_NOTIONAL must be a positive number, got %q", cfg.vooNotional)
	}

	return cfg, nil
}

func getEnvDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
