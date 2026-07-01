# Changelog

All notable changes to golazo-tui will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [Unreleased]

## [0.6.1] — 2026-07-02

### Added
- **Rich match timeline** — opening a match now shows tournament form, stakes, H2H record, goal-by-goal commentary (first blood, equalizer, late drama), HT scoreline, second-half kick-off, explosive-half callouts, and full-time summaries (comebacks, clean sheets, knockout advancement)
- **Match preview** — upcoming fixtures show a preview block instead of a blank events list
- **Goal metadata** — penalty and own-goal tags, stoppage-time minutes (e.g. `90+2'`) parsed from API scorer strings

## [0.6.0] — 2026-07-02

### Added
- **3D splash screen** — GOLAZO block-letter logo with a large ASCII soccer ball spinning in 3D at startup (`GOLAZO_NO_SPLASH=1` to skip)
- **Golazo goal celebration** — 10-second overlay on any screen showing “Golazo!”, scorer/score line, and a mini spinning ball; all goals in a poll batch are celebrated
- **Desktop goal alerts** — Windows toast and `notify-send` on Linux/macOS when alerts are enabled (`!`)
- **Help overlay** (`?`) — full keyboard shortcut reference
- **Team hub** (`t`) — favorite team fixtures and WC form; cycle teams with `t`; Enter on standings/scorers opens team context
- **Match predictions** (`p`) — form-weighted forecasts with xG, confidence, low-data badges, accuracy badges after FT, H2H block in detail; predictions cached in SQLite
- **Knockout bracket** (`b`) — ASCII tree with responsive widths, winner hints, Enter → match detail
- **Match-day digest** (`d`) — newspaper layout with top story, biggest result, next kickoff, recent goals; Enter → match detail
- **Golden Boot scorers** (`s`) — leaderboard with cursor selection; Enter → team hub
- **Fixtures overhaul** — scrollable list with cursor, Enter → detail, stage filter (`1` all / `2` group / `3` knockout)
- **Timezone preference** (`z`) — cycle Amsterdam, local, or UTC; persisted in SQLite prefs
- **Mouse navigation** — click sidebar items to switch screens
- **Staleness banner** — header note when cache is old or fetch fails
- **Sidebar panel** — dark background styling for the nav column
- **Embedded changelog** — in-app changelog works outside the project root (build-time embed)
- **Fetcher token from cache** — `golazo-fetcher` reads JWT from SQLite when `GOLAZO_API_TOKEN` is unset
- **API resilience** — HTTP retry with backoff, 30s stadium map cache, 401 auto re-register, atomic SQLite batch writes on fetch
- **Goal detection** — seeds from `lastscore` on restart to avoid false celebrations; prefers API scorer timeline over index guessing
- **Fetcher tests** — goal dedup and scorer attribution unit tests

### Fixed
- `getGames` mutex deadlock that hung `TestFetchMatchesFromMockAPI` and could stall live fetches
- Live status line now correctly says data polls every 5s (UI tick remains 1s)
- Team form stats use finished matches only (no mid-game skew)
- Predictions subtitle no longer says “knockout only” when group fixtures are included
- Bracket and predictions lists auto-scroll to keep cursor visible

### Changed
- Goal celebration text is **Golazo!** (was “GOOOAL!”) and lasts **10 seconds** (was 2s)
- `esc` backs out of detail/help; `b` opens bracket only from main screens (no longer dual back/bracket on detail)
- Footer and README document all screens and keybindings
- `j`/`k` on digest, fixtures, standings, and scorers moves cursor; changelog/detail still scroll
- Live dashboard shows ★ on favorite team rows; up to 12 upcoming fixtures are selectable
- SQLite writer uses `busy_timeout` for safer TUI + fetcher coexistence

## [0.5.0] — 2026-06-14

### Added
- Zero-config startup — auto-registers a JWT with worldcup26.ir on first launch; token stored in SQLite so subsequent launches use it directly; no manual `GOLAZO_API_TOKEN` setup needed
- Background auto-fetch — TUI polls the API every 5 seconds internally (no external fetcher process required); scores and standings update live while the app is open
- Match detail screen — press Enter on any live match for a full view: teams, gold score, pulsing ● minute, venue/group info, goal event liveblog, match progress bar; press b to return
- Goal event tracking — detects score changes between polls via both worldcup26.ir live data and the ESPN fallback; stores `GoalEvent` records per match and renders them as `⚽` entries in the detail screen timeline
- ESPN public scoreboard fallback — when worldcup26.ir is unreachable or auth fails, ESPN's unauthenticated scoreboard API patches live and finished scores directly in the DB; no token required
- Live screen cursor — j/k moves selection through live/upcoming/finished rows; selected row highlighted with `>`; Enter opens detail screen
- Last-score cache — finished-score is preserved across live→FT status transitions so scores never revert to `– –`

### Fixed
- Scores missing for matches with millisecond-precision timestamps (RFC3339Nano now tried first in `parseKickoff`)
- Nil scores on time-promoted matches caused by duplicate entries from multiple API buckets (seenIDs deduplication in `Live.Load`)
- `golazo.bat` written with CRLF line endings so it runs correctly on Windows without modification

### Changed
- j/k on the live screen now moves cursor; standings/changelog scroll unchanged
- Header shows no error note when ESPN successfully provides live data during primary API outage

## [0.4.0] — 2026-06-14

### Added
- Blinking ● live indicator — animates red ↔ dim at 1 s interval when matches are live; only the ● character re-renders (no full-screen flicker)
- Sidebar live count badge — LIVE nav item shows a pulsing ● when matches are in progress, visible from any screen
- Live match scores styled in gold (`GoldBold`) for visual prominence on the live dashboard

### Changed
- Data refresh interval reduced from 30 s to 5 s — live data appears within 5 s of the fetcher writing it; SQLite WAL read-through confirmed correct
- Verified auto-refresh chain: `Init → TickCmd → TickMsg → Load(db) → TickCmd`; no bugs found

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
