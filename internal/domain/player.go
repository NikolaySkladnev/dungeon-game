package domain

const MaxHealth = 100

type State string

const (
	StateSuccess State = "SUCCESS"
	StateFail    State = "FAIL"
	StateDisqual State = "DISQUAL"
)

type Player struct {
	ID int

	Registered   bool
	Disqualified bool
	Ended        bool
	InDungeon    bool

	EnteredAt int
	EndedAt   int

	Health       int
	CurrentFloor int

	Floors []FloorProgress
	Boss   BossProgress
}

type FloorProgress struct {
	Visited        bool
	Active         bool
	ActiveFrom     int
	SpentBefore    int
	MonstersKilled int
	Cleared        bool
	ClearTime      int
}

type BossProgress struct {
	Entered     bool
	Active      bool
	ActiveFrom  int
	SpentBefore int
	Killed      bool
	KillTime    int
}

func NewPlayer(id int, regularFloors int) *Player {
	return &Player{
		ID:     id,
		Health: MaxHealth,
		Floors: make([]FloorProgress, regularFloors+1),
	}
}
