package screens

import (
	"testing"
	"time"

	"github.com/djmelvee/golazo-tui/internal/tz"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

func TestClassifyDigestLiveAnyDay(t *testing.T) {
	now := time.Date(2026, 7, 2, 1, 0, 0, 0, tz.Amsterdam)
	kick := time.Date(2026, 7, 1, 22, 0, 0, 0, tz.Amsterdam).UTC()
	hs, as, min := 1, 0, 67

	day := ClassifyDigestDay(now, []wc.Match{{
		ID: 1, Status: wc.StatusLive, KickoffAt: kick,
		HomeTeam: wc.Team{Name: "Belgium"}, AwayTeam: wc.Team{Name: "Senegal"},
		HomeScore: &hs, AwayScore: &as, Minute: &min,
	}}, nil, nil)

	if len(day.Live) != 1 {
		t.Fatalf("expected live match, got %d", len(day.Live))
	}
}

func TestClassifyDigestFinishedCarryOver(t *testing.T) {
	now := time.Date(2026, 7, 2, 1, 0, 0, 0, tz.Amsterdam)
	kick := time.Date(2026, 7, 1, 22, 0, 0, 0, tz.Amsterdam).UTC()
	hs, as := 2, 1

	day := ClassifyDigestDay(now, nil, []wc.Match{{
		ID: 2, Status: wc.StatusFinished, KickoffAt: kick,
		HomeTeam: wc.Team{Name: "England"}, AwayTeam: wc.Team{Name: "DR Congo"},
		HomeScore: &hs, AwayScore: &as,
	}}, nil)

	if len(day.Finished) != 1 {
		t.Fatalf("expected carried-over FT, got %d", len(day.Finished))
	}
}

func TestClassifyDigestUpcomingToday(t *testing.T) {
	now := time.Date(2026, 7, 2, 10, 0, 0, 0, tz.Amsterdam)
	kick := time.Date(2026, 7, 2, 21, 0, 0, 0, tz.Amsterdam).UTC()

	day := ClassifyDigestDay(now, nil, nil, []wc.Match{{
		ID: 3, Status: wc.StatusUpcoming, KickoffAt: kick,
		HomeTeam: wc.Team{Name: "Spain"}, AwayTeam: wc.Team{Name: "Austria"},
	}})

	if len(day.Upcoming) != 1 {
		t.Fatalf("expected upcoming today, got %d", len(day.Upcoming))
	}
}

func TestClassifyDigestFinishedOnToday(t *testing.T) {
	now := time.Date(2026, 7, 1, 23, 0, 0, 0, tz.Amsterdam)
	kick := time.Date(2026, 7, 1, 18, 0, 0, 0, tz.Amsterdam).UTC()
	hs, as := 1, 0

	day := ClassifyDigestDay(now, nil, []wc.Match{{
		ID: 4, Status: wc.StatusFinished, KickoffAt: kick,
		HomeTeam: wc.Team{Name: "England"}, AwayTeam: wc.Team{Name: "DR Congo"},
		HomeScore: &hs, AwayScore: &as,
	}}, nil)

	if len(day.Finished) != 1 {
		t.Fatalf("expected FT today, got %d", len(day.Finished))
	}
}