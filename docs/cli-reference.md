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
skylight calendar get --event-id ID
skylight calendar create --title TITLE --start-at DATETIME [--end-at DATETIME] [--all-day] [--category-id ID] [--color COLOR]
skylight calendar update --event-id ID [--title TITLE] [--start-at DATETIME] [--end-at DATETIME] [--all-day] [--category-id ID] [--color COLOR]
skylight calendar delete --event-id ID
skylight calendar day [--date DATE]           # default: today
skylight calendar sources
skylight calendar source-enable --source-id ID
skylight calendar source-disable --source-id ID
skylight calendar create-countdown --title TITLE --date DATE
skylight calendar week [--date DATE]
```

## Chores

```bash
skylight chore list [--date DATE] [--assignee-id ID] [--status S] [--after DATE] [--before DATE] [--include-late] [--up-for-grabs] [--week [DATE]]
skylight chore get --chore-id ID
skylight chore search --query TERM [--assignee-id ID] [--status S] [--after DATE] [--before DATE]
skylight chore create --title TITLE [--description D] [--points N] [--assignee-id ID] [--date DATE] [--recurring] [--up-for-grabs] \
    [--frequency daily|weekly|monthly] [--interval N] [--recurrence-days mon,wed] [--end-date DATE] [--recur-from scheduled|completed]
skylight chore update --chore-id ID [--title T] [--description D] [--status S] [--points N] [--assignee-id ID] [--date DATE] [--up-for-grabs] \
    [--frequency daily|weekly|monthly] [--interval N] [--recurrence-days mon,wed] [--end-date DATE] [--recur-from scheduled|completed]
skylight chore delete --chore-id ID
skylight chore complete --chore-id ID
skylight chore skip --chore-id ID [--defer-until DATE]
skylight chore claim --chore-id ID --assignee-id ID
skylight chore streak [--days N] [--assignee-id ID]
```

> **Note:** The Skylight API requires a date window for every chore query.
> When `--after`/`--before` are omitted, go-skylight supplies one automatically:
> `--date DATE` becomes a same-day window, `--up-for-grabs` uses the next 7 days,
> and bare `chore list` / `chore search` use the current calendar month. When
> passed explicitly, `--after` and `--before` must be used together.

## Rewards

```bash
skylight reward list
skylight reward get --reward-id ID
skylight reward create --title TITLE --points N [--emoji-icon EMOJI] [--no-respawn] [--category-ids 1,2]
skylight reward update --reward-id ID [--title T] [--points N] [--emoji-icon EMOJI] [--no-respawn] [--category-ids 1,2]
skylight reward delete --reward-id ID
skylight reward redeem   --reward-id ID
skylight reward unredeem --reward-id ID
skylight reward points
skylight reward remove-stars --assignee-id ID --points N
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
skylight list delete-section --list-id ID --section-id SECTION_ID [--dry-run] [--yes]
skylight list reorder-item  --list-id ID --item-id ITEM_ID --position N
skylight list clear-completed --list-id ID
skylight list task-box-item --title TITLE
```

## Meals

```bash
skylight meal categories
skylight meal create-category --name NAME [--color COLOR]
skylight meal update-category --category-id ID [--name NAME] [--color COLOR]
skylight meal delete-category --category-id ID [--dry-run] [--yes]
skylight meal recipes [--title FILTER]
skylight meal recipe-info --recipe-id ID
skylight meal create-recipe --title TITLE [--description D] [--ingredients a,b] [--url URL] [--meal-category-id ID]
skylight meal update-recipe --recipe-id ID [--title T] [--description D] [--ingredients a,b] [--url URL]
skylight meal delete-recipe --recipe-id ID [--dry-run] [--yes]
skylight meal sittings [--date-min DATE] [--date-max DATE]
skylight meal get-sitting --sitting-id ID
skylight meal create-sitting --recipe-id ID --date DATE [--summary S] [--meal-category-id ID]
skylight meal update-sitting --sitting-id ID [--date DATE] [--summary S] [--meal-category-id ID] [--recipe-id ID]
skylight meal delete-sitting --sitting-id ID [--date DATE] [--dry-run] [--yes]
skylight meal sitting-recipe --sitting-id ID
skylight meal add-to-grocery --recipe-id ID
skylight meal plan --recipes ID,ID --start-date DATE [--categories ID,ID]
```

## Photos

```bash
skylight photo list [--page-token TOKEN]
skylight photo upload --file PATH [--caption TEXT]
skylight photo delete --photo-id ID [--photo-id ID ...]
skylight photo download [--photo-id ID ...] [--all] [--output-dir DIR]
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

A routine is a recurring chore with a fixed time-of-day slot (morning, afternoon, or evening) rather than a separate resource. `routine list` shows routines active or starting in the next 30 days.

```bash
skylight routine list [--assignee-id ID]
skylight routine get --routine-id ID
skylight routine create --title TITLE --time-of-day morning|afternoon|evening --category-id ID --start-date DATE
skylight routine update --routine-id ID [--title T] [--time-of-day morning|afternoon|evening]
skylight routine delete --routine-id ID
```

## Grocery

```bash
skylight grocery list
skylight grocery show     --list-id ID
skylight grocery create   --title TITLE
skylight grocery delete   --list-id ID [--dry-run] [--yes]
skylight grocery add      --list-id ID --items "Milk,Eggs"
skylight grocery add-recipe --recipe-id ID
skylight grocery update-item --list-id ID --item-id ITEM_ID [--title T] [--completed]
skylight grocery delete-item --list-id ID --item-id ITEM_ID [--dry-run] [--yes]
skylight grocery organize --list-id ID
skylight grocery order    --list-id ID --retailer costco
skylight grocery clear    --list-id ID
```

## Status, Home, Analytics & Watch

```bash
skylight status                     # quick overview of the connected frame
skylight home [--no-tasks] [--no-lists] [--no-meals] [--no-routines]   # weekly combined view of events, tasks, lists, meals, routines
skylight analytics [--days N] [--start-date DATE] [--end-date DATE]    # date range is mutually exclusive with --days
skylight watch [--interval SECONDS] [--resources rewards,chores,calendar,lists,routines,meals,photos] [--persist]
```

## Templates

Templates are stored locally in `~/.skylight/templates/` as JSON files. Use `template save` to capture the current frame's chores and rewards, then `template apply` to recreate them on any frame.

```bash
skylight template save  --name NAME [--resources chores,rewards]
skylight template apply --name NAME [--start-date DATE] [--resources chores,rewards] [--dry-run]
skylight template list
skylight template delete --name NAME
```

## Export & Import

```bash
skylight export [--output-file PATH] [--resources chores,rewards,lists,recipes,sittings,calendar,routines,bounties,categories,photos] [--days N]
skylight import --file PATH [--dry-run] [--resources all]
```

## Shell Completion

Cobra auto-generates shell completion scripts for bash, zsh, and fish. Enable completion once by adding the appropriate line to your shell profile:

```bash
# bash (~/.bashrc or ~/.bash_profile)
source <(skylight completion bash)

# zsh (~/.zshrc) — also enable compinit if not already done
source <(skylight completion zsh)

# fish (~/.config/fish/config.fish)
skylight completion fish | source
```

After sourcing, tab-complete commands, subcommands, and enum flags like `--output` and `--status`.
