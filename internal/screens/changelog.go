package screens

import (
	"os"
	"strings"

	"github.com/djmelvee/golazo-tui/internal/styles"
)

// Changelog renders CHANGELOG.md inside the TUI.
type Changelog struct {
	w, h   int
	lines  []string
	scroll int
}

func (c *Changelog) SetSize(w, h int) {
	c.w = w
	c.h = h
}

// Load reads CHANGELOG.md from the working directory and pre-renders each line.
func (c *Changelog) Load() {
	raw, err := os.ReadFile("CHANGELOG.md")
	if err != nil {
		c.lines = []string{
			styles.DimText.Render("  CHANGELOG.md not found."),
			styles.DimText.Render("  Run golazo-tui from the project root directory."),
		}
		c.scroll = 0
		return
	}

	rawLines := strings.Split(strings.ReplaceAll(string(raw), "\r\n", "\n"), "\n")
	c.lines = make([]string, 0, len(rawLines))
	for _, l := range rawLines {
		c.lines = append(c.lines, renderChangelogLine(l))
	}
	c.scroll = 0
}

func (c *Changelog) ScrollDown() {
	visible := c.h - 6
	if visible < 1 {
		visible = 20
	}
	max := len(c.lines) - visible
	if max < 0 {
		max = 0
	}
	if c.scroll < max {
		c.scroll++
	}
}

func (c *Changelog) ScrollUp() {
	if c.scroll > 0 {
		c.scroll--
	}
}

func (c Changelog) View() string {
	var sb strings.Builder
	sb.WriteString(styles.Heading.Render("  ─── CHANGELOG  ·  FIFA WORLD CUP 2026"))
	sb.WriteString("\n\n")

	if len(c.lines) == 0 {
		sb.WriteString(styles.DimText.Render("  No changelog data.\n"))
		return sb.String()
	}

	visible := c.h - 6
	if visible < 1 {
		visible = 20
	}

	start := c.scroll
	end := start + visible
	if end > len(c.lines) {
		end = len(c.lines)
	}

	for _, line := range c.lines[start:end] {
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	if len(c.lines) > visible {
		sb.WriteString("\n")
		sb.WriteString(styles.DimText.Render(
			"  j/k or ↑↓ to scroll",
		))
	}

	return sb.String()
}

func renderChangelogLine(line string) string {
	switch {
	case strings.HasPrefix(line, "# "):
		return styles.GoldBold.Render("  " + line)
	case strings.HasPrefix(line, "## "):
		return styles.GoldBold.Render("  " + line)
	case strings.HasPrefix(line, "### "):
		return styles.GoldText.Render("  " + line)
	case strings.HasPrefix(line, "- "):
		return styles.DimText.Render("  " + line)
	case line == "":
		return ""
	default:
		return styles.MainText.Render("  " + line)
	}
}
