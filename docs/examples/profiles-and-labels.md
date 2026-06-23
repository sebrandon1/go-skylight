# Profiles and labels

The Skylight app only lets you merge or convert profiles and labels — it won't let you delete them outright. The CLI can.

## List profiles and labels

```bash
skylight category list
```

```json
[
  { "id": "6750875", "name": "Team Blah Blah", "color": "#FF5733" },
  { "id": "6750876", "name": "Emma", "color": "#33A1FF" }
]
```

## Delete a profile or label

```bash
skylight category delete --category-id 6750875
```

> **Warning:** Deleting a profile also deletes all chores and rewards assigned exclusively to it. Tasks assigned to multiple profiles remain for the other assignees.

## Rename a profile or label

```bash
skylight category update --category-id 6750875 --name "Saturday Crew" --color "#22C55E"
```

## Create a new profile or label

```bash
skylight category create --name "Alex" --color "#A855F7"
```
