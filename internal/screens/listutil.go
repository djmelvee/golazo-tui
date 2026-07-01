package screens

import "strings"

func wrapLines(s string, width int) []string {
	if width <= 20 {
		width = 72
	}
	max := width - 4
	words := strings.Fields(s)
	if len(words) == 0 {
		return nil
	}
	var line string
	var out []string
	for _, w := range words {
		if line == "" {
			line = w
			continue
		}
		if len(line)+1+len(w) > max {
			out = append(out, line)
			line = w
		} else {
			line += " " + w
		}
	}
	if line != "" {
		out = append(out, line)
	}
	return out
}

// ScrollToItem adjusts scroll so the item at cursor stays visible.
// itemStarts[i] is the line index where item i begins; itemLines[i] is its height.
func ScrollToItem(cursor int, itemStarts []int, itemLines []int, visible, scroll int) int {
	if cursor < 0 || cursor >= len(itemStarts) {
		return scroll
	}
	start := itemStarts[cursor]
	end := start + itemLines[cursor]
	if start < scroll {
		return start
	}
	if end > scroll+visible {
		return end - visible
	}
	return scroll
}