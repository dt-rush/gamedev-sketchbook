package main

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

func drawPoint(r *sdl.Renderer, p Vec2D, c sdl.Color, sz int) {
	sspx, sspy := GRIDSpaceToScreenSpace(p)
	r.SetDrawColor(c.R, c.G, c.B, 255)
	r.FillRect(&sdl.Rect{
		int32(sspx - sz/2),
		int32(sspy - sz/2),
		int32(sz), int32(sz)})
}

func drawVector(r *sdl.Renderer, pos Vec2D, v Vec2D, c sdl.Color) {
	// screen-space position
	sspx, sspy := GRIDSpaceToScreenSpace(pos)
	// screen-space vector tip
	ssvtx, ssvty := GRIDSpaceToScreenSpace(pos.Add(v))
	gfx.LineColor(r,
		int32(sspx),
		int32(sspy),
		int32(ssvtx),
		int32(ssvty),
		sdl.Color{R: c.R, G: c.G, B: c.B, A: 255})
}

func drawRect(r *sdl.Renderer, rect Rect2D, c sdl.Color) {
	r.SetDrawColor(c.R, c.G, c.B, 255)
	ssr := rect.ToScreenSpaceSdlRect()
	r.FillRect(&ssr)
}
