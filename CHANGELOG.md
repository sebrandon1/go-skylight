# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Changed
- `status` counts list items from `ListLists` instead of one `GetList` per list; its JSON output no longer has `list_errors` (#494)

### Added
- `ListItem`/`ListItemData`: `Section`, read from the API and sent by `AddListItem` and `UpdateListItem`

### Fixed
- `export`: lists were saved without their items; `ListLists` now returns each list's items, so the `list all`, `grocery list` and `home` ITEMS column is no longer always 0 and their JSON output includes `list_items` (#493, #494)
- `import`: list items are recreated in their original order with their completed status and section (#493, #494)

## [v0.2.5] - 2026-09-30

### Changed
- Default API version bumped from `2026-06-01` to `2026-08-05` to support upcoming linked tasks feature (multi-assignee chores) (#491)

## [v0.2.4] - 2026-09-30

### Added
- `Chore`/`ChoreData`: `StartTime` (due time, `HH:MM`) and `EmojiIcon`, read from the API and sent on create/update

### Fixed
- Integration tests: add refresh token authentication to bypass Cloudflare blocks; remove CI integration workflow as tests require personal credentials (#487)
- `import`: recurring chores were recreated as one one-off chore per exported occurrence; now each series is recreated once from its rule with its due time, icon, description and up-for-grabs flag. Ended series and completed one-offs are skipped, and the `--dry-run` count matches what is created (#488, #489)
- `export`: up-for-grabs chores were missing, and chore due time and icon were not captured; now included (#488, #489)

## [v0.2.3] - 2026-09-29

### Added
- `calendar schedule` — agenda-style view of upcoming events with date grouping and time formatting (#473, #486)
- `frame list-albums` — list all photo albums available on the frame (#485)

### Fixed
- `chore create`/`chore update`: recurrence flags (`--frequency`, `--recurrence-days`, `--interval`, `--end-date`) were silently ignored; now properly sent as an RRULE via the `create_multiple` endpoint (#481)
- `GetChore` rewritten to use the list endpoint with date-window filter instead of searching the full month; composite instance IDs now extract the embedded date for the query window (#482, #484)
- `DeleteChore` integration tests now properly sweep recurring chores with `DeleteRecurringChore` instead of leaving orphaned instances (#483)
- `--recur-from` documented as having no effect (the API does not store it) (#481)

## [v0.2.2] - 2026-09-21

### Fixed
- `DeleteChore` split into `DeleteChore` (one-time) and `DeleteRecurringChore` (recurring) — the Skylight API now rejects `apply_to=all` on non-recurring chores (#475)
- Bump default `skylight-api-version` header from `2026-03-01` to `2026-06-01` to support deletion of Up for Grabs chores (#476)
- Integration test cleanup failures now surface as `t.Errorf` instead of silent `t.Logf`; `IsNotFound` typed checks replace fragile string-contains error matching (#475)

### Added
- Integration test pre-run sweep (`TestMain`) removes `integration-test-*` artifacts left by interrupted runs (#475)

### Changed
- Go version updated to 1.27.1 (`go.mod`, CI, Dockerfile) (#472)
- Dependency: `golang.org/x/time` bumped to v0.16.0 (#474)

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
