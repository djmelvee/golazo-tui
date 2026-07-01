package ascii

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestHeaderBannerWidth(t *testing.T) {
	for _, w := range []int{60, 80, 120} {
		b := HeaderBanner(w)
		if b == "" {
			t.Fatalf("empty banner at width %d", w)
		}
		lines := strings.Split(b, "\n")
		if len(lines) != HeaderLineCount(w) {
			t.Fatalf("width %d: got %d lines, want %d", w, len(lines), HeaderLineCount(w))
		}
	}
}

func TestHeaderBoxAlignment(t *testing.T) {
	b := HeaderBanner(100)
	for _, line := range strings.Split(b, "\n") {
		if runewidth.StringWidth(line) > 102 {
			t.Fatalf("line wider than box: %q", line)
		}
	}
}

func TestSplashFrames(t *testing.T) {
	frames := SplashFrames(100)
	if len(frames) < 16 {
		t.Fatalf("expected spin frames, got %d", len(frames))
	}
	if frames[0] == frames[12] {
		t.Fatal("splash frames should animate")
	}
}

func TestSpinBall3D(t *testing.T) {
	a := SpinBall3D(0)
	b := SpinBall3D(8)
	if len(a) < 10 {
		t.Fatalf("spin ball too short: %d", len(a))
	}
	if strings.Join(a, "") == strings.Join(b, "") {
		t.Fatal("3D ball should change between frames")
	}
	shaded := strings.ContainsAny(strings.Join(a, ""), " .'`^\",:;!iIlL|/\\|)(][}{*#")
	if !shaded {
		t.Fatal("expected shaded ASCII luminance chars")
	}
}

func TestGoalBurst(t *testing.T) {
	if GoalBurst(80) == "" {
		t.Fatal("empty burst")
	}
}