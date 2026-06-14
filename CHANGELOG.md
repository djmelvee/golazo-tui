# Changelog

All notable changes to golazo-tui will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [0.1.0] — 2026-06-14

### Added
- Live dashboard (`h`) — current matches with LIVE minute badge, score, and venue
- Group standings (`g`) — all 12 groups A–L, 48 teams, GD column, top-2 rows in gold
- Upcoming fixtures (`f`) — NS matches grouped by date with MATCHDAY labels
- Changelog viewer (`c`) — renders this file inside the TUI (scrollable with j/k)
- Auto-refresh every 30 s on the live screen; last-updated timestamp shown
- Full FIFA World Cup 2026 branding: header, sidebar, WC vocabulary throughout
- `golazo-seed` — offline sample data with 48 teams across 12 groups A–L
- `golazo-fetcher` — polls `worldcup26.ir` REST API; `--watch` mode with configurable interval
- SQLite KV cache (WAL mode) — fetcher writes; TUI is read-only
- `golazo.bat` — Windows double-click launcher (build → seed → run)

### Fixed
- Match status codes updated to standard football API values: `LIVE`, `FT`, `NS`
- Only `FT` matches carry scores; `NS` matches show kickoff time only
- Removed Cameroon duplicate: Group F correctly uses `CMR`; Group K uses Indonesia (`IDN`)
- Renamed confusing map keys: `MAR2` → `CMR`, `MEX2` → `ECU`, `MEX3` → `KSA`, `POR2` → `COL`
