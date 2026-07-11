# CLI Reference

Resource commands are top-level (e.g. `skylight chore list`).

## Global flags

All commands accept these persistent flags:

| Flag | Notes |
|------|--------|
| `--refresh-token` | OAuth2 refresh token (recommended) |
| `--user-id` / `--token` | Bearer access token pair (alternative to refresh token) |
| `--frame-id` | Target frame for resource commands |
| `--config` | Config file path (default `~/.skylight/config`) |
| `--device-fingerprint` | Stable device UUID used with OAuth |
| `--output` / `-o` | `json` (default) or `table` |
| `--quiet` / `-q` | Suppress non-essential success messages |
| `--email` / `--password` | **Deprecated** — use `--refresh-token` instead |

## Config

```bash
skylight config show
skylight config get <key>
skylight config set <key> <value>
skylight config unset <key>
skylight config edit
```

## Login

```bash
skylight login [--save]
```

## Calendar

```bash
skylight calendar list [--start-date DATE] [--end-date DATE]
skylight calendar create --title TITLE --start-at DATETIME [--end-at DATETIME] [--all-day]
skylight calendar update --event-id ID [--title TITLE] [--start-at DATETIME] [--end-at DATETIME]
skylight calendar delete --event-id ID
skylight calendar sources
skylight calendar create-countdown --title TITLE --date DATE
skylight calendar week [--date DATE]
```

## Chores

```bash
skylight chore list [--date DATE] [--assignee-id ID] [--status S] [--after DATE] [--before DATE] [--include-late] [--up-for-grabs] [--week [DATE]]
skylight chore create --title TITLE [--description D] [--points N] [--assignee-id ID] [--date DATE] [--recurring] [--up-for-grabs]
skylight chore update --chore-id ID [--title T] [--description D] [--status S] [--points N] [--assignee-id ID] [--date DATE]
skylight chore delete --chore-id ID
skylight chore complete --chore-id ID
skylight chore skip --chore-id ID
skylight chore claim --chore-id ID --assignee-id ID
skylight chore streak [--days N]
```

## Rewards

```bash
skylight reward list
skylight reward create --title TITLE --points N [--emoji-icon EMOJI] [--no-respawn] [--category-ids 1,2]
skylight reward update --reward-id ID [--title T] [--points N] [--emoji-icon EMOJI] [--no-respawn] [--category-ids 1,2]
skylight reward delete --reward-id ID
skylight reward redeem   --reward-id ID
skylight reward unredeem --reward-id ID
skylight reward points
skylight reward remove-stars --category-id ID --points N
```

## Lists

```bash
skylight list all
skylight list info       --list-id ID
skylight list create     --title TITLE [--color COLOR] [--hide-from-frame]
skylight list update     --list-id ID [--title T] [--color C] [--hide-from-frame]
skylight list delete     --list-id ID
skylight list add-item   --list-id ID --title TITLE [--position N]
skylight list update-item --list-id ID --item-id ITEM_ID [--title T] [--completed] [--position N]
skylight list delete-item --list-id ID --item-id ITEM_ID
skylight list clear-completed --list-id ID
skylight list task-box-item --title TITLE [--date DATE]
```

## Meals

```bash
skylight meal categories
skylight meal recipes
skylight meal recipe-info --recipe-id ID
skylight meal create-recipe --title TITLE [--description D] [--ingredients a,b] [--url URL] [--meal-category-id ID]
skylight meal update-recipe --recipe-id ID [--title T] [--description D] [--ingredients a,b] [--url URL]
skylight meal delete-recipe --recipe-id ID
skylight meal sittings [--date-min DATE] [--date-max DATE]
skylight meal get-sitting --sitting-id ID
skylight meal create-sitting --recipe-id ID --date DATE [--summary S] [--meal-category-id ID]
skylight meal delete-sitting --sitting-id ID [--date DATE]
skylight meal sitting-recipe --sitting-id ID
skylight meal add-to-grocery --recipe-id ID
skylight meal plan --recipes ID,ID --start-date DATE [--categories ID,ID]
```

## Photos

```bash
skylight photo list [--page-token TOKEN]
skylight photo upload --file PATH [--caption TEXT]
skylight photo delete --message-id ID [--message-id ID ...]
skylight photo download [--message-id ID ...] [--all] [--output-dir DIR]
```

## Categories (Profiles & Labels)

```bash
skylight category list
skylight category create --name NAME [--color COLOR]
skylight category update --category-id ID [--name NAME] [--color COLOR]
skylight category delete --category-id ID
```

## Frame

```bash
skylight frame list
skylight frame info
skylight frame devices
skylight frame avatars
skylight frame colors
skylight frame set-album --album-id ID   # -1 for all photos
```

## Add-ons

```bash
skylight addon list
```

## Bounties & Rotations

```bash
skylight bounty create --title TITLE --points N --reward-title R [--assignee-id ID] [--due-date DATE] [--emoji-icon EMOJI] [--recurring]
skylight bounty list
skylight bounty update --chore-id ID --reward-id ID [--title T] [--reward-title R] [--points N] [--due-date DATE] [--emoji-icon EMOJI]
skylight bounty delete --chore-id ID --reward-id ID

skylight rotation create --chores "Dishes,Vacuum" --assignee-ids "id1,id2" \
    --start-date DATE --weeks N --points N
```

## Routines

```bash
skylight routine list
skylight routine create --title TITLE [--assignee-id ID] [--steps "Step 1,Step 2"]
skylight routine update --routine-id ID [--title T] [--assignee-id ID] [--steps "..."]
skylight routine delete --routine-id ID
skylight routine reorder --routine-ids "id1,id2,id3"
```

## Grocery

```bash
skylight grocery list
skylight grocery show     --list-id ID
skylight grocery create   --title TITLE
skylight grocery add      --list-id ID --items "Milk,Eggs"
skylight grocery add-recipe --recipe-id ID
skylight grocery organize --list-id ID
skylight grocery order    --list-id ID --retailer costco
skylight grocery clear    --list-id ID
```

## Status, Home, Analytics & Watch

```bash
skylight status                     # quick overview of the connected frame
skylight home [--no-tasks] [--no-lists] [--no-meals]   # weekly combined view of events, tasks, lists, meals
skylight analytics [--days N]       # family activity statistics over a time period
skylight watch [--interval SECONDS] [--resources rewards,chores,calendar,lists,routines] [--persist]
```

## Export & Import

```bash
skylight export [--output-file PATH] [--resources chores,rewards,lists,recipes,sittings,calendar] [--days N]
skylight import --file PATH [--dry-run] [--resources all]
```
