# Backup and restore

## Export your frame data

```bash
skylight export --output-file skylight-backup.json
```

Exports chores, rewards, lists, recipes, meal sittings, and calendar events for a 90-day window around today. Useful before making bulk changes.

Export specific resource types only:

```bash
skylight export --resources chores,rewards --output-file chores-rewards.json
```

## Restore from a backup

```bash
skylight import --input-file skylight-backup.json
```

Preview what would be imported without making changes:

```bash
skylight import --input-file skylight-backup.json --dry-run
```
