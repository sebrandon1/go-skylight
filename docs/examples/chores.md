# Chores

## List chores

```bash
# All pending chores in an explicit date window (--after/--before must be used together)
skylight chore list --status pending --after 2026-08-01 --before 2026-08-31

# Without flags, list defaults to the current calendar month
# (--date DATE alone becomes a same-day window)
skylight chore list --status pending

# Completed chores this week (uses GNU date; on macOS substitute the date manually)
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

skylight chore complete --chore-id CHORE_ID
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


## Delete a chore

```bash
skylight chore delete --chore-id CHORE_ID
```
