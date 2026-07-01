//go:build !windows

package notify

import (
	"fmt"
	"os"
)

func GoalBell() {
	fmt.Fprint(os.Stdout, "\a")
	fmt.Fprint(os.Stderr, "\a")
}