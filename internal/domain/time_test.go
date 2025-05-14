package domain

import "testing"

func TestParseClockValid(t *testing.T) {
	got, err := ParseClock("14:05:09")
	if err != nil {
		t.Fatalf("ParseClock returned error: %v", err)
	}

	want := 14*SecondsInHour + 5*60 + 9
	if got != want {
		t.Fatalf("ParseClock() = %d, want %d", got, want)
	}
}

func TestParseClockInvalid(t *testing.T) {
	tests := []string{
		"14:05",
		"24:00:00",
		"14:5:00",
		"bad-time",
	}

	for _, tc := range tests {
		t.Run(tc, func(t *testing.T) {
			if _, err := ParseClock(tc); err == nil {
				t.Fatalf("ParseClock(%q) expected error", tc)
			}
		})
	}
}

func TestFormatClock(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want string
	}{
		{name: "usual", in: 14*SecondsInHour + 5*60 + 9, want: "14:05:09"},
		{name: "wrap after day", in: SecondsInDay + 61, want: "00:01:01"},
		{name: "negative wrap", in: -1, want: "23:59:59"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatClock(tc.in); got != tc.want {
				t.Fatalf("FormatClock(%d) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatDuration(t *testing.T) {
	tests := []struct {
		name string
		in   int
		want string
	}{
		{name: "zero", in: 0, want: "00:00:00"},
		{name: "usual", in: 2*SecondsInHour + 3*60 + 4, want: "02:03:04"},
		{name: "more than day", in: 27*SecondsInHour + 5, want: "27:00:05"},
		{name: "negative becomes zero", in: -10, want: "00:00:00"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := FormatDuration(tc.in); got != tc.want {
				t.Fatalf("FormatDuration(%d) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}
