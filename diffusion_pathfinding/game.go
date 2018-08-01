package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"time"
)

type Game struct {
	w  *World
	c  *Controls
	ui *UI

	paused   bool
	showData bool

	r *sdl.Renderer
	f *ttf.Font
}

func NewGame(r *sdl.Renderer, f *ttf.Font) *Game {
	g := &Game{r: r, f: f, showData: true}

	g.w = NewWorld(g)
	g.c = NewControls()
	g.ui = NewUI(r, f)

	g.ui.UpdateMsg(0, "i: show data, m: toggle mode, p: pause")
	g.ui.UpdateMsg(1, "g: place random obstacles, c: clear obstacles")
	g.ui.UpdateMsg(2, fmt.Sprintf("grid dimension: %d", GRID_CELL_DIMENSION))
	g.ui.UpdateMsg(3, MODENAMES[g.c.mode])
	return g
}

func (g *Game) HandleKeyEvents(e sdl.Event) {
	switch e.(type) {
	case *sdl.KeyboardEvent:
		ke := e.(*sdl.KeyboardEvent)
		if ke.Type == sdl.KEYDOWN {
			if ke.Keysym.Sym == sdl.K_m {
				g.c.ToggleMode()
				g.ui.UpdateMsg(3, MODENAMES[g.c.mode])
			}
			if ke.Keysym.Sym == sdl.K_c {
				g.w.ClearObstacles()
			}
			if ke.Keysym.Sym == sdl.K_g {
				g.w.RandomObstacles()
			}
			if ke.Keysym.Sym == sdl.K_i {
				g.showData = !g.showData
			}
			if ke.Keysym.Sym == sdl.K_p {
				g.paused = !g.paused
			}
		}
	}
}

func (g *Game) HandleMouseMotionEvents(me *sdl.MouseMotionEvent) {
	p := MouseMotionEventToVec2D(me)
	if me.State&sdl.ButtonLMask() != 0 {
		if g.c.mode == MODE_PLACING_WAYPOINT {
			g.HandleWayPointInput(sdl.BUTTON_LEFT, p)
		}
	}
	if me.State&sdl.ButtonRMask() != 0 {
		if g.c.mode == MODE_PLACING_WAYPOINT {
			g.HandleWayPointInput(sdl.BUTTON_RIGHT, p)
		} else if g.c.mode == MODE_PLACING_OBSTACLE {
			g.HandleObstacleInput(sdl.BUTTON_LEFT, p)
		}
	}
}

func (g *Game) HandleMouseButtonEvents(me *sdl.MouseButtonEvent) {
	p := MouseButtonEventToVec2D(me)
	if me.Type == sdl.MOUSEBUTTONDOWN {
		if g.c.mode == MODE_PLACING_WAYPOINT {
			g.HandleWayPointInput(me.Button, p)
		} else if g.c.mode == MODE_PLACING_OBSTACLE {
			g.HandleObstacleInput(me.Button, p)
		} else if g.c.mode == MODE_PLACING_CHASER {
			g.HandleChaserInput(me.Button, p)
		}
	}
}

func (g *Game) HandleChaserInput(button uint8, p Vec2D) {
	pos := Position{
		int(p.X / GRIDCELL_WORLD_W),
		int(p.Y / GRIDCELL_WORLD_H),
	}
	if !g.w.dm.InGrid(pos.X, pos.Y) ||
		g.w.dm.CellHasObstacle(pos.X, pos.Y) {
		return
	}
	p = Vec2D{
		float64(pos.X*GRIDCELL_WORLD_W + GRIDCELL_WORLD_W/2),
		float64(pos.Y*GRIDCELL_WORLD_H + GRIDCELL_WORLD_H/2)}
	if button == sdl.BUTTON_LEFT {
		g.w.chasers = append(g.w.chasers, NewChaser(p, g.w))
	}
}

func (g *Game) HandleWayPointInput(button uint8, p Vec2D) {
	pos := Position{
		int(p.X / GRIDCELL_WORLD_W),
		int(p.Y / GRIDCELL_WORLD_H),
	}
	if !g.w.dm.InGrid(pos.X, pos.Y) ||
		g.w.dm.CellHasObstacle(pos.X, pos.Y) {
		return
	}
	p = Vec2D{
		float64(pos.X*GRIDCELL_WORLD_W + GRIDCELL_WORLD_W/2),
		float64(pos.Y*GRIDCELL_WORLD_H + GRIDCELL_WORLD_H/2)}
	if button == sdl.BUTTON_LEFT {
		g.w.e = NewEntity(p, g.w)
	}
	if button == sdl.BUTTON_RIGHT {
		if g.w.e != nil {
			g.w.e.moveTarget = &p

			startCell := g.w.dm.ToGridSpace(g.w.e.pos)
			endCell := g.w.dm.ToGridSpace(*g.w.e.moveTarget)
			t0 := time.Now()
			path := g.w.pc.Path(startCell, endCell)
			msg := fmt.Sprintf("path compute took %.3f ms",
				float64(time.Since(t0).Nanoseconds()/1e6)/1000.0)
			g.ui.UpdateMsg(5, msg)
			g.ui.UpdateMsg(6, fmt.Sprintf("path length: %d", len(path)))

			g.w.e.path = g.w.e.path[:0]
			for _, p := range path {
				g.w.e.path = append(g.w.e.path, Vec2D{
					float64(p.X*GRIDCELL_WORLD_W + GRIDCELL_WORLD_W/2),
					float64(p.Y*GRIDCELL_WORLD_H + GRIDCELL_WORLD_H/2),
				})
			}
		}
	}
}

func (g *Game) HandleObstacleInput(button uint8, pos Vec2D) {
	if button == sdl.BUTTON_LEFT {
		pos := Position{
			int(pos.X / GRIDCELL_WORLD_W),
			int(pos.Y / GRIDCELL_WORLD_H),
		}
		if !g.w.dm.InGrid(pos.X, pos.Y) {
			return
		}
		g.w.AddObstacle(pos)
	}
	if button == sdl.BUTTON_RIGHT {

	}
}

func (g *Game) HandleQuit(e sdl.Event) bool {
	switch e.(type) {
	case *sdl.QuitEvent:
		return false
	case *sdl.KeyboardEvent:
		ke := e.(*sdl.KeyboardEvent)
		if ke.Keysym.Sym == sdl.K_ESCAPE ||
			ke.Keysym.Sym == sdl.K_q {
			return true
		}
	}
	return false
}

func (g *Game) HandleEvents() bool {
	for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
		if g.HandleQuit(e) {
			return false
		}
		switch e.(type) {
		case *sdl.KeyboardEvent:
			g.HandleKeyEvents(e)
		case *sdl.MouseMotionEvent:
			g.HandleMouseMotionEvents(e.(*sdl.MouseMotionEvent))
		case *sdl.MouseButtonEvent:
			g.HandleMouseButtonEvents(e.(*sdl.MouseButtonEvent))
		}
	}
	return true
}

func (g *Game) gameloop() int {

	fpsTicker := time.NewTicker(time.Millisecond * (1000 / FPS))

gameloop:
	for {
		// try to draw
		select {
		case _ = <-fpsTicker.C:
			sdl.Do(func() {
				g.r.Clear()
				g.DrawGrid()
				g.DrawUI()
				g.r.Present()
			})
		default:
		}

		// handle input
		if !g.HandleEvents() {
			break gameloop
		}

		// update grid
		if !g.paused {
			g.w.Update()
		}

		sdl.Delay(1000 / (2 * FPS))
	}
	return 0
}
