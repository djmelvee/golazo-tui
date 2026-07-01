package espn

import "testing"

func TestParseESPNMinute(t *testing.T) {
	cases := []struct {
		in   string
		want int
		ok   bool
	}{
		{"64'", 64, true},
		{"90'+7'", 90, true},
		{"45'+2'", 45, true},
		{"", 0, false},
		{"HT", 0, false},
	}
	for _, tc := range cases {
		got, ok := parseESPNMinute(tc.in)
		if ok != tc.ok || got != tc.want {
			t.Fatalf("parseESPNMinute(%q) = (%d, %v), want (%d, %v)", tc.in, got, ok, tc.want, tc.ok)
		}
	}
}