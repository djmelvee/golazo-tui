package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/fetcher"
	"github.com/djmelvee/golazo-tui/internal/screens"
	"github.com/djmelvee/golazo-tui/internal/styles"
	"github.com/djmelvee/golazo-tui/internal/wc"
)

const sidebarWidth = 18

// FetchMsg is returned by the background API poller after each poll attempt.
type FetchMsg struct{ err error }

// fetchNow runs a single API poll immediately (no delay).
func fetchNow(client *wc.Client, db *data.Store) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return FetchMsg{err: fetcher.Fetch(ctx, client, db)}
	}
}

// fetchAfter schedules an API poll after a 5-second delay.
func fetchAfter(client *wc.Client, db *data.Store) tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return FetchMsg{err: fetcher.Fetch(ctx, client, db)}
	})
}

type Model struct {
	w, h      int
	route     string // "live" | "standings" | "fixtures" | "changelog" | "detail"
	live      screens.Live
	standings screens.Standings
	fixtures  screens.Fixtures
	changelog screens.Changelog
	detail    screens.MatchDetail
	db        *data.Store
	client    *wc.Client // nil when GOLAZO_API_TOKEN is not set
}

// New creates the root Bubble Tea model.
func New(db *data.Store, client *wc.Client) Model {
	m := Model{
		route:  "live",
		db:     db,
		client: client,
	}
	m.live.Load(db)
	m.standings.Load(db)
	m.fixtures.Load(db)
	m.changelog.Load()
	return m
}

// WithNote sets an initial status note in the live header (e.g. a startup error).
// It is replaced on the first successful or failed fetch.
func (m Model) WithNote(note string) Model {
	m.live.SetFetchNote(note)
	return m
}

func (m Model) Init() tea.Cmd {
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
		contentW := m.w - sidebarWidth
		if contentW < 40 {
			contentW = 40
		}
		m.live.SetSize(contentW, m.h-6)
		m.standings.SetSize(contentW, m.h-6)
		m.fixtures.SetSize(contentW, m.h-6)
		m.changelog.SetSize(contentW, m.h-6)
		m.detail.SetSize(contentW, m.h-4)

	case FetchMsg:
		if msg.err != nil {
			m.live.SetFetchNote("fetch error · retrying")
		} else {
			m.live.SetFetchNote("")
		}
		m.live.Load(m.db)
		if m.route == "detail" && m.detail.MatchID() > 0 {
			mid := m.detail.MatchID()
			if match := m.live.FindMatch(mid); match != nil {
				m.detail.Update(*match, m.db.GetEvents(mid))
			}
		}
		return m, fetchAfter(m.client, m.db)

	case screens.TickMsg:
		m.live.Load(m.db)
		if m.route == "detail" && m.detail.MatchID() > 0 {
			mid := m.detail.MatchID()
			if match := m.live.FindMatch(mid); match != nil {
				m.detail.Update(*match, m.db.GetEvents(mid))
			}
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
		case "b", "esc":
			if m.route == "detail" {
				m.route = "live"
			}
		case "enter":
			if m.route == "live" {
				if sel := m.live.SelectedMatch(); sel != nil {
					m.detail.Set(*sel, m.db.GetEvents(sel.ID))
					m.detail.SetSize(m.w-sidebarWidth, m.h-4)
					m.route = "detail"
				}
			}
		case "h":
			m.route = "live"
		case "g":
			if m.route != "detail" {
				m.route = "standings"
				m.standings.Load(m.db)
			}
		case "f":
			if m.route != "detail" {
				m.route = "fixtures"
				m.fixtures.Load(m.db)
			}
		case "c":
			if m.route != "detail" {
				m.route = "changelog"
				m.changelog.Load()
			}
		case "j", "down":
			switch m.route {
			case "live":
				m.live.CursorDown()
			case "detail":
				m.detail.ScrollDown()
			case "standings":
				m.standings.ScrollDown()
			case "changelog":
				m.changelog.ScrollDown()
			}
		case "k", "up":
			switch m.route {
			case "live":
				m.live.CursorUp()
			case "detail":
				m.detail.ScrollUp()
			case "standings":
				m.standings.ScrollUp()
			case "changelog":
				m.changelog.ScrollUp()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m Model) View() tea.View {
	var v tea.View
	v.AltScreen = true
	v.SetContent(m.render())
	return v
}

func (m Model) render() string {
	header := styles.HeaderBar(m.w)
	// detail is a sub-view of live — highlight LIVE in sidebar
	sidebarRoute := m.route
	if sidebarRoute == "detail" {
		sidebarRoute = "live"
	}
	sidebar := renderSidebar(sidebarRoute, m.h, m.live.LiveCount(), m.live.Blink())

	var content string
	switch m.route {
	case "live":
		content = m.live.View()
	case "standings":
		content = m.standings.View()
	case "fixtures":
		content = m.fixtures.View()
	case "changelog":
		content = m.changelog.View()
	case "detail":
		content = m.detail.View()
	}

	contentW := m.w - sidebarWidth
	if contentW < 40 {
		contentW = 40
	}

	contentHeight := m.h - 4 // header(1) + blank(1) + footer(1) + blank(1)
	if contentHeight < 1 {
		contentHeight = 20
	}
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
	}
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, "")
	}

	sidebarLines := strings.Split(sidebar, "\n")
	for len(sidebarLines) < contentHeight {
		sidebarLines = append(sidebarLines, strings.Repeat(" ", sidebarWidth))
	}

	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.Join(sidebarLines, "\n"),
		strings.Join(contentLines, "\n"),
	)

	footer := renderFooter(m.route, m.w)
	return header + "\n" + body + "\n" + footer
}

