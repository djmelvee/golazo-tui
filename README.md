# ⚽ Golazo TUI

A terminal dashboard for the **2026 FIFA World Cup** — live scores, group standings, and upcoming fixtures, all in your terminal.

Built with [Bubble Tea v2](https://charm.land/bubbletea) + [Lip Gloss v2](https://charm.land/lipgloss) + SQLite (pure Go, no CGO).

```
  ╔══════════════════════════════════════════════════════════════════════╗
  ║  ⚽  FIFA WORLD CUP 2026  ·  🇺🇸 USA  🇨🇦 CANADA  🇲🇽 MEXICO         ║
  ╠══════════════╦═══════════════════════════════════════════════════════╣
  ║ ⚽ GOLAZO    ║  ● LIVE MATCHES                                       ║
  ║ ──────────── ║                                                       ║
  ║ ● LIVE   [h] ║  ● 74'  🇩🇪 Germany  3 – 1  🇷🇸 Serbia               ║
  ║   Standings  ║  ● 67'  🏴󠁧󠁢󠁥󠁮󠁧󠁿 England  0 – 0  🇸🇳 Senegal               ║
  ║   Fixtures   ║  ● 45'  🇧🇷 Brazil   2 – 1  🇲🇽 Mexico                ║
  ║ ──────────── ║  ● 22'  🇺🇸 USA      1 – 0  🇵🇦 Panama                ║
  ║ 🏆 WC 2026   ║                                                       ║
  ║ 48 Teams     ║  FULL TIME                                            ║
  ║ 12 Groups    ║  FT  🇦🇷 Argentina  2 – 0  🇨🇱 Chile                   ║
  ╠══════════════╩═══════════════════════════════════════════════════════╣
  ║  h live · g standings · f fixtures · q quit                         ║
  ╚══════════════════════════════════════════════════════════════════════╝
```

## Features

- **Live dashboard** — current matches with live minute counter and score
- **Group standings** — all 12 groups (A–L), 48 teams, GD column, top-2 highlighted in gold
- **Upcoming fixtures** — next matches grouped by date with matchday labels
- **Auto-refresh** — live screen polls the cache every 30 seconds
- **Offline seed** — works without API access using realistic WC2026 sample data

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
| `j` / `k` | Scroll standings down / up |
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
