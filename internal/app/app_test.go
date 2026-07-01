package app_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/djmelvee/golazo-tui/internal/app"
	"github.com/djmelvee/golazo-tui/internal/data"
)

func openTestDB(t *testing.T) *data.Store {
	t.Helper()
	home, _ := os.UserHomeDir()
	dbPath := filepath.Join(home, ".cache", "golazo-tui", "cache.db")
	db, err := data.OpenRO(dbPath)
	if err != nil {
		t.Skipf("cache DB not found at %s (run golazo-seed first): %v", dbPath, err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}

func render(m tea.Model) string {
	return m.View().Content
}

func TestLiveDashboard(t *testing.T) {
	t.Setenv("GOLAZO_NO_SPLASH", "1")
	db := openTestDB(t)
	m := app.New(db, nil)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	out := render(m2)

	checks := []string{
		"WORLD CUP",
		"2026",
		"●",
		"FULL TIME",
		"GOLAZO",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("live screen missing %q", want)
		}
	}
	t.Logf("Live screen output length: %d chars", len(out))
}

func TestStandingsScreen(t *testing.T) {
	t.Setenv("GOLAZO_NO_SPLASH", "1")
	db := openTestDB(t)
	m := app.New(db, nil)
	// Use a tall window so all 12 groups fit without scrolling
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 150})
	m3, _ := m2.Update(tea.KeyPressMsg{Code: 'g'})

	out := render(m3)

	checks := []string{
		"WORLD CUP",
		"GROUP A",
		"GROUP L",
		"Brazil",
		"Pts",
		"Top 2 advanc",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("standings screen missing %q", want)
		}
	}
	t.Logf("Standings screen output length: %d chars", len(out))
}

func TestFixturesScreen(t *testing.T) {
	t.Setenv("GOLAZO_NO_SPLASH", "1")
	db := openTestDB(t)
	m := app.New(db, nil)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m3, _ := m2.Update(tea.KeyPressMsg{Code: 'f'})

	out := render(m3)

	checks := []string{
		"WORLD CUP",
		"UPCOMING FIXTURES",
		"MATCHDAY",
		"vs",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("fixtures screen missing %q", want)
		}
	}
	t.Logf("Fixtures screen output length: %d chars", len(out))
}

func TestPredictionsScreen(t *testing.T) {
	t.Setenv("GOLAZO_NO_SPLASH", "1")
	db := openTestDB(t)
	m := app.New(db, nil)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 80})
	m3, _ := m2.Update(tea.KeyPressMsg{Code: 'p'})

	out := render(m3)

	checks := []string{
		"MATCH PREDICTIONS",
		"HT ",
		"draw ",
		"breakdown",
		"Predict",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("predictions screen missing %q", want)
		}
	}
}

func TestPredictionDetailScreen(t *testing.T) {
	t.Setenv("GOLAZO_NO_SPLASH", "1")
	db := openTestDB(t)
	m := app.New(db, nil)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 80})
	m3, _ := m2.Update(tea.KeyPressMsg{Code: 'p'})
	m4, _ := m3.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

	out := render(m4)
	checks := []string{
		"PREDICTION BREAKDOWN",
		"SUMMARY",
		"WHY THIS PREDICTION",
		"Expected goals",
		"Other likely scores",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("prediction detail missing %q", want)
		}
	}
}

func TestQuitKey(t *testing.T) {
	t.Setenv("GOLAZO_NO_SPLASH", "1")
	db := openTestDB(t)
	m := app.New(db, nil)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q'})
	if cmd == nil {
		t.Error("expected Quit cmd on 'q' press, got nil")
	}
}
