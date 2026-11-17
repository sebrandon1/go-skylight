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
skylight config get <key> [--reveal]   # sensitive values are masked unless --reveal
skylight config set <key> <value>
skylight config unset <key>
skylight config edit
```

## Config file keys

| Key | Description |
|-----|-------------|
| `SKYLIGHT_EMAIL` | User email address (deprecated; use refresh token) |
| `SKYLIGHT_PASSWORD` | User password (deprecated; use refresh token) |
| `SKYLIGHT_TOKEN` | Bearer access token |
| `SKYLIGHT_USER_ID` | User ID (used with `SKYLIGHT_TOKEN`) |
| `SKYLIGHT_FRAME_ID` | Default frame ID |
| `SKYLIGHT_REFRESH_TOKEN` | OAuth2 refresh token (recommended) |
| `SKYLIGHT_DEVICE_FINGERPRINT` | Stable device UUID used with OAuth |
| `SKYLIGHT_OUTPUT` | Output format (`json` or `table`) |
| `SKYLIGHT_QUIET` | Suppress non-essential success messages (`true`/`false`) |

Each key can also be set as an environment variable of the same name.

> **Note:** Settings resolve in this order: CLI flag, environment variable, config file. `SKYLIGHT_REFRESH_TOKEN` is the exception. A token already saved in the config file beats the environment value, because refresh tokens are single-use and each run saves its replacement to the file.
>
> A container with only `SKYLIGHT_REFRESH_TOKEN` and no persistent config can run once, because that token is spent on the first run. Mount `~/.skylight` so the replacement token is kept.

## Login

```bash
skylight login [--save]
```

With `--save`, tokens go to the config file and are not printed. If the save fails, the tokens are printed so the session isn't lost, and the command exits non-zero. Without `--save`, the tokens are printed.

## Calendar

```bash
skylight calendar list [--start-date DATE] [--end-date DATE]  # default: this month; end date is exclusive
skylight calendar get --event-id ID
skylight calendar create --title TITLE --start-at DATETIME [--end-at DATETIME] [--all-day] [--category-id ID] [--color COLOR]
skylight calendar update --event-id ID [--title TITLE] [--start-at DATETIME] [--end-at DATETIME] [--all-day] [--category-id ID] [--color COLOR]
skylight calendar delete --event-id ID [--dry-run] [--yes]
skylight calendar day [--date DATE]           # default: today
skylight calendar sources
skylight calendar source-enable --source-id ID
skylight calendar source-disable --source-id ID
skylight calendar create-countdown --title TITLE --date DATE
skylight calendar week [--date DATE]
skylight calendar schedule [--date DATE] [--days N]   # agenda view; default: today (UTC), 3 days
```

## Chores

```bash
skylight chore list [--date DATE] [--assignee-id ID] [--status S] [--after DATE] [--before DATE] [--include-late] [--up-for-grabs] [--week [DATE]]
skylight chore get --chore-id ID
skylight chore search --query TERM [--assignee-id ID] [--status S] [--after DATE] [--before DATE]
skylight chore create --title TITLE [--description D] [--points N] [--assignee-id ID | --assignee-ids ID1,ID2] [--date DATE] [--due-time HH:MM] [--recurring] [--up-for-grabs] \
    [--frequency daily|weekly|monthly] [--interval N] [--recurrence-days mon,wed] [--end-date DATE]
skylight chore update --chore-id ID [--title T] [--description D] [--status S] [--points N] [--assignee-id ID] [--assignee-ids ID1,ID2] [--apply-to-profiles all|one] [--date DATE] [--due-time HH:MM] [--up-for-grabs] \
    [--frequency daily|weekly|monthly] [--interval N] [--recurrence-days mon,wed] [--end-date DATE]
skylight chore notifications show
skylight chore notifications set [--task-due on|off] [--due-category-id ID,...] [--due-up-for-grabs on|off] \
    [--task-completed on|off] [--completed-category-id ID,...] [--completed-up-for-grabs on|off] [--dry-run]
