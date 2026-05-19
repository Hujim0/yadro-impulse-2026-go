package monitor

import (
	"dungeonGameLib/lib/engine/event"
	"dungeonGameLib/lib/engine/gamestate"
	"fmt"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type clickable struct {
	bounds rl.Rectangle
	label  string
	action func()
}

type inputHandler struct {
	buttons     []clickable
	progressBar rl.Rectangle
	state       *MonitorState
}

func Run(config Config, evts []event.GameEvent) {
	state := New(config, evts)

	rl.InitWindow(100, 100, "Dungeon Monitor")
	mon := rl.GetCurrentMonitor()
	windowW = int32(float32(rl.GetMonitorWidth(mon)) * 0.88)
	windowH = int32(float32(rl.GetMonitorHeight(mon)) * 0.84)
	rl.SetWindowSize(int(windowW), int(windowH))
	rl.SetWindowPosition(
		(rl.GetMonitorWidth(mon)-int(windowW))/2,
		(rl.GetMonitorHeight(mon)-int(windowH))/2,
	)
	defer rl.CloseWindow()

	loadEmojiTextures()

	rl.SetTargetFPS(60)

	playAccum := float32(0)

	for !rl.WindowShouldClose() {
		dt := rl.GetFrameTime()

		if state.IsPlaying {
			playAccum += dt
			for playAccum >= 0.8 {
				playAccum -= 0.8
				state.Step(1)
			}
		}

		ih := buildInputState(state)
		ih.handleInput()

		rl.BeginDrawing()
		rl.ClearBackground(colBgDark)

		drawHeader(state, ih)
		drawDungeonView(state)
		drawSidebar(state)

		rl.EndDrawing()
	}
}

func buildInputState(state *MonitorState) inputHandler {
	var ih inputHandler
	ih.state = state

	btnY := float32(10)
	btnH := float32(30)
	btnW := float32(40)

	prevX := float32(windowW-sideW) - 320
	playX := prevX + btnW + 8
	nextX := playX + btnW + 8

	ih.buttons = []clickable{
		{rl.NewRectangle(prevX, btnY, btnW, btnH), "<", func() { state.Step(-1) }},
		{rl.NewRectangle(playX, btnY, btnW, btnH), playLabel(state), func() { state.TogglePlay() }},
		{rl.NewRectangle(nextX, btnY, btnW, btnH), ">", func() { state.Step(1) }},
	}

	barX := nextX + btnW + 15
	ih.progressBar = rl.NewRectangle(barX, btnY+10, 200, 8)

	return ih
}

func playLabel(state *MonitorState) string {
	if state.IsPlaying {
		return "||"
	}
	return ">"
}

func (ih *inputHandler) handleInput() {
	if rl.IsMouseButtonPressed(rl.MouseLeftButton) {
		mpos := rl.GetMousePosition()

		for _, b := range ih.buttons {
			if rl.CheckCollisionPointRec(mpos, b.bounds) {
				b.action()
				return
			}
		}

		if rl.CheckCollisionPointRec(mpos, ih.progressBar) {
			relX := mpos.X - ih.progressBar.X
			pct := relX / ih.progressBar.Width
			total := len(ih.state.Events)
			idx := int(pct * float32(total))
			if idx < 0 {
				idx = 0
			}
			if idx >= total {
				idx = total - 1
			}
			ih.state.Step(idx - ih.state.CurrentIdx)
			return
		}

		cardX := float32(sideViewX() + 8)
		for i, pair := range ih.state.PlayersSorted() {
			cardY := float32(headerH + 38 + int32(i)*56)
			if mpos.X >= cardX && mpos.X < cardX+float32(sideW-16) &&
				mpos.Y >= cardY && mpos.Y < cardY+50 {
				ih.state.SelectPlayer(pair.Id)
				return
			}
		}
	}

	if rl.IsKeyPressed(rl.KeyLeft) {
		ih.state.Step(-1)
	}
	if rl.IsKeyPressed(rl.KeyRight) || rl.IsKeyPressed(rl.KeySpace) {
		ih.state.Step(1)
	}
	if rl.IsKeyPressed(rl.KeyP) {
		ih.state.TogglePlay()
	}
}

func sideViewX() int32 {
	return windowW - sideW
}

func drawHeader(state *MonitorState, ih inputHandler) {
	rl.DrawRectangle(0, 0, windowW, headerH, colBgUI)
	rl.DrawLine(0, headerH, windowW, headerH, colBorder)

	rl.DrawText("Dungeon Monitor", 20, 15, 16, colAccent)

	for _, b := range ih.buttons {
		isPlay := b.label == ">" || b.label == "||"
		bg := colBgElement
		tc := colTextMain
		if isPlay {
			bg = colAccent
			tc = colBgDark
		}
		rl.DrawRectangleRec(b.bounds, bg)
		rl.DrawRectangleLinesEx(b.bounds, 1, colTextMuted)
		ts := rl.MeasureText(b.label, 14)
		lx := int32(b.bounds.X) + (int32(b.bounds.Width)-ts)/2
		ly := int32(b.bounds.Y) + (int32(b.bounds.Height)-14)/2
		rl.DrawText(b.label, lx, ly, 14, tc)
	}

	rl.DrawRectangleRec(ih.progressBar, colBgDark)
	rl.DrawRectangleLinesEx(ih.progressBar, 1, colTextMuted)
	total := len(state.Events)
	if total > 0 {
		pct := float32(state.CurrentIdx+1) / float32(total)
		fill := rl.NewRectangle(ih.progressBar.X+1, ih.progressBar.Y+1, (ih.progressBar.Width-2)*pct, ih.progressBar.Height-2)
		rl.DrawRectangleRec(fill, colAccent)
	}

	counterStr := fmt.Sprintf("%d/%d", state.CurrentIdx+1, total)
	counterX := int32(ih.progressBar.X+ih.progressBar.Width) + 12
	rl.DrawText(counterStr, counterX, int32(ih.progressBar.Y)-2, 11, colTextMuted)
}

func drawDungeonView(state *MonitorState) {
	viewW := windowW - sideW
	rl.DrawRectangle(0, headerH, viewW, windowH-headerH, colBgDark)

	views := state.FloorViews()
	yOff := headerH + 20
	floorW := viewW - 80
	if floorW > 600 {
		floorW = 600
	}
	floorX := (viewW - floorW) / 2

	for _, fv := range views {
		rl.DrawRectangle(floorX, yOff, floorW, 110, colBgElement)
		borderCol := colTextMuted
		if fv.Cleared {
			borderCol = colAccent
		}
		if fv.IsBoss {
			borderCol = colDanger
		}
		rl.DrawRectangleLines(floorX, yOff, floorW, 110, borderCol)

		title := fmt.Sprintf("Level %d", fv.Num)
		rl.DrawText(title, floorX+12, yOff+8, 14, colTextMain)

		badgeX := floorX + rl.MeasureText(title, 14) + 20
		if fv.IsBoss {
			drawBadge(badgeX, yOff+8, "BOSS", colBgDark, colDanger, colDanger)
			badgeX += 55
		} else {
			drawBadge(badgeX, yOff+8, "FLOOR", colBgDark, colTextMuted, colTextMuted)
			badgeX += 55
		}
		if fv.Cleared {
			drawBadge(badgeX, yOff+8, "CLEARED", colAccent, colBgDark, colAccent)
		}

		gridY := yOff + 36
		gridX := floorX + 12

		entCount := fv.MonstersLeft
		if fv.IsBoss && !fv.BossDefeated {
			entCount = 1
		}
		for i := 0; i < entCount; i++ {
			ew := entityW
			eh := entityH
			label := "M"
			eCol := colDanger
			if fv.IsBoss {
				ew = bossW
				eh = bossH
				label = "B"
			}
			drawEntity(gridX+int32(i)*(ew+10), gridY, ew, eh, label, eCol, colTextMain, false)
		}

		entStart := entCount
		for _, oid := range fv.Occupants {
			ex := gridX + int32(entStart)*(entityW+10)
			p, ok := state.GetPlayer(oid)
			if !ok {
				continue
			}
			dead := p.State == gamestate.FAIL || p.Hp <= 0
			drawEntity(ex, gridY, entityW, entityH, "P", colAccent, colTextMain, dead)
			drawHealthBar(ex, gridY, entityW, p.Hp)
			entStart++
		}

		if entStart == 0 {
			rl.DrawText("Empty", gridX, gridY+15, 12, colTextMuted)
		}

		yOff += 130
	}
}

func drawBadge(x, y int32, text string, bg, fg, border rl.Color) {
	tw := rl.MeasureText(text, 10)
	rl.DrawRectangle(x, y, tw+12, 16, bg)
	rl.DrawRectangleLines(x, y, tw+12, 16, border)
	rl.DrawText(text, x+6, y+3, 10, fg)
}

func drawEntity(x, y, w, h int32, label string, border, text rl.Color, dead bool) {
	bg := colBgDark
	if dead {
		bg = rl.Fade(bg, 0.2)
	}
	rl.DrawRectangle(x, y, w, h, bg)
	rl.DrawRectangleLines(x, y, w, h, border)

	var tex rl.Texture2D
	switch label {
	case "M":
		tex = texMonster
	case "B":
		tex = texBoss
	case "P":
		if dead {
			tex = texDead
		} else {
			tex = texPlayer
		}
	}

	if tex.ID > 0 {
		tint := rl.White
		if dead {
			tint = rl.Fade(rl.White, 0.3)
		}
		pad := float32(8)
		src := rl.NewRectangle(0, 0, float32(tex.Width), float32(tex.Height))
		dst := rl.NewRectangle(float32(x)+pad, float32(y)+pad, float32(w)-pad*2, float32(h)-pad*2)
		rl.DrawTexturePro(tex, src, dst, rl.NewVector2(0, 0), 0, tint)
	} else {
		if dead {
			text = rl.Fade(text, 0.2)
		}
		ts := rl.MeasureText(label, 20)
		rl.DrawText(label, x+(w-ts)/2, y+(h-20)/2, 20, text)
	}
}

func drawHealthBar(x, y, w int32, hp int) {
	barY := y + entityH - 6
	rl.DrawRectangle(x+2, barY, w-4, 4, rl.Black)
	fillW := int32(float32(w-4) * float32(hp) / 100.0)
	fillCol := colAccent
	if hp < 30 {
		fillCol = colDanger
	}
	rl.DrawRectangle(x+2, barY, fillW, 4, fillCol)
}

func drawSidebar(state *MonitorState) {
	sx := sideViewX()
	rl.DrawRectangle(sx, headerH, sideW, windowH-headerH, colBgUI)
	rl.DrawLine(sx, headerH, sx+sideW, headerH, colBorder)

	rl.DrawRectangle(sx, headerH, sideW, 32, colBgElement)
	rl.DrawLine(sx, headerH+32, sx+sideW, headerH+32, colBorder)
	t := "PLAYERS"
	rl.DrawText(t, sx+(sideW-rl.MeasureText(t, 12))/2, headerH+8, 12, colAccent)

	pY := headerH + 42
	for _, pair := range state.PlayersSorted() {
		p := pair.Player
		dead := p.Hp <= 0 || p.State == gamestate.FAIL
		isActive := p.Id == state.SelectedPid
		cardCol := colBgDark
		cborder := colBorder
		if isActive {
			cardCol = colBgElement
			cborder = colAccent
		}
		card := rl.NewRectangle(float32(sx+8), float32(pY-4), float32(sideW-16), 50)
		rl.DrawRectangleRec(card, cardCol)
		rl.DrawRectangleLinesEx(card, 1, cborder)

		ix := sx + 16
		if dead && texDead.ID > 0 {
			src := rl.NewRectangle(0, 0, float32(texDead.Width), float32(texDead.Height))
			dst := rl.NewRectangle(float32(ix+1), float32(pY), 22, 22)
			rl.DrawTexturePro(texDead, src, dst, rl.NewVector2(0, 0), 0, rl.White)
		} else {
			iconStr := fmt.Sprintf("%d", p.Id)
			rl.DrawCircle(ix+12, pY+8, 12, colBgUI)
			rl.DrawCircleLines(ix+12, pY+8, 12, cborder)
			ts := rl.MeasureText(iconStr, 12)
			rl.DrawText(iconStr, ix+12-ts/2, pY+2, 12, colTextMain)
		}

		rl.DrawText(fmt.Sprintf("Player %d", p.Id), ix+30, pY, 13, colTextMain)

		hpCol := colAccent
		if p.Hp < 30 {
			hpCol = colDanger
		}
		hpStr := fmt.Sprintf("HP %d", p.Hp)
		rl.DrawText(hpStr, sx+sideW-rl.MeasureText(hpStr, 12)-8, pY, 12, hpCol)

		st := playerStatusString(p, state)
		rl.DrawText(st, ix+30, pY+18, 11, statusColor(st))

		pY += 56
	}

	logY := pY + 8
	rl.DrawRectangle(sx, logY, sideW, 28, colBgElement)
	rl.DrawLine(sx, logY+28, sx+sideW, logY+28, colBorder)
	t2 := "EVENT LOG"
	rl.DrawText(t2, sx+(sideW-rl.MeasureText(t2, 12))/2, logY+6, 12, colAccent)

	logContentY := logY + 30
	entries := state.LogEntries()
	maxVisible := int((windowH - logContentY - 80) / 20)
	startIdx := 0
	if len(entries) > maxVisible {
		startIdx = len(entries) - maxVisible
	}

	for idx := startIdx; idx < len(entries) && idx-startIdx < maxVisible; idx++ {
		entry := entries[idx]
		ey := logContentY + int32(idx-startIdx)*20

		bgCol := colBgUI
		if entry.IsActive {
			bgCol = colBgFloor
		}
		rl.DrawRectangle(sx+4, ey, sideW-8, 19, bgCol)

		if entry.IsActive {
			rl.DrawRectangle(sx+4, ey, 3, 19, colAccent)
		} else if entry.IsAlert {
			rl.DrawRectangle(sx+4, ey, 3, 19, colDanger)
		}

		rl.DrawText(entry.Time, sx+12, ey+2, 11, colAccent)
		timeW := rl.MeasureText("00:00:00 ", 11)
		pidStr := fmt.Sprintf("P%d:", entry.PlayerID)
		rl.DrawText(pidStr, sx+12+timeW, ey+2, 11, colTextMain)
		pidW := rl.MeasureText("P99: ", 11)
		tc := colTextMain
		if entry.IsAlert {
			tc = colDanger
		}
		rl.DrawText(entry.Message, sx+12+timeW+pidW, ey+2, 11, tc)
	}

	if state.DungeonClosed() {
		reportY := windowH - 160
		rl.DrawRectangle(sx, reportY, sideW, 160, colBgElement)
		rl.DrawLine(sx, reportY, sx+sideW, reportY, colBorder)
		t3 := "FINAL REPORT"
		rl.DrawText(t3, sx+(sideW-rl.MeasureText(t3, 12))/2, reportY+6, 12, colAccent)

		ry := reportY + 28
		for _, pair := range state.PlayersSorted() {
			st := playerStatusString(pair.Player, state)
			rl.DrawText(fmt.Sprintf("[%s]", st), sx+8, ry, 10, statusColor(st))
			bw := rl.MeasureText(fmt.Sprintf("[%s] ", st), 10)
			rl.DrawText(pair.Player.String(), sx+8+bw, ry, 10, colTextMain)
			ry += 16
		}
	}
}

func playerStatusString(p *gamestate.Player, state *MonitorState) string {
	if p.State == gamestate.IN_GAME {
		allCleared := true
		for _, f := range p.Floors {
			if !f.IsCompleted() {
				allCleared = false
				break
			}
		}
		if allCleared {
			return "SUCCESS"
		}
		return "IN DUNGEON"
	}
	return p.State.String()
}

func statusColor(st string) rl.Color {
	switch st {
	case "SUCCESS":
		return colAccent
	case "FAIL":
		return colDanger
	case "DISQUAL":
		return colTextMuted
	case "IN DUNGEON":
		return colTextMain
	case "IDLE", "REGISTERED":
		return colTextMuted
	}
	return colTextMuted
}
