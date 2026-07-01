// Package fetcher polls the WC2026 API and writes results to the SQLite cache.
// It is shared by both the golazo-fetcher binary and the built-in TUI background poller.
package fetcher

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/espn"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// NewGoal is emitted when a poll detects a new goal.
type NewGoal struct {
	MatchID    int
	HomeTeam   string
	AwayTeam   string
	HomeFlag   string
	AwayFlag   string
	ScorerName string
	Minute     int
	HomeScore  int
	AwayScore  int
	ScoredBy   string // "home" or "away"
}

// Fetch polls the API once: fetches all match buckets + standings,
// detects goal events from score changes, and writes everything to db.
// Returns newly detected goals for UI celebrations.
func Fetch(ctx context.Context, client *wc.Client, db *data.Store) ([]NewGoal, error) {
	live, err := client.FetchMatches(ctx, string(wc.StatusLive))
	if err != nil {
		espnErr := espn.PatchScores(ctx, db)
		newGoals := detectGoals(db.LiveMatches(), db)
		if espnErr == nil {
			return newGoals, nil
		}
		return newGoals, fmt.Errorf("fetch live: %w", err)
	}

	newGoals := detectGoals(live, db)

	upcoming, err := client.FetchMatches(ctx, string(wc.StatusUpcoming))
	if err != nil {
		return newGoals, fmt.Errorf("fetch upcoming: %w", err)
	}

	finished, err := client.FetchMatches(ctx, string(wc.StatusFinished))
	if err != nil {
		return newGoals, fmt.Errorf("fetch finished: %w", err)
	}

	standings, err := client.FetchStandings(ctx)
	if err != nil {
		return newGoals, fmt.Errorf("fetch standings: %w", err)
	}

	batch, err := db.BeginBatch()
	if err != nil {
		return newGoals, fmt.Errorf("begin batch: %w", err)
	}
	if err := batch.Set("matches:live", live); err != nil {
		_ = batch.Rollback()
		return newGoals, fmt.Errorf("set live: %w", err)
	}
	if err := batch.Set("matches:upcoming", upcoming); err != nil {
		_ = batch.Rollback()
		return newGoals, fmt.Errorf("set upcoming: %w", err)
	}
	if err := batch.Set("matches:finished", finished); err != nil {
		_ = batch.Rollback()
		return newGoals, fmt.Errorf("set finished: %w", err)
	}
	if err := batch.Set("standings", standings); err != nil {
		_ = batch.Rollback()
		return newGoals, fmt.Errorf("set standings: %w", err)
	}
	if err := batch.Commit(); err != nil {
		return newGoals, fmt.Errorf("commit batch: %w", err)
	}

	_ = espn.PatchScores(ctx, db)
	extra := detectGoals(db.LiveMatches(), db)
	newGoals = append(newGoals, extra...)

	return newGoals, nil
}

func detectGoals(matches []wc.Match, db *data.Store) []NewGoal {
	var fresh []NewGoal
	for _, m := range matches {
		if m.HomeScore == nil || m.AwayScore == nil {
			continue
		}
		_ = db.SetLastScore(m.ID, *m.HomeScore, *m.AwayScore)

		existing := db.GetEvents(m.ID)
		prevHome, prevAway := 0, 0
		if len(existing) > 0 {
			last := existing[len(existing)-1]
			prevHome, prevAway = last.HomeScore, last.AwayScore
		} else if h, a, ok := db.GetLastScore(m.ID); ok {
			// Seed from cached score on restart — avoids false goal celebrations.
			prevHome, prevAway = h, a
			if prevHome > *m.HomeScore {
				prevHome = *m.HomeScore
			}
			if prevAway > *m.AwayScore {
				prevAway = *m.AwayScore
			}
		}
		newHome, newAway := *m.HomeScore, *m.AwayScore
		minute := 0
		if m.Minute != nil {
			minute = *m.Minute
		}

		changed := false
		for prevHome < newHome {
			prevHome++
			name := scorerForGoal(m, "home", prevHome)
			existing = append(existing, wc.GoalEvent{
				MatchID: m.ID, HomeScore: prevHome, AwayScore: prevAway,
				Minute: minute, ScoredBy: "home", ScorerName: name, DetectedAt: time.Now(),
			})
			fresh = append(fresh, NewGoal{
				MatchID: m.ID, HomeTeam: m.HomeTeam.Name, AwayTeam: m.AwayTeam.Name,
				HomeFlag: m.HomeTeam.Flag, AwayFlag: m.AwayTeam.Flag,
				ScorerName: name, Minute: minute, HomeScore: prevHome, AwayScore: prevAway,
				ScoredBy: "home",
			})
			changed = true
		}
		for prevAway < newAway {
			prevAway++
			name := scorerForGoal(m, "away", prevAway)
			existing = append(existing, wc.GoalEvent{
				MatchID: m.ID, HomeScore: prevHome, AwayScore: prevAway,
				Minute: minute, ScoredBy: "away", ScorerName: name, DetectedAt: time.Now(),
			})
			fresh = append(fresh, NewGoal{
				MatchID: m.ID, HomeTeam: m.HomeTeam.Name, AwayTeam: m.AwayTeam.Name,
				HomeFlag: m.HomeTeam.Flag, AwayFlag: m.AwayTeam.Flag,
				ScorerName: name, Minute: minute, HomeScore: prevHome, AwayScore: prevAway,
				ScoredBy: "away",
			})
			changed = true
		}
		if changed {
			if err := db.SetEvents(m.ID, existing); err != nil {
				log.Printf("set events %d: %v", m.ID, err)
			}
		}
	}
	return fresh
}

func scorerForGoal(m wc.Match, side string, goalNum int) string {
	scorers := m.HomeScorers
	if side == "away" {
		scorers = m.AwayScorers
	}
	if goalNum > 0 && goalNum <= len(scorers) && scorers[goalNum-1].Name != "" {
		return scorers[goalNum-1].Name
	}
	return wc.ScorerNameForGoal(m, side, goalNum)
}