skylight chore delete --chore-id ID [--recurring | --apply-to-profiles all|one] [--dry-run] [--yes]   # --recurring deletes every instance
skylight chore complete --chore-id ID
skylight chore skip --chore-id ID [--defer-until DATE]
skylight chore claim --chore-id ID --assignee-id ID
skylight chore streak [--days N] [--assignee-id ID]
skylight chore all [--limit N] [--include-up-for-grabs=true|false]   # All Tasks view: one entry per series, bucketed late/today/today_timed/any_day/future; aliases: skylight task all, skylight chore all-tasks; --limit must be >= 0
```

> **Note:** The Skylight API requires a date window for every chore query.
> When `--after`/`--before` are omitted, go-skylight supplies one automatically:
> `--date DATE` becomes a same-day window, `--up-for-grabs` uses the next 7 days,
> and bare `chore list` / `chore search` use the current calendar month. When
> passed explicitly, `--after` and `--before` must be used together.
>
> **Due times and task notifications.** `--due-time` sets the chore's time of day
> (`HH:MM`, 24-hour, `00:00`-`23:59`; anything else is rejected before the request
> because the API silently accepts junk). There is no per-task reminder: task
> notifications are a frame-wide setting, and a "task due" notification fires for
> tasks that have a due time. `chore notifications show` prints the current
> settings; `chore notifications set` changes only the flags you pass, and a rule
> you do not mention is not sent. `--due-category-id` and
> `--completed-category-id` take family member IDs from `skylight profile list`
> (repeatable or comma-separated; pass an empty value to clear the filter).
> With `--output table` the due time shows in a DUE TIME column, and
> `chore notifications show -o table` shows `-` when a rule has no category
> filter.
>
> Notification rule behavior (confirmed against the live API):
>
> - A disabled rule is sent without filters, because the API rejects filters on a
>   disabled rule (422) and clears them. Disabling a rule therefore does not
>   preserve its category filter or up-for-grabs setting.
> - Enabling a rule needs at least one category ID or up-for-grabs. Otherwise the
>   command fails locally, because the API would silently store the rule as
>   disabled.
> - Changing filters (`--*-category-id`, `--*-up-for-grabs`) on a rule that stays
>   disabled fails locally; pass `--task-due on` or `--task-completed on` in the
>   same command.
>
> On `chore update`, `--frequency`/`--recurrence-days` replace the whole schedule,
> so pass `--interval` and `--end-date` again if the chore uses them.
>
> **Linked tasks.** `chore create --assignee-ids A,B` (two or more profile IDs) creates one
> task linked across those profiles and prints every copy; the `LINKED` table column and the
> `linked_task_group_id` JSON field show the shared group. Linking needs API version
> `2026-08-05` or later (the default is newer); on older versions the API silently creates
> unlinked copies. `chore update --apply-to-profiles all --date DATE` edits every linked copy
> (`--date` is the instance date); add `--assignee-ids` with the full profile list to link
> another profile or drop one (dropping deletes that profile's copy). `--apply-to-profiles one`
> edits only that chore and detaches it from the group. `chore delete --apply-to-profiles
> all|one` removes the task for everyone or just that profile's copy; a plain delete removes it
> for everyone. Detaching one copy from a two-copy group dissolves the group, so the other
> copy ends up unlinked too. `--assignee-id` and `--assignee-ids` are mutually exclusive.
>
> Constraints: on create, `--assignee-ids` cannot be combined with `--assignee-id` or `--up-for-grabs`; on delete, `--apply-to-profiles` cannot be combined with `--recurring`; on update, `--assignee-ids` requires `--apply-to-profiles all`.

## Rewards

```bash
skylight reward list [--assignee-id ID] [--points-min N] [--points-max N] [--status redeemed|available]
skylight reward get --reward-id ID
skylight reward create --title TITLE --points N [--emoji-icon EMOJI] [--no-respawn] [--category-ids 1,2]
skylight reward update --reward-id ID [--title T] [--points N] [--emoji-icon EMOJI] [--no-respawn] [--category-ids 1,2]
skylight reward delete --reward-id ID [--dry-run] [--yes]
skylight reward redeem   --reward-id ID [--dry-run] [--yes]
skylight reward unredeem --reward-id ID [--dry-run] [--yes]
skylight reward points
skylight reward remove-stars --assignee-id ID --points N [--dry-run] [--yes]
```

## Lists

```bash
skylight list all
skylight list info       --list-id ID
skylight list create     --title TITLE [--color COLOR] [--hide-from-frame]
skylight list update     --list-id ID [--title T] [--color C] [--hide-from-frame]
skylight list delete     --list-id ID [--dry-run] [--yes]
skylight list add-item   --list-id ID --title TITLE [--position N]
skylight list update-item --list-id ID --item-id ITEM_ID [--title T] [--completed] [--position N]
skylight list delete-item --list-id ID --item-id ITEM_ID [--dry-run] [--yes]
skylight list delete-section --list-id ID --section-id SECTION_ID [--dry-run] [--yes]
skylight list reorder-item  --list-id ID --item-id ITEM_ID --position N
skylight list clear-completed --list-id ID [--dry-run] [--yes]
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
skylight meal sittings [--date-min DATE] [--date-max DATE]  # default: this month; end date is exclusive
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
skylight photo list [--page-token TOKEN] [--limit N]
skylight photo upload --file PATH [--caption TEXT]
skylight photo delete --photo-id ID [--photo-id ID ...] [--dry-run] [--yes]
skylight photo download [--photo-id ID ...] [--all] [--output-dir DIR]
```

## Categories (Profiles & Labels)

```bash
skylight category list
skylight category create --name NAME [--color COLOR]
skylight category update --category-id ID [--name NAME] [--color COLOR]
skylight category delete --category-id ID [--dry-run] [--yes]
```

## Profiles

```bash
skylight profile list
skylight profile create --name NAME [--color COLOR]
skylight profile update --profile-id ID [--name NAME] [--color COLOR]
skylight profile delete --profile-id ID [--dry-run] [--yes]
```

## Labels

```bash
skylight label list
skylight label create --name NAME [--color COLOR]
skylight label update --label-id ID [--name NAME] [--color COLOR]
skylight label delete --label-id ID [--dry-run] [--yes]
```

## Messages

Household text messages (Calendar Plus). `--to` takes family member category IDs from `skylight profile list`.

```bash
skylight message list [--start-date YYYY-MM-DD] [--end-date YYYY-MM-DD]   # default: today + 30 days
skylight message send --text TEXT --to ID[,ID...] [--send-at YYYY-MM-DDTHH:MM] [--repeat RRULE]
skylight message update --message-id ID [--text TEXT] [--to ID[,ID...]] [--send-at TIME] [--repeat RRULE]
skylight message delete --message-id ID [--dry-run] [--yes]
```

- With no `--send-at`, the message is delivered now. `--send-at` takes local time (`2026-11-07T08:00`) or RFC 3339.
- `--text` is capped at 220 characters, the same as the app.
- `message send` shows the message on every device on the frame. `message update` keeps the devices the message already targets.
- `--repeat` takes an RFC 5545 rule without the `RRULE:` prefix (e.g. `FREQ=WEEKLY;BYDAY=SA`) and needs `--send-at`.
- `list` prints one row per delivery, so a recurring message shows once for each occurrence, all with the same ID.
- Recurring deliveries keep a fixed UTC time, so the local delivery time moves an hour across a daylight saving change. This is server behavior.

## Frame

```bash
skylight frame list
skylight frame info
skylight frame devices
skylight frame tabs [--device-id ID]     # which navigation tabs each device shows
skylight frame avatars
skylight frame colors
skylight frame list-albums
skylight frame set-album --album-id ID   # -1 for all photos; use list-albums to find IDs
skylight frame update [--screensaver-show-weather=true|false] [--screensaver-show-events=true|false]
                      [--hide-tabs TABS] [--show-tabs TABS] [--device-id ID]
