package wc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// flexInt unmarshals JSON numbers or numeric strings.
func flexInt(raw json.RawMessage) (int, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return 0, nil
	}
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return n, nil
	}
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return 0, fmt.Errorf("flexInt: %s", string(raw))
	}
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "null") {
		return 0, nil
	}
	return strconv.Atoi(s)
}

func flexBool(raw json.RawMessage) bool {
	if len(raw) == 0 || string(raw) == "null" {
		return false
	}
	var b bool
	if json.Unmarshal(raw, &b) == nil {
		return b
	}
	var s string
	if json.Unmarshal(raw, &s) != nil {
		return false
	}
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "true", "1", "yes":
		return true
	default:
		return false
	}
}

func flexString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var s string
	if json.Unmarshal(raw, &s) == nil {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(string(raw))
}

type elapsedState struct {
	finished   bool
	live       bool
	notStarted bool
	minute     int
}

func parseElapsed(raw json.RawMessage) elapsedState {
	if len(raw) == 0 || string(raw) == "null" {
		return elapsedState{notStarted: true}
	}
	var n int
	if json.Unmarshal(raw, &n) == nil {
		if n > 0 {
			return elapsedState{live: true, minute: n}
		}
		return elapsedState{notStarted: true}
	}
	s := strings.ToLower(strings.TrimSpace(flexString(raw)))
	switch s {
	case "finished", "ft", "fulltime", "full_time":
		return elapsedState{finished: true}
	case "live", "in", "playing":
		return elapsedState{live: true}
	case "notstarted", "not_started", "ns", "pre", "":
		return elapsedState{notStarted: true}
	default:
		if m, err := strconv.Atoi(s); err == nil && m > 0 {
			return elapsedState{live: true, minute: m}
		}
		return elapsedState{notStarted: true}
	}
}

type rawGame struct {
	ID             json.RawMessage `json:"id"`
	HomeTeamID     json.RawMessage `json:"home_team_id"`
	AwayTeamID     json.RawMessage `json:"away_team_id"`
	HomeScore      json.RawMessage `json:"home_score"`
	AwayScore      json.RawMessage `json:"away_score"`
	HomeTeamNameEn string          `json:"home_team_name_en"`
	AwayTeamNameEn string          `json:"away_team_name_en"`
	Group          string          `json:"group"`
	Matchday       json.RawMessage `json:"matchday"`
	LocalDate      string          `json:"local_date"`
	Finished       json.RawMessage `json:"finished"`
	TimeElapsed    json.RawMessage `json:"time_elapsed"`
	StadiumID      json.RawMessage `json:"stadium_id"`
	Type           string          `json:"type"`
	HomeScorers    json.RawMessage `json:"home_scorers"`
	AwayScorers    json.RawMessage `json:"away_scorers"`
}

func parseRawGame(r rawGame) (apiGame, error) {
	id, err := flexInt(r.ID)
	if err != nil {
		return apiGame{}, fmt.Errorf("id: %w", err)
	}
	homeID, err := flexInt(r.HomeTeamID)
	if err != nil {
		return apiGame{}, fmt.Errorf("home_team_id: %w", err)
	}
	awayID, err := flexInt(r.AwayTeamID)
	if err != nil {
		return apiGame{}, fmt.Errorf("away_team_id: %w", err)
	}
	homeScore, err := flexInt(r.HomeScore)
	if err != nil {
		return apiGame{}, fmt.Errorf("home_score: %w", err)
	}
	awayScore, err := flexInt(r.AwayScore)
	if err != nil {
		return apiGame{}, fmt.Errorf("away_score: %w", err)
	}
	matchday, err := flexInt(r.Matchday)
	if err != nil {
		return apiGame{}, fmt.Errorf("matchday: %w", err)
	}
	stadiumID, err := flexInt(r.StadiumID)
	if err != nil {
		return apiGame{}, fmt.Errorf("stadium_id: %w", err)
	}

	finished := flexBool(r.Finished)
	elapsed := parseElapsed(r.TimeElapsed)
	if elapsed.finished {
		finished = true
	}

	return apiGame{
		ID:             id,
		HomeTeamID:     homeID,
		AwayTeamID:     awayID,
		HomeScore:      homeScore,
		AwayScore:      awayScore,
		HomeTeamNameEn: r.HomeTeamNameEn,
		AwayTeamNameEn: r.AwayTeamNameEn,
		Group:          r.Group,
		Matchday:       matchday,
		LocalDate:      r.LocalDate,
		Finished:       finished,
		TimeElapsed:    elapsed.minute,
		IsLive:         elapsed.live,
		StadiumID:      stadiumID,
		Type:           r.Type,
		HomeScorers:    ParseScorers(flexString(r.HomeScorers), r.HomeTeamNameEn),
		AwayScorers:    ParseScorers(flexString(r.AwayScorers), r.AwayTeamNameEn),
	}, nil
}

