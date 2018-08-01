package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

func drawRect(r *sdl.Renderer, pos *Position, c sdl.Color) {
	px := int32(float64(pos.X) * WORLD_CELL_PIXEL_WIDTH)
	py := int32(float64((WORLD_CELLHEIGHT-1)-pos.Y) * WORLD_CELL_PIXEL_HEIGHT)
	px1 := int32(float64(pos.X+1) * WORLD_CELL_PIXEL_WIDTH)
	py1 := int32(float64((WORLD_CELLHEIGHT-1)-(pos.Y-1)) * WORLD_CELL_PIXEL_HEIGHT)
	r.SetDrawColor(c.R, c.G, c.B, 255)
	r.FillRect(&sdl.Rect{
		px, py,
		px1 - px,
		py1 - py})
}

func (w *World) DrawWorldMap(r *sdl.Renderer) {
	for y := 0; y < WORLD_CELLHEIGHT; y++ {
		for x := 0; x < WORLD_CELLWIDTH; x++ {
			pos := Position{x, y}
			drawRect(r, &pos, w.m.cells[y][x].color)
		}
	}
}

func (w *World) DrawEntityAndPath(r *sdl.Renderer) {

	if w.e != nil {
		if w.e.path != nil {
			for _, pos := range w.e.path {
				drawRect(r, &pos, sdl.Color{R: 255, G: 255, B: 255})
			}
		}
		if w.e.moveTarget != nil {
			drawRect(r, w.e.moveTarget, sdl.Color{R: 0, G: 255, B: 255})
		}
		drawRect(r, &w.e.pos, sdl.Color{R: 255, G: 0, B: 0})
	}
}
