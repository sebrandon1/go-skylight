# Changelog

All notable changes to this project will be documented in this file.

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
