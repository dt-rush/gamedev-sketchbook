package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

func (g *Game) DrawGrid() {
	g.r.SetDrawColor(0, 0, 0, 255)
	g.r.FillRect(nil)
	if g.showData {
		g.DrawDiffusionMap()
	}
	g.DrawObstacles()
	g.DrawEntityAndPath()
	g.DrawChasers()
}

func (g *Game) DrawChasers() {
	for _, c := range g.w.chasers {
		drawPoint(g.r, c.pos, sdl.Color{R: 128, G: 128, B: 255}, POINTSZ)
		if c.moveTarget != nil {
			drawPoint(g.r, *c.moveTarget, sdl.Color{R: 128, G: 128, B: 255}, POINTSZ/2)
		}
	}
}

func (g *Game) DrawEntityAndPath() {

	if g.w.e != nil {

		drawPoint(g.r, g.w.e.pos,
			sdl.Color{R: 0, G: 255, B: 0}, ENTITYSZ)

		if g.w.e.moveTarget != nil {

			drawPoint(g.r, *g.w.e.moveTarget,
				sdl.Color{R: 255, G: 255, B: 255}, POINTSZ)

			if g.showData {
				toward := g.w.e.moveTarget.Sub(g.w.e.pos).Unit().Scale(VECLENGTH)
				drawVector(g.r, g.w.e.pos, toward,
					sdl.Color{R: 255, G: 255, B: 0})
				drawVector(g.r, g.w.e.pos, g.w.e.vel.Scale(VECLENGTH),
					sdl.Color{R: 255, G: 0, B: 0})
				drawVector(g.r, g.w.e.pos, g.w.e.steer.Scale(VECLENGTH),
					sdl.Color{R: 255, G: 0, B: 255})

				for _, p := range g.w.e.path {
					drawPoint(g.r, p, sdl.Color{R: 255, G: 255, B: 255}, POINTSZ/2)
				}

			}

		}

	}
}

func (g *Game) DrawObstacles() {
	for _, o := range g.w.obstacles {
		oc := Vec2D{
			float64(GRIDCELL_WORLD_W*o.X + GRIDCELL_WORLD_W/2),
			float64(GRIDCELL_WORLD_H*o.Y + GRIDCELL_WORLD_H/2)}
		orec := Rect2D{
			float64(GRIDCELL_WORLD_W * o.X),
			float64(GRIDCELL_WORLD_H * o.Y),
			GRIDCELL_WORLD_W,
			GRIDCELL_WORLD_H}
		bodyColor := sdl.Color{R: 255, G: 0, B: 0}
		pointColor := sdl.Color{R: 0, G: 0, B: 0}
		if g.w.e != nil && g.w.e.moveTarget != nil {
			toward := g.w.e.moveTarget.Sub(g.w.e.pos)
			d := toward.Magnitude()
			ovec := oc.Sub(g.w.e.pos)
			oIsAhead := g.w.e.vel.Project(ovec) > 0
			closeToO := ovec.Magnitude() < GRIDCELL_WORLD_W*1.5
			if d > GRIDCELL_WORLD_W && oIsAhead && closeToO {
				bodyColor = sdl.Color{R: 128, G: 0, B: 0}
			}
		}
		drawRect(g.r, orec, bodyColor)
		drawPoint(g.r, oc, pointColor, 3)
	}
}

func (g *Game) DrawDiffusionMap() {
	g.r.Copy(g.w.dm.st, nil, nil)
}

func (g *Game) DrawUI() {
	g.r.Copy(g.ui.st, nil, nil)
}
