package main

import (
	"os"
	"runtime/pprof"

	"github.com/dt-rush/sameriver/v3"
)

func main() {
	f, err := os.Create("profile.pprof")
	if err != nil {
		panic(err)
	}
	if err := pprof.StartCPUProfile(f); err != nil {
		panic(err)
	}
	defer pprof.StopCPUProfile()

	sameriver.RunGame(sameriver.GameInitSpec{
		WindowSpec: sameriver.WindowSpec{
			Title:      "grids and circles",
			Width:      800,
			Height:     800,
			Fullscreen: false},
		LoadingScene: &LoadingScene{},
		FirstScene:   &GameScene{},
	})
}
