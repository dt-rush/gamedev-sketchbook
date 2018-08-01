package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

type Rect2D struct {
	X, Y float64
	W, H float64
}

func CenteredSquare(p Vec2D, r float64) Rect2D {
	return Rect2D{p.X - r/2, p.Y - r/2, r, r}
}

func (r Rect2D) ToScreenSpaceSdlRect() sdl.Rect {
	// set the corner to top-left instead of bottom-left
	r.Y += r.H
	x, y := GRIDSpaceToScreenSpace(Vec2D{r.X, r.Y})
	w := WINDOW_WIDTH * (r.W / float64(GRID_WORLD_DIMENSION))
	h := WINDOW_WIDTH * (r.H / float64(GRID_WORLD_DIMENSION))
	return sdl.Rect{int32(x), int32(y), int32(w), int32(h)}
}

func (r Rect2D) Contains(p Vec2D) bool {
	return r.X <= p.X && p.X <= (r.X+r.W) && r.Y <= p.Y && p.Y <= (r.Y+r.H)
}

func (r1 Rect2D) Overlaps(r2 Rect2D) bool {
	return !(r1.X > r2.X+r2.W ||
		r1.X+r1.W < r2.X ||
		r1.Y > r2.Y+r2.H ||
		r1.Y+r1.H < r2.Y)
}

func (r Rect2D) Add(v Vec2D) Rect2D {
	return Rect2D{r.X + v.X, r.Y + v.Y, r.W, r.H}
}
