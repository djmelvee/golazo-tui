package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/djmelvee/golazo-tui/internal/ascii"
	"github.com/djmelvee/golazo-tui/internal/auth"
	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/fetcher"
	"github.com/djmelvee/golazo-tui/internal/notify"
	"github.com/djmelvee/golazo-tui/internal/screens"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/tz"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

const sidebarWidth = 20

// FetchMsg is returned by the background API poller after each poll attempt.
type FetchMsg struct {
	err      error
	newGoals []fetcher.NewGoal
}

func fetchNow(client *wc.Client, db *data.Store) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		goals, err := fetchWithAuth(ctx, client, db)
		return FetchMsg{err: err, newGoals: goals}
	}
}

func fetchAfter(client *wc.Client, db *data.Store) tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		goals, err := fetchWithAuth(ctx, client, db)
		return FetchMsg{err: err, newGoals: goals}
	})
}

func fetchWithAuth(ctx context.Context, client *wc.Client, db *data.Store) ([]fetcher.NewGoal, error) {
	goals, err := fetcher.Fetch(ctx, client, db)
	if err == nil || client == nil {
		return goals, err
	}
	if strings.Contains(err.Error(), "401") && db != nil {
		tok, regErr := auth.Register(ctx, client.BaseURL())
		if regErr == nil {
			client.SetToken(tok)
			_ = db.SetToken(tok)
			return fetcher.Fetch(ctx, client, db)
		}
	}
	return goals, err
}

type Model struct {
	w, h      int
	route     string
	live      screens.Live
	standings screens.Standings
	fixtures  screens.Fixtures
	changelog screens.Changelog
	bracket   screens.Bracket
	scorers   screens.Scorers
	digest       screens.Digest
	predictions      screens.Predictions
	predictionDetail screens.PredictionDetail
	splash           screens.Splash
	detail           screens.MatchDetail
	help             screens.Help
	teamHub          screens.TeamHub
	db        *data.Store
	client    *wc.Client

	celebrating       bool
	celebrationText   string
	celebrationUntil  time.Time
	celebrationFrame  int
	toastNote         string
	toastUntil        time.Time
	staleNote         string
}

// New creates the root Bubble Tea model.
func New(db *data.Store, client *wc.Client) Model {
	route := "live"
	if !screens.SkipSplash() {
		route = "splash"
	}
	applyTZPref(db)
	m := Model{
		route:  route,
		db:     db,
		client: client,
	}
	m.live.Load(db)
	m.standings.Load(db)
	m.fixtures.Load(db)
	m.bracket.Load(db)
	m.scorers.Load(db)
	m.digest.Load(db)
	m.predictions.Load(db)
	m.teamHub.Load(db)
	m.changelog.Load()
	return m
}

func applyTZPref(db *data.Store) {
	switch db.GetPrefString("timezone", "amsterdam") {
	case "local":
		tz.SetDisplayMode(tz.DisplayLocal)
	case "utc":
		tz.SetDisplayMode(tz.DisplayUTC)
	default:
		tz.SetDisplayMode(tz.DisplayAmsterdam)
	}
}

func (m Model) WithNote(note string) Model {
	m.live.SetFetchNote(note)
	return m
}

