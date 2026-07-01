//go:build windows

package notify

import (
	"fmt"
	"os"
	"syscall"
)

var (
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	procBeep = kernel32.NewProc("Beep")
)

func GoalBell() {
	_, _, _ = procBeep.Call(880, 200)
	_, _, _ = procBeep.Call(1100, 150)
	fmt.Fprint(os.Stdout, "\a")
	fmt.Fprint(os.Stderr, "\a")
}