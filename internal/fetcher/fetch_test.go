package fetcher

import (
	"path/filepath"
	"testing"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

func TestDetectGoalsSeedsFromLastScore(t *testing.T) {
	dir := t.TempDir()
	db, err := data.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	hs, as := 2, 1
	m := wc.Match{
		ID: 99, Status: wc.StatusLive,
		HomeTeam: wc.Team{Name: "A"}, AwayTeam: wc.Team{Name: "B"},
		HomeScore: &hs, AwayScore: &as,
	}
	_ = db.SetLastScore(99, 2, 1)

	goals := detectGoals([]wc.Match{m}, db)
	if len(goals) != 0 {
		t.Fatalf("expected no false goals on restart, got %d", len(goals))
	}
}

func TestScorerForGoalPrefersAPI(t *testing.T) {
	m := wc.Match{
		HomeScorers: []wc.Scorer{{Name: "API Scorer", Minute: 10}},
	}
	if got := scorerForGoal(m, "home", 1); got != "API Scorer" {
		t.Fatalf("got %q", got)
	}
}

