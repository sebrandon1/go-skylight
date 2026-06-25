# Chores

## List chores

```bash
# All pending chores
skylight chore list --status pending

# Completed chores this week (--after/--before are required for status=complete)
skylight chore list --status complete \
  --after $(date -d 'last monday' +%Y-%m-%d) \
  --before $(date +%Y-%m-%d)

# Only up-for-grabs (unassigned) chores
skylight chore list --up-for-grabs
```

## Create a chore

```bash
# Get profile IDs first
skylight category list

# Create and assign to a profile
skylight chore create \
  --title "Take out trash" \
  --points 5 \
  --assignee-id 6750876

# Create unassigned (up-for-grabs)
skylight chore create --title "Vacuum living room" --points 10
```

## Complete a chore

```bash
skylight chore list  # get the chore ID

skylight chore complete --chore-id CHORE_ID --assignee-id 6750876
```

## Claim an up-for-grabs chore

```bash
skylight chore list --up-for-grabs  # find the chore ID

skylight chore claim --chore-id CHORE_ID --assignee-id YOUR_CATEGORY_ID
```

## Skip a recurring chore

```bash
skylight chore skip --chore-id CHORE_ID
```

## Skip an up-for-grabs chore instance

Skylight's app UI does not show a skip button for up-for-grabs (unassigned)
chores, but the API accepts a `skipped` status for them. Use the CLI to skip
a specific instance without deleting and recreating the chore.

```bash
# List up-for-grabs chores to find the instance ID
skylight chore list --up-for-grabs

# Skip today's instance — the date suffix targets only this occurrence
skylight chore skip --chore-id 12345678-2026-06-25
```

The chore continues repeating on its normal schedule; only the targeted
instance is skipped. This is useful for weather-dependent chores (e.g. skip
"water outside plants" on a rainy day) without losing the recurring pattern.

## Delete a chore

```bash
skylight chore delete --chore-id CHORE_ID
```