```

Show and hide tabs:

- Tabs: `home`, `calendar`, `lists`, `tasks`, `rewards`, `meals`, `recipes`, `photos`, `sleep`. Pass several as a comma-separated list, e.g. `--hide-tabs meals,recipes`.
- Tab visibility is set per device, not per frame. `--device-id` is needed only when the frame has more than one device; `frame devices` lists their IDs.
- Hiding a tab does not delete any of its data. Tabs you don't name are left unchanged.

## Add-ons

```bash
skylight addon list
```

## Bounties & Rotations

```bash
skylight bounty create --title TITLE --points N --reward-title R [--assignee-id ID] [--due-date DATE] [--emoji-icon EMOJI] [--recurring]
skylight bounty list [--after DATE] [--before DATE]
skylight bounty update --chore-id ID --reward-id ID [--title T] [--reward-title R] [--points N] [--due-date DATE] [--emoji-icon EMOJI]
skylight bounty delete (--chore-id ID --reward-id ID | --all) [--after DATE] [--before DATE] [--dry-run] [--yes]

skylight rotation create --chores "Dishes,Vacuum" --assignee-ids "id1,id2" \
    --start-date DATE --weeks N --points N
```

## Routines

A routine is a recurring chore with a fixed time-of-day slot (morning, afternoon, or evening) rather than a separate resource. `routine list` shows routines active or starting in the next 30 days.

```bash
skylight routine list [--assignee-id ID]
skylight routine habits [--assignee-id ID]
skylight routine get --routine-id ID
skylight routine create --title TITLE --time-of-day morning|afternoon|evening --category-id ID --start-date DATE [--emoji-icon E] [--points N] [--track-habit]
skylight routine update --routine-id ID [--title T] [--time-of-day morning|afternoon|evening] [--emoji-icon E] [--points N] [--track-habit=true|false]
skylight routine delete --routine-id ID [--dry-run] [--yes]
```

`--points` sets the stars awarded each time the routine is done and must be greater than 0. `--track-habit=false` removes the routine's habit tracker; turning it back on starts a new one.

`routine habits` lists only routines with habit tracking on, adding the current streak, longest streak, and progress for the current time range (e.g. `2/2 this week`). In JSON output each routine carries a `habit` object (`current_streak`, `longest_streak`, `time_range_progress`); `routine list` includes it too. There is no habits endpoint: the data comes from the `habit_tracker` record the chores list returns for a tracked routine. With `--output table`, `routine habits` adds STREAK, LONGEST and PROGRESS columns, and so does `routine list` whenever any routine tracks a habit.

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
skylight template delete --name NAME [--yes]
```

## Export & Import

```bash
skylight export [--output-file PATH] [--resources chores,rewards,lists,recipes,sittings,calendar,routines,bounties,categories,photos] [--days N]
skylight import --file PATH [--dry-run] [--resources all] [--skip-existing]
```

`--dry-run` changes nothing. Combined with `--skip-existing`, it reads the frame and reports how many items would be created and how many are already there.

`--skip-existing` skips resources already on the frame, so re-running an import does not create duplicates. Resources match by title, and meal sittings and calendar events match by title and date. If any lookup fails, the import stops before creating anything. Photos are the exception: they are always uploaded again.

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
