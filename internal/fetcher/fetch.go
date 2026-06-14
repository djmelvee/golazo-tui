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

// Fetch polls the API once: fetches all match buckets + standings,
// detects goal events from score changes, and writes everything to db.
// On failure, ESPN's public scoreboard API is used as a fallback to patch
// any nil scores in the cache without requiring authentication.
func Fetch(ctx context.Context, client *wc.Client, db *data.Store) error {
	live, err := client.FetchMatches(ctx, string(wc.StatusLive))
	if err != nil {
		// Primary API failed — patch DB with ESPN scores, then detect goals
		// from whatever ESPN wrote to matches:live (including any just-moved
		// upcoming matches that are now live with real scores).
		espnErr := espn.PatchScores(ctx, db)
		detectGoals(db.LiveMatches(), db)
		if espnErr == nil {
			// ESPN covered the gap — treat as a successful partial refresh so
			// the header stays clean rather than showing "fetch error".
			return nil
		}
		return fmt.Errorf("fetch live: %w", err)
	}

	// Detect score changes on primary-API live data.
	detectGoals(live, db)

	if err := db.Set("matches:live", live); err != nil {
		return fmt.Errorf("set live: %w", err)
	}

	upcoming, err := client.FetchMatches(ctx, string(wc.StatusUpcoming))
	if err != nil {
		return fmt.Errorf("fetch upcoming: %w", err)
	}
	if err := db.Set("matches:upcoming", upcoming); err != nil {
		return fmt.Errorf("set upcoming: %w", err)
	}

	finished, err := client.FetchMatches(ctx, string(wc.StatusFinished))
	if err != nil {
		return fmt.Errorf("fetch finished: %w", err)
	}
	if err := db.Set("matches:finished", finished); err != nil {
		return fmt.Errorf("set finished: %w", err)
	}

	standings, err := client.FetchStandings(ctx)
	if err != nil {
		return fmt.Errorf("fetch standings: %w", err)
	}
	if err := db.Set("standings", standings); err != nil {
		return fmt.Errorf("set standings: %w", err)
	}

	// Run ESPN patch after a successful primary fetch to catch any scores the
	// worldcup26.ir API is slow to report. Then re-run goal detection so any
	// score changes ESPN applied also produce goal events.
	_ = espn.PatchScores(ctx, db)
	detectGoals(db.LiveMatches(), db)

	return nil
}

// detectGoals compares each live match's current score against the last stored
// goal events and appends a new GoalEvent for every unanswered score increment.
// It is idempotent: re-running with the same scores produces no new events.
func detectGoals(matches []wc.Match, db *data.Store) {
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
		}
		newHome, newAway := *m.HomeScore, *m.AwayScore
		minute := 0
		if m.Minute != nil {
			minute = *m.Minute
		}

		changed := false
		for prevHome < newHome {
			prevHome++
			existing = append(existing, wc.GoalEvent{
				MatchID: m.ID, HomeScore: prevHome, AwayScore: prevAway,
				Minute: minute, ScoredBy: "home", DetectedAt: time.Now(),
			})
			changed = true
		}
		for prevAway < newAway {
			prevAway++
			existing = append(existing, wc.GoalEvent{
				MatchID: m.ID, HomeScore: prevHome, AwayScore: prevAway,
				Minute: minute, ScoredBy: "away", DetectedAt: time.Now(),
			})
			changed = true
		}
		if changed {
			if err := db.SetEvents(m.ID, existing); err != nil {
				log.Printf("set events %d: %v", m.ID, err)
			}
		}
	}
}
