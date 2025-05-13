package usecase

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"dungeon-game/internal/domain"
)

type Processor struct {
	cfg     domain.DungeonConfig
	players map[int]*domain.Player
	logs    []string
}

func NewProcessor(cfg domain.DungeonConfig) *Processor {
	return &Processor{
		cfg:     cfg,
		players: make(map[int]*domain.Player),
	}
}

func (p *Processor) Run(events []domain.Event) ([]string, []domain.ReportRow) {
	for _, event := range events {
		p.finishExpiredPlayers(event.At)
		p.Process(event)
	}

	p.finishExpiredPlayers(p.cfg.CloseAtSeconds())
	return p.logs, p.Report()
}

func (p *Processor) Process(event domain.Event) {
	player := p.getPlayer(event.PlayerID)

	if player.Ended {
		return
	}

	if event.ID != domain.EventRegister && !player.Registered {
		p.disqualify(player, event.At)
		return
	}

	switch event.ID {
	case domain.EventRegister:
		p.register(player, event.At)
	case domain.EventEnterDungeon:
		p.enterDungeon(player, event.At)
	case domain.EventKillMonster:
		p.killMonster(player, event.At)
	case domain.EventNextFloor:
		p.nextFloor(player, event.At)
	case domain.EventPreviousFloor:
		p.previousFloor(player, event.At)
	case domain.EventEnterBossFloor:
		p.enterBossFloor(player, event.At)
	case domain.EventKillBoss:
		p.killBoss(player, event.At)
	case domain.EventLeaveDungeon:
		p.leaveDungeon(player, event.At)
	case domain.EventCannotContinue:
		p.cannotContinue(player, event.At, event.Extra)
	case domain.EventRestoreHealth:
		p.restoreHealth(player, event.At, event.Extra)
	case domain.EventReceiveDamage:
		p.receiveDamage(player, event.At, event.Extra)
	default:
		p.impossibleMove(player, event.At, event.ID)
	}
}

