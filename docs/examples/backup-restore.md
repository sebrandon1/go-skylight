# Backup and restore

## Export your frame data

```bash
skylight export --output-file skylight-backup.json
```

Exports chores, rewards, lists, recipes, meal sittings, calendar events, routines, bounties, categories, and photos. Chores, meal sittings, and calendar events cover 90 days either side of today (change it with `--days`). Useful before making bulk changes.

Export specific resource types only:

```bash
skylight export --resources chores,rewards --output-file chores-rewards.json
```

## Restore from a backup

```bash
skylight import --file skylight-backup.json
```

Preview what would be imported without making changes:

```bash
skylight import --file skylight-backup.json --dry-run
```

## What a restore recreates

- Recurring chores come back as one chore per series with their original schedule, starting from the first exported occurrence on or after the day you import that isn't done yet. If the backup has none (for example, it's older than `--days`), the series starts on its latest exported occurrence, so past dates may show as late. Series that have already ended are skipped.
- One-off chores come back only if they were still pending. Completion history and streaks are not restored.
- Lists come back with their items in order, including completed items and sections.
- Rewards come back with their assignee, description and repeat-after-redeem setting. Backups from older versions don't record repeat-after-redeem, so their rewards come back as one-time.
- Assignees are matched by ID, and categories are not imported, so restore into the frame the backup came from.
- Import doesn't check what's already on the frame. Importing the same file twice creates everything twice.
