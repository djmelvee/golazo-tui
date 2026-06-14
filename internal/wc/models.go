package wc

import "time"

// GoalEvent records a detected score change during a live match.
// Minute is the API's time_elapsed at the moment the change was detected —
// approximate but typically accurate to within the fetcher poll interval.
type GoalEvent struct {
	MatchID    int       `json:"match_id"`
	HomeScore  int       `json:"home_score"`
	AwayScore  int       `json:"away_score"`
	Minute     int       `json:"minute"`
	ScoredBy   string    `json:"scored_by"` // "home" or "away"
	DetectedAt time.Time `json:"detected_at"`
}

type MatchStatus string

const (
	StatusLive     MatchStatus = "LIVE"
	StatusFinished MatchStatus = "FT"
	StatusUpcoming MatchStatus = "NS"
)

type Team struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Flag  string `json:"flag"`  // emoji e.g. "🇧🇷"
	Group string `json:"group"` // "A"–"L"
}

type Match struct {
	ID        int         `json:"id"`
	HomeTeam  Team        `json:"home_team"`
	AwayTeam  Team        `json:"away_team"`
	HomeScore *int        `json:"home_score"`
	AwayScore *int        `json:"away_score"`
	Status    MatchStatus `json:"status"`
	Minute    *int        `json:"minute"`   // nil unless live
	KickoffAt time.Time   `json:"kickoff_at"`
	Venue     string      `json:"venue"`
	Group     string      `json:"group"`    // "" for knockouts
	Stage     string      `json:"stage"`    // "group", "r32", "r16", "qf", "sf", "final"
	Matchday  int         `json:"matchday"` // 1–3 for group stage
}

type GroupRow struct {
	Team    Team `json:"team"`
	Played  int  `json:"played"`
	W       int  `json:"w"`
	D       int  `json:"d"`
	L       int  `json:"l"`
	GF      int  `json:"gf"`
	GA      int  `json:"ga"`
	GD      int  `json:"gd"` // GF - GA
	Pts     int  `json:"pts"`
}
