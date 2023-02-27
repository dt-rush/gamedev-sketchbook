package main

import (
	"github.com/veandco/go-sdl2/sdl"

	"github.com/dt-rush/sameriver/v3"
)

type LoadingScene struct{}

func (s *LoadingScene) Name() string {
	return "loading-scene"
}

func (s *LoadingScene) Init(game *sameriver.Game, config map[string]string) {

}

func (s *LoadingScene) Update(dtMs float64, allowanceMs float64) {
}

func (s *LoadingScene) Draw(window *sdl.Window, renderer *sdl.Renderer) {
}

func (s *LoadingScene) HandleKeyboardState(kb []uint8) {
}
func (s *LoadingScene) HandleKeyboardEvent(ke *sdl.KeyboardEvent) {
}

func (s *LoadingScene) IsDone() bool {
	return false
}

func (s *LoadingScene) NextScene() sameriver.Scene {
	return nil
}

func (s *LoadingScene) End() {
}

func (s *LoadingScene) IsTransient() bool {
	return false
}

func (s *LoadingScene) Destroy() {
}
