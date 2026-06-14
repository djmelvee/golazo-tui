// Package espn provides a fallback score source using ESPN's public scoreboard API.
// No authentication or API key required.
package espn

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// PatchScores fetches WC2026 results from ESPN and patches any nil-score or
// misclassified records in the local DB. Moves finished matches from upcoming
// to finished with real scores, updates live scores in place.
// Called when the primary worldcup26.ir API is unavailable.
func PatchScores(ctx context.Context, db *data.Store) error {
	now := time.Now().UTC()
	today := now.Format("20060102")
	yesterday := now.Add(-24 * time.Hour).Format("20060102")

	var all []espnEvent
	for _, date := range []string{today, yesterday} {
		evs, err := fetchDay(ctx, date)
		if err != nil {
			continue
		}
		all = append(all, evs...)
	}
	if len(all) == 0 {
		return fmt.Errorf("espn: no events returned")
	}
	return applyToCache(db, all)
}

// ── ESPN API structs ────────────────────────────────────────────────────────

type espnResponse struct {
	Events []espnEvent `json:"events"`
}

type espnEvent struct {
	Date         string       `json:"date"`
	Competitions []espnComp   `json:"competitions"`
}

type espnComp struct {
	Status      espnCompStatus `json:"status"`
	Venue       espnVenue      `json:"venue"`
	Competitors []espnTeamComp `json:"competitors"`
	Notes       []espnNote     `json:"notes"`
	Groups      espnGroup      `json:"groups"`
}

type espnCompStatus struct {
	DisplayClock string         `json:"displayClock"`
	Period       int            `json:"period"`
	Type         espnStatusType `json:"type"`
}

type espnStatusType struct {
	State     string `json:"state"` // "pre", "in", "post"
	Completed bool   `json:"completed"`
}

type espnVenue struct {
	FullName string   `json:"fullName"`
	Address  espnAddr `json:"address"`
}

type espnAddr struct {
	City string `json:"city"`
}

type espnTeamComp struct {
	HomeAway string   `json:"homeAway"`
	Score    string   `json:"score"`
	Team     espnTeam `json:"team"`
}

type espnTeam struct {
	DisplayName string `json:"displayName"`
}

type espnNote struct {
	Headline string `json:"headline"`
}

type espnGroup struct {
	Name string `json:"name"`
}

// ── ESPN fetch ──────────────────────────────────────────────────────────────

var scoreboardURLs = []string{
	"https://site.api.espn.com/apis/site/v2/sports/soccer/fifa.world/scoreboard",
	"https://site.api.espn.com/apis/site/v2/sports/soccer/fifa.worldcup/scoreboard",
}

func fetchDay(ctx context.Context, date string) ([]espnEvent, error) {
	hc := &http.Client{Timeout: 8 * time.Second}
	var lastErr error
	for _, base := range scoreboardURLs {
		url := base + "?dates=" + date
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			lastErr = err
			continue
		}
		req.Header.Set("User-Agent", "golazo-tui/1.0")
		resp, err := hc.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			lastErr = fmt.Errorf("espn: HTTP %d for %s", resp.StatusCode, url)
			continue
		}
		var r espnResponse
		if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
			lastErr = err
			continue
		}
		if len(r.Events) > 0 {
			return r.Events, nil
		}
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("espn: no events for %s", date)
}

// ── DB patching ─────────────────────────────────────────────────────────────

