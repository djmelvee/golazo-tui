//go:build !windows

package notify

import (
	"os/exec"
)

// DesktopAlert shows a desktop notification via notify-send when available.
func DesktopAlert(title, body string) {
	if _, err := exec.LookPath("notify-send"); err != nil {
		return
	}
	_ = exec.Command("notify-send", "-a", "Golazo TUI", title, body).Run()
}