func (m Model) Init() tea.Cmd {
	if m.route == "splash" {
		return tea.Batch(screens.SplashTickCmd(), screens.SplashDoneCmd())
	}
	cmds := []tea.Cmd{screens.TickCmd(), screens.BlinkCmd()}
	if m.client != nil {
		cmds = append(cmds, fetchNow(m.client, m.db))
	}
	return tea.Batch(cmds...)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w = msg.Width
		m.h = msg.Height
		m.splash.SetSize(m.w, m.h)
		m.help.SetSize(m.w-sidebarWidth, m.bodyHeight())
		m.setContentSizes()

	case tea.MouseClickMsg:
		if m.route != "splash" && msg.X < sidebarWidth {
			if route := sidebarRouteAt(msg.Y); route != "" {
				m.navigate(route)
			}
		}

	case screens.SplashTickMsg:
		if m.route == "splash" {
			m.splash.Advance()
			return m, screens.SplashTickCmd()
		}

	case screens.SplashDoneMsg:
		if m.route == "splash" {
			m.route = "live"
			m.setContentSizes()
			cmds := []tea.Cmd{screens.TickCmd(), screens.BlinkCmd()}
			if m.client != nil {
				cmds = append(cmds, fetchNow(m.client, m.db))
			}
			return m, tea.Batch(cmds...)
		}

	case screens.CelebrationDoneMsg:
		if m.celebrating && !m.celebrationUntil.IsZero() && time.Now().After(m.celebrationUntil) {
			m.celebrating = false
			m.celebrationText = ""
			m.celebrationUntil = time.Time{}
		}

	case FetchMsg:
		if msg.err != nil {
			m.live.SetFetchNote("fetch error · retrying")
			m.staleNote = "API fetch error · showing cached data"
		} else {
			m.live.SetFetchNote("")
			m.staleNote = ""
			if !m.db.IsFresh("matches:live", 10*time.Minute) {
				m.staleNote = "cache may be stale · last update >10m ago"
			}
		}
		hadGoals := len(msg.newGoals) > 0
		m.handleNewGoals(msg.newGoals)
		m.reloadActiveScreens()
		if m.route == "detail" && m.detail.MatchID() > 0 {
			mid := m.detail.MatchID()
			if match := m.db.FindMatch(mid); match != nil {
				m.detail.Update(*match, m.db.GetEvents(mid))
			}
		}
		var cmds []tea.Cmd
		if m.client != nil {
			cmds = append(cmds, fetchAfter(m.client, m.db))
		}
		if hadGoals {
			cmds = append(cmds, screens.CelebrationDoneCmd())
		}
		return m, tea.Batch(cmds...)

	case screens.TickMsg:
		m.reloadActiveScreens()
		if m.celebrating {
			m.celebrationFrame++
		}
		if m.route == "detail" && m.detail.MatchID() > 0 {
			mid := m.detail.MatchID()
			if match := m.db.FindMatch(mid); match != nil {
				m.detail.Update(*match, m.db.GetEvents(mid))
			}
		}
		if !m.toastUntil.IsZero() && time.Now().After(m.toastUntil) {
			m.toastNote = ""
		}
		if m.celebrating && !m.celebrationUntil.IsZero() && time.Now().After(m.celebrationUntil) {
			m.celebrating = false
			m.celebrationText = ""
			m.celebrationUntil = time.Time{}
		}
		return m, screens.TickCmd()

	case screens.BlinkMsg:
		m.live.ToggleBlink()
		if m.route == "detail" {
			m.detail.ToggleBlink()
		}
		return m, screens.BlinkCmd()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "esc":
			switch m.route {
			case "detail":
				m.route = "live"
			case "prediction-detail":
				m.route = "predictions"
			case "help":
				m.route = "live"
			}
		case "b":
			if m.route != "splash" && m.route != "detail" && m.route != "prediction-detail" && m.route != "help" {
				m.navigate("bracket")
			}
		case "?":
			if m.route != "splash" {
				m.route = "help"
			}
		case "t":
			if m.route != "splash" {
				if m.route == "team" {
					m.teamHub.CycleFavorite(m.db)
				} else {
					m.route = "team"
					m.teamHub.Load(m.db)
				}
			}
		case "z":
			if m.route != "splash" {
				m.cycleTimezone()
			}
		case "1", "2", "3":
			if m.route == "fixtures" {
				m.fixtures.SetFilter(int(msg.String()[0] - '1'))
			}
		case "enter":
			switch m.route {
			case "live":
				if sel := m.live.SelectedMatch(); sel != nil {
					m.openDetail(*sel)
				}
			case "bracket":
				if id := m.bracket.SelectedMatchID(); id > 0 {
					if match := m.db.FindMatch(id); match != nil {
						m.openDetail(*match)
					}
				}
			case "predictions":
				if pr := m.predictions.SelectedPrediction(); pr != nil {
					m.openPredictionDetail(*pr)
				}
			case "fixtures":
				if sel := m.fixtures.SelectedMatch(); sel != nil {
					m.openDetail(*sel)
				}
			case "digest":
				if id := m.digest.SelectedMatchID(); id > 0 {
					if match := m.db.FindMatch(id); match != nil {
						m.openDetail(*match)
					}
				}
			case "standings":
				if team := m.standings.SelectedTeam(); team != nil {
					m.teamHub.SetTeam(*team, m.db)
					m.route = "team"
				}
			case "scorers":
				if team := m.scorers.SelectedTeam(); team != "" {
					m.teamHub.SetTeam(wc.Team{Name: team}, m.db)
					m.route = "team"
				}
			case "team":
				if id := m.teamHub.SelectedMatchID(); id > 0 {
					if match := m.db.FindMatch(id); match != nil {
						m.openDetail(*match)
					}
				}
			}
		case "h":
			m.navigate("live")
		case "g":
			if m.route != "detail" && m.route != "splash" {
				m.navigate("standings")
			}
		case "f":
			if m.route != "detail" && m.route != "splash" {
				m.navigate("fixtures")
			}
		case "c":
			if m.route != "detail" && m.route != "splash" {
				m.navigate("changelog")
			}
		case "s":
			if m.route != "detail" && m.route != "splash" {
				m.navigate("scorers")
			}
		case "d":
			if m.route != "detail" && m.route != "splash" {
				m.navigate("digest")
			}
		case "p":
			if m.route != "detail" && m.route != "splash" {
				m.navigate("predictions")
			}
		case "!":
			if m.route != "splash" {
				on := !m.db.GetPrefBool("bell_enabled", false)
				_ = m.db.SetPrefBool("bell_enabled", on)
				if on {
					m.toastNote = "goal alerts on (bell + desktop)"
				} else {
					m.toastNote = "goal alerts off"
				}
				m.toastUntil = time.Now().Add(3 * time.Second)
			}
		case "j", "down":
			m.cursorDown()
		case "k", "up":
			m.cursorUp()
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m *Model) navigate(route string) {
	m.route = route
	switch route {
	case "standings":
		m.standings.Load(m.db)
	case "fixtures":
		m.fixtures.Load(m.db)
	case "bracket":
		m.bracket.Load(m.db)
	case "scorers":
		m.scorers.Load(m.db)
	case "digest":
		m.digest.Load(m.db)
	case "predictions":
		m.predictions.Load(m.db)
	case "team":
		m.teamHub.Load(m.db)
	case "changelog":
		m.changelog.Load()
	}
}

