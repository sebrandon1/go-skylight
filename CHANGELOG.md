# Changelog

All notable changes to this project will be documented in this file.

## [v0.0.12] - 2026-03-18

### Changed
- Bump `github.com/cpuguy83/go-md2man/v2` from v2.0.6 to v2.0.7
- Bump `github.com/spf13/pflag` from v1.0.9 to v1.0.10

## [1.0.0] - 2026-03-16

### Added
- **Functional options** for `NewClient` / `NewClientWithToken`: `WithBaseURL`, `WithHTTPClient`, `WithLogger`, `WithRateLimit`, `WithRetry`
- **Retry logic** with exponential backoff and full jitter using `crypto/rand`; respects `Retry-After` headers on 429 responses
- **Token-bucket rate limiter** (`golang.org/x/time/rate`) wrapping all outgoing HTTP calls
- **Typed errors**: `AuthError`, `NotFoundError`, `RateLimitError`, `NetworkError` — use `errors.As` to inspect
- **slog-based debug logging** middleware; authorization headers are always redacted
- **`RewardsPoller`** — goroutine-based poller that streams `RedemptionEvent` values for newly redeemed rewards
- **Local JSON deduplication state** for `RewardsPoller` so restarts do not re-fire events
- **Table-driven tests** for all exported library functions
- **`Example_` functions** in `lib/example_test.go` for pkg.go.dev rendering
- **Go version matrix** in CI (`1.25.x` + `1.26.x` × ubuntu + macos)
- `CONTRIBUTING.md` with breaking-change policy and conventional commit format

### Changed
- All resource methods now resolve URLs via `c.effectiveURL()` so `WithBaseURL` is honoured without swapping the package-level `SkylightURL`
- README rewritten with architecture overview, quick-start, and full API reference table

## [v0.0.8] - 2026-03-12

### Added
- `--version` flag with build-time version injection via ldflags
- Govulncheck step in CI pipeline
- SHA256 checksums.txt uploaded as release asset

### Changed
- Consolidated Ubuntu and macOS CI workflows into a single matrix-based workflow
- Release binaries now include version string via ldflags

## [v0.0.7] - 2026-03-10

### Added
- Dashboard command aggregating today's events, chores, points, meals, and lists
- Bounty commands (create chore + paired reward together, list matched pairs)
- Chore rotation command for rotating assignments across family members
- CLAUDE.md project instructions

## [v0.0.6] - 2026-03-10

### Fixed
- Recurring field omitted from JSON when set to false (changed from `omitempty` to pointer)

## [v0.0.5] - 2026-03-09

### Fixed
- Request body format: send flat JSON instead of wrapped objects to match API expectations

## [v0.0.4] - 2026-03-09

### Fixed
- JSON-API response parsing for sessions, chores, and rewards (envelope unwrapping)

## [v0.0.3] - 2026-03-09

### Added
- Chore list filters (date, status, assignee, after, before, include-late)
- Reward create options (emoji-icon, no-respawn, category-ids)
- Auto-login when email/password flags are set
- Client reuse across commands via `getClient()` helper

### Fixed
- Goconst lint: reference `loginCmd.Name()` instead of string literal

## [v0.0.2] - 2026-03-09

### Added
- macOS CI workflow and nightly schedule

## [v0.0.1] - 2026-03-06

### Added
- Initial release with CLI and Go library for Skylight Calendar API
- Session login (email/password authentication)
- Calendar events (list, create, update, delete)
- Source calendars (list)
- Chores (list, create, update, delete)
- Lists and list items (CRUD operations)
- Rewards (list, create, update, delete, redeem, unredeem, points)
- Recipes and meals (CRUD, sittings, grocery list)
- Categories, frame info, devices, avatars, colors
- Comprehensive test coverage for lib and cmd packages
