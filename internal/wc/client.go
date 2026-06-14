package wc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"sync"
	"time"
)

// Client is a JWT-authenticated HTTP client for the worldcup26.ir API.
type Client struct {
	base         string
	token        string
	http         *http.Client
	mu           sync.Mutex
	gameCache    []apiGame
	gameCachedAt time.Time
}

// New creates a new Client. base is the API root URL, token is the JWT bearer token.
func New(base, token string) *Client {
	return &Client{
		base:  base,
		token: token,
		http:  &http.Client{Timeout: 10 * time.Second},
	}
}

// apiGame mirrors the JSON returned by GET /get/games on worldcup26.ir:3050.
// Status is derived from finished + time_elapsed — the API has no status string.
type apiGame struct {
	ID             int    `json:"id"`
	HomeTeamID     int    `json:"home_team_id"`
	AwayTeamID     int    `json:"away_team_id"`
	HomeScore      int    `json:"home_score"`
	AwayScore      int    `json:"away_score"`
	HomeTeamNameEn string `json:"home_team_name_en"`
	AwayTeamNameEn string `json:"away_team_name_en"`
	Group          string `json:"group"`
	Matchday       int    `json:"matchday"`
	LocalDate      string `json:"local_date"` // RFC3339 e.g. "2026-06-11T20:00:00Z"
	Finished       bool   `json:"finished"`
	TimeElapsed    int    `json:"time_elapsed"`
	StadiumID      int    `json:"stadium_id"`
	Type           string `json:"type"` // "group", "round_of_32", etc.
}

// apiTeam mirrors the JSON returned by GET /get/teams.
type apiTeam struct {
	ID     int    `json:"id"`
	NameEn string `json:"name_en"`
	Group  string `json:"group"`
}

// apiStadium mirrors the JSON returned by GET /get/stadiums.
type apiStadium struct {
	ID     int    `json:"id"`
	NameEn string `json:"name_en"`
	City   string `json:"city_en"`
}

// flagMap maps official English team names to flag emojis for all 48 WC 2026
// qualified nations. The worldcup26.ir API does not supply flag emojis.
var flagMap = map[string]string{
	// Group A
	"Mexico":                 "🇲🇽",
	"South Korea":            "🇰🇷",
	"Czech Republic":         "🇨🇿",
	"South Africa":           "🇿🇦",
	// Group B
	"Canada":                 "🇨🇦",
	"Bosnia and Herzegovina": "🇧🇦",
	"Qatar":                  "🇶🇦",
	"Switzerland":            "🇨🇭",
	// Group C
	"Brazil":                 "🇧🇷",
	"Morocco":                "🇲🇦",
	"Scotland":               "🏴󠁧󠁢󠁳󠁣󠁴󠁿",
	"Haiti":                  "🇭🇹",
	// Group D
	"United States":          "🇺🇸",
	"Paraguay":               "🇵🇾",
	"Australia":              "🇦🇺",
	"Türkiye":                "🇹🇷",
	// Group E
	"Germany":                "🇩🇪",
	"Curaçao":                "🇨🇼",
	"Côte d'Ivoire":          "🇨🇮",
	"Ecuador":                "🇪🇨",
	// Group F
	"Netherlands":            "🇳🇱",
	"Japan":                  "🇯🇵",
	"Sweden":                 "🇸🇪",
	"Tunisia":                "🇹🇳",
	// Group G
	"Belgium":                "🇧🇪",
	"Egypt":                  "🇪🇬",
	"Iran":                   "🇮🇷",
	"New Zealand":            "🇳🇿",
	// Group H
	"Spain":                  "🇪🇸",
	"Cape Verde":             "🇨🇻",
	"Saudi Arabia":           "🇸🇦",
	"Uruguay":                "🇺🇾",
	// Group I
	"France":                 "🇫🇷",
	"Senegal":                "🇸🇳",
	"Iraq":                   "🇮🇶",
	"Norway":                 "🇳🇴",
	// Group J
	"Argentina":              "🇦🇷",
	"Algeria":                "🇩🇿",
	"Austria":                "🇦🇹",
	"Jordan":                 "🇯🇴",
	// Group K
	"Portugal":               "🇵🇹",
	"DR Congo":               "🇨🇩",
	"Uzbekistan":             "🇺🇿",
	"Colombia":               "🇨🇴",
	// Group L
	"England":                "🏴󠁧󠁢󠁥󠁮󠁧󠁿",
	"Croatia":                "🇭🇷",
	"Ghana":                  "🇬🇭",
	"Panama":                 "🇵🇦",
}

func (c *Client) get(ctx context.Context, path string, dest any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.base+path, nil)
	if err != nil {
		return err
	}
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API returned %d for %s", resp.StatusCode, path)
	}

	return json.NewDecoder(resp.Body).Decode(dest)
}

// getGames fetches all 104 games with a 30-second in-memory cache so that
// FetchMatches (called 3× per cycle) and FetchStandings share one HTTP call.
func (c *Client) getGames(ctx context.Context) ([]apiGame, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.gameCache != nil && time.Since(c.gameCachedAt) < 30*time.Second {
		return c.gameCache, nil
	}

	var games []apiGame
	if err := c.get(ctx, "/get/games", &games); err != nil {
		return nil, fmt.Errorf("fetch games: %w", err)
	}

	c.gameCache = games
	c.gameCachedAt = time.Now()
	return games, nil
}

// deriveStatus maps worldcup26.ir game fields to our MatchStatus constants.
func deriveStatus(g apiGame) MatchStatus {
	if g.Finished {
		return StatusFinished
	}
	if g.TimeElapsed > 0 {
		return StatusLive
	}
	return StatusUpcoming
}

