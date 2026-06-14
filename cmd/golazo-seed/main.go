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

	// Relative to today (2026-06-14 = day 4 of WC2026, matchday 1 ongoing)
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	yesterday := today.Add(-24 * time.Hour)
	twoDaysAgo := today.Add(-48 * time.Hour)
	threeDaysAgo := today.Add(-72 * time.Hour)
	tomorrow := today.Add(24 * time.Hour)
	twoDays := today.Add(48 * time.Hour)

	// ── TEAMS (48 teams, groups A–L) ─────────────────────────────────────────

	teams := map[string]wc.Team{
		// Group A — host nation USA
		"USA": {ID: 1, Name: "United States", Flag: "🇺🇸", Group: "A"},
		"PAN": {ID: 2, Name: "Panama", Flag: "🇵🇦", Group: "A"},
		"URU": {ID: 3, Name: "Uruguay", Flag: "🇺🇾", Group: "A"},
		"BEL": {ID: 4, Name: "Belgium", Flag: "🇧🇪", Group: "A"},
		// Group B
		"ARG": {ID: 5, Name: "Argentina", Flag: "🇦🇷", Group: "B"},
		"CHI": {ID: 6, Name: "Chile", Flag: "🇨🇱", Group: "B"},
		"POL": {ID: 7, Name: "Poland", Flag: "🇵🇱", Group: "B"},
		"NGA": {ID: 8, Name: "Nigeria", Flag: "🇳🇬", Group: "B"},
		// Group C
		"BRA": {ID: 9, Name: "Brazil", Flag: "🇧🇷", Group: "C"},
		"MEX": {ID: 10, Name: "Mexico", Flag: "🇲🇽", Group: "C"},
		"SUI": {ID: 11, Name: "Switzerland", Flag: "🇨🇭", Group: "C"},
		"ALG": {ID: 12, Name: "Algeria", Flag: "🇩🇿", Group: "C"},
		// Group D
		"FRA": {ID: 13, Name: "France", Flag: "🇫🇷", Group: "D"},
		"AUS": {ID: 14, Name: "Australia", Flag: "🇦🇺", Group: "D"},
		"MAR": {ID: 15, Name: "Morocco", Flag: "🇲🇦", Group: "D"},
		"DEN": {ID: 16, Name: "Denmark", Flag: "🇩🇰", Group: "D"},
		// Group E
		"GER": {ID: 17, Name: "Germany", Flag: "🇩🇪", Group: "E"},
		"SRB": {ID: 18, Name: "Serbia", Flag: "🇷🇸", Group: "E"},
		"CRC": {ID: 19, Name: "Costa Rica", Flag: "🇨🇷", Group: "E"},
		"NZL": {ID: 20, Name: "New Zealand", Flag: "🇳🇿", Group: "E"},
		// Group F
		"ESP": {ID: 21, Name: "Spain", Flag: "🇪🇸", Group: "F"},
		"CRO": {ID: 22, Name: "Croatia", Flag: "🇭🇷", Group: "F"},
		"CMR": {ID: 23, Name: "Cameroon", Flag: "🇨🇲", Group: "F"},
		"TUN": {ID: 24, Name: "Tunisia", Flag: "🇹🇳", Group: "F"},
		// Group G
		"ENG": {ID: 25, Name: "England", Flag: "🏴󠁧󠁢󠁥󠁮󠁧󠁿", Group: "G"},
		"SEN": {ID: 26, Name: "Senegal", Flag: "🇸🇳", Group: "G"},
		"IRN": {ID: 27, Name: "Iran", Flag: "🇮🇷", Group: "G"},
		"VEN": {ID: 28, Name: "Venezuela", Flag: "🇻🇪", Group: "G"},
		// Group H
		"POR": {ID: 29, Name: "Portugal", Flag: "🇵🇹", Group: "H"},
		"TUR": {ID: 30, Name: "Türkiye", Flag: "🇹🇷", Group: "H"},
		"GHA": {ID: 31, Name: "Ghana", Flag: "🇬🇭", Group: "H"},
		"HND": {ID: 32, Name: "Honduras", Flag: "🇭🇳", Group: "H"},
		// Group I
		"NET": {ID: 33, Name: "Netherlands", Flag: "🇳🇱", Group: "I"},
		"ECU": {ID: 34, Name: "Ecuador", Flag: "🇪🇨", Group: "I"},
		"EGY": {ID: 35, Name: "Egypt", Flag: "🇪🇬", Group: "I"},
		"QAT": {ID: 36, Name: "Qatar", Flag: "🇶🇦", Group: "I"},
		// Group J
		"COL": {ID: 37, Name: "Colombia", Flag: "🇨🇴", Group: "J"},
		"CAN": {ID: 38, Name: "Canada", Flag: "🇨🇦", Group: "J"},
		"CIV": {ID: 39, Name: "Côte d'Ivoire", Flag: "🇨🇮", Group: "J"},
		"SVK": {ID: 40, Name: "Slovakia", Flag: "🇸🇰", Group: "J"},
		// Group K
		"JPN": {ID: 41, Name: "Japan", Flag: "🇯🇵", Group: "K"},
		"KOR": {ID: 42, Name: "South Korea", Flag: "🇰🇷", Group: "K"},
		"PER": {ID: 43, Name: "Peru", Flag: "🇵🇪", Group: "K"},
		"IDN": {ID: 44, Name: "Indonesia", Flag: "🇮🇩", Group: "K"},
		// Group L
		"ITA": {ID: 45, Name: "Italy", Flag: "🇮🇹", Group: "L"},
		"KSA": {ID: 46, Name: "Saudi Arabia", Flag: "🇸🇦", Group: "L"},
		"RSA": {ID: 47, Name: "South Africa", Flag: "🇿🇦", Group: "L"},
		"ROM": {ID: 48, Name: "Romania", Flag: "🇷🇴", Group: "L"},
	}

	// ── LIVE MATCHES (June 14, in progress) ──────────────────────────────────
	// Only LIVE matches carry scores + a minute counter.

	min34, min67 := 34, 67
	s0_0 := [2]int{0, 0}
	s5_2 := [2]int{5, 2}

	liveMatches := []wc.Match{
		{
			ID: 1, HomeTeam: teams["ENG"], AwayTeam: teams["SEN"],
			HomeScore: &s0_0[0], AwayScore: &s0_0[1],
			Status: wc.StatusLive, Minute: &min34,
			KickoffAt: today.Add(14 * time.Hour),
			Venue:     "SoFi Stadium, Inglewood CA",
			Group: "G", Stage: "group", Matchday: 1,
		},
		{
			ID: 2, HomeTeam: teams["ITA"], AwayTeam: teams["RSA"],
			HomeScore: &s5_2[0], AwayScore: &s5_2[1],
			Status: wc.StatusLive, Minute: &min67,
			KickoffAt: today.Add(17 * time.Hour),
			Venue:     "Estadio BBVA, Monterrey",
			Group: "L", Stage: "group", Matchday: 1,
		},
	}

	// ── FINISHED MATCHES (FT — June 11–13, matchday 1) ───────────────────────
	// All FT matches carry a score. No score fields on any other status.

	s2_0 := [2]int{2, 0}
	s3_1 := [2]int{3, 1}
	s2_1 := [2]int{2, 1}
	s1_1 := [2]int{1, 1}
	s4_0 := [2]int{4, 0}
	s3_0 := [2]int{3, 0}
	s2_2 := [2]int{2, 2}

	finishedMatches := []wc.Match{
		// June 11 — WC2026 opening day
		{
			ID: 101, HomeTeam: teams["USA"], AwayTeam: teams["PAN"],
			HomeScore: &s2_0[0], AwayScore: &s2_0[1],
			Status:    wc.StatusFinished,
			KickoffAt: threeDaysAgo.Add(20 * time.Hour),
			Venue:     "MetLife Stadium, New Jersey",
			Group: "A", Stage: "group", Matchday: 1,
		},
		{
			ID: 102, HomeTeam: teams["BRA"], AwayTeam: teams["MEX"],
			HomeScore: &s3_1[0], AwayScore: &s3_1[1],
			Status:    wc.StatusFinished,
			KickoffAt: threeDaysAgo.Add(23 * time.Hour),
			Venue:     "AT&T Stadium, Arlington TX",
			Group: "C", Stage: "group", Matchday: 1,
		},
		// June 12
		{
			ID: 103, HomeTeam: teams["ARG"], AwayTeam: teams["CHI"],
			HomeScore: &s2_0[0], AwayScore: &s2_0[1],
			Status:    wc.StatusFinished,
			KickoffAt: twoDaysAgo.Add(17 * time.Hour),
			Venue:     "Rose Bowl, Pasadena CA",
			Group: "B", Stage: "group", Matchday: 1,
		},
		{
			ID: 104, HomeTeam: teams["FRA"], AwayTeam: teams["DEN"],
			HomeScore: &s1_1[0], AwayScore: &s1_1[1],
			Status:    wc.StatusFinished,
			KickoffAt: twoDaysAgo.Add(20 * time.Hour),
			Venue:     "Mercedes-Benz Stadium, Atlanta GA",
			Group: "D", Stage: "group", Matchday: 1,
		},
		// June 13
		{
			ID: 105, HomeTeam: teams["GER"], AwayTeam: teams["SRB"],
			HomeScore: &s2_1[0], AwayScore: &s2_1[1],
			Status:    wc.StatusFinished,
			KickoffAt: yesterday.Add(14 * time.Hour),
			Venue:     "Levi's Stadium, Santa Clara CA",
			Group: "E", Stage: "group", Matchday: 1,
		},
		{
			ID: 106, HomeTeam: teams["ESP"], AwayTeam: teams["TUN"],
			HomeScore: &s4_0[0], AwayScore: &s4_0[1],
			Status:    wc.StatusFinished,
			KickoffAt: yesterday.Add(17 * time.Hour),
			Venue:     "NRG Stadium, Houston TX",
			Group: "F", Stage: "group", Matchday: 1,
		},
		{
			ID: 107, HomeTeam: teams["POR"], AwayTeam: teams["GHA"],
			HomeScore: &s3_0[0], AwayScore: &s3_0[1],
			Status:    wc.StatusFinished,
			KickoffAt: yesterday.Add(20 * time.Hour),
			Venue:     "BMO Field, Toronto",
			Group: "H", Stage: "group", Matchday: 1,
		},
		{
			ID: 108, HomeTeam: teams["JPN"], AwayTeam: teams["KOR"],
			HomeScore: &s2_2[0], AwayScore: &s2_2[1],
			Status:    wc.StatusFinished,
			KickoffAt: yesterday.Add(23 * time.Hour),
			Venue:     "BC Place, Vancouver",
			Group: "K", Stage: "group", Matchday: 1,
		},
	}

	// ── NOT STARTED / UPCOMING (NS — no scores) ───────────────────────────────
	// NS matches carry no HomeScore/AwayScore — only kickoff time and venue.

	upcomingMatches := []wc.Match{
		// Later today (June 14)
		{
			ID: 201, HomeTeam: teams["URU"], AwayTeam: teams["BEL"],
			Status:    wc.StatusUpcoming,
			KickoffAt: today.Add(21 * time.Hour),
			Venue:     "Estadio Azteca, Mexico City",
			Group: "A", Stage: "group", Matchday: 1,
		},
		// June 15
		{
			ID: 202, HomeTeam: teams["COL"], AwayTeam: teams["CIV"],
			Status:    wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(13 * time.Hour),
			Venue:     "Gillette Stadium, Foxborough MA",
			Group: "J", Stage: "group", Matchday: 1,
		},
		{
			ID: 203, HomeTeam: teams["MAR"], AwayTeam: teams["AUS"],
			Status:    wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(17 * time.Hour),
			Venue:     "AT&T Stadium, Arlington TX",
			Group: "D", Stage: "group", Matchday: 1,
		},
		{
			ID: 204, HomeTeam: teams["POL"], AwayTeam: teams["NGA"],
			Status:    wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(20 * time.Hour),
			Venue:     "MetLife Stadium, New Jersey",
			Group: "B", Stage: "group", Matchday: 1,
		},
		{
			ID: 205, HomeTeam: teams["PER"], AwayTeam: teams["IDN"],
			Status:    wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(23 * time.Hour),
			Venue:     "SoFi Stadium, Inglewood CA",
			Group: "K", Stage: "group", Matchday: 1,
		},
		// June 16
		{
			ID: 206, HomeTeam: teams["CRO"], AwayTeam: teams["CMR"],
			Status:    wc.StatusUpcoming,
			KickoffAt: twoDays.Add(14 * time.Hour),
			Venue:     "NRG Stadium, Houston TX",
			Group: "F", Stage: "group", Matchday: 1,
		},
		{
			ID: 207, HomeTeam: teams["TUR"], AwayTeam: teams["HND"],
			Status:    wc.StatusUpcoming,
			KickoffAt: twoDays.Add(17 * time.Hour),
			Venue:     "Estadio Azteca, Mexico City",
			Group: "H", Stage: "group", Matchday: 1,
		},
		{
			ID: 208, HomeTeam: teams["IRN"], AwayTeam: teams["VEN"],
			Status:    wc.StatusUpcoming,
			KickoffAt: twoDays.Add(20 * time.Hour),
			Venue:     "Rose Bowl, Pasadena CA",
			Group: "G", Stage: "group", Matchday: 1,
		},
	}

	// ── STANDINGS (reflects FT results only — LIVE games not yet counted) ─────

	standings := map[string][]wc.GroupRow{
		"A": {
			{Team: teams["USA"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 0, GD: 2, Pts: 3},
			{Team: teams["BEL"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["URU"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["PAN"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
		},
		"B": {
			{Team: teams["ARG"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 0, GD: 2, Pts: 3},
			{Team: teams["POL"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["NGA"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CHI"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
		},
		"C": {
			{Team: teams["BRA"], Played: 1, W: 1, D: 0, L: 0, GF: 3, GA: 1, GD: 2, Pts: 3},
			{Team: teams["SUI"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["ALG"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["MEX"], Played: 1, W: 0, D: 0, L: 1, GF: 1, GA: 3, GD: -2, Pts: 0},
		},
		"D": {
			{Team: teams["FRA"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["DEN"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["MAR"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["AUS"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"E": {
			{Team: teams["GER"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 1, GD: 1, Pts: 3},
			{Team: teams["CRC"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["NZL"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["SRB"], Played: 1, W: 0, D: 0, L: 1, GF: 1, GA: 2, GD: -1, Pts: 0},
		},
		"F": {
			{Team: teams["ESP"], Played: 1, W: 1, D: 0, L: 0, GF: 4, GA: 0, GD: 4, Pts: 3},
			{Team: teams["CRO"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CMR"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["TUN"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 4, GD: -4, Pts: 0},
		},
		"G": {
			// England vs Senegal in progress — standings unchanged until FT
			{Team: teams["ENG"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["SEN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["IRN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["VEN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"H": {
			{Team: teams["POR"], Played: 1, W: 1, D: 0, L: 0, GF: 3, GA: 0, GD: 3, Pts: 3},
			{Team: teams["TUR"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["HND"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["GHA"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 3, GD: -3, Pts: 0},
		},
		"I": {
			// No matchday 1 fixtures played yet in Group I
			{Team: teams["NET"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["ECU"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["EGY"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["QAT"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"J": {
			{Team: teams["COL"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CAN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CIV"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["SVK"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"K": {
			{Team: teams["JPN"], Played: 1, W: 0, D: 1, L: 0, GF: 2, GA: 2, GD: 0, Pts: 1},
			{Team: teams["KOR"], Played: 1, W: 0, D: 1, L: 0, GF: 2, GA: 2, GD: 0, Pts: 1},
			{Team: teams["PER"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["IDN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"L": {
			// Italy vs South Africa in progress — standings unchanged until FT
			{Team: teams["ITA"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["ROM"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["RSA"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["KSA"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
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