func (p *Processor) Report() []domain.ReportRow {
	ids := make([]int, 0, len(p.players))
	for id := range p.players {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	rows := make([]domain.ReportRow, 0, len(ids))
	for _, id := range ids {
		player := p.players[id]
		rows = append(rows, domain.ReportRow{
			State:        p.state(player),
			PlayerID:     player.ID,
			DungeonTime:  p.dungeonTime(player),
			AvgFloorTime: p.averageFloorClearTime(player),
			BossTime:     player.Boss.KillTime,
			Health:       player.Health,
		})
	}

	return rows
}

func (p *Processor) getPlayer(id int) *domain.Player {
	if player, ok := p.players[id]; ok {
		return player
	}

	player := domain.NewPlayer(id, p.cfg.RegularFloors())
	p.players[id] = player
	return player
}

func (p *Processor) register(player *domain.Player, at int) {
	if player.Registered || player.InDungeon {
		p.impossibleMove(player, at, domain.EventRegister)
		return
	}

	player.Registered = true
	p.addLog(at, "Player [%d] registered", player.ID)
}

func (p *Processor) enterDungeon(player *domain.Player, at int) {
	if player.InDungeon || at < p.cfg.OpenAtSeconds || at >= p.cfg.CloseAtSeconds() {
		p.impossibleMove(player, at, domain.EventEnterDungeon)
		return
	}

	player.InDungeon = true
	player.EnteredAt = at
	player.CurrentFloor = 1
	p.enterRegularFloor(player, 1, at)

	p.addLog(at, "Player [%d] entered the dungeon", player.ID)
}

func (p *Processor) killMonster(player *domain.Player, at int) {
	if !player.InDungeon || !p.isRegularFloor(player.CurrentFloor) {
		p.impossibleMove(player, at, domain.EventKillMonster)
		return
	}

	floor := &player.Floors[player.CurrentFloor]
	if floor.Cleared || floor.MonstersKilled >= p.cfg.Monsters {
		p.impossibleMove(player, at, domain.EventKillMonster)
		return
	}

	floor.MonstersKilled++
	p.addLog(at, "Player [%d] killed the monster", player.ID)

	if floor.MonstersKilled == p.cfg.Monsters {
		p.clearRegularFloor(player, player.CurrentFloor, at)
	}
}

func (p *Processor) nextFloor(player *domain.Player, at int) {
	if !player.InDungeon || player.CurrentFloor < 1 || player.CurrentFloor >= p.cfg.Floors {
		p.impossibleMove(player, at, domain.EventNextFloor)
		return
	}

	if p.isRegularFloor(player.CurrentFloor) && !player.Floors[player.CurrentFloor].Cleared {
		p.impossibleMove(player, at, domain.EventNextFloor)
		return
	}

	p.leaveCurrentFloor(player, at)
	player.CurrentFloor++
	if p.isRegularFloor(player.CurrentFloor) {
		p.enterRegularFloor(player, player.CurrentFloor, at)
	}

	p.addLog(at, "Player [%d] went to the next floor", player.ID)
}

func (p *Processor) previousFloor(player *domain.Player, at int) {
	if !player.InDungeon || player.CurrentFloor <= 1 {
		p.impossibleMove(player, at, domain.EventPreviousFloor)
		return
	}

	p.leaveCurrentFloor(player, at)
	player.CurrentFloor--
	if p.isRegularFloor(player.CurrentFloor) {
		p.enterRegularFloor(player, player.CurrentFloor, at)
	}

	p.addLog(at, "Player [%d] went to the previous floor", player.ID)
}

func (p *Processor) enterBossFloor(player *domain.Player, at int) {
	if !player.InDungeon || player.CurrentFloor != p.cfg.Floors || player.Boss.Active || player.Boss.Killed || !p.allRegularFloorsCleared(player) {
		p.impossibleMove(player, at, domain.EventEnterBossFloor)
		return
	}

	player.Boss.Entered = true
	player.Boss.Active = true
	player.Boss.ActiveFrom = at
	p.addLog(at, "Player [%d] entered the boss's floor", player.ID)
}

func (p *Processor) killBoss(player *domain.Player, at int) {
	if !player.InDungeon || player.CurrentFloor != p.cfg.Floors || !player.Boss.Active || player.Boss.Killed {
		p.impossibleMove(player, at, domain.EventKillBoss)
		return
	}

	player.Boss.SpentBefore += at - player.Boss.ActiveFrom
	player.Boss.Active = false
	player.Boss.Killed = true
	player.Boss.KillTime = player.Boss.SpentBefore
	p.addLog(at, "Player [%d] killed the boss", player.ID)
}

func (p *Processor) leaveDungeon(player *domain.Player, at int) {
	if !player.InDungeon {
		p.impossibleMove(player, at, domain.EventLeaveDungeon)
		return
	}

	p.leaveCurrentFloor(player, at)
	player.InDungeon = false
	player.Ended = true
	player.EndedAt = at
	p.addLog(at, "Player [%d] left the dungeon", player.ID)
}

func (p *Processor) cannotContinue(player *domain.Player, at int, reason string) {
	if reason == "" {
		reason = "unknown"
	}

	if player.InDungeon {
		p.leaveCurrentFloor(player, at)
		player.InDungeon = false
		player.EndedAt = at
	}
	player.Disqualified = true
	player.Ended = true
	p.addLog(at, "Player [%d] cannot continue due to [%s]", player.ID, reason)
}

func (p *Processor) restoreHealth(player *domain.Player, at int, extra string) {
	if !player.InDungeon {
		p.impossibleMove(player, at, domain.EventRestoreHealth)
		return
	}

	health, ok := parseNonNegativeInt(extra)
	if !ok {
		p.impossibleMove(player, at, domain.EventRestoreHealth)
		return
	}

	player.Health += health
	if player.Health > domain.MaxHealth {
		player.Health = domain.MaxHealth
	}
	p.addLog(at, "Player [%d] has restored [%d] of health", player.ID, health)
}

func (p *Processor) receiveDamage(player *domain.Player, at int, extra string) {
	if !player.InDungeon {
		p.impossibleMove(player, at, domain.EventReceiveDamage)
		return
	}

	damage, ok := parseNonNegativeInt(extra)
	if !ok {
		p.impossibleMove(player, at, domain.EventReceiveDamage)
		return
	}

	player.Health -= damage
	if player.Health < 0 {
		player.Health = 0
	}
	p.addLog(at, "Player [%d] recieved [%d] of damage", player.ID, damage)

	if player.Health == 0 {
		p.leaveCurrentFloor(player, at)
		player.InDungeon = false
		player.Ended = true
		player.EndedAt = at
		p.addLog(at, "Player [%d] is dead", player.ID)
	}
}

func (p *Processor) disqualify(player *domain.Player, at int) {
	player.Disqualified = true
	player.Ended = true
	player.InDungeon = false
	p.addLog(at, "Player [%d] is disqualified", player.ID)
}

func (p *Processor) impossibleMove(player *domain.Player, at int, eventID domain.EventID) {
	p.addLog(at, "Player [%d] makes imposible move [%d]", player.ID, eventID)
}

func (p *Processor) finishExpiredPlayers(now int) {
	closeAt := p.cfg.CloseAtSeconds()
	if now < closeAt {
		return
	}

	for _, player := range p.players {
		if player.InDungeon && !player.Ended {
			p.leaveCurrentFloor(player, closeAt)
			player.InDungeon = false
			player.Ended = true
			player.EndedAt = closeAt
		}
	}
}

func (p *Processor) enterRegularFloor(player *domain.Player, floorNumber int, at int) {
	floor := &player.Floors[floorNumber]
	floor.Visited = true
	if !floor.Cleared && !floor.Active {
		floor.Active = true
		floor.ActiveFrom = at
	}
}

func (p *Processor) clearRegularFloor(player *domain.Player, floorNumber int, at int) {
	floor := &player.Floors[floorNumber]
	if floor.Active {
		floor.SpentBefore += at - floor.ActiveFrom
		floor.Active = false
	}
	floor.Cleared = true
	floor.ClearTime = floor.SpentBefore
}

func (p *Processor) leaveCurrentFloor(player *domain.Player, at int) {
	if p.isRegularFloor(player.CurrentFloor) {
		floor := &player.Floors[player.CurrentFloor]
		if !floor.Cleared && floor.Active {
			floor.SpentBefore += at - floor.ActiveFrom
			floor.Active = false
		}
		return
	}

	if player.CurrentFloor == p.cfg.Floors && player.Boss.Active && !player.Boss.Killed {
		player.Boss.SpentBefore += at - player.Boss.ActiveFrom
		player.Boss.Active = false
	}
}

func (p *Processor) isRegularFloor(floorNumber int) bool {
	return floorNumber >= 1 && floorNumber <= p.cfg.RegularFloors()
}

func (p *Processor) allRegularFloorsCleared(player *domain.Player) bool {
	for floor := 1; floor <= p.cfg.RegularFloors(); floor++ {
		if !player.Floors[floor].Cleared {
			return false
		}
	}
	return true
}

func (p *Processor) completed(player *domain.Player) bool {
	return p.allRegularFloorsCleared(player) && player.Boss.Killed
}

func (p *Processor) state(player *domain.Player) domain.State {
	if player.Disqualified {
		return domain.StateDisqual
	}
	if p.completed(player) {
		return domain.StateSuccess
	}
	return domain.StateFail
}

func (p *Processor) dungeonTime(player *domain.Player) int {
	if player.EnteredAt == 0 && player.EndedAt == 0 && !player.InDungeon {
		return 0
	}
	if player.EndedAt > 0 {
		return player.EndedAt - player.EnteredAt
	}
	if player.InDungeon {
		return p.cfg.CloseAtSeconds() - player.EnteredAt
	}
	return 0
}

func (p *Processor) averageFloorClearTime(player *domain.Player) int {
	total := 0
	count := 0
	for floor := 1; floor <= p.cfg.RegularFloors(); floor++ {
		if player.Floors[floor].Cleared {
			total += player.Floors[floor].ClearTime
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return total / count
}

func (p *Processor) addLog(at int, format string, args ...any) {
	p.logs = append(p.logs, fmt.Sprintf("[%s] %s", domain.FormatClock(at), fmt.Sprintf(format, args...)))
}

func parseNonNegativeInt(value string) (int, bool) {
	if value == "" || strings.Contains(value, " ") {
		return 0, false
	}

	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}
