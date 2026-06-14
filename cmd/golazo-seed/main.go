// golazo-seed writes offline sample data to the cache DB so the TUI
// works without hitting the live API.
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

func main() {
	dbPath := os.Getenv("GOLAZO_DB")
	if dbPath == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatal(err)
		}
		dbPath = filepath.Join(home, ".cache", "golazo-tui", "cache.db")
	}

	db, err := data.Open(dbPath)
	if err != nil {
		log.Fatalf("open DB: %v", err)
	}
	defer db.Close()

	// Relative to 2026-06-14 (tournament day 4, matchday 1 still in progress)
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.Add(-24 * time.Hour)
	twoDaysAgo := today.Add(-48 * time.Hour)
	threeDaysAgo := today.Add(-72 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	twoDays := today.Add(48 * time.Hour)
	threeDays := today.Add(72 * time.Hour)

	// ── TEAMS (48 confirmed WC 2026 qualified nations) ────────────────────────
	// Source: UEFA, CONMEBOL, CONCACAF, CAF, AFC, OFC final draw results.
	// Italy did NOT qualify — lost UEFA playoff to Bosnia and Herzegovina.

	teams := map[string]wc.Team{
		// Group A — hosts: USA, Canada, Mexico
		"MEX": {ID: 1, Name: "Mexico", Flag: "🇲🇽", Group: "A"},
		"KOR": {ID: 2, Name: "South Korea", Flag: "🇰🇷", Group: "A"},
		"CZE": {ID: 3, Name: "Czech Republic", Flag: "🇨🇿", Group: "A"},
		"RSA": {ID: 4, Name: "South Africa", Flag: "🇿🇦", Group: "A"},
		// Group B
		"CAN": {ID: 5, Name: "Canada", Flag: "🇨🇦", Group: "B"},
		"BIH": {ID: 6, Name: "Bosnia and Herzegovina", Flag: "🇧🇦", Group: "B"},
		"QAT": {ID: 7, Name: "Qatar", Flag: "🇶🇦", Group: "B"},
		"SUI": {ID: 8, Name: "Switzerland", Flag: "🇨🇭", Group: "B"},
		// Group C
		"BRA": {ID: 9, Name: "Brazil", Flag: "🇧🇷", Group: "C"},
		"MAR": {ID: 10, Name: "Morocco", Flag: "🇲🇦", Group: "C"},
		"SCO": {ID: 11, Name: "Scotland", Flag: "🏴󠁧󠁢󠁳󠁣󠁴󠁿", Group: "C"},
		"HAI": {ID: 12, Name: "Haiti", Flag: "🇭🇹", Group: "C"},
		// Group D
		"USA": {ID: 13, Name: "United States", Flag: "🇺🇸", Group: "D"},
		"PAR": {ID: 14, Name: "Paraguay", Flag: "🇵🇾", Group: "D"},
		"AUS": {ID: 15, Name: "Australia", Flag: "🇦🇺", Group: "D"},
		"TUR": {ID: 16, Name: "Türkiye", Flag: "🇹🇷", Group: "D"},
		// Group E
		"GER": {ID: 17, Name: "Germany", Flag: "🇩🇪", Group: "E"},
		"CUW": {ID: 18, Name: "Curaçao", Flag: "🇨🇼", Group: "E"},
		"CIV": {ID: 19, Name: "Côte d'Ivoire", Flag: "🇨🇮", Group: "E"},
		"ECU": {ID: 20, Name: "Ecuador", Flag: "🇪🇨", Group: "E"},
		// Group F
		"NED": {ID: 21, Name: "Netherlands", Flag: "🇳🇱", Group: "F"},
		"JPN": {ID: 22, Name: "Japan", Flag: "🇯🇵", Group: "F"},
		"SWE": {ID: 23, Name: "Sweden", Flag: "🇸🇪", Group: "F"},
		"TUN": {ID: 24, Name: "Tunisia", Flag: "🇹🇳", Group: "F"},
		// Group G
		"BEL": {ID: 25, Name: "Belgium", Flag: "🇧🇪", Group: "G"},
		"EGY": {ID: 26, Name: "Egypt", Flag: "🇪🇬", Group: "G"},
		"IRN": {ID: 27, Name: "Iran", Flag: "🇮🇷", Group: "G"},
		"NZL": {ID: 28, Name: "New Zealand", Flag: "🇳🇿", Group: "G"},
		// Group H
		"ESP": {ID: 29, Name: "Spain", Flag: "🇪🇸", Group: "H"},
		"CPV": {ID: 30, Name: "Cape Verde", Flag: "🇨🇻", Group: "H"},
		"KSA": {ID: 31, Name: "Saudi Arabia", Flag: "🇸🇦", Group: "H"},
		"URU": {ID: 32, Name: "Uruguay", Flag: "🇺🇾", Group: "H"},
		// Group I
		"FRA": {ID: 33, Name: "France", Flag: "🇫🇷", Group: "I"},
		"SEN": {ID: 34, Name: "Senegal", Flag: "🇸🇳", Group: "I"},
		"IRQ": {ID: 35, Name: "Iraq", Flag: "🇮🇶", Group: "I"},
		"NOR": {ID: 36, Name: "Norway", Flag: "🇳🇴", Group: "I"},
		// Group J
		"ARG": {ID: 37, Name: "Argentina", Flag: "🇦🇷", Group: "J"},
		"ALG": {ID: 38, Name: "Algeria", Flag: "🇩🇿", Group: "J"},
		"AUT": {ID: 39, Name: "Austria", Flag: "🇦🇹", Group: "J"},
		"JOR": {ID: 40, Name: "Jordan", Flag: "🇯🇴", Group: "J"},
		// Group K
		"POR": {ID: 41, Name: "Portugal", Flag: "🇵🇹", Group: "K"},
		"COD": {ID: 42, Name: "DR Congo", Flag: "🇨🇩", Group: "K"},
		"UZB": {ID: 43, Name: "Uzbekistan", Flag: "🇺🇿", Group: "K"},
		"COL": {ID: 44, Name: "Colombia", Flag: "🇨🇴", Group: "K"},
		// Group L
		"ENG": {ID: 45, Name: "England", Flag: "🏴󠁧󠁢󠁥󠁮󠁧󠁿", Group: "L"},
		"CRO": {ID: 46, Name: "Croatia", Flag: "🇭🇷", Group: "L"},
		"GHA": {ID: 47, Name: "Ghana", Flag: "🇬🇭", Group: "L"},
		"PAN": {ID: 48, Name: "Panama", Flag: "🇵🇦", Group: "L"},
	}

	// ── LIVE MATCHES — none (honest: seed reflects June 14 afternoon) ─────────
	liveMatches := []wc.Match{}

	// ── FINISHED MATCHES (8 confirmed FT results, June 11–14) ─────────────────
	// Scores sourced from ESPN/FIFA official WC 2026 match reports.

	s2_0 := [2]int{2, 0}
	s2_1 := [2]int{2, 1}
	s1_1 := [2]int{1, 1}
	s4_1 := [2]int{4, 1}
	s1_0 := [2]int{1, 0}

	finishedMatches := []wc.Match{
		// June 11
		{
			ID: 101, HomeTeam: teams["MEX"], AwayTeam: teams["RSA"],
			HomeScore: &s2_0[0], AwayScore: &s2_0[1],
			Status:    wc.StatusFinished,
			KickoffAt: threeDaysAgo.Add(20 * time.Hour),
			Venue:     "AT&T Stadium, Arlington TX",
			Group: "A", Stage: "group", Matchday: 1,
		},
		// June 12
		{
			ID: 102, HomeTeam: teams["KOR"], AwayTeam: teams["CZE"],
			HomeScore: &s2_1[0], AwayScore: &s2_1[1],
			Status:    wc.StatusFinished,
			KickoffAt: twoDaysAgo.Add(17 * time.Hour),
			Venue:     "Rose Bowl, Pasadena CA",
			Group: "A", Stage: "group", Matchday: 1,
		},
		{
			ID: 103, HomeTeam: teams["CAN"], AwayTeam: teams["BIH"],
			HomeScore: &s1_1[0], AwayScore: &s1_1[1],
			Status:    wc.StatusFinished,
			KickoffAt: twoDaysAgo.Add(20 * time.Hour),
			Venue:     "BC Place, Vancouver",
			Group: "B", Stage: "group", Matchday: 1,
		},
		{
			ID: 104, HomeTeam: teams["USA"], AwayTeam: teams["PAR"],
			HomeScore: &s4_1[0], AwayScore: &s4_1[1],
			Status:    wc.StatusFinished,
			KickoffAt: twoDaysAgo.Add(23 * time.Hour),
			Venue:     "MetLife Stadium, New Jersey",
			Group: "D", Stage: "group", Matchday: 1,
		},
		// June 13
		{
			ID: 105, HomeTeam: teams["QAT"], AwayTeam: teams["SUI"],
			HomeScore: &s1_1[0], AwayScore: &s1_1[1],
			Status:    wc.StatusFinished,
			KickoffAt: yesterday.Add(14 * time.Hour),
			Venue:     "Levi's Stadium, Santa Clara CA",
			Group: "B", Stage: "group", Matchday: 1,
		},
		{
			ID: 106, HomeTeam: teams["BRA"], AwayTeam: teams["MAR"],
			HomeScore: &s1_1[0], AwayScore: &s1_1[1],
			Status:    wc.StatusFinished,
			KickoffAt: yesterday.Add(20 * time.Hour),
			Venue:     "SoFi Stadium, Inglewood CA",
			Group: "C", Stage: "group", Matchday: 1,
		},
		{
			ID: 107, HomeTeam: teams["SCO"], AwayTeam: teams["HAI"],
			HomeScore: &s1_0[0], AwayScore: &s1_0[1],
			Status:    wc.StatusFinished,
			KickoffAt: yesterday.Add(23 * time.Hour),
			Venue:     "Gillette Stadium, Foxborough MA",
			Group: "C", Stage: "group", Matchday: 1,
		},
		// June 14 (earlier today)
		{
			ID: 108, HomeTeam: teams["AUS"], AwayTeam: teams["TUR"],
			HomeScore: &s2_0[0], AwayScore: &s2_0[1],
			Status:    wc.StatusFinished,
			KickoffAt: today.Add(4 * time.Hour),
			Venue:     "Mercedes-Benz Stadium, Atlanta GA",
			Group: "D", Stage: "group", Matchday: 1,
		},
	}

	// ── UPCOMING / NS (June 14 later + next days) ─────────────────────────────
	// All times UTC. Sources: ESPN/CBS Sports confirmed schedule.
	// CET = UTC+1, CEST (June) = UTC+2. Dutch viewers add 2h to UTC times.

	upcomingMatches := []wc.Match{
		// June 14 — 17:00 UTC = 19:00 CEST
		{
			ID: 201, HomeTeam: teams["GER"], AwayTeam: teams["CUW"],
			Status:    wc.StatusUpcoming,
			KickoffAt: today.Add(17 * time.Hour),
			Venue:     "NRG Stadium, Houston TX",
			Group: "E", Stage: "group", Matchday: 1,
		},
		// June 14 — 20:00 UTC = 22:00 CEST
		{
			ID: 202, HomeTeam: teams["NED"], AwayTeam: teams["JPN"],
			Status:    wc.StatusUpcoming,
			KickoffAt: today.Add(20 * time.Hour),
			Venue:     "AT&T Stadium, Arlington TX",
			Group: "F", Stage: "group", Matchday: 1,
		},
		// June 14 — 23:00 UTC = 01:00 CEST (Jun 15)
		{
			ID: 203, HomeTeam: teams["CIV"], AwayTeam: teams["ECU"],
			Status:    wc.StatusUpcoming,
			KickoffAt: today.Add(23 * time.Hour),
			Venue:     "Lincoln Financial Field, Philadelphia PA",
			Group: "E", Stage: "group", Matchday: 1,
		},
		// June 15 — 02:00 UTC = 04:00 CEST
		{
			ID: 204, HomeTeam: teams["SWE"], AwayTeam: teams["TUN"],
			Status:    wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(2 * time.Hour),
			Venue:     "Estadio BBVA, Monterrey",
			Group: "F", Stage: "group", Matchday: 1,
		},
		// June 15 — 16:00 UTC = 18:00 CEST
		{
			ID: 205, HomeTeam: teams["ESP"], AwayTeam: teams["CPV"],
			Status:    wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(16 * time.Hour),
			Venue:     "Mercedes-Benz Stadium, Atlanta GA",
			Group: "H", Stage: "group", Matchday: 1,
		},
		// June 15 — 22:00 UTC = 00:00 CEST (Jun 16)
		{
			ID: 206, HomeTeam: teams["BEL"], AwayTeam: teams["EGY"],
			Status:    wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(22 * time.Hour),
			Venue:     "Lumen Field, Seattle WA",
			Group: "G", Stage: "group", Matchday: 1,
		},
		// June 16 — 19:00 UTC = 21:00 CEST
		{
			ID: 207, HomeTeam: teams["FRA"], AwayTeam: teams["SEN"],
			Status:    wc.StatusUpcoming,
			KickoffAt: twoDays.Add(19 * time.Hour),
			Venue:     "MetLife Stadium, New Jersey",
			Group: "I", Stage: "group", Matchday: 1,
		},
		// June 17 — 01:00 UTC = 03:00 CEST
		{
			ID: 208, HomeTeam: teams["ARG"], AwayTeam: teams["ALG"],
			Status:    wc.StatusUpcoming,
			KickoffAt: threeDays.Add(1 * time.Hour),
			Venue:     "Arrowhead Stadium, Kansas City MO",
			Group: "J", Stage: "group", Matchday: 1,
		},
		// June 17 — 17:00 UTC = 19:00 CEST
		{
			ID: 209, HomeTeam: teams["POR"], AwayTeam: teams["COD"],
			Status:    wc.StatusUpcoming,
			KickoffAt: threeDays.Add(17 * time.Hour),
			Venue:     "NRG Stadium, Houston TX",
			Group: "K", Stage: "group", Matchday: 1,
		},
		// June 17 — 20:00 UTC = 22:00 CEST
		{
			ID: 210, HomeTeam: teams["ENG"], AwayTeam: teams["CRO"],
			Status:    wc.StatusUpcoming,
			KickoffAt: threeDays.Add(20 * time.Hour),
			Venue:     "AT&T Stadium, Arlington TX",
			Group: "L", Stage: "group", Matchday: 1,
		},
	}

	// ── STANDINGS (computed from 8 FT results; groups E–L all 0pts) ──────────
	// Sorted by Pts DESC, GD DESC per group.

	standings := map[string][]wc.GroupRow{
		// Group A: MEX 2-0 RSA (Jun 11), KOR 2-1 CZE (Jun 12)
		"A": {
			{Team: teams["MEX"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 0, GD: 2, Pts: 3},
			{Team: teams["KOR"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 1, GD: 1, Pts: 3},
			{Team: teams["CZE"], Played: 1, W: 0, D: 0, L: 1, GF: 1, GA: 2, GD: -1, Pts: 0},
			{Team: teams["RSA"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
		},
		// Group B: CAN 1-1 BIH (Jun 12), QAT 1-1 SUI (Jun 13)
		"B": {
			{Team: teams["CAN"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["BIH"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["QAT"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["SUI"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
		},
		// Group C: BRA 1-1 MAR (Jun 13), SCO 1-0 HAI (Jun 13)
		"C": {
			{Team: teams["SCO"], Played: 1, W: 1, D: 0, L: 0, GF: 1, GA: 0, GD: 1, Pts: 3},
			{Team: teams["BRA"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["MAR"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["HAI"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 1, GD: -1, Pts: 0},
		},
		// Group D: USA 4-1 PAR (Jun 12), AUS 2-0 TUR (Jun 14)
		"D": {
			{Team: teams["USA"], Played: 1, W: 1, D: 0, L: 0, GF: 4, GA: 1, GD: 3, Pts: 3},
			{Team: teams["AUS"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 0, GD: 2, Pts: 3},
			{Team: teams["TUR"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
			{Team: teams["PAR"], Played: 1, W: 0, D: 0, L: 1, GF: 1, GA: 4, GD: -3, Pts: 0},
		},
		// Groups E–L: no matchday 1 results yet
		"E": {
			{Team: teams["GER"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CUW"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CIV"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["ECU"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"F": {
			{Team: teams["NED"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["JPN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["SWE"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["TUN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"G": {
			{Team: teams["BEL"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["EGY"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["IRN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["NZL"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"H": {
			{Team: teams["ESP"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CPV"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["KSA"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["URU"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"I": {
			{Team: teams["FRA"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["SEN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["IRQ"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["NOR"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"J": {
			{Team: teams["ARG"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["ALG"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["AUT"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["JOR"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"K": {
			{Team: teams["POR"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["COD"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["UZB"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["COL"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"L": {
			{Team: teams["ENG"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CRO"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["GHA"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["PAN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
	}

	// ── WRITE TO CACHE ────────────────────────────────────────────────────────

	if err := db.Set("matches:live", liveMatches); err != nil {
		log.Fatalf("set live: %v", err)
	}
	if err := db.Set("matches:finished", finishedMatches); err != nil {
		log.Fatalf("set finished: %v", err)
	}
	if err := db.Set("matches:upcoming", upcomingMatches); err != nil {
		log.Fatalf("set upcoming: %v", err)
	}
	if err := db.Set("standings", standings); err != nil {
		log.Fatalf("set standings: %v", err)
	}

	fmt.Printf("✓ Seeded %d LIVE, %d FT, %d NS matches + 12 groups (48 teams)\n",
		len(liveMatches), len(finishedMatches), len(upcomingMatches))
	fmt.Printf("  DB: %s\n", dbPath)
}