func applyToCache(db *data.Store, events []espnEvent) error {
	upcoming := db.UpcomingMatches()
	finished := db.FinishedMatches()
	live := db.LiveMatches()

	upcomingChanged := false
	finishedChanged := false
	liveChanged := false

	// Map ESPN events to simple score records, keyed by normalised team pair.
	type scoreRec struct {
		homeTeam  string
		awayTeam  string
		homeScore *int
		awayScore *int
		minute    *int
		status    wc.MatchStatus
	}

	var scores []scoreRec
	for _, ev := range events {
		if len(ev.Competitions) == 0 {
			continue
		}
		comp := ev.Competitions[0]
		var home, away *espnTeamComp
		for i := range comp.Competitors {
			c := &comp.Competitors[i]
			if c.HomeAway == "home" {
				home = c
			} else {
				away = c
			}
		}
		if home == nil || away == nil {
			continue
		}

		status := wc.StatusUpcoming
		switch comp.Status.Type.State {
		case "in":
			status = wc.StatusLive
		case "post":
			status = wc.StatusFinished
		}

		var hs, as *int
		if h, err := strconv.Atoi(home.Score); err == nil {
			hs = &h
		}
		if a, err := strconv.Atoi(away.Score); err == nil {
			as = &a
		}

		var minute *int
		if status == wc.StatusLive {
			parts := strings.SplitN(comp.Status.DisplayClock, ":", 2)
			if m, err := strconv.Atoi(parts[0]); err == nil {
				minute = &m
			}
		}

		scores = append(scores, scoreRec{
			homeTeam:  home.Team.DisplayName,
			awayTeam:  away.Team.DisplayName,
			homeScore: hs,
			awayScore: as,
			minute:    minute,
			status:    status,
		})
	}

	// For each ESPN score record, find a matching DB match and patch it.
	removeFromUpcoming := make(map[int]bool)

	for _, sc := range scores {
		switch sc.status {
		case wc.StatusFinished:
			if sc.homeScore == nil || sc.awayScore == nil {
				continue
			}
			// 1. Move from upcoming → finished if found there.
			for i, m := range upcoming {
				if !removeFromUpcoming[i] && teamsMatch(m.HomeTeam.Name, m.AwayTeam.Name, sc.homeTeam, sc.awayTeam) {
					m.Status = wc.StatusFinished
					m.HomeScore = sc.homeScore
					m.AwayScore = sc.awayScore
					finished = append(finished, m)
					removeFromUpcoming[i] = true
					finishedChanged = true
					upcomingChanged = true
				}
			}
			// 2. Patch nil scores in existing finished records.
			for i := range finished {
				if finished[i].HomeScore == nil && teamsMatch(finished[i].HomeTeam.Name, finished[i].AwayTeam.Name, sc.homeTeam, sc.awayTeam) {
					finished[i].HomeScore = sc.homeScore
					finished[i].AwayScore = sc.awayScore
					finishedChanged = true
				}
			}

		case wc.StatusLive:
			if sc.homeScore == nil {
				continue
			}
			// Move upcoming → live.
			for i, m := range upcoming {
				if !removeFromUpcoming[i] && teamsMatch(m.HomeTeam.Name, m.AwayTeam.Name, sc.homeTeam, sc.awayTeam) {
					m.Status = wc.StatusLive
					m.HomeScore = sc.homeScore
					m.AwayScore = sc.awayScore
					m.Minute = sc.minute
					live = append(live, m)
					removeFromUpcoming[i] = true
					liveChanged = true
					upcomingChanged = true
				}
			}
			// Patch nil scores in existing live records.
			for i := range live {
				if live[i].HomeScore == nil && teamsMatch(live[i].HomeTeam.Name, live[i].AwayTeam.Name, sc.homeTeam, sc.awayTeam) {
					live[i].HomeScore = sc.homeScore
					live[i].AwayScore = sc.awayScore
					live[i].Minute = sc.minute
					liveChanged = true
				}
			}
		}
	}

	if upcomingChanged {
		var kept []wc.Match
		for i, m := range upcoming {
			if !removeFromUpcoming[i] {
				kept = append(kept, m)
			}
		}
		if err := db.Set("matches:upcoming", kept); err != nil {
			return err
		}
	}
	if finishedChanged {
		if err := db.Set("matches:finished", finished); err != nil {
			return err
		}
	}
	if liveChanged {
		if err := db.Set("matches:live", live); err != nil {
			return err
		}
	}
	return nil
}

// teamsMatch returns true when two team-name pairs refer to the same match
// (either same order or swapped home/away), using normalised names.
func teamsMatch(dbHome, dbAway, espnHome, espnAway string) bool {
	dH, dA := norm(dbHome), norm(dbAway)
	eH, eA := norm(espnHome), norm(espnAway)
	return (dH == eH && dA == eA) || (dH == eA && dA == eH)
}

// norm lowercases and strips accents + known name aliases for comparison.
func norm(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	// Strip common accented characters so "Curaçao" == "Curacao"
	replacer := strings.NewReplacer(
		"ç", "c", "é", "e", "ô", "o", "ü", "u",
		"á", "a", "í", "i", "ó", "o", "ú", "u",
		"ñ", "n",
	)
	s = replacer.Replace(s)
	// Known aliases
	switch s {
	case "usa", "u.s.", "us", "united states of america":
		return "united states"
	case "turkey":
		return "turkiye"
	case "ivory coast":
		return "cote d'ivoire"
	case "bosnia-herzegovina", "bosnia & herzegovina":
		return "bosnia and herzegovina"
	case "dr congo", "congo dr", "democratic republic of the congo", "democratic republic of congo", "drc":
		return "dr congo"
	}
	return s
}
