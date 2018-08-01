package main

import (
	"flag"
	"fmt"
	"github.com/veandco/go-sdl2/sdl"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"sync"
	"time"
)

var cpuprofile = flag.String("cpuprofile", "", "if provided, use as filename of prof output")

func init() {
	flag.Parse()
	rand.Seed(time.Now().UnixNano())
}

type KeyEventHandler struct {
	once  map[int]*sync.Once
	mutex map[int]*sync.Mutex
}

func handleQuit(e sdl.Event) bool {
	switch e.(type) {
	case *sdl.QuitEvent:
		return false
	case *sdl.KeyboardEvent:
		ke := e.(*sdl.KeyboardEvent)
		if ke.Keysym.Sym == sdl.K_ESCAPE ||
			ke.Keysym.Sym == sdl.K_q {
			return false
		}
	}
	return true
}

func doRegen(w *World, keh KeyEventHandler) {
	if _, ok := keh.once[sdl.K_g]; !ok {
		keh.once[sdl.K_g] = &sync.Once{}
		keh.mutex[sdl.K_g] = &sync.Mutex{}
	}
	keh.mutex[sdl.K_g].Lock()
	go func() {
		keh.once[sdl.K_g].Do(func() {
			w.mapMutex.Lock()
			w.RegenMap()
			ms := w.ComputePath()
			w.mapMutex.Unlock()
			fmt.Printf("path calculation took %.3f ms\n", ms)
			keh.mutex[sdl.K_g].Lock()
			time.Sleep(time.Second)
			keh.once[sdl.K_g] = &sync.Once{}
			keh.mutex[sdl.K_g].Unlock()
		})
	}()
	keh.mutex[sdl.K_g].Unlock()
}

func handleKeyEvents(w *World, e sdl.Event, keh KeyEventHandler) {
	switch e.(type) {
	case *sdl.KeyboardEvent:
		ke := e.(*sdl.KeyboardEvent)
		if ke.Keysym.Sym == sdl.K_p && ke.Type == sdl.KEYDOWN {
			w.param++
			fmt.Printf("Param: %d\n", w.param)
			doRegen(w, keh)
		}
		if ke.Keysym.Sym == sdl.K_o && ke.Type == sdl.KEYDOWN {
			w.param--
			if w.param < 0 {
				w.param = 0
			}
			fmt.Printf("Param: %d\n", w.param)
			doRegen(w, keh)
		}
		if ke.Keysym.Sym == sdl.K_g && ke.Type == sdl.KEYDOWN {
			w.param = 0
			doRegen(w, keh)
		}
	}
}

func handleMouseEvents(w *World, e sdl.Event) {
	switch e.(type) {
	case *sdl.MouseButtonEvent:
		me := e.(*sdl.MouseButtonEvent)
		if me.Type == sdl.MOUSEBUTTONDOWN {
			pos := screenSpaceToWorldSpace(Point2D{int(me.X), int(me.Y)})
			if me.Button == sdl.BUTTON_LEFT {
				e := Entity{pos: pos}
				w.e = &e
				for i, lake := range w.m.Lakes {
					if lake.containsPoint2D(e.pos) {
						for _, lake := range w.m.Lakes {
							lake.highlighted = false
						}
						lake.highlighted = true
						fmt.Printf("entity is inside lake %d\n", i)
						break
					}
				}
			}
			if me.Button == sdl.BUTTON_RIGHT {
				if w.e != nil {
					w.e.moveTarget = &pos
					ms := w.ComputePath()
					fmt.Printf("path calculation took %.3f ms\n", ms)
				}
			}
		}
	}
}

func gameloop() int {

	sdl.Init(sdl.INIT_EVERYTHING)
	r, exitcode := GetRenderer()
	if exitcode != 0 {
		return exitcode
	}

	w := NewWorld(r)

	moveTicker := time.NewTicker(10 * time.Millisecond)
	fpsTicker := time.NewTicker(time.Millisecond * (1000 / FPS))
	keh := KeyEventHandler{
		make(map[int]*sync.Once),
		make(map[int]*sync.Mutex)}
	running := true
	go func() {
		for {
			select {
			case _ = <-fpsTicker.C:
				sdl.Do(func() {
					r.Clear()

					w.mapMutex.Lock()
					w.DrawWorldMap(r)
					w.mapMutex.Unlock()

					w.entityMutex.Lock()
					w.DrawEntityAndPath(r)
					w.entityMutex.Unlock()

					r.Present()
				})
			}
		}
	}()

	for running {
		for e := sdl.PollEvent(); e != nil; e = sdl.PollEvent() {
			running = handleQuit(e)
			handleKeyEvents(w, e, keh)
			handleMouseEvents(w, e)
		}

		select {
		case _ = <-moveTicker.C:
			w.entityMutex.Lock()
			w.MoveEntity()
			w.entityMutex.Unlock()
		default:
		}
		sdl.Delay(1000 / (2 * FPS))
	}
	fmt.Println("Done")
	return 0
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
		exitcode = gameloop()
	})
	os.Exit(exitcode)

}
