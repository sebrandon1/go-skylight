# Examples

Real-world scenarios showing how to use the Skylight CLI for common household tasks.

## Getting started

### 1. Install

```bash
go install github.com/sebrandon1/go-skylight@latest
```

### 2. Log in and save your credentials

```bash
skylight login --email you@example.com --password yourpassword --save
```

This saves your credentials to `~/.skylight/config` so you don't need to pass flags on every command.

### 3. Find your frame ID

```bash
skylight frame list
```

```json
[
  { "id": "3136444", "name": "Kitchen Frame", "timezone": "America/Chicago" }
]
```

Save it to your config so you never have to type it again:

```bash
skylight config set SKYLIGHT_FRAME_ID 3136444
```

From here every command automatically uses this frame. You're ready.

---

## Profiles and labels

### Delete a profile

The Skylight app only lets you merge or convert profiles — it won't let you delete one outright. The CLI can.

First, list your profiles to get the ID of the one you want to remove:

```bash
skylight category list
```

```json
[
  { "id": "6750875", "name": "Team Blah Blah", "color": "#FF5733" },
  { "id": "6750876", "name": "Emma", "color": "#33A1FF" }
]
```

Then delete by ID:

```bash
skylight category delete --category-id 6750875
```

> **Warning:** Deleting a profile also deletes all chores and rewards assigned exclusively to it. If you have tasks assigned to multiple profiles, those tasks remain for the other assignees.

### Delete a label

Labels are managed through the same `category` commands (the API treats profiles and labels as the same resource type):

```bash
# List to find the label ID
skylight category list

# Delete it
skylight category delete --category-id LABEL_ID
```

### Rename a profile or label

```bash
skylight category update --category-id 6750875 --name "Saturday Crew" --color "#22C55E"
```

### Create a new profile

```bash
skylight category create --name "Alex" --color "#A855F7"
```

---

## Chores

### List today's chores

```bash
skylight chore list --status pending
```

### Create a chore and assign it

```bash
# First get the category ID of the person to assign it to
skylight category list

# Create the chore
skylight chore create \
  --title "Take out trash" \
  --points 5 \
  --assignee-id 6750876
```

### Mark a chore complete

```bash
# List chores to get the chore ID
skylight chore list

# Complete it
skylight chore complete --chore-id CHORE_ID --assignee-id 6750876
```

### Put a chore up for grabs

Up-for-grabs chores are unassigned and claimable by any family member.

```bash
# Create an unassigned chore (no --assignee-id)
skylight chore create --title "Vacuum living room" --points 10

# List all up-for-grabs chores
skylight chore list --up-for-grabs

# Claim one
skylight chore claim --chore-id CHORE_ID --assignee-id YOUR_CATEGORY_ID
```

### Delete a chore you no longer need

```bash
skylight chore delete --chore-id CHORE_ID
```

---

## Rewards

### List rewards and current star balances

```bash
skylight reward list
skylight reward points
```

### Create a reward

```bash
skylight reward create --title "Movie night pick" --points 50 --emoji-icon 🎬
```

### Redeem a reward for a family member

```bash
# Get the reward ID from `reward list`, category ID from `category list`
skylight reward redeem --reward-id REWARD_ID --category-id CATEGORY_ID
```

---

## Calendar

### See this week's events

```bash
skylight calendar list --start-date $(date +%Y-%m-%d)
```

### Add an event

```bash
skylight calendar create \
  --title "Soccer practice" \
  --start-at "2026-06-28T15:00:00" \
  --end-at "2026-06-28T16:30:00" \
  --category-id 6750876
```

### Delete an event

```bash
skylight calendar delete --event-id EVENT_ID
```

---

## Backup and restore

### Export everything to a file

```bash
skylight export --output-file skylight-backup.json
```

This exports chores, rewards, lists, recipes, meal sittings, and calendar events for a 90-day window around today. Useful before making bulk changes.

### Restore from a backup

```bash
skylight import --input-file skylight-backup.json
```

---

## Scripting and automation

### Use JSON output for scripting

All commands support `--output json` (the default). Pipe to `jq` to filter:

```bash
# Get all pending chores assigned to a specific profile
skylight chore list --output json | jq '[.[] | select(.assignee_id == "6750876" and .status == "pending")]'

# Count total stars across all profiles
skylight reward points | jq '[.[].current_point_balance] | add'
```

### Suppress success messages (for scripts)

Pass credentials via environment variables to avoid interactive prompts in CI/cron:

```bash
export SKYLIGHT_REFRESH_TOKEN=your_token
export SKYLIGHT_FRAME_ID=3136444
skylight chore list
```

### Watch for changes in real time

```bash
# Prints a diff whenever chores, rewards, or calendar change
skylight watch
```
