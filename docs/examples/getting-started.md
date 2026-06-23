# Getting started

## 1. Install

```bash
go install github.com/sebrandon1/go-skylight@latest
```

## 2. Log in and save your credentials

```bash
skylight login --email you@example.com --password yourpassword --save
```

This saves your credentials to `~/.skylight/config` so you don't need to pass flags on every command.

## 3. Find your frame ID

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

From here, every command automatically uses this frame.
