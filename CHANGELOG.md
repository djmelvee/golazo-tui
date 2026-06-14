# Changelog

All notable changes to golazo-tui will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [0.3.0] — 2026-06-14

### Added
- GitHub Actions CI — `go build ./...` and `go vet ./...` on every push to main and every pull request (`.github/workflows/ci.yml`)
- `--version` flag on `golazo-tui` — prints `golazo-tui v0.3.0` and exits; version defined as a single constant in `cmd/golazo-tui/main.go`
- `go install github.com/djmelvee/golazo-tui/cmd/golazo-tui@latest` support — documented in README under new Install section

## [0.2.0] — 2026-06-14

### Added
- `WORKFLOW.md` — project workflow documentation: issue creation, labels, topics, and commit conventions
- Responsive layout — team names and venues scale with terminal width; works from 80 cols to fullscreen
- Issue confirmation rule (rule 5) and README currency rule (rule 6) in `WORKFLOW.md`

### Fixed
- Replaced incorrect seed data: removed Italy (did not qualify), added Bosnia and Herzegovina, Scotland, Curaçao, DR Congo, Uzbekistan, Cape Verde, and all other correct WC 2026 nations
- 8 confirmed FT results June 11–14 with real scores; 0 LIVE (honest); 10 NS upcoming
- Fixed API client: `GET /get/games` replaces non-existent `/matches?status=` endpoint
- Fixed API client: `GET /get/groups` replaces non-existent `/standings` endpoint
- LIVE/FT/NS status now derived from `finished` + `time_elapsed` fields (API has no status string)
- Standings W/D/L computed from finished match data; `/get/groups` not used
- Added `flagMap` for all 48 WC 2026 teams (API does not supply flag emojis)
- Added 30-second in-memory game cache to avoid 4× duplicate API calls per fetch cycle
- Replaced `log.Fatal` on missing `GOLAZO_API_TOKEN` with human-readable registration instructions
- Distinguish "no data at all" from "no live matches right now" on the live screen
- Corrected match kickoff times; all times displayed in CET (`Europe/Amsterdam` timezone)
- README preview corrected with real FT results; feature list updated to reflect actual app state

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
