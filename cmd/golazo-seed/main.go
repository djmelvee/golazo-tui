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

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	// ── TEAMS ────────────────────────────────────────────────────────────────

	teams := map[string]wc.Team{
		// Group A
		"USA":         {ID: 1, Name: "United States", Flag: "🇺🇸", Group: "A"},
		"PAN":         {ID: 2, Name: "Panama", Flag: "🇵🇦", Group: "A"},
		"URU":         {ID: 3, Name: "Uruguay", Flag: "🇺🇾", Group: "A"},
		"BEL":         {ID: 4, Name: "Belgium", Flag: "🇧🇪", Group: "A"},
		// Group B
		"ARG":         {ID: 5, Name: "Argentina", Flag: "🇦🇷", Group: "B"},
		"CHI":         {ID: 6, Name: "Chile", Flag: "🇨🇱", Group: "B"},
		"POL":         {ID: 7, Name: "Poland", Flag: "🇵🇱", Group: "B"},
		"NGA":         {ID: 8, Name: "Nigeria", Flag: "🇳🇬", Group: "B"},
		// Group C
		"BRA":         {ID: 9, Name: "Brazil", Flag: "🇧🇷", Group: "C"},
		"MEX":         {ID: 10, Name: "Mexico", Flag: "🇲🇽", Group: "C"},
		"SUI":         {ID: 11, Name: "Switzerland", Flag: "🇨🇭", Group: "C"},
		"ALG":         {ID: 12, Name: "Algeria", Flag: "🇩🇿", Group: "C"},
		// Group D
		"FRA":         {ID: 13, Name: "France", Flag: "🇫🇷", Group: "D"},
		"AUS":         {ID: 14, Name: "Australia", Flag: "🇦🇺", Group: "D"},
		"MAR":         {ID: 15, Name: "Morocco", Flag: "🇲🇦", Group: "D"},
		"DEN":         {ID: 16, Name: "Denmark", Flag: "🇩🇰", Group: "D"},
		// Group E
		"GER":         {ID: 17, Name: "Germany", Flag: "🇩🇪", Group: "E"},
		"SRB":         {ID: 18, Name: "Serbia", Flag: "🇷🇸", Group: "E"},
		"CRC":         {ID: 19, Name: "Costa Rica", Flag: "🇨🇷", Group: "E"},
		"NZL":         {ID: 20, Name: "New Zealand", Flag: "🇳🇿", Group: "E"},
		// Group F
		"ESP":         {ID: 21, Name: "Spain", Flag: "🇪🇸", Group: "F"},
		"CRO":         {ID: 22, Name: "Croatia", Flag: "🇭🇷", Group: "F"},
		"CMR":         {ID: 23, Name: "Cameroon", Flag: "🇨🇲", Group: "F"},
		"TUN":         {ID: 24, Name: "Tunisia", Flag: "🇹🇳", Group: "F"},
		// Group G
		"ENG":         {ID: 25, Name: "England", Flag: "🏴󠁧󠁢󠁥󠁮󠁧󠁿", Group: "G"},
		"SEN":         {ID: 26, Name: "Senegal", Flag: "🇸🇳", Group: "G"},
		"IRN":         {ID: 27, Name: "Iran", Flag: "🇮🇷", Group: "G"},
		"VEN":         {ID: 28, Name: "Venezuela", Flag: "🇻🇪", Group: "G"},
		// Group H
		"POR":         {ID: 29, Name: "Portugal", Flag: "🇵🇹", Group: "H"},
		"TUR":         {ID: 30, Name: "Türkiye", Flag: "🇹🇷", Group: "H"},
		"GHA":         {ID: 31, Name: "Ghana", Flag: "🇬🇭", Group: "H"},
		"HND":         {ID: 32, Name: "Honduras", Flag: "🇭🇳", Group: "H"},
		// Group I
		"NET":         {ID: 33, Name: "Netherlands", Flag: "🇳🇱", Group: "I"},
		"ECU":         {ID: 34, Name: "Ecuador", Flag: "🇪🇨", Group: "I"},
		"EGY":         {ID: 35, Name: "Egypt", Flag: "🇪🇬", Group: "I"},
		"QAT":         {ID: 36, Name: "Qatar", Flag: "🇶🇦", Group: "I"},
		// Group J
		"COL":         {ID: 37, Name: "Colombia", Flag: "🇨🇴", Group: "J"},
		"CAN":         {ID: 38, Name: "Canada", Flag: "🇨🇦", Group: "J"},
		"CIV":         {ID: 39, Name: "Côte d'Ivoire", Flag: "🇨🇮", Group: "J"},
		"SVK":         {ID: 40, Name: "Slovakia", Flag: "🇸🇰", Group: "J"},
		// Group K
		"JPN":         {ID: 41, Name: "Japan", Flag: "🇯🇵", Group: "K"},
		"KOR":         {ID: 42, Name: "South Korea", Flag: "🇰🇷", Group: "K"},
		"PER":         {ID: 43, Name: "Peru", Flag: "🇵🇪", Group: "K"},
		"IDN":         {ID: 44, Name: "Indonesia", Flag: "🇮🇩", Group: "K"},
		// Group L
		"ITA":         {ID: 45, Name: "Italy", Flag: "🇮🇹", Group: "L"},
		"KSA":         {ID: 46, Name: "Saudi Arabia", Flag: "🇸🇦", Group: "L"},
		"RSA":         {ID: 47, Name: "South Africa", Flag: "🇿🇦", Group: "L"},
		"ROM":         {ID: 48, Name: "Romania", Flag: "🇷🇴", Group: "L"},
	}

	// ── LIVE MATCHES ─────────────────────────────────────────────────────────

	minute22, minute45, minute67, minute74 := 22, 45, 67, 74
	score1_0 := [2]int{1, 0}
	score2_1 := [2]int{2, 1}
	score0_0 := [2]int{0, 0}
	score3_1 := [2]int{3, 1}

	liveMatches := []wc.Match{
		{
			ID: 1, HomeTeam: teams["USA"], AwayTeam: teams["PAN"],
			HomeScore: &score1_0[0], AwayScore: &score1_0[1],
			Status: wc.StatusLive, Minute: &minute22,
			KickoffAt: today.Add(13 * time.Hour), Venue: "MetLife Stadium, New Jersey",
			Group: "A", Stage: "group", Matchday: 2,
		},
		{
			ID: 2, HomeTeam: teams["BRA"], AwayTeam: teams["MEX"],
			HomeScore: &score2_1[0], AwayScore: &score2_1[1],
			Status: wc.StatusLive, Minute: &minute45,
			KickoffAt: today.Add(16 * time.Hour), Venue: "AT&T Stadium, Arlington TX",
			Group: "C", Stage: "group", Matchday: 2,
		},
		{
			ID: 3, HomeTeam: teams["ENG"], AwayTeam: teams["SEN"],
			HomeScore: &score0_0[0], AwayScore: &score0_0[1],
			Status: wc.StatusLive, Minute: &minute67,
			KickoffAt: today.Add(19 * time.Hour), Venue: "SoFi Stadium, Inglewood CA",
			Group: "G", Stage: "group", Matchday: 2,
		},
		{
			ID: 4, HomeTeam: teams["GER"], AwayTeam: teams["SRB"],
			HomeScore: &score3_1[0], AwayScore: &score3_1[1],
			Status: wc.StatusLive, Minute: &minute74,
			KickoffAt: today.Add(20 * time.Hour), Venue: "Levi's Stadium, Santa Clara CA",
			Group: "E", Stage: "group", Matchday: 2,
		},
	}

	// ── FINISHED MATCHES ─────────────────────────────────────────────────────

	score2_0 := [2]int{2, 0}
	score1_1 := [2]int{1, 1}
	score4_0 := [2]int{4, 0}
	score0_1 := [2]int{0, 1}
	score3_0 := [2]int{3, 0}
	score2_2 := [2]int{2, 2}
	score1_2 := [2]int{1, 2}
	score5_2 := [2]int{5, 2}

	yesterday := today.Add(-24 * time.Hour)

	finishedMatches := []wc.Match{
		{
			ID: 101, HomeTeam: teams["ARG"], AwayTeam: teams["CHI"],
			HomeScore: &score2_0[0], AwayScore: &score2_0[1],
			Status: wc.StatusFinished,
			KickoffAt: yesterday.Add(13 * time.Hour), Venue: "Rose Bowl, Pasadena CA",
			Group: "B", Stage: "group", Matchday: 1,
		},
		{
			ID: 102, HomeTeam: teams["FRA"], AwayTeam: teams["DEN"],
			HomeScore: &score1_1[0], AwayScore: &score1_1[1],
			Status: wc.StatusFinished,
			KickoffAt: yesterday.Add(16 * time.Hour), Venue: "Mercedes-Benz Stadium, Atlanta GA",
			Group: "D", Stage: "group", Matchday: 1,
		},
		{
			ID: 103, HomeTeam: teams["ESP"], AwayTeam: teams["TUN"],
			HomeScore: &score4_0[0], AwayScore: &score4_0[1],
			Status: wc.StatusFinished,
			KickoffAt: yesterday.Add(19 * time.Hour), Venue: "NRG Stadium, Houston TX",
			Group: "F", Stage: "group", Matchday: 1,
		},
		{
			ID: 104, HomeTeam: teams["URU"], AwayTeam: teams["BEL"],
			HomeScore: &score0_1[0], AwayScore: &score0_1[1],
			Status: wc.StatusFinished,
			KickoffAt: yesterday.Add(19 * time.Hour), Venue: "Estadio Azteca, Mexico City",
			Group: "A", Stage: "group", Matchday: 1,
		},
		{
			ID: 105, HomeTeam: teams["POR"], AwayTeam: teams["GHA"],
			HomeScore: &score3_0[0], AwayScore: &score3_0[1],
			Status: wc.StatusFinished,
			KickoffAt: yesterday.Add(13 * time.Hour), Venue: "BMO Field, Toronto",
			Group: "H", Stage: "group", Matchday: 1,
		},
		{
			ID: 106, HomeTeam: teams["JPN"], AwayTeam: teams["KOR"],
			HomeScore: &score2_2[0], AwayScore: &score2_2[1],
			Status: wc.StatusFinished,
			KickoffAt: yesterday.Add(16 * time.Hour), Venue: "BC Place, Vancouver",
			Group: "K", Stage: "group", Matchday: 1,
		},
		{
			ID: 107, HomeTeam: teams["NET"], AwayTeam: teams["QAT"],
			HomeScore: &score1_2[0], AwayScore: &score1_2[1],
			Status: wc.StatusFinished,
			KickoffAt: yesterday.Add(21 * time.Hour), Venue: "Gillette Stadium, Foxborough MA",
			Group: "I", Stage: "group", Matchday: 1,
		},
		{
			ID: 108, HomeTeam: teams["ITA"], AwayTeam: teams["RSA"],
			HomeScore: &score5_2[0], AwayScore: &score5_2[1],
			Status: wc.StatusFinished,
			KickoffAt: yesterday.Add(21 * time.Hour), Venue: "Estadio BBVA, Monterrey",
			Group: "L", Stage: "group", Matchday: 1,
		},
	}

	// ── UPCOMING MATCHES ─────────────────────────────────────────────────────

	tomorrow := today.Add(24 * time.Hour)
	twoDays := today.Add(48 * time.Hour)

	upcomingMatches := []wc.Match{
		{
			ID: 201, HomeTeam: teams["POL"], AwayTeam: teams["NGA"],
			Status: wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(13 * time.Hour), Venue: "MetLife Stadium, New Jersey",
			Group: "B", Stage: "group", Matchday: 2,
		},
		{
			ID: 202, HomeTeam: teams["MAR"], AwayTeam: teams["AUS"],
			Status: wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(16 * time.Hour), Venue: "AT&T Stadium, Arlington TX",
			Group: "D", Stage: "group", Matchday: 2,
		},
		{
			ID: 203, HomeTeam: teams["TUR"], AwayTeam: teams["HND"],
			Status: wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(19 * time.Hour), Venue: "Estadio Azteca, Mexico City",
			Group: "H", Stage: "group", Matchday: 2,
		},
		{
			ID: 204, HomeTeam: teams["PER"], AwayTeam: teams["IDN"],
			Status: wc.StatusUpcoming,
			KickoffAt: tomorrow.Add(21 * time.Hour), Venue: "SoFi Stadium, Inglewood CA",
			Group: "K", Stage: "group", Matchday: 2,
		},
		{
			ID: 205, HomeTeam: teams["CAN"], AwayTeam: teams["SVK"],
			Status: wc.StatusUpcoming,
			KickoffAt: twoDays.Add(13 * time.Hour), Venue: "BC Place, Vancouver",
			Group: "J", Stage: "group", Matchday: 2,
		},
		{
			ID: 206, HomeTeam: teams["ITA"], AwayTeam: teams["ROM"],
			Status: wc.StatusUpcoming,
			KickoffAt: twoDays.Add(16 * time.Hour), Venue: "Mercedes-Benz Stadium, Atlanta GA",
			Group: "L", Stage: "group", Matchday: 2,
		},
		{
			ID: 207, HomeTeam: teams["CRO"], AwayTeam: teams["CMR"],
			Status: wc.StatusUpcoming,
			KickoffAt: twoDays.Add(19 * time.Hour), Venue: "NRG Stadium, Houston TX",
			Group: "F", Stage: "group", Matchday: 2,
		},
		{
			ID: 208, HomeTeam: teams["IRN"], AwayTeam: teams["VEN"],
			Status: wc.StatusUpcoming,
			KickoffAt: twoDays.Add(21 * time.Hour), Venue: "Rose Bowl, Pasadena CA",
			Group: "G", Stage: "group", Matchday: 2,
		},
	}

	// ── STANDINGS ─────────────────────────────────────────────────────────────

	standings := map[string][]wc.GroupRow{
		"A": {
			{Team: teams["USA"], Played: 2, W: 1, D: 1, L: 0, GF: 2, GA: 1, GD: 1, Pts: 4},
			{Team: teams["BEL"], Played: 2, W: 1, D: 1, L: 0, GF: 2, GA: 1, GD: 1, Pts: 4},
			{Team: teams["PAN"], Played: 2, W: 0, D: 1, L: 1, GF: 1, GA: 2, GD: -1, Pts: 1},
			{Team: teams["URU"], Played: 2, W: 0, D: 1, L: 1, GF: 1, GA: 2, GD: -1, Pts: 1},
		},
		"B": {
			{Team: teams["ARG"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 0, GD: 2, Pts: 3},
			{Team: teams["NGA"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
			{Team: teams["POL"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CHI"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
		},
		"C": {
			{Team: teams["BRA"], Played: 2, W: 2, D: 0, L: 0, GF: 5, GA: 1, GD: 4, Pts: 6},
			{Team: teams["SUI"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 0, GD: 2, Pts: 3},
			{Team: teams["MEX"], Played: 2, W: 0, D: 0, L: 2, GF: 1, GA: 4, GD: -3, Pts: 0},
			{Team: teams["ALG"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
		},
		"D": {
			{Team: teams["MAR"], Played: 2, W: 1, D: 1, L: 0, GF: 2, GA: 1, GD: 1, Pts: 4},
			{Team: teams["FRA"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["DEN"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["AUS"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 4, GD: -4, Pts: 0},
		},
		"E": {
			{Team: teams["GER"], Played: 2, W: 2, D: 0, L: 0, GF: 6, GA: 2, GD: 4, Pts: 6},
			{Team: teams["SRB"], Played: 2, W: 1, D: 0, L: 1, GF: 3, GA: 5, GD: -2, Pts: 3},
			{Team: teams["CRC"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
			{Team: teams["NZL"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 3, GD: -3, Pts: 0},
		},
		"F": {
			{Team: teams["ESP"], Played: 1, W: 1, D: 0, L: 0, GF: 4, GA: 0, GD: 4, Pts: 3},
			{Team: teams["CRO"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["CMR"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["TUN"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 4, GD: -4, Pts: 0},
		},
		"G": {
			{Team: teams["ENG"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 0, GD: 2, Pts: 3},
			{Team: teams["SEN"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
			{Team: teams["IRN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["VEN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"H": {
			{Team: teams["POR"], Played: 1, W: 1, D: 0, L: 0, GF: 3, GA: 0, GD: 3, Pts: 3},
			{Team: teams["TUR"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["GHA"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 3, GD: -3, Pts: 0},
			{Team: teams["HND"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"I": {
			{Team: teams["QAT"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 1, GD: 1, Pts: 3},
			{Team: teams["EGY"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["ECU"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["NET"], Played: 1, W: 0, D: 0, L: 1, GF: 1, GA: 2, GD: -1, Pts: 0},
		},
		"J": {
			{Team: teams["COL"], Played: 1, W: 1, D: 0, L: 0, GF: 2, GA: 0, GD: 2, Pts: 3},
			{Team: teams["CIV"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["CAN"], Played: 1, W: 0, D: 1, L: 0, GF: 1, GA: 1, GD: 0, Pts: 1},
			{Team: teams["SVK"], Played: 1, W: 0, D: 0, L: 1, GF: 0, GA: 2, GD: -2, Pts: 0},
		},
		"K": {
			{Team: teams["JPN"], Played: 1, W: 0, D: 1, L: 0, GF: 2, GA: 2, GD: 0, Pts: 1},
			{Team: teams["KOR"], Played: 1, W: 0, D: 1, L: 0, GF: 2, GA: 2, GD: 0, Pts: 1},
			{Team: teams["PER"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["IDN"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
		},
		"L": {
			{Team: teams["ITA"], Played: 1, W: 1, D: 0, L: 0, GF: 5, GA: 2, GD: 3, Pts: 3},
			{Team: teams["ROM"], Played: 0, W: 0, D: 0, L: 0, GF: 0, GA: 0, GD: 0, Pts: 0},
			{Team: teams["RSA"], Played: 1, W: 0, D: 0, L: 1, GF: 2, GA: 5, GD: -3, Pts: 0},
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

	fmt.Printf("✓ Seeded %d live, %d FT, %d upcoming matches + 12 groups (48 teams)\n",
		len(liveMatches), len(finishedMatches), len(upcomingMatches))
	fmt.Printf("  DB: %s\n", dbPath)
}
