package domain

type EventID int

const (
	EventRegister       EventID = 1
	EventEnterDungeon   EventID = 2
	EventKillMonster    EventID = 3
	EventNextFloor      EventID = 4
	EventPreviousFloor  EventID = 5
	EventEnterBossFloor EventID = 6
	EventKillBoss       EventID = 7
	EventLeaveDungeon   EventID = 8
	EventCannotContinue EventID = 9
	EventRestoreHealth  EventID = 10
	EventReceiveDamage  EventID = 11
)

type Event struct {
	At       int
	PlayerID int
	ID       EventID
	Extra    string
}