func (m *Model) cycleTimezone() {
	order := []tz.DisplayMode{tz.DisplayAmsterdam, tz.DisplayLocal, tz.DisplayUTC}
	cur := tz.GetDisplayMode()
	next := order[0]
	for i, mode := range order {
		if mode == cur {
			next = order[(i+1)%len(order)]
			break
		}
	}
	tz.SetDisplayMode(next)
	_ = m.db.SetPrefString("timezone", string(next))
	m.toastNote = "timezone: " + tz.DisplayLabel()
	m.toastUntil = time.Now().Add(3 * time.Second)
	m.reloadActiveScreens()
}

func (m *Model) cursorDown() {
	switch m.route {
	case "live":
		m.live.CursorDown()
	case "detail":
		m.detail.ScrollDown()
	case "prediction-detail":
		m.predictionDetail.ScrollDown()
	case "standings":
		m.standings.CursorDown()
	case "changelog":
		m.changelog.ScrollDown()
	case "bracket":
		m.bracket.CursorDown()
	case "scorers":
		m.scorers.CursorDown()
	case "digest":
		m.digest.CursorDown()
	case "predictions":
		m.predictions.CursorDown()
	case "fixtures":
		m.fixtures.CursorDown()
	case "team":
		m.teamHub.CursorDown()
	}
}

