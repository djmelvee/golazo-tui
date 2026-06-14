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
	db := openTestDB(t)
	m := app.New(db)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	out := render(m2)

	checks := []string{
		"FIFA WORLD CUP 2026",
		"LIVE MATCHES",
		"●",
		"FULL TIME",
		"GOLAZO TUI",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("live screen missing %q", want)
		}
	}
	t.Logf("Live screen output length: %d chars", len(out))
}

func TestStandingsScreen(t *testing.T) {
	db := openTestDB(t)
	m := app.New(db)
	// Use a tall window so all 12 groups fit without scrolling
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 120})
	m3, _ := m2.Update(tea.KeyPressMsg{Code: 'g'})

	out := render(m3)

	checks := []string{
		"FIFA WORLD CUP 2026",
		"GROUP A",
		"GROUP L",
		"Brazil",
		"Pts",
		"Top 2 advance",
	}
	for _, want := range checks {
		if !strings.Contains(out, want) {
			t.Errorf("standings screen missing %q", want)
		}
	}
	t.Logf("Standings screen output length: %d chars", len(out))
}

func TestFixturesScreen(t *testing.T) {
	db := openTestDB(t)
	m := app.New(db)
	m2, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m3, _ := m2.Update(tea.KeyPressMsg{Code: 'f'})

	out := render(m3)

	checks := []string{
		"FIFA WORLD CUP 2026",
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

func TestQuitKey(t *testing.T) {
	db := openTestDB(t)
	m := app.New(db)
	_, cmd := m.Update(tea.KeyPressMsg{Code: 'q'})
	if cmd == nil {
		t.Error("expected Quit cmd on 'q' press, got nil")
	}
}
