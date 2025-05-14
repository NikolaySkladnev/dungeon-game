package output

import (
	"bytes"
	"errors"
	"testing"

	"dungeon-game/internal/domain"
)

type brokenWriter struct{}

func (brokenWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestPresenterFormatReport(t *testing.T) {
	presenter := NewPresenter()
	rows := []domain.ReportRow{
		{
			State:        domain.StateSuccess,
			PlayerID:     1,
			DungeonTime:  24 * 60,
			AvgFloorTime: 5 * 60,
			BossTime:     11 * 60,
			Health:       35,
		},
	}

	got := presenter.FormatReport(rows)
	want := []string{
		"Final report:",
		"[SUCCESS] 1 [00:24:00, 00:05:00, 00:11:00] HP:35",
	}

	if len(got) != len(want) {
		t.Fatalf("len(got) = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestPresenterWrite(t *testing.T) {
	presenter := NewPresenter()
	var buf bytes.Buffer

	err := presenter.Write(&buf, []string{"[14:00:00] Player [1] registered"}, []domain.ReportRow{
		{State: domain.StateFail, PlayerID: 1, DungeonTime: 0, AvgFloorTime: 0, BossTime: 0, Health: 100},
	})
	if err != nil {
		t.Fatalf("Write returned error: %v", err)
	}

	want := "[14:00:00] Player [1] registered\nFinal report:\n[FAIL] 1 [00:00:00, 00:00:00, 00:00:00] HP:100\n"
	if got := buf.String(); got != want {
		t.Fatalf("Write output = %q, want %q", got, want)
	}
}

func TestPresenterWriteReturnsError(t *testing.T) {
	presenter := NewPresenter()
	if err := presenter.Write(brokenWriter{}, []string{"x"}, nil); err == nil {
		t.Fatal("Write expected error")
	}
}
