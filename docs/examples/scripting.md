# Scripting

## Filter output with jq

All commands output JSON by default. Pipe to `jq` to extract what you need:

```bash
# Pending chores for a specific profile
skylight chore list | jq '[.[] | select(.assignee_id == "6750876" and .status == "pending")]'

# Total stars across all profiles
skylight reward points | jq '[.[].current_point_balance] | add'

# Names of all profiles
skylight category list | jq '[.[].name]'
```

## Use environment variables (CI, cron, Docker)

Pass credentials without a config file:

```bash
export SKYLIGHT_REFRESH_TOKEN=your_token
export SKYLIGHT_FRAME_ID=3136444

skylight chore list
```

Or inline for a one-off:

```bash
SKYLIGHT_REFRESH_TOKEN=your_token SKYLIGHT_FRAME_ID=3136444 skylight chore list
```

## Watch for live changes

Prints a diff to stdout whenever chores, rewards, or calendar events change:

```bash
skylight watch
```

## Table output

Switch any command to table format for human-readable output:

```bash
skylight chore list --output table
skylight reward list --output table
```
