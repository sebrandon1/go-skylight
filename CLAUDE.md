# go-skylight

Go CLI and client library for the [Skylight Calendar](https://www.ourskylight.com/) API. Provides command-line access to manage frames, calendars, chores, rewards, lists, meals, and family member categories on Skylight devices.

## Go Version

Go 1.26.1 (see `go.mod`)

## Dependencies

- `github.com/spf13/cobra` -- CLI framework

## Build / Test / Lint

```bash
make build          # go build -o go-skylight
make test           # go test ./... -v
make lint           # golangci-lint run ./...
make vet            # go vet ./...
make clean          # rm -f go-skylight
```

Run `make lint` before committing and fix any issues.

## Project Structure

```
main.go                 # Entrypoint, calls cmd.Execute()
cmd/                    # Cobra command definitions
  root.go               # Root command, persistent flags (--email, --password, --token, --user-id, --frame-id)
  session.go            # login command
  frame.go              # frame info, devices, avatars, colors
  calendar.go           # calendar list, create, delete, sources
  chore.go              # chore list, create, delete
  reward.go             # reward list, create, delete, redeem, unredeem, points
  list.go               # list all, info, create, delete, add-item, delete-item
  meal.go               # meal categories, recipes, sittings, grocery list
  category.go           # list family member categories
  dashboard.go          # today command (aggregates events, chores, points, meals, lists)
  bounty.go             # bounty create/list (chore + reward pairs)
  rotation.go           # chore rotation create (rotating assignments across members)
  helpers.go            # printJSON utility
lib/                    # API client library
  client.go             # HTTP client, auth, request helpers (get/post/put/patch/delete)
  session.go            # Login (POST /api/sessions)
  structs.go            # All API types and request/response structs
  calendar.go           # Calendar event CRUD, source calendars
  category.go           # List categories
  chore.go              # Chore CRUD (JSON-API format)
  frame.go              # Frame info, devices, avatars, colors
  list.go               # List CRUD, list item CRUD, task box items
  meal.go               # Recipes, meal sittings, meal categories, grocery list
  reward.go             # Reward CRUD, redeem/unredeem, points (JSON-API format)
  bounty.go             # Bounty (chore + reward pair) create and list
  dashboard.go          # Dashboard aggregator (today's events, chores, points, meals, lists)
  rotation.go           # Chore rotation generator (rotating assignments across members)
  *_test.go             # Unit tests using httptest mock servers
```

## Authentication

Two modes:
1. **Email/password:** `--email` + `--password` (auto-login via `POST /api/sessions`)
2. **Token:** `--user-id` + `--token` (skip login, use pre-existing credentials)

Most commands require `--frame-id` to identify the target Skylight frame.

## API Base URL

`https://app.ourskylight.com/api` (set in `lib/client.go` as `SkylightURL`; overridden in tests)

## CI/CD

- GitHub Actions workflow at `.github/workflows/release-binaries.yaml`
- Builds cross-platform binaries (linux/darwin, amd64/arm64) on release publish
- Dependabot configured for dependency updates

## Notes

- The Skylight API uses JSON-API format for chores and rewards; the library flattens these into simpler structs.
- Tests use `httptest.NewServer` with a swapped `SkylightURL` for isolation.
- Do not add `Co-Authored-By` lines to commit messages.
