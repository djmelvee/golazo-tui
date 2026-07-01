package ascii

import (
	"strings"

	"github.com/mattn/go-runewidth"
)

var golazoLogo = []string{
	`  ██████╗  ██████╗ ██╗      █████╗ ███████╗ ██████╗ `,
	` ██╔════╝ ██╔═══██╗██║     ██╔══██╗╚══███╔╝██╔═══██╗`,
	` ██║  ███╗██║   ██║██║     ███████║  ███╔╝ ██║   ██║`,
	` ██║   ██║██║   ██║██║     ██╔══██║ ███╔╝  ██║   ██║`,
	` ╚██████╔╝╚██████╔╝███████╗██║  ██║███████╗╚██████╔╝`,
	`  ╚═════╝  ╚═════╝ ╚══════╝╚═╝  ╚═╝╚══════╝ ╚═════╝ `,
}

var golazoLogoCompact = []string{
	`  ___   ___  _     ___  ___  ___ `,
	` / __| | _ \| |   / _ \| _ \/ _ \`,
	`| (__  |   /| |__| (_) |   / (_) |`,
	` \___| |_|_\|____|\___/|_|_\___/ `,
}

// SplashFrames returns boot animation frames: GOLAZO logo + 3D spinning soccer ball.
func SplashFrames(width int) []string {
	if width < 50 {
		width = 50
	}
	const spinFrames = 24
	var frames []string
	for f := 0; f < spinFrames; f++ {
		frames = append(frames, composeSplash(width, f))
	}
	return frames
}

func composeSplash(width, spinFrame int) string {
	var lines []string
	logo := golazoLogo
	if width < 72 {
		logo = golazoLogoCompact
	}
	lines = append(lines, logo...)
	lines = append(lines, "")
	lines = append(lines, SpinBall3D(spinFrame)...)
	lines = append(lines, "")
	lines = append(lines, "* FIFA WORLD CUP 2026 *")
	lines = append(lines, "USA  ·  CANADA  ·  MEXICO")
	lines = append(lines, strings.Repeat("─", clamp(width-4, 24, 56)))
	return centerBlock(width, lines)
}

// HeaderBanner returns a compact World Cup 2026 ASCII top bar.
func HeaderBanner(width int) string {
	if width < 56 {
		return compactHeader(width)
	}
	return wideHeader(width)
}

func wideHeader(width int) string {
	art := []string{
		boxLine(width, "╔", "═", "╗"),
		boxPad(width, " ⚽ FIFA WORLD CUP 2026 · GOLAZO · USA · CAN · MEX ⚽"),
		boxLine(width, "╚", "═", "╝"),
	}
	return strings.Join(art, "\n")
}

func compactHeader(width int) string {
	art := []string{
		boxLine(width, "╔", "═", "╗"),
		boxPad(width, " ⚽ WC 2026 · GOLAZO LIVE ⚽"),
		boxLine(width, "╚", "═", "╝"),
	}
	return strings.Join(art, "\n")
}

func boxLine(width int, left, fill, right string) string {
	inner := width - 2
	if inner < 20 {
		inner = 20
	}
	if inner > 120 {
		inner = 120
	}
	return left + strings.Repeat(fill, inner) + right
}

func boxPad(width int, content string) string {
	inner := width - 2
	if inner < 20 {
		inner = 20
	}
	if inner > 120 {
		inner = 120
	}
	pad := inner - runewidth.StringWidth(content)
	if pad < 0 {
		pad = 0
	}
	return "║" + content + strings.Repeat(" ", pad) + "║"
}

// GoalBurst returns ASCII art for a goal celebration overlay.
func GoalBurst(width int) string {
	lines := []string{
		"",
		"   *   Golazo!   *   ",
		"  \\\\ | //  \\\\ | //  ",
		"   \\|/    \\|/    ",
		"",
	}
	return centerBlock(width, lines)
}

func centerBlock(width int, lines []string) string {
	var out []string
	for _, l := range lines {
		w := runewidth.StringWidth(l)
		if w >= width {
			out = append(out, l)
			continue
		}
		pad := (width - w) / 2
		out = append(out, strings.Repeat(" ", pad)+l)
	}
	return strings.Join(out, "\n")
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// HeaderLineCount returns how many lines HeaderBanner produces.
func HeaderLineCount(width int) int {
	return 3
}