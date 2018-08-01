package main

import (
	"flag"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "if provided, use as filename of prof output")

func init() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
}

func InitSDL() (*sdl.Renderer, *ttf.Font) {
	// init SDL
	sdl.Init(sdl.INIT_EVERYTHING)
	// init SDL TTF
	err := ttf.Init()
	if err != nil {
		panic(err)
	}
	r, rendererError := GetRenderer()
	if rendererError != 0 {
		log.Fatalf("failed to build renderer (reason %d)\n", rendererError)
	}
	f, err := GetFont()
	if err != nil {
		log.Fatalf("couldn't load font: %v\n", err)
	}
	return r, f
}

func main() {

	var exitcode int
	sdl.Main(func() {
		if *cpuprofile != "" {
			f, err := os.Create(*cpuprofile)
			if err != nil {
				log.Fatal("could not create CPU profile: ", err)
			}
			if err := pprof.StartCPUProfile(f); err != nil {
				log.Fatal("could not start CPU profile: ", err)
			}
			defer pprof.StopCPUProfile()
		}
		r, f := InitSDL()
		g := NewGame(r, f)
		exitcode = g.gameloop()
	})
	os.Exit(exitcode)

}
