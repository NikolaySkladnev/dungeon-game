package usecase_test

import (
	"strings"
	"testing"

	"dungeon-game/internal/adapter/input"
	"dungeon-game/internal/adapter/output"
	"dungeon-game/internal/domain"
	"dungeon-game/internal/usecase"
)

func TestProcessorSample(t *testing.T) {
	openAt, err := domain.ParseClock("14:05:00")
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := domain.NewDungeonConfig(2, 2, openAt, 2)
	if err != nil {
		t.Fatal(err)
	}

	inputData := `[14:00:00] 1 1
[14:00:00] 2 1
[14:10:00] 2 2
[14:10:00] 3 2
[14:11:00] 2 5
[14:12:00] 3 3
[14:14:00] 2 3
[14:27:00] 2 11 60
[14:29:00] 2 11 50
[14:40:00] 1 2
[14:41:00] 1 3
[14:44:00] 1 11 50
[14:45:00] 1 3
[14:48:00] 1 4
[14:48:00] 1 6
[14:49:00] 1 11 25
[14:49:02] 1 10 80
[14:50:00] 1 11 65
[14:59:00] 1 7
[15:04:00] 1 8`

	reader := input.NewEventReader()
	events, err := reader.Read(strings.NewReader(inputData))
	if err != nil {
		t.Fatal(err)
	}

	processor := usecase.NewProcessor(cfg)
	logs, report := processor.Run(events)

	presenter := output.NewPresenter()
	got := strings.Join(append(logs, presenter.FormatReport(report)...), "\n")

	want := `[14:00:00] Player [1] registered
[14:00:00] Player [2] registered
[14:10:00] Player [2] entered the dungeon
[14:10:00] Player [3] is disqualified
[14:11:00] Player [2] makes imposible move [5]
[14:14:00] Player [2] killed the monster
[14:27:00] Player [2] recieved [60] of damage
[14:29:00] Player [2] recieved [50] of damage
[14:29:00] Player [2] is dead
[14:40:00] Player [1] entered the dungeon
[14:41:00] Player [1] killed the monster
[14:44:00] Player [1] recieved [50] of damage
[14:45:00] Player [1] killed the monster
[14:48:00] Player [1] went to the next floor
[14:48:00] Player [1] entered the boss's floor
[14:49:00] Player [1] recieved [25] of damage
[14:49:02] Player [1] has restored [80] of health
[14:50:00] Player [1] recieved [65] of damage
[14:59:00] Player [1] killed the boss
[15:04:00] Player [1] left the dungeon
Final report:
[SUCCESS] 1 [00:24:00, 00:05:00, 00:11:00] HP:35
[FAIL] 2 [00:19:00, 00:00:00, 00:00:00] HP:0
[DISQUAL] 3 [00:00:00, 00:00:00, 00:00:00] HP:100`

	if got != want {
		t.Fatalf("unexpected output\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestProcessorDisqualifiesUnregisteredPlayer(t *testing.T) {
	cfg := mustConfig(t, "14:05:00", 2, 2, 2)
	processor := usecase.NewProcessor(cfg)

	logs, report := processor.Run([]domain.Event{
		{At: mustClock(t, "14:10:00"), PlayerID: 10, ID: domain.EventEnterDungeon},
	})

	if len(logs) != 1 || logs[0] != "[14:10:00] Player [10] is disqualified" {
		t.Fatalf("logs = %#v", logs)
	}
	row := findReportRow(t, report, 10)
	if row.State != domain.StateDisqual {
		t.Fatalf("state = %s, want %s", row.State, domain.StateDisqual)
	}
	if row.DungeonTime != 0 || row.Health != domain.MaxHealth {
		t.Fatalf("row = %+v, want zero dungeon time and full hp", row)
	}
}

func TestProcessorHealthCannotExceed100(t *testing.T) {
	cfg := mustConfig(t, "14:00:00", 2, 1, 1)
	processor := usecase.NewProcessor(cfg)

	_, report := processor.Run([]domain.Event{
		{At: mustClock(t, "14:00:00"), PlayerID: 1, ID: domain.EventRegister},
		{At: mustClock(t, "14:00:01"), PlayerID: 1, ID: domain.EventEnterDungeon},
		{At: mustClock(t, "14:01:00"), PlayerID: 1, ID: domain.EventReceiveDamage, Extra: "20"},
		{At: mustClock(t, "14:02:00"), PlayerID: 1, ID: domain.EventRestoreHealth, Extra: "50"},
		{At: mustClock(t, "14:03:00"), PlayerID: 1, ID: domain.EventLeaveDungeon},
	})

	row := findReportRow(t, report, 1)
	if row.Health != domain.MaxHealth {
		t.Fatalf("health = %d, want %d", row.Health, domain.MaxHealth)
	}
	if row.State != domain.StateFail {
		t.Fatalf("state = %s, want %s because dungeon was not completed", row.State, domain.StateFail)
	}
}

func TestProcessorStopsActivePlayerAtDungeonClose(t *testing.T) {
	cfg := mustConfig(t, "14:00:00", 2, 2, 1)
	processor := usecase.NewProcessor(cfg)

	_, report := processor.Run([]domain.Event{
		{At: mustClock(t, "14:00:00"), PlayerID: 1, ID: domain.EventRegister},
		{At: mustClock(t, "14:10:00"), PlayerID: 1, ID: domain.EventEnterDungeon},
	})

	row := findReportRow(t, report, 1)
	wantDungeonTime := 50 * 60
	if row.DungeonTime != wantDungeonTime {
		t.Fatalf("DungeonTime = %d, want %d", row.DungeonTime, wantDungeonTime)
	}
	if row.State != domain.StateFail {
		t.Fatalf("state = %s, want %s", row.State, domain.StateFail)
	}
}

func TestProcessorCannotContinueWithMultiWordReason(t *testing.T) {
	cfg := mustConfig(t, "14:00:00", 2, 1, 1)
	processor := usecase.NewProcessor(cfg)

	logs, report := processor.Run([]domain.Event{
		{At: mustClock(t, "14:00:00"), PlayerID: 1, ID: domain.EventRegister},
		{At: mustClock(t, "14:01:00"), PlayerID: 1, ID: domain.EventEnterDungeon},
		{At: mustClock(t, "14:06:00"), PlayerID: 1, ID: domain.EventCannotContinue, Extra: "lost internet connection"},
	})

	joinedLogs := strings.Join(logs, "\n")
	if !strings.Contains(joinedLogs, "[14:06:00] Player [1] cannot continue due to [lost internet connection]") {
		t.Fatalf("logs do not contain cannot-continue reason:\n%s", joinedLogs)
	}

	row := findReportRow(t, report, 1)
	if row.State != domain.StateDisqual {
		t.Fatalf("state = %s, want %s", row.State, domain.StateDisqual)
	}
	if row.DungeonTime != 5*60 {
		t.Fatalf("DungeonTime = %d, want %d", row.DungeonTime, 5*60)
	}
}

func TestProcessorAverageFloorTimeForMultipleFloorsExcludesBoss(t *testing.T) {
	cfg := mustConfig(t, "14:00:00", 3, 1, 2)
	processor := usecase.NewProcessor(cfg)

	_, report := processor.Run([]domain.Event{
		{At: mustClock(t, "14:00:00"), PlayerID: 1, ID: domain.EventRegister},
		{At: mustClock(t, "14:00:00"), PlayerID: 1, ID: domain.EventEnterDungeon},
		{At: mustClock(t, "14:05:00"), PlayerID: 1, ID: domain.EventKillMonster},
		{At: mustClock(t, "14:10:00"), PlayerID: 1, ID: domain.EventNextFloor},
		{At: mustClock(t, "14:18:00"), PlayerID: 1, ID: domain.EventKillMonster},
		{At: mustClock(t, "14:18:00"), PlayerID: 1, ID: domain.EventNextFloor},
		{At: mustClock(t, "14:18:00"), PlayerID: 1, ID: domain.EventEnterBossFloor},
		{At: mustClock(t, "14:28:00"), PlayerID: 1, ID: domain.EventKillBoss},
		{At: mustClock(t, "14:30:00"), PlayerID: 1, ID: domain.EventLeaveDungeon},
	})

	row := findReportRow(t, report, 1)
	if row.State != domain.StateSuccess {
		t.Fatalf("state = %s, want %s", row.State, domain.StateSuccess)
	}
	if row.DungeonTime != 30*60 {
		t.Fatalf("DungeonTime = %d, want %d", row.DungeonTime, 30*60)
	}
	if row.AvgFloorTime != 6*60+30 {
		t.Fatalf("AvgFloorTime = %d, want %d", row.AvgFloorTime, 6*60+30)
	}
	if row.BossTime != 10*60 {
		t.Fatalf("BossTime = %d, want %d", row.BossTime, 10*60)
	}
}

func mustConfig(t *testing.T, openAt string, floors int, monsters int, duration int) domain.DungeonConfig {
	t.Helper()
	openAtSeconds := mustClock(t, openAt)
	cfg, err := domain.NewDungeonConfig(floors, monsters, openAtSeconds, duration)
	if err != nil {
		t.Fatalf("NewDungeonConfig: %v", err)
	}
	return cfg
}

func mustClock(t *testing.T, value string) int {
	t.Helper()
	seconds, err := domain.ParseClock(value)
	if err != nil {
		t.Fatalf("ParseClock(%q): %v", value, err)
	}
	return seconds
}

func findReportRow(t *testing.T, rows []domain.ReportRow, playerID int) domain.ReportRow {
	t.Helper()
	for _, row := range rows {
		if row.PlayerID == playerID {
			return row
		}
	}
	t.Fatalf("player %d not found in report: %+v", playerID, rows)
	return domain.ReportRow{}
}
