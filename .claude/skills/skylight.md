---
name: skylight
description: View and manage Skylight chores, rewards, meals, grocery lists, and family activity via the go-skylight CLI
allowed-tools: [Bash, AskUserQuestion]
---

# Skylight Manager

Manage the family Skylight frame using the `skylight` CLI binary. All operations go through the binary — no MCP server required.

## Binary

The binary is `skylight`. If it's not in PATH, use `./skylight` from the repo root after running `make build`.

Config is loaded automatically from `~/.skylight/config` (refresh token, frame ID, etc.).

## Usage

- `/skylight` or `/skylight status` — today's chore dashboard
- `/skylight chores` — list pending chores
- `/skylight rewards` — reward balances and available rewards
- `/skylight grocery` — show the grocery list
- `/skylight analytics` — family activity stats (last 30 days)
- `/skylight home` — weekly combined view of events, tasks, and lists
- `/skylight watch` — live polling for new completions and events

## Discovering Family Members

Always resolve family member names to category IDs dynamically — never hardcode them:

```bash
skylight category -o json
```

Returns an array with `id`, `name`, `color`, and `avatar_url` per member. Use the `id` field when commands require `--assignee-id`.

## Operations

### Status Dashboard (default)

Run `skylight status` for a quick frame overview, then fetch today's chores and points for a per-member view:

```bash
skylight status
skylight chore list --include-late -o json   # all pending + late
skylight reward points -o json               # per-category balances
```

Present grouped per family member:
```
### [Name] (X pts)
- [x] Completed Chore
- [ ] Due Today
- [ ] 2 days late: Late Chore
```

### Chores

**List:**
```bash
skylight chore list [--date YYYY-MM-DD] [--status pending|complete] [--assignee-id ID] [--include-late]
```

**Create:**
```bash
skylight chore create --title "Make Bed" --assignee-id ID --points 1 --date YYYY-MM-DD
skylight chore create --title "Take Out Trash" --assignee-id ID --points 2 --recurring
skylight chore create --title "Free Pick" --up-for-grabs --points 3
```

**Complete / Skip:**
```bash
skylight chore complete --chore-id ID
skylight chore skip --chore-id ID
```

**Update:**
```bash
skylight chore update --chore-id ID [--title] [--points] [--status pending|complete] [--assignee-id] [--date]
```

**Delete:**
```bash
skylight chore delete --chore-id ID
```

When creating interactively: ask for title, who it's assigned to (or up-for-grabs), points (default 1), due date, and recurring vs one-time.

### Rewards

```bash
skylight reward list -o json        # all rewards
skylight reward points -o json      # per-category point balances
skylight reward redeem --reward-id ID
skylight reward unredeem --reward-id ID
```

Cross-reference balances with reward costs to show what each member can afford.

### Grocery

```bash
skylight grocery show               # current grocery list with items
skylight grocery add --title "Milk" --title "Eggs"
skylight grocery add-recipe --recipe-id ID
skylight grocery clear              # remove completed items
skylight grocery organize           # deduplicate and sort by aisle
skylight grocery list               # list all grocery lists (to find list ID)
```

### Meals

```bash
skylight meal recipes               # list all recipes
skylight meal recipe-info --recipe-id ID
skylight meal sittings [--date-min YYYY-MM-DD] [--date-max YYYY-MM-DD]
skylight meal create-sitting --recipe-id ID --date YYYY-MM-DD --meal-category-id ID
skylight meal categories            # list meal categories (breakfast/lunch/dinner)
```

### Analytics

```bash
skylight analytics [--days 30]
```

Shows per-member chore completion rates, top chores, reward stats, and calendar density.

### Home / Weekly View

```bash
skylight home [--no-lists] [--no-tasks]
```

Weekly combined view of calendar events, pending tasks, and active lists.

### Watch (Live Polling)

```bash
skylight watch [--interval 60] [--resources rewards,chores,calendar]
```

Prints new completions and upcoming events as they happen. Ctrl+C to stop.

### Bounty (Chore + Reward Pairs)

```bash
skylight bounty list
skylight bounty create --title "Clean Room" --points 5 --assignee-id ID --reward-title "Extra Screen Time" --date YYYY-MM-DD
```

### Rotation (Rotating Chore Assignments)

```bash
skylight rotation create --chores "Dishes,Vacuum,Laundry" --assignee-ids "ID1,ID2,ID3" --start-date YYYY-MM-DD --weeks 4 --points 2
```

### Export / Import

```bash
skylight export [--resources chores,rewards,lists,recipes,sittings,calendar] [--output-file backup.json]
skylight import --file backup.json [--dry-run] [--resources chores,rewards]
```

## Output Format

Use `-o json` when you need to parse data programmatically. Use `-o table` (or omit `-o`) for human-readable display. Default is `json`.

## Notes

- Chore IDs for recurring instances follow the format `ID-YYYY-MM-DD` or `ID-YYYY-MM-DD-HHMM`
- `--recurring` chores carry over as late if not completed; one-time chores expire at end of day
- Redeemed rewards don't re-appear in the list after redemption
- The `skylight category` command is the source of truth for family member IDs — always fetch fresh rather than hardcoding