func (m *Model) cursorUp() {
	switch m.route {
	case "live":
		m.live.CursorUp()
	case "detail":
		m.detail.ScrollUp()
	case "prediction-detail":
		m.predictionDetail.ScrollUp()
	case "standings":
		m.standings.CursorUp()
	case "changelog":
		m.changelog.ScrollUp()
	case "bracket":
		m.bracket.CursorUp()
	case "scorers":
		m.scorers.CursorUp()
	case "digest":
		m.digest.CursorUp()
	case "predictions":
		m.predictions.CursorUp()
	case "fixtures":
		m.fixtures.CursorUp()
	case "team":
		m.teamHub.CursorUp()
	}
}

func sidebarRouteAt(y int) string {
	// approximate line → route from sidebar layout
	routes := []string{"", "live", "predictions", "bracket", "digest", "scorers", "standings", "fixtures", "changelog"}
	line := y - 3
	if line < 0 || line >= len(routes) {
		return ""
	}
	return routes[line]
}

func (m *Model) setContentSizes() {
	contentW := m.w - sidebarWidth
	if contentW < 40 {
		contentW = 40
	}
	bodyH := m.bodyHeight()
	m.live.SetSize(contentW, bodyH)
	m.standings.SetSize(contentW, bodyH)
	m.fixtures.SetSize(contentW, bodyH)
	m.changelog.SetSize(contentW, bodyH)
	m.bracket.SetSize(contentW, bodyH)
	m.scorers.SetSize(contentW, bodyH)
	m.digest.SetSize(contentW, bodyH)
	m.predictions.SetSize(contentW, bodyH)
	m.predictionDetail.SetSize(contentW, bodyH)
	m.teamHub.SetSize(contentW, bodyH)
	m.help.SetSize(contentW, bodyH)
	m.detail.SetSize(contentW, bodyH-2)
}

func (m *Model) bodyHeight() int {
	h := m.h - ascii.HeaderLineCount(m.w) - 3
	if h < 10 {
		h = 10
	}
	return h
}

func (m *Model) openDetail(match wc.Match) {
	m.detail.Set(match, m.db.GetEvents(match.ID))
	m.detail.SetSize(m.w-sidebarWidth, m.bodyHeight())
	m.route = "detail"
}

func (m *Model) openPredictionDetail(pr wc.MatchPrediction) {
	m.predictionDetail.Set(pr)
	m.predictionDetail.SetSize(m.w-sidebarWidth, m.bodyHeight())
	m.route = "prediction-detail"
}

func (m *Model) handleNewGoals(goals []fetcher.NewGoal) {
	if len(goals) == 0 {
		return
	}
	if m.db.GetPrefBool("bell_enabled", false) {
		for range goals {
			notify.GoalBell()
		}
	}
	var parts []string
	for _, g := range goals {
		label := g.HomeFlag + " " + g.HomeTeam
		if g.ScoredBy == "away" {
			label = g.AwayFlag + " " + g.AwayTeam
		}
		if g.ScorerName != "" {
			label += " — " + g.ScorerName
		}
		parts = append(parts, fmt.Sprintf("%s  %d – %d", label, g.HomeScore, g.AwayScore))
		if m.db.GetPrefBool("bell_enabled", false) {
			notify.DesktopAlert("Golazo!", fmt.Sprintf("%s  %d – %d", label, g.HomeScore, g.AwayScore))
		}
	}
	m.celebrating = true
	if len(parts) == 1 {
		m.celebrationText = parts[0]
	} else {
		m.celebrationText = strings.Join(parts, "  ·  ")
	}
	m.celebrationUntil = time.Now().Add(screens.CelebrationDuration)
	m.celebrationFrame = 0
}

