package screens

import (
	"fmt"
	"sort"
	"strings"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/tz"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

// TeamHub shows context for the favorite team across the tournament.
type TeamHub struct {
	w, h       int
	team       wc.Team
	scroll     int
	lines      []string
	cursor     int
	matchIDs   []int
	allMatches []wc.Match
}

func (t *TeamHub) SetSize(w, h int) {
	t.w = w
	t.h = h
	t.rebuild()
}

func (t *TeamHub) Load(db *data.Store) {
	t.allMatches = db.AllMatches()
	name := db.GetPrefString("favorite_team", "")
	t.team = wc.Team{Name: name}
	t.matchIDs = nil
	t.cursor = 0
	t.scroll = 0

	if name == "" {
		t.lines = []string{
			styles.Heading.Render("  TEAM HUB"),
			"",
			styles.DimText.Render("  No favorite team yet."),
			styles.DimText.Render("  Press t to cycle through all teams."),
		}
		return
	}
	for _, m := range t.allMatches {
		if m.HomeTeam.Name == name {
			t.team = m.HomeTeam
			break
		}
		if m.AwayTeam.Name == name {
			t.team = m.AwayTeam
			break
		}
	}
	t.rebuild()
}

func (t *TeamHub) CycleFavorite(db *data.Store) {
	names := uniqueTeamNames(t.allMatches)
	if len(names) == 0 {
		t.allMatches = db.AllMatches()
		names = uniqueTeamNames(t.allMatches)
	}
	if len(names) == 0 {
		return
	}
	cur := db.GetPrefString("favorite_team", "")
	idx := 0
	for i, n := range names {
		if n == cur {
			idx = (i + 1) % len(names)
			break
		}
	}
	_ = db.SetPrefString("favorite_team", names[idx])
	t.Load(db)
}

func uniqueTeamNames(matches []wc.Match) []string {
	seen := make(map[string]struct{})
	var names []string
	for _, m := range matches {
		for _, team := range []wc.Team{m.HomeTeam, m.AwayTeam} {
			if team.Name == "" {
				continue
			}
			if _, ok := seen[team.Name]; !ok {
				seen[team.Name] = struct{}{}
				names = append(names, team.Name)
			}
		}
	}
	sort.Strings(names)
	return names
}

func (t *TeamHub) CursorDown() {
	if t.cursor < len(t.matchIDs)-1 {
		t.cursor++
	}
}

func (t *TeamHub) CursorUp() {
	if t.cursor > 0 {
		t.cursor--
	}
}

func (t *TeamHub) SelectedMatchID() int {
	if t.cursor >= 0 && t.cursor < len(t.matchIDs) {
		return t.matchIDs[t.cursor]
	}
	return 0
}

func (t *TeamHub) SetTeam(team wc.Team, db *data.Store) {
	_ = db.SetPrefString("favorite_team", team.Name)
	t.Load(db)
}

func (t *TeamHub) View() string {
	vis := t.h - 8
	if vis < 4 {
		vis = 4
	}
	start := t.scroll
	end := start + vis
	if end > len(t.lines) {
		end = len(t.lines)
	}
	var sb strings.Builder
	if start < len(t.lines) {
		sb.WriteString(strings.Join(t.lines[start:end], "\n"))
	}
	sb.WriteString("\n")
	sb.WriteString(styles.DimText.Render("  t cycle · enter match · j/k pick") + "\n")
	return sb.String()
}

func (t *TeamHub) rebuild() {
	if t.team.Name == "" && len(t.lines) > 0 {
		return
	}
	var lines []string
	lines = append(lines, styles.Heading.Render(fmt.Sprintf("  %s %s", t.team.Flag, t.team.Name)))
	if t.team.Group != "" {
		lines = append(lines, styles.GoldText.Render("  Group "+t.team.Group))
	}

	forms := wc.BuildTeamForms(t.allMatches)
	if f := lookupTeamForm(forms, t.team); f != nil && f.Played > 0 {
		lines = append(lines, styles.DimText.Render(fmt.Sprintf(
			"  WC form: %dW-%dD-%dL · %.1f gpg · %.1f conceded",
			f.W, f.D, f.L, f.GoalsPerGame, f.ConcededPerGame)))
	}
	lines = append(lines, "")

	var teamMatches []wc.Match
	for _, m := range t.allMatches {
		if m.HomeTeam.Name == t.team.Name || m.AwayTeam.Name == t.team.Name {
			teamMatches = append(teamMatches, m)
		}
	}
	sort.Slice(teamMatches, func(i, j int) bool {
		return teamMatches[i].KickoffAt.Before(teamMatches[j].KickoffAt)
	})

	t.matchIDs = nil
	lines = append(lines, styles.GoldText.Render("  MATCHES"))
	for i, m := range teamMatches {
		t.matchIDs = append(t.matchIDs, m.ID)
		score := "– –"
		if m.HomeScore != nil && m.AwayScore != nil {
			score = fmt.Sprintf("%d – %d", *m.HomeScore, *m.AwayScore)
		}
		prefix := "  "
		if i == t.cursor {
			prefix = styles.GoldText.Render("> ")
		}
		status := string(m.Status)
		if status == "" {
			status = "NS"
		}
		lines = append(lines, fmt.Sprintf("%s%s %s vs %s %s  %s  %s",
			prefix, tz.FormatKickoffShort(m.KickoffAt), m.HomeTeam.Flag, m.AwayTeam.Flag,
			styles.GoldBold.Render(score), status, wc.StageLabel(m.Stage)))
	}
	if len(teamMatches) == 0 {
		lines = append(lines, styles.DimText.Render("  — no fixtures —"))
	}
	t.lines = lines
}

func lookupTeamForm(forms map[int]*wc.TeamForm, team wc.Team) *wc.TeamForm {
	if f, ok := forms[team.ID]; ok {
		return f
	}
	for _, f := range forms {
		if f.Team.Name == team.Name {
			return f
		}
	}
	return nil
}

// FavoriteName returns the stored favorite team name, if any.
func FavoriteName(db *data.Store) string {
	return db.GetPrefString("favorite_team", "")
}

// HighlightFavorite wraps a team name in gold when it matches the favorite.
func HighlightFavorite(name, favorite string) string {
	if favorite != "" && name == favorite {
		return styles.GoldBold.Render(name)
	}
	return name
}