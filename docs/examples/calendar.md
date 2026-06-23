# Calendar

## List upcoming events

```bash
skylight calendar list --start-date $(date +%Y-%m-%d)
```

## Add an event

```bash
skylight calendar create \
  --title "Soccer practice" \
  --start-at "2026-06-28T15:00:00" \
  --end-at "2026-06-28T16:30:00" \
  --category-id 6750876
```

## Add an all-day event

```bash
skylight calendar create \
  --title "School holiday" \
  --start-at "2026-07-04" \
  --end-at "2026-07-04" \
  --all-day
```

## Update an event

```bash
skylight calendar update --event-id EVENT_ID --title "Soccer practice (rescheduled)" --start-at "2026-06-29T15:00:00"
```

## Delete an event

```bash
skylight calendar delete --event-id EVENT_ID
```