func (m *Model) reloadActiveScreens() {
	m.live.Load(m.db)
	switch m.route {
	case "standings":
		m.standings.Load(m.db)
	case "fixtures":
		m.fixtures.Load(m.db)
	case "bracket":
		m.bracket.Load(m.db)
	case "scorers":
		m.scorers.Load(m.db)
	case "digest":
		m.digest.Load(m.db)
	case "predictions":
		m.predictions.Load(m.db)
	case "team":
		m.teamHub.Load(m.db)
	}
}

func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.SetContent(m.render())
	return v
}

func (m Model) render() string {
	if m.route == "splash" {
		return m.splash.View()
	}

	header := styles.RenderHeader(m.w)
	if m.staleNote != "" {
		header += "\n" + styles.DimText.Render("  "+m.staleNote)
	}
	sidebarRoute := m.route
	switch sidebarRoute {
	case "detail":
		sidebarRoute = "live"
	case "prediction-detail":
		sidebarRoute = "predictions"
	case "help":
		sidebarRoute = "live"
	}
	sidebar := renderSidebar(sidebarRoute, m.bodyHeight(), m.live.LiveCount(), m.live.Blink())

	var content string
	switch m.route {
	case "live":
		content = m.live.View()
		if m.toastNote != "" {
			content = styles.GoldText.Render("  "+m.toastNote+"\n") + content
		}
	case "standings":
		content = m.standings.View()
	case "fixtures":
		content = m.fixtures.View()
	case "changelog":
		content = m.changelog.View()
	case "bracket":
		content = m.bracket.View()
	case "scorers":
		content = m.scorers.View()
	case "digest":
		content = m.digest.View()
	case "predictions":
		content = m.predictions.View()
	case "prediction-detail":
		content = m.predictionDetail.View()
	case "detail":
		content = m.detail.View()
	case "help":
		content = m.help.View()
	case "team":
		content = m.teamHub.View()
	}

	contentW := m.w - sidebarWidth
	if contentW < 40 {
		contentW = 40
	}
	contentHeight := m.bodyHeight()

	if m.route != "help" && m.route != "prediction-detail" && m.route != "detail" {
		contentLines := strings.Split(content, "\n")
		if len(contentLines) > contentHeight {
			contentLines = contentLines[:contentHeight]
		}
		for len(contentLines) < contentHeight {
			contentLines = append(contentLines, "")
		}
		content = strings.Join(contentLines, "\n")
	}

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebar, content)

	if m.celebrating && m.route != "splash" {
		body = m.overlayCelebration(body, contentW, contentHeight)
	}

	footer := renderFooter(m.route, m.w)
	return header + "\n" + body + "\n" + footer
}

func (m Model) overlayCelebration(body string, contentW, height int) string {
	lines := strings.Split(body, "\n")
	if len(lines) < height {
		for len(lines) < height {
			lines = append(lines, "")
		}
	}
	overlay := m.renderCelebration(contentW, height)
	sidebarLines := 0
	if len(lines) > 0 {
		// sidebar is first sidebarWidth chars - overlay only content area from col sidebarWidth
		for i := range overlay {
			if i >= len(lines) {
				break
			}
			if overlay[i] != "" {
				pad := strings.Repeat(" ", sidebarWidth)
				if len(lines[i]) > sidebarWidth {
					lines[i] = pad + overlay[i]
				} else {
					lines[i] = pad + overlay[i]
				}
			}
		}
	}
	_ = sidebarLines
	return strings.Join(lines, "\n")
}

