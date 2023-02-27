package main

import (
	"fmt"

	"github.com/dt-rush/sameriver/v3"
)

func main() {
	fmt.Println("vim-go")
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