func renderSidebar(route string, h int, liveCount int, blink bool) string {
	nav := func(key, label, r string) string {
		if r == route {
			prefix := styles.LiveBadge.Render("●")
			return styles.ActiveNav.Render(fmt.Sprintf(" %s %-11s [%s]", prefix, label, key))
		}
		if r == "live" && liveCount > 0 {
			dot := styles.DimText.Render("●")
			if blink {
				dot = styles.LiveBadge.Render("●")
			}
			return " " + dot + styles.InactiveNav.Render(fmt.Sprintf(" %-11s [%s]", label, key))
		}
		return styles.InactiveNav.Render(fmt.Sprintf("   %-11s [%s]", label, key))
	}

	rule := styles.DimText.Render(strings.Repeat("─", sidebarWidth-2))

	var sb strings.Builder
	sb.WriteString("\n")
	sb.WriteString(styles.GoldBold.Render(" ⚽ GOLAZO TUI") + "\n")
	sb.WriteString(" " + rule + "\n")
	sb.WriteString(nav("h", "LIVE", "live") + "\n")
	sb.WriteString(nav("g", "Standings", "standings") + "\n")
	sb.WriteString(nav("f", "Fixtures", "fixtures") + "\n")
	sb.WriteString(nav("c", "Changelog", "changelog") + "\n")
	sb.WriteString(" " + rule + "\n")
	sb.WriteString(styles.GoldText.Render(" 🏆 WC 2026") + "\n")
	sb.WriteString(styles.DimText.Render(" 48 Teams") + "\n")
	sb.WriteString(styles.DimText.Render(" 12 Groups") + "\n")
	sb.WriteString(" " + rule + "\n")
	sb.WriteString(styles.DimText.Render(" created by") + "\n")
	sb.WriteString(styles.DimText.Render(" Melvin Nijholt") + "\n")
	sb.WriteString(styles.GoldText.Render(" djMelvee") + "\n")

	return lipgloss.NewStyle().Width(sidebarWidth).Render(sb.String())
}

func renderFooter(route string, w int) string {
	var items []string
	if route == "detail" {
		items = []string{
			styles.GoldText.Render("b back"),
			styles.DimText.Render("j/k scroll events"),
			styles.DimText.Render("q quit"),
		}
	} else {
		items = []string{
			hintKey("h", "live", route == "live"),
			hintKey("g", "standings", route == "standings"),
			hintKey("f", "fixtures", route == "fixtures"),
			hintKey("c", "changelog", route == "changelog"),
			styles.DimText.Render("enter detail"),
			styles.DimText.Render("q quit"),
		}
	}
	bar := "  " + strings.Join(items, styles.DimText.Render("  ·  "))
	return lipgloss.NewStyle().Width(w).Render(bar)
}

func hintKey(key, label string, active bool) string {
	if active {
		return styles.GoldText.Render(key + " " + label)
	}
	return styles.DimText.Render(key + " " + label)
}