func (m Model) renderCelebration(contentW, height int) []string {
	burst := strings.Split(ascii.GoalBurst(contentW), "\n")
	ball := ascii.SpinBall3D(m.celebrationFrame % 24)
	mid := height / 2
	out := make([]string, height)
	for i := range out {
		out[i] = ""
	}
	text := styles.GoldBold.Render("  " + m.celebrationText)
	if mid < height {
		out[mid] = text
	}
	for i, l := range burst {
		idx := mid - len(burst)/2 + i
		if idx >= 0 && idx < height && l != "" {
			out[idx] = styles.GoldText.Render(l)
		}
	}
	ballMid := mid + 3
	for i, l := range ball {
		if l == "" {
			continue
		}
		idx := ballMid - len(ball)/2 + i
		if idx >= 0 && idx < height {
			out[idx] = styles.DimText.Render(l)
		}
	}
	return out
}

func renderSidebar(route string, h int, liveCount int, blink bool) string {
	nav := func(key, label, r string) string {
		if r == route {
			prefix := styles.LiveBadge.Render("●")
			return styles.ActiveNav.Render(fmt.Sprintf(" %s %-10s[%s]", prefix, label, key))
		}
		if r == "live" && liveCount > 0 {
			dot := styles.DimText.Render("●")
			if blink {
				dot = styles.LiveBadge.Render("●")
			}
			return " " + dot + styles.InactiveNav.Render(fmt.Sprintf(" %-10s[%s]", label, key))
		}
		return styles.InactiveNav.Render(fmt.Sprintf("   %-10s[%s]", label, key))
	}

	rule := styles.DimText.Render(strings.Repeat("─", sidebarWidth-2))

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(styles.GoldBold.Render(" ⚽ GOLAZO") + "\n")
	sb.WriteString(" " + rule + "\n")
	sb.WriteString(nav("h", "LIVE", "live") + "\n")
	sb.WriteString(nav("p", "Predict", "predictions") + "\n")
	sb.WriteString(nav("b", "Bracket", "bracket") + "\n")
	sb.WriteString(nav("d", "Digest", "digest") + "\n")
	sb.WriteString(nav("s", "Scorers", "scorers") + "\n")
	sb.WriteString(nav("g", "Groups", "standings") + "\n")
	sb.WriteString(nav("f", "Fixtures", "fixtures") + "\n")
	sb.WriteString(nav("t", "Team", "team") + "\n")
	sb.WriteString(nav("c", "Changelog", "changelog") + "\n")
	sb.WriteString(" " + rule + "\n")
	sb.WriteString(styles.GoldText.Render(" 🏆 WC 2026") + "\n")
	sb.WriteString(styles.DimText.Render(" ! alerts · ? help") + "\n")
	sb.WriteString(" " + rule + "\n")
	sb.WriteString(styles.DimText.Render(" djMelvee") + "\n")

	return lipgloss.NewStyle().
		Width(sidebarWidth).
		Background(styles.SidebarBg).
		Foreground(styles.TextDim).
		Render(sb.String())
}

func renderFooter(route string, w int) string {
	var items []string
	switch route {
	case "detail", "prediction-detail":
		items = []string{
			styles.GoldText.Render("esc back"),
			styles.DimText.Render("j/k scroll"),
			styles.DimText.Render("q quit"),
		}
	case "help":
		items = []string{
			styles.GoldText.Render("?/esc close"),
			styles.DimText.Render("q quit"),
		}
	case "splash":
		items = []string{styles.DimText.Render("loading...")}
	default:
		items = []string{
			hintKey("h", "live", route == "live"),
			hintKey("p", "predict", route == "predictions"),
			hintKey("b", "bracket", route == "bracket"),
			hintKey("d", "digest", route == "digest"),
			hintKey("f", "fixtures", route == "fixtures"),
			hintKey("t", "team", route == "team"),
			styles.DimText.Render("enter detail"),
			styles.DimText.Render("! alerts"),
			styles.DimText.Render("? help"),
			styles.DimText.Render("q quit"),
		}
	}
	bar := "  " + strings.Join(items, styles.DimText.Render(" · "))
	return lipgloss.NewStyle().Width(w).Render(bar)
}

func hintKey(key, label string, active bool) string {
	if active {
		return styles.GoldText.Render(key + " " + label)
	}
	return styles.DimText.Render(key + " " + label)
}