# ⚽ Golazo TUI

A terminal dashboard for the **2026 FIFA World Cup** — live scores, predictions, bracket, digest, and more.

Built with [Bubble Tea v2](https://charm.land/bubbletea) + [Lip Gloss v2](https://charm.land/lipgloss) + SQLite (pure Go, no CGO).

## Features

- **Live dashboard** (`h`) — live minute, gold scores, blinking ●, favorite team ★, selectable upcoming
- **Predictions** (`p`) — form-weighted forecasts with xG, confidence, accuracy badges, full breakdown on Enter
- **Knockout bracket** (`b`) — ASCII tree with Enter → match detail
- **Match-day digest** (`d`) — newspaper-style today view with headlines and recent goals
- **Golden Boot** (`s`) — top scorers table
- **Group standings** (`g`) — all 12 groups; Enter on a team → team hub
- **Fixtures** (`f`) — scrollable list with group/knockout filter (`1`/`2`/`3`)
- **Team hub** (`t`) — favorite team fixtures and form; cycle teams with `t`
- **Match detail** — goal timeline, progress bar, event liveblog
- **Golazo celebration** — 10s overlay with spinning ASCII ball on goals (any screen)
- **Goal alerts** (`!`) — terminal bell + desktop notification (Windows toast / notify-send)
- **Timezone** (`z`) — Amsterdam (default), local, or UTC
- **Help** (`?`) — full keybinding reference
- **Splash** — GOLAZO logo + 3D spinning soccer ball (skip with `GOLAZO_NO_SPLASH=1`)
- **Auto-fetch** — built-in API poll every 5s; optional `golazo-fetcher` daemon
- **Offline seed** — `golazo-seed` for demo data without API access

## Install

```bash
go install github.com/djmelvee/golazo-tui/cmd/golazo-tui@latest
```

Requires Go 1.23+.

## Quick Start (Windows)

Double-click **`golazo.bat`** — builds, seeds, and launches the TUI.

## Manual Setup

```bash
go build -o bin/golazo-seed    ./cmd/golazo-seed
go build -o bin/golazo-fetcher ./cmd/golazo-fetcher
go build -o bin/golazo-tui     ./cmd/golazo-tui
./bin/golazo-seed
./bin/golazo-tui
```

## Live Data

The TUI auto-registers a JWT on first launch and polls the API internally. Optionally run the fetcher daemon:

```bash
export GOLAZO_API_TOKEN=your_token   # optional — also read from SQLite cache
./bin/golazo-fetcher --watch --interval 5
```

| Env var | Default | Description |
|---|---|---|
| `GOLAZO_DB` | `~/.cache/golazo-tui/cache.db` | SQLite cache path |
| `GOLAZO_API` | `http://worldcup26.ir:3050` | API base URL |
| `GOLAZO_API_TOKEN` | auto / cached | JWT bearer token |
| `GOLAZO_NO_SPLASH` | — | Set to `1` to skip boot animation |

## Keybindings

| Key | Action |
|---|---|
| `h` | Live dashboard |
| `p` | Match predictions |
| `b` | Knockout bracket |
| `d` | Match-day digest |
| `s` | Golden Boot scorers |
| `g` | Group standings |
| `f` | Upcoming fixtures |
| `t` | Team hub (cycle favorite with `t`) |
| `c` | Changelog |
| `?` | Help overlay |
| `z` | Cycle timezone |
| `!` | Toggle goal alerts |
| `j` / `k` | Cursor or scroll (screen-dependent) |
| `Enter` | Open match / prediction / team detail |
| `esc` | Back from detail or help |
| `1` `2` `3` | Fixtures filter: all / group / knockout |
| `q` | Quit |

## Architecture

```
worldcup26.ir API ──▶ fetcher (atomic SQLite writes) ──▶ cache.db ──▶ golazo-tui
ESPN fallback      ──▶ patch scores when API down
golazo-seed        ──▶ offline demo data
```

## Stack

- [Bubble Tea v2](https://charm.land/bubbletea) — TUI framework
- [Lip Gloss v2](https://charm.land/lipgloss) — terminal styling
- [modernc.org/sqlite](https://pkg.go.dev/modernc.org/sqlite) — pure-Go SQLite