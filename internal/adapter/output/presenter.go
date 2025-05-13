package output

import (
	"fmt"
	"io"

	"dungeon-game/internal/domain"
)

type Presenter struct{}

func NewPresenter() Presenter {
	return Presenter{}
}

func (p Presenter) Write(w io.Writer, logs []string, report []domain.ReportRow) error {
	for _, line := range logs {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	for _, line := range p.FormatReport(report) {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}

	return nil
}

func (p Presenter) FormatReport(rows []domain.ReportRow) []string {
	lines := []string{"Final report:"}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf(
			"[%s] %d [%s, %s, %s] HP:%d",
			row.State,
			row.PlayerID,
			domain.FormatDuration(row.DungeonTime),
			domain.FormatDuration(row.AvgFloorTime),
			domain.FormatDuration(row.BossTime),
			row.Health,
		))
	}
	return lines
}
