package domain

type ReportRow struct {
	State        State
	PlayerID     int
	DungeonTime  int
	AvgFloorTime int
	BossTime     int
	Health       int
}
