# Alpaca Integration

`alpaca-trigger` watches for Skylight reward redemptions and places a notional
VOO market buy on Alpaca Markets every time a reward is redeemed.

## Setup

```bash
make build-trigger
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `ALPACA_API_KEY` | Yes | — | Alpaca API key ID |
| `ALPACA_API_SECRET` | Yes | — | Alpaca API secret key |
| `ALPACA_BASE_URL` | — | `https://paper-api.alpaca.markets` | **Paper trading by default** |
| `SKYLIGHT_USER_ID` | Yes | — | Skylight user ID |
| `SKYLIGHT_TOKEN` | Yes | — | Skylight API token |
| `SKYLIGHT_FRAME_ID` | Yes | — | Skylight frame ID to watch |
| `POLLER_INTERVAL` | — | `60s` | Poll frequency (Go duration string) |
| `POLLER_STATE_FILE` | — | `~/.skylight/poller-state.json` | Deduplication state |
| `VOO_NOTIONAL` | — | `1.00` | Dollar amount per redemption |

> **Warning:** Set `ALPACA_BASE_URL=https://api.alpaca.markets` only when you intend to place real orders. The default is the paper-trading endpoint.

## Run

```bash
export ALPACA_API_KEY=your_key
export ALPACA_API_SECRET=your_secret
export SKYLIGHT_USER_ID=uid
export SKYLIGHT_TOKEN=tok
export SKYLIGHT_FRAME_ID=fid

./alpaca-trigger
```
