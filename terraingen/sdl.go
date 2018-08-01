package main

import (
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"os"
)

func GetRenderer() (*sdl.Renderer, int) {
	var err error
	var window *sdl.Window
	var renderer *sdl.Renderer

	sdl.Do(func() {
		window, err = sdl.CreateWindow(
			"WORLDMAP TEST",
			sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
			WINDOW_WIDTH, WINDOW_HEIGHT, sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL)
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create window: %s\n", err)
		return nil, 1
	}

	sdl.Do(func() {
		renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
	})
	if err != nil {
		fmt.Fprint(os.Stderr, "Failed to create renderer: %s\n", err)
		return nil, 2
	}

	sdl.Do(func() {
		renderer.Clear()
	})

	return renderer, 0
}
