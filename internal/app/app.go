package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/djmelvee/golazo-tui/internal/data"
	"github.com/djmelvee/golazo-tui/internal/screens"
	"github.com/djmelvee/golazo-tui/internal/styles"
)

const sidebarWidth = 18

type Model struct {
	w, h      int
	route     string // "live" | "standings" | "fixtures"
	live      screens.Live
	standings screens.Standings
	fixtures  screens.Fixtures
	db        *data.Store
}

// New creates the root Bubble Tea model.
func New(db *data.Store) Model {
	m := Model{
		route: "live",
		db:    db,
	}
	m.live.Load(db)
	m.standings.Load(db)
	m.fixtures.Load(db)
	return m
}

func (m Model) Init() tea.Cmd {
	return screens.TickCmd()
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

	case screens.TickMsg:
		m.live.Load(m.db)
		return m, screens.TickCmd()

	case tea.KeyPressMsg:
		switch msg.String() {
		case "h":
			m.route = "live"
		case "g":
			m.route = "standings"
			m.standings.Load(m.db)
		case "f":
			m.route = "fixtures"
			m.fixtures.Load(m.db)
		case "j", "down":
			if m.route == "standings" {
				m.standings.ScrollDown()
			}
		case "k", "up":
			if m.route == "standings" {
				m.standings.ScrollUp()
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
	// Header
	header := styles.HeaderBar(m.w)

	// Sidebar
	sidebar := renderSidebar(m.route, m.h)

	// Content
	var content string
	switch m.route {
	case "live":
		content = m.live.View()
	case "standings":
		content = m.standings.View()
	case "fixtures":
		content = m.fixtures.View()
	}

	contentW := m.w - sidebarWidth
	if contentW < 40 {
		contentW = 40
	}

	// Crop content to available height
	contentHeight := m.h - 4 // header(1) + blank(1) + footer(1) + blank(1)
	if contentHeight < 1 {
		contentHeight = 20
	}
	contentLines := strings.Split(content, "\n")
	if len(contentLines) > contentHeight {
		contentLines = contentLines[:contentHeight]
	}
	// Pad to fill height
	for len(contentLines) < contentHeight {
		contentLines = append(contentLines, "")
	}

	// Pad sidebar lines to same height
	sidebarLines := strings.Split(sidebar, "\n")
	for len(sidebarLines) < contentHeight {
		sidebarLines = append(sidebarLines, strings.Repeat(" ", sidebarWidth))
	}

	// Join sidebar + content side by side
	body := lipgloss.JoinHorizontal(
		lipgloss.Top,
		strings.Join(sidebarLines, "\n"),
		strings.Join(contentLines, "\n"),
	)

	// Footer
	footer := renderFooter(m.route, m.w)

	return header + "\n" + body + "\n" + footer
}

func renderSidebar(route string, h int) string {
	nav := func(key, label, r string) string {
		var prefix string
		if r == route {
			prefix = styles.LiveBadge.Render("●")
			return styles.ActiveNav.Render(fmt.Sprintf(" %s %-11s [%s]", prefix, label, key))
		}
		prefix = " "
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
	sb.WriteString(" " + rule + "\n")
	sb.WriteString(styles.GoldText.Render(" 🏆 WC 2026") + "\n")
	sb.WriteString(styles.DimText.Render(" 48 Teams") + "\n")
	sb.WriteString(styles.DimText.Render(" 12 Groups") + "\n")

	return lipgloss.NewStyle().Width(sidebarWidth).Render(sb.String())
}

func renderFooter(route string, w int) string {
	items := []string{
		hintKey("h", "live", route == "live"),
		hintKey("g", "standings", route == "standings"),
		hintKey("f", "fixtures", route == "fixtures"),
		styles.DimText.Render("q quit"),
	}
	bar := "  " + strings.Join(items, styles.DimText.Render("  ·  "))
	return lipgloss.NewStyle().Width(w).Render(bar)
}

func hintKey(key, label string, active bool) string {
	if active {
		return styles.GoldText.Render(key+" "+label)
	}
	return styles.DimText.Render(key + " " + label)
}
