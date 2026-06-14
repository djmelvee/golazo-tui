# ⚽ Golazo TUI

A terminal dashboard for the **2026 FIFA World Cup** — live scores, group standings, and upcoming fixtures, all in your terminal.

Built with [Bubble Tea v2](https://charm.land/bubbletea) + [Lip Gloss v2](https://charm.land/lipgloss) + SQLite (pure Go, no CGO).

```
  ╔══════════════════════════════════════════════════════════════════════╗
  ║  ⚽  FIFA WORLD CUP 2026  ·  🇺🇸 USA  🇨🇦 CANADA  🇲🇽 MEXICO         ║
  ╠══════════════╦═══════════════════════════════════════════════════════╣
  ║ ⚽ GOLAZO    ║  Updated 21:04 CET  ·  auto-refreshes every 5s       ║
  ║ ──────────── ║                                                       ║
  ║ ● LIVE   [h] ║  No matches currently live · check back during        ║
  ║   Standings  ║  match hours                                          ║
  ║   Fixtures   ║                                                       ║
  ║   Changelog  ║  FULL TIME                                            ║
  ║ ──────────── ║  FT  🇲🇽 Mexico        2 – 0  🇿🇦 South Africa        ║
  ║ 🏆 WC 2026   ║  FT  🇰🇷 South Korea   2 – 1  🇨🇿 Czech Republic      ║
  ║ 48 Teams     ║  FT  🇺🇸 United States  4 – 1  🇵🇾 Paraguay           ║
  ║ 12 Groups    ║  FT  🇦🇺 Australia      2 – 0  🇹🇷 Türkiye             ║
  ╠══════════════╩═══════════════════════════════════════════════════════╣
  ║  h live · g standings · f fixtures · c changelog · q quit           ║
  ╚══════════════════════════════════════════════════════════════════════╝
```

## Features

- **Live dashboard** — current matches with live minute counter, gold score, and blinking ● indicator
- **Match detail** — press Enter on a live match for a full detail view: score, goal timeline, progress bar
- **Group standings** — all 12 groups (A–L), 48 teams, GD column, top-2 highlighted in gold
- **Upcoming fixtures** — next matches grouped by date with matchday labels
- **Changelog** — in-app changelog viewer (scrollable)
- **Auto-refresh** — live screen polls the cache every 5 seconds; live indicator pulses every second
- **Goal tracking** — fetcher detects score changes and logs goal events with minute
- **Offline seed** — works without API access using realistic WC2026 sample data
- **Responsive layout** — names and venues scale with terminal width; works at 80 cols and fullscreen

## Install

```bash
go install github.com/djmelvee/golazo-tui/cmd/golazo-tui@latest
```

Installs the TUI binary to `$GOPATH/bin`. Requires Go 1.23+. For seeded or live data, build the helper binaries manually (see **Manual Setup** below).

## Quick Start (Windows)

Double-click **`golazo.bat`** — it builds the binaries, seeds match data, and launches the TUI in one step. Requires [Go 1.23+](https://go.dev/dl/).

## Manual Setup

```bash
# Build all three binaries
go build -o bin/golazo-seed    ./cmd/golazo-seed
go build -o bin/golazo-fetcher ./cmd/golazo-fetcher
go build -o bin/golazo-tui     ./cmd/golazo-tui

# Populate the cache with offline sample data
./bin/golazo-seed

# Launch the dashboard
./bin/golazo-tui
```

## Live Data (optional)

The fetcher polls `http://worldcup26.ir:3050`. Set your token and run:

```bash
export GOLAZO_API_TOKEN=your_token_here
./bin/golazo-fetcher --watch --interval 60
```

| Env var | Default | Description |
|---|---|---|
| `GOLAZO_DB` | `~/.cache/golazo-tui/cache.db` | SQLite cache path |
| `GOLAZO_API` | `http://worldcup26.ir:3050` | API base URL |
| `GOLAZO_API_TOKEN` | *(required for fetcher)* | JWT bearer token |

## Keybindings

| Key | Action |
|---|---|
| `h` | Live dashboard |
| `g` | Group standings |
| `f` | Upcoming fixtures |
| `c` | Changelog |
| `j` / `k` | Move cursor (live) · scroll (standings / changelog / detail) |
| `Enter` | Open match detail for selected live match |
| `b` | Back from match detail to live dashboard |
| `q` | Quit |

## Architecture

```
golazo-fetcher  ──writes──▶  SQLite (WAL)  ──reads──▶  golazo-tui
golazo-seed     ──writes──▶  SQLite (WAL)
```

The TUI never touches the network — it reads only from the local SQLite cache. The fetcher and seed write to the same cache file. This decoupled design means the TUI stays fast and never blocks on a slow API.

## Stack

- [`charm.land/bubbletea/v2`](https://charm.land/bubbletea) — TUI framework
- [`charm.land/lipgloss/v2`](https://charm.land/lipgloss) — terminal styling
- [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) — pure-Go SQLite (no CGO)
