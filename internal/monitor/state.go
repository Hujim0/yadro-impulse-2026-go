package monitor

import (
	"dungeonGameLib/lib/engine/event"
	"dungeonGameLib/lib/engine/formatting"
	"dungeonGameLib/lib/engine/gamestate"
	"fmt"
	"log"
	"strings"
)

type Config struct {
	Floors   int    `json:"Floors"`
	Monsters int    `json:"Monsters"`
	OpenAt   string `json:"OpenAt"`
	Duration int    `json:"Duration"`
}

type LogEntry struct {
	Time     string
	PlayerID int
	Message  string
	IsActive bool
	IsAlert  bool
}

type FloorView struct {
	Num          int
	IsBoss       bool
	Cleared      bool
	MonstersLeft int
	BossDefeated bool
	Occupants    []int
}

type MonitorState struct {
	Config      Config
	Events      []event.GameEvent
	game        *gamestate.GameInstance
	CurrentIdx  int
	SelectedPid int
	IsPlaying   bool
	PlayElapsed float64

	entries []LogEntry
	players []gamestate.PlayerAndIdPair
}

func New(config Config, evts []event.GameEvent) *MonitorState {
	s := &MonitorState{
		Config:     config,
		Events:     evts,
		CurrentIdx: -1,
	}
	s.ReplayUpTo(-1)
	return s
}

func (s *MonitorState) Reset() {
	gi, err := gamestate.CreateGameInstance(s.Config.Floors, s.Config.Monsters, s.Config.OpenAt, s.Config.Duration)
	if err != nil {
		panic(err)
	}
	gi.SetSyncMode()
	s.game = gi
}

func (s *MonitorState) ReplayUpTo(idx int) {
	s.Reset()
	for i := 0; i <= idx && i < len(s.Events); i++ {
		s.game.ApplyEventSync(s.Events[i])
	}
	s.rebuildCache()
}

func (s *MonitorState) rebuildCache() {
	s.entries = nil
	deadFlagged := make(map[int]bool)

	for i, ev := range s.Events {
		if i > s.CurrentIdx {
			break
		}
		msg := eventDesc(ev)
		s.entries = append(s.entries, LogEntry{
			Time:     formatting.SecondsToFormattedDate(ev.OccuredAtSecond),
			PlayerID: ev.PlayerId,
			Message:  msg,
			IsActive: i == s.CurrentIdx,
		})

		if ev.EventId == event.PlayerGotDamaged {
			hp := 100
			if p, ok := s.game.GetPlayer(ev.PlayerId); ok {
				hp = p.Hp
			}
			if hp <= 0 && !deadFlagged[ev.PlayerId] {
				deadFlagged[ev.PlayerId] = true
				s.entries = append(s.entries, LogEntry{
					Time:     formatting.SecondsToFormattedDate(ev.OccuredAtSecond),
					PlayerID: ev.PlayerId,
					Message:  "is dead",
					IsAlert:  true,
				})
			}
		}
		if ev.EventId == event.PlayerDisqualified {
			s.entries = append(s.entries, LogEntry{
				Time:     formatting.SecondsToFormattedDate(ev.OccuredAtSecond),
				PlayerID: ev.PlayerId,
				Message:  "disqualified",
				IsAlert:  true,
			})
		}
		if ev.EventId == event.PlayerMakesImposibleMove {
			s.entries = append(s.entries, LogEntry{
				Time:     formatting.SecondsToFormattedDate(ev.OccuredAtSecond),
				PlayerID: ev.PlayerId,
				Message:  "impossible move",
				IsAlert:  true,
			})
		}
	}

	s.players = s.game.CompilePlayerData()

	if s.SelectedPid == 0 && len(s.players) > 0 {
		s.SelectedPid = s.players[0].Id
	}
}

func (s *MonitorState) FloorViews() []FloorView {
	views := make([]FloorView, s.Config.Floors)
	for i := range views {
		views[i].Num = i + 1
		views[i].IsBoss = i == s.Config.Floors-1
	}

	selPlayer, ok := s.game.GetPlayer(s.SelectedPid)
	if ok && len(selPlayer.Floors) > 0 {
		for i, f := range selPlayer.Floors {
			views[i].MonstersLeft = f.MonsterCount
			views[i].BossDefeated = f.BossDefeated
			views[i].Cleared = f.IsCompleted()
		}
		if selPlayer.CurrentFloor < len(views) {
			if selPlayer.ExitedDungeonAtSeconds == 0 || selPlayer.Hp <= 0 {
				views[selPlayer.CurrentFloor].Occupants = append(views[selPlayer.CurrentFloor].Occupants, selPlayer.Id)
			}
		}
	}

	return views
}

func (s *MonitorState) Step(dir int) {
	nxt := s.CurrentIdx + dir
	log.Printf("Step(dir=%d) currentIdx=%d -> nxt=%d", dir, s.CurrentIdx, nxt)
	if nxt < -1 {
		nxt = -1
	}
	if nxt >= len(s.Events) {
		log.Printf("Step: at end (nxt=%d >= %d), stopping play", nxt, len(s.Events))
		if dir > 0 {
			s.IsPlaying = false
		}
		return
	}
	s.CurrentIdx = nxt
	s.ReplayUpTo(nxt)
}

func (s *MonitorState) TogglePlay() {
	s.IsPlaying = !s.IsPlaying
	log.Printf("TogglePlay: %v (idx=%d, events=%d)", s.IsPlaying, s.CurrentIdx, len(s.Events))
	if !s.IsPlaying {
		s.PlayElapsed = 0
	}
	if s.IsPlaying && s.CurrentIdx >= len(s.Events)-1 {
		log.Printf("TogglePlay: resetting to start")
		s.CurrentIdx = -1
		s.ReplayUpTo(-1)
	}
}

func (s *MonitorState) SelectPlayer(id int) {
	s.SelectedPid = id
}

func (s *MonitorState) DungeonClosed() bool {
	return len(s.Events) > 0 && s.CurrentIdx >= len(s.Events)-1
}

func (s *MonitorState) PlayersSorted() []gamestate.PlayerAndIdPair {
	return s.players
}

func (s *MonitorState) LogEntries() []LogEntry {
	return s.entries
}

func (s *MonitorState) GetPlayer(id int) (*gamestate.Player, bool) {
	return s.game.GetPlayer(id)
}

func eventDesc(ev event.GameEvent) string {
	str, ok := event.EventTypeToFormatString[ev.EventId]
	if !ok {
		return "unknown"
	}
	raw := strings.Split(fmt.Sprintf(str+"\n", ev.PlayerId, ev.ExtraParam), "\n")[0]
	formatted := formatting.RemoveExtraSprintfParam(raw)
	parts := strings.SplitN(formatted, " ", 3)
	if len(parts) >= 3 {
		return parts[2]
	}
	if len(parts) >= 2 {
		return parts[1]
	}
	return formatted
}
