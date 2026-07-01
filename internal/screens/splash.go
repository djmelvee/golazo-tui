package screens

import (
	"os"
	"strings"

	"github.com/djmelvee/golazo-tui/internal/ascii"
	"github.com/djmelvee/golazo-tui/internal/styles"
)

// Splash is the boot animation screen.
type Splash struct {
	w, h   int
	frame  int
	frames []string
}

// SkipSplash returns true when GOLAZO_NO_SPLASH is set.
func SkipSplash() bool {
	return os.Getenv("GOLAZO_NO_SPLASH") == "1"
}

func NewSplash(w, h int) Splash {
	frames := ascii.SplashFrames(w)
	return Splash{w: w, h: h, frames: frames}
}

func (s *Splash) SetSize(w, h int) {
	s.w = w
	s.h = h
	s.frames = ascii.SplashFrames(w)
}

func (s *Splash) Advance() {
	s.frame = (s.frame + 1) % len(s.frames)
}

func (s *Splash) View() string {
	if len(s.frames) == 0 {
		return ""
	}
	body := s.frames[s.frame%len(s.frames)]
	lines := strings.Split(body, "\n")
	pad := (s.h - len(lines) - 2) / 2
	if pad < 0 {
		pad = 0
	}
	dots := strings.Repeat(".", (s.frame%3)+1)
	var sb strings.Builder
	for i := 0; i < pad; i++ {
		sb.WriteString("\n")
	}
	sb.WriteString(styles.GoldBold.Render(body))
	sb.WriteString("\n\n")
	sb.WriteString(styles.DimText.Render("  Loading World Cup data" + dots))
	return sb.String()
}