// parseKickoff tries RFC3339, then datetime-without-zone, then date-only.
func parseKickoff(s string) time.Time {
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02T15:04:05", s); err == nil {
		return t
	}
	if t, err := time.Parse("2006-01-02", s); err == nil {
		return t
	}
	return time.Time{}
}

func gameToMatch(g apiGame, venue string) Match {
	home := Team{ID: g.HomeTeamID, Name: g.HomeTeamNameEn, Flag: flagMap[g.HomeTeamNameEn], Group: g.Group}
	away := Team{ID: g.AwayTeamID, Name: g.AwayTeamNameEn, Flag: flagMap[g.AwayTeamNameEn], Group: g.Group}

	status := deriveStatus(g)

	var homeScore, awayScore *int
	var minute *int

	if status == StatusFinished || status == StatusLive {
		hs, as := g.HomeScore, g.AwayScore
		homeScore, awayScore = &hs, &as
	}
	if status == StatusLive && g.TimeElapsed > 0 {
		m := g.TimeElapsed
		minute = &m
	}

	stage := g.Type
	if stage == "" {
		stage = "group"
	}

	return Match{
		ID:        g.ID,
		HomeTeam:  home,
		AwayTeam:  away,
		HomeScore: homeScore,
		AwayScore: awayScore,
		Status:    status,
		Minute:    minute,
		KickoffAt: parseKickoff(g.LocalDate),
		Venue:     venue,
		Group:     g.Group,
		Stage:     stage,
		Matchday:  g.Matchday,
	}
}

// fetchStadiums returns a stadiumID→"Name, City" map. Non-fatal on error.
func (c *Client) fetchStadiums(ctx context.Context) map[int]string {
	var stads []apiStadium
	if err := c.get(ctx, "/get/stadiums", &stads); err != nil {
		return map[int]string{}
	}
	m := make(map[int]string, len(stads))
	for _, s := range stads {
		name := s.NameEn
		if s.City != "" {
			name += ", " + s.City
		}
		m[s.ID] = name
	}
	return m
}

// FetchMatches calls GET /get/games and returns matches with the given status.
// status must be one of the StatusXxx constants ("LIVE", "FT", "NS").
func (c *Client) FetchMatches(ctx context.Context, status string) ([]Match, error) {
	games, err := c.getGames(ctx)
	if err != nil {
		return nil, err
	}

	stadiums := c.fetchStadiums(ctx)

	want := MatchStatus(status)
	var matches []Match
	for _, g := range games {
		if deriveStatus(g) != want {
			continue
		}
		matches = append(matches, gameToMatch(g, stadiums[g.StadiumID]))
	}
	return matches, nil
}

// FetchStandings derives group standings from finished games (GET /get/games).
// Teams that have not yet played appear with 0 stats via GET /get/teams.
// W/D/L are computed here — the /get/groups endpoint does not include them.
func (c *Client) FetchStandings(ctx context.Context) (map[string][]GroupRow, error) {
	games, err := c.getGames(ctx)
	if err != nil {
		return nil, err
	}

	type stats struct {
		team                     Team
		played, w, d, l, gf, ga int
	}

	groups := make(map[string]map[int]*stats)

	// Pre-populate all teams so unplayed teams appear at 0
	var apiTeams []apiTeam
	if err := c.get(ctx, "/get/teams", &apiTeams); err == nil {
		for _, t := range apiTeams {
			grp := t.Group
			if groups[grp] == nil {
				groups[grp] = make(map[int]*stats)
			}
			groups[grp][t.ID] = &stats{
				team: Team{ID: t.ID, Name: t.NameEn, Flag: flagMap[t.NameEn], Group: grp},
			}
		}
	}

	for _, g := range games {
		if !g.Finished {
			continue
		}
		grp := g.Group
		if groups[grp] == nil {
			groups[grp] = make(map[int]*stats)
		}
		if groups[grp][g.HomeTeamID] == nil {
			groups[grp][g.HomeTeamID] = &stats{
				team: Team{ID: g.HomeTeamID, Name: g.HomeTeamNameEn, Flag: flagMap[g.HomeTeamNameEn], Group: grp},
			}
		}
		if groups[grp][g.AwayTeamID] == nil {
			groups[grp][g.AwayTeamID] = &stats{
				team: Team{ID: g.AwayTeamID, Name: g.AwayTeamNameEn, Flag: flagMap[g.AwayTeamNameEn], Group: grp},
			}
		}

		home := groups[grp][g.HomeTeamID]
		away := groups[grp][g.AwayTeamID]

		home.played++
		home.gf += g.HomeScore
		home.ga += g.AwayScore
		away.played++
		away.gf += g.AwayScore
		away.ga += g.HomeScore

		switch {
		case g.HomeScore > g.AwayScore:
			home.w++
			away.l++
		case g.HomeScore < g.AwayScore:
			home.l++
			away.w++
		default:
			home.d++
			away.d++
		}
	}

	result := make(map[string][]GroupRow, len(groups))
	for grp, teamMap := range groups {
		rows := make([]GroupRow, 0, len(teamMap))
		for _, s := range teamMap {
			pts := s.w*3 + s.d
			rows = append(rows, GroupRow{
				Team:   s.team,
				Played: s.played,
				W: s.w, D: s.d, L: s.l,
				GF: s.gf, GA: s.ga,
				GD: s.gf - s.ga,
				Pts: pts,
			})
		}
		sort.Slice(rows, func(i, j int) bool {
			if rows[i].Pts != rows[j].Pts {
				return rows[i].Pts > rows[j].Pts
			}
			return rows[i].GD > rows[j].GD
		})
		result[grp] = rows
	}
	return result, nil
}