func decodeGames(body []byte) ([]apiGame, error) {
	var wrapped struct {
		Games []rawGame `json:"games"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && len(wrapped.Games) > 0 {
		return parseRawGames(wrapped.Games)
	}

	var bare []rawGame
	if err := json.Unmarshal(body, &bare); err == nil && len(bare) > 0 {
		return parseRawGames(bare)
	}

	return nil, fmt.Errorf("unrecognized games response shape")
}

func parseRawGames(raw []rawGame) ([]apiGame, error) {
	out := make([]apiGame, 0, len(raw))
	for i, r := range raw {
		g, err := parseRawGame(r)
		if err != nil {
			return nil, fmt.Errorf("game %d: %w", i, err)
		}
		out = append(out, g)
	}
	return out, nil
}

type rawTeam struct {
	ID     json.RawMessage `json:"id"`
	NameEn string          `json:"name_en"`
	Group  string          `json:"group"`
	Groups string          `json:"groups"`
}

func parseRawTeam(r rawTeam) (apiTeam, error) {
	id, err := flexInt(r.ID)
	if err != nil {
		return apiTeam{}, err
	}
	grp := r.Group
	if grp == "" {
		grp = r.Groups
	}
	return apiTeam{ID: id, NameEn: r.NameEn, Group: grp}, nil
}

func decodeTeams(body []byte) ([]apiTeam, error) {
	var wrapped struct {
		Teams []rawTeam `json:"teams"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && len(wrapped.Teams) > 0 {
		return parseRawTeams(wrapped.Teams)
	}
	var bare []rawTeam
	if err := json.Unmarshal(body, &bare); err == nil && len(bare) > 0 {
		return parseRawTeams(bare)
	}
	return nil, fmt.Errorf("unrecognized teams response shape")
}

func parseRawTeams(raw []rawTeam) ([]apiTeam, error) {
	out := make([]apiTeam, 0, len(raw))
	for _, r := range raw {
		t, err := parseRawTeam(r)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, nil
}

type rawStadium struct {
	ID     json.RawMessage `json:"id"`
	NameEn string          `json:"name_en"`
	City   string          `json:"city_en"`
}

func parseRawStadium(r rawStadium) (apiStadium, error) {
	id, err := flexInt(r.ID)
	if err != nil {
		return apiStadium{}, err
	}
	return apiStadium{ID: id, NameEn: r.NameEn, City: r.City}, nil
}

func decodeStadiums(body []byte) ([]apiStadium, error) {
	var wrapped struct {
		Stadiums []rawStadium `json:"stadiums"`
	}
	if err := json.Unmarshal(body, &wrapped); err == nil && len(wrapped.Stadiums) > 0 {
		return parseRawStadiums(wrapped.Stadiums)
	}
	var bare []rawStadium
	if err := json.Unmarshal(body, &bare); err == nil && len(bare) > 0 {
		return parseRawStadiums(bare)
	}
	return nil, fmt.Errorf("unrecognized stadiums response shape")
}

func parseRawStadiums(raw []rawStadium) ([]apiStadium, error) {
	out := make([]apiStadium, 0, len(raw))
	for _, r := range raw {
		s, err := parseRawStadium(r)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// parseKickoff parses kickoff without stadium city (tests / fallback).
func parseKickoff(s string) time.Time {
	return parseKickoffAt(s, "")
}

// teamFlag returns the flag emoji for a team name, with alias fallbacks.
func teamFlag(name string) string {
	if f, ok := flagMap[name]; ok {
		return f
	}
	switch name {
	case "Turkey":
		return flagMap["Türkiye"]
	case "Ivory Coast":
		return flagMap["Côte d'Ivoire"]
	case "Democratic Republic of the Congo":
		return flagMap["DR Congo"]
	}
	return ""
}