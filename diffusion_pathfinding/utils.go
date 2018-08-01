package main

import (
	"github.com/veandco/go-sdl2/sdl"
)

func ScreenSpaceToGRIDSpace(p Vec2D) (x int, y int) {
	return int(GRID_WORLD_DIMENSION * (p.X / float64(WINDOW_WIDTH))),
		int(GRID_WORLD_DIMENSION * (1.0 - p.Y/float64(WINDOW_HEIGHT)))
}

func GRIDSpaceToScreenSpace(p Vec2D) (x int, y int) {
	return int(WINDOW_WIDTH * (p.X / float64(GRID_WORLD_DIMENSION))),
		int(WINDOW_HEIGHT * (1.0 - p.Y/float64(GRID_WORLD_DIMENSION)))
}

func MouseButtonEventToVec2D(me *sdl.MouseButtonEvent) Vec2D {
	x, y := ScreenSpaceToGRIDSpace(Vec2D{float64(me.X), float64(me.Y)})
	return Vec2D{float64(x), float64(y)}
}

func MouseMotionEventToVec2D(me *sdl.MouseMotionEvent) Vec2D {
	x, y := ScreenSpaceToGRIDSpace(Vec2D{float64(me.X), float64(me.Y)})
	return Vec2D{float64(x), float64(y)}
}
