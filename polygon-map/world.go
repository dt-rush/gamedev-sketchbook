package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"sync"
)

type World struct {
	m           *WorldMap
	mapMutex    sync.Mutex
	e           *Entity
	entityMutex sync.Mutex
	c           *PathCalculator
	r           *sdl.Renderer
	param       int
}

func NewWorld(r *sdl.Renderer) *World {
	w := World{}
	w.r = r
	w.param = 0
	w.m = GenerateWorldMap(r)
	w.c = NewPathCalculator(w.m)
	return &w
}

func (w *World) RegenMap() {
	//if w.m != nil { w.m.perlinTexture.Destroy() }
	// w.m = GenerateWorldMap(w.r)
	// fmt.Printf("seed: %d\n", w.m.seed)
	w.m.Regen(w.param)
}
