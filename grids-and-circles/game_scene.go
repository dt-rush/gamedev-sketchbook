package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"

	"github.com/dt-rush/sameriver/v3"
)

const gridx = 5
const gridy = 5
const radius = 100
const playerRadius = 100
const vel = 0.08

var Logger = log.New(
	os.Stdout,
	"",
	log.Lmicroseconds)

type GameScene struct {
	ended     bool
	destroyed bool
	game      *sameriver.Game

	avgDrawMs *float64

	w         *sameriver.World
	player    *sameriver.Entity
	particles []*sameriver.Entity

	UIFont *ttf.Font

	info            string
	infoSurface     *sdl.Surface
	infoTexture     *sdl.Texture
	infoRect        sdl.Rect
	infoDisplayTick sameriver.TimeAccumulator
}

func (s *GameScene) Init(game *sameriver.Game, config map[string]string) {
	var err error
	// set scene "abstract base class" members
	s.destroyed = false
	s.game = game

	// set up score font
	if s.UIFont, err = ttf.OpenFont("test.ttf", 16); err != nil {
		panic(err)
	}
	s.updateInfoTexture()
	s.infoDisplayTick = sameriver.NewTimeAccumulator(1000)

	// build world and spawn entities
	s.buildWorld()
}

func (s *GameScene) buildWorld() {
	// construct world object
	s.w = sameriver.NewWorld(map[string]any{
		"width":               s.game.WindowSpec.Width,
		"height":              s.game.WindowSpec.Height,
		"distanceHasherGridX": gridx,
		"distanceHasherGridY": gridy,
	})
	ps := sameriver.NewPhysicsSystem()
	s.w.RegisterSystems(ps)
	Logger.Println("spawning player")
	s.player = s.w.Spawn(map[string]any{
		"components": map[string]any{
			"Vec2D,Position": sameriver.Vec2D{s.w.Width / 2, s.w.Height / 2},
			"Vec2D,Box":      sameriver.Vec2D{playerRadius * 2, playerRadius * 2},
			"Vec2D,Velocity": sameriver.Vec2D{0, 0},
			"Vec2D,Mass":     1.0,
		},
	})
	s.particles = make([]*sameriver.Entity, 0)
	for x := 0; x < s.game.WindowSpec.Width; x += 19 {
		for y := 0; y < s.game.WindowSpec.Height; y += 19 {
			particle := s.w.Spawn(map[string]any{
				"components": map[string]any{
					"Vec2D,Position": sameriver.Vec2D{float64(x), float64(y)},
					"Vec2D,Box":      sameriver.Vec2D{3, 3},
				},
			})
			s.particles = append(s.particles, particle)
		}
	}
}

func (s *GameScene) drawRect(
	r *sdl.Renderer, pos, box sameriver.Vec2D, c sdl.Color) {

	pos = pos.ShiftedCenterToBottomLeft(box)
	r.SetDrawColor(c.R, c.G, c.B, c.A)
	s.game.Screen.FillRect(r, &pos, &box)
}

func (s *GameScene) updateInfoTexture() {
	if s.infoSurface != nil {
		s.infoSurface.Free()
	}
	if s.infoTexture != nil {
		s.infoTexture.Destroy()
	}
	// render message ("press space") surface
	var err error
	s.infoSurface, err = s.UIFont.RenderUTF8Solid(
		fmt.Sprintf("%s ", s.info),
		sdl.Color{255, 255, 255, 255})
	if err != nil {
		panic(err)
	}
	// create the texture
	s.infoTexture, err = s.game.Renderer.CreateTextureFromSurface(s.infoSurface)
	if err != nil {
		panic(err)
	}
	// set the width of the texture on screen
	w, h, err := s.UIFont.SizeUTF8(s.info)
	if err != nil {
		panic(err)
	}
	s.infoRect = sdl.Rect{10, 10, int32(w), int32(h)}
}

func (s *GameScene) Name() string {
	return "game-scene"
}

func (s *GameScene) Update(dt_ms float64, allowance_ms float64) {
	s.w.Update(allowance_ms)
	if s.infoDisplayTick.Tick(dt_ms) {
		s.updateInfoTexture()
	}
}

func (s *GameScene) drawCell(r *sdl.Renderer, cell [2]int, color sdl.Color) {
	x := float64(1+cell[0])*s.w.SpatialHasher.CellSizeX - s.w.SpatialHasher.CellSizeX/2
	y := float64(1+cell[1])*s.w.SpatialHasher.CellSizeY - s.w.SpatialHasher.CellSizeY/2
	pos := sameriver.Vec2D{x, y}
	box := sameriver.Vec2D{s.w.SpatialHasher.CellSizeX, s.w.SpatialHasher.CellSizeY}
	bottomLeft := pos.ShiftedCenterToBottomLeft(box)
	cellRect := s.game.Screen.ScreenSpaceRect(&bottomLeft, &box)
	// draw top
	gfx.ThickLineColor(r,
		int32(x-s.w.SpatialHasher.CellSizeX/2),
		int32(cellRect.Y),
		int32(x+s.w.SpatialHasher.CellSizeX/2),
		int32(cellRect.Y),
		3,
		color)
	// draw left
	gfx.ThickLineColor(r,
		int32(x-s.w.SpatialHasher.CellSizeX/2),
		int32(cellRect.Y),
		int32(x-s.w.SpatialHasher.CellSizeX/2),
		int32(cellRect.Y+int32(s.w.SpatialHasher.CellSizeY)),
		3,
		color)
	// draw right
	gfx.ThickLineColor(r,
		int32(x+s.w.SpatialHasher.CellSizeX/2),
		int32(cellRect.Y),
		int32(x+s.w.SpatialHasher.CellSizeX/2),
		int32(cellRect.Y+int32(s.w.SpatialHasher.CellSizeY)),
		3,
		color)
	// draw bottom
	gfx.ThickLineColor(r,
		int32(x-s.w.SpatialHasher.CellSizeX/2),
		int32(cellRect.Y+int32(s.w.SpatialHasher.CellSizeY)),
		int32(x+s.w.SpatialHasher.CellSizeX/2),
		int32(cellRect.Y+int32(s.w.SpatialHasher.CellSizeY)),
		3,
		color)
}

func (s *GameScene) Draw(w *sdl.Window, r *sdl.Renderer) {
	t0 := time.Now()
	defer func() {
		dtMs := float64(time.Since(t0).Nanoseconds()) / 1.0e6
		if s.avgDrawMs == nil {
			s.avgDrawMs = &dtMs
		} else {
			*s.avgDrawMs = (*s.avgDrawMs + dtMs) / 2.0
		}
	}()
	// draw the info
	r.Copy(s.infoTexture, nil, &s.infoRect)
	// draw the player
	playerPos := *s.player.GetVec2D("Position")
	playerBox := *s.player.GetVec2D("Box")

	// draw particles
	for _, p := range s.particles {
		s.drawRect(r,
			*p.GetVec2D("Position"),
			*p.GetVec2D("Box"),
			sdl.Color{255, 255, 255, 90})
	}
	// draw particles in distance
	for _, p := range s.w.EntitiesWithinDistance(playerPos, playerBox, radius) {
		s.drawRect(r,
			*p.GetVec2D("Position"),
			*p.GetVec2D("Box"),
			sdl.Color{0, 255, 0, 200})
	}

	// draw cells
	cells := s.w.CellsWithinDistanceApprox(playerPos, playerBox, radius)
	for _, cell := range cells {
		s.drawCell(r, cell, sdl.Color{255, 200, 0, 70})
	}
	cells = s.w.CellsWithinDistance(playerPos, playerBox, radius)
	for _, cell := range cells {
		s.drawCell(r, cell, sdl.Color{0, 255, 255, 200})

	}

	// draw gridlines
	r.SetDrawColor(255, 255, 255, 255)
	for x := 0; x < s.w.SpatialHasher.GridX; x++ {
		gfx.ThickLineColor(r,
			int32(float64(x)*s.w.SpatialHasher.CellSizeX),
			int32(0),
			int32(float64(x)*s.w.SpatialHasher.CellSizeX),
			int32(s.game.WindowSpec.Height),
			2,
			sdl.Color{255, 255, 255, 100},
		)
	}
	for y := 0; y < s.w.SpatialHasher.GridY; y++ {
		gfx.ThickLineColor(r,
			int32(0),
			int32(float64(y)*s.w.SpatialHasher.CellSizeY),
			int32(s.game.WindowSpec.Width),
			int32(float64(y)*s.w.SpatialHasher.CellSizeY),
			2,
			sdl.Color{255, 255, 255, 100},
		)
	}
	s.drawRect(r,
		playerPos,
		playerBox,
		sdl.Color{255, 255, 255, 255})

	// TODO: this doesn't change properly when the player resizes their box
	s.drawRect(r,
		playerPos.Add(sameriver.Vec2D{radius + playerRadius, 0}),
		sameriver.Vec2D{5, 5},
		sdl.Color{255, 0, 0, 255})
	s.drawRect(r,
		playerPos.Add(sameriver.Vec2D{-radius - playerRadius, 0}),
		sameriver.Vec2D{5, 5},
		sdl.Color{255, 0, 0, 255})
	s.drawRect(r,
		playerPos.Add(sameriver.Vec2D{0, radius + playerRadius}),
		sameriver.Vec2D{5, 5},
		sdl.Color{255, 0, 0, 255})
	s.drawRect(r,
		playerPos.Add(sameriver.Vec2D{0, -radius - playerRadius}),
		sameriver.Vec2D{5, 5},
		sdl.Color{255, 0, 0, 255})
}

func (s *GameScene) handleKeyboardState(kb []uint8) {
	v := s.player.GetVec2D("Velocity")
	// get player v1
	v.X = vel * float64(
		int8(kb[sdl.SCANCODE_D]|kb[sdl.SCANCODE_RIGHT])-
			int8(kb[sdl.SCANCODE_A]|kb[sdl.SCANCODE_LEFT]))
	v.Y = vel * float64(
		int8(kb[sdl.SCANCODE_W]|kb[sdl.SCANCODE_UP])-
			int8(kb[sdl.SCANCODE_S]|kb[sdl.SCANCODE_DOWN]))
	box := s.player.GetVec2D("Box")
	box.X = box.X + 5*vel*float64(
		int8(kb[sdl.SCANCODE_R])-
			int8(kb[sdl.SCANCODE_F]))
	box.Y = box.Y + 5*vel*float64(
		int8(kb[sdl.SCANCODE_R])-
			int8(kb[sdl.SCANCODE_F]))
}

func (s *GameScene) HandleKeyboardState(kb []uint8) {
	s.handleKeyboardState(kb)
}

func (s *GameScene) HandleKeyboardEvent(ke *sdl.KeyboardEvent) {
	if ke.Type == sdl.KEYDOWN {
		if ke.Keysym.Sym == sdl.K_SPACE {
			fmt.Println("you pressed space")
		}
	}
}

func (s *GameScene) IsDone() bool {
	return s.ended
}

func (s *GameScene) NextScene() sameriver.Scene {
	return nil
}

func (s *GameScene) End() {
	fmt.Println(s.w.DumpStatsString())
	if s.avgDrawMs != nil {
		nEntities, _ := s.w.NumEntities()
		Logger.Printf("%d entities", nEntities)
		Logger.Printf("Avg draw time ms: %f", *s.avgDrawMs)
	}
}

func (s *GameScene) IsTransient() bool {
	return true
}

func (s *GameScene) Destroy() {
	if !s.destroyed {
		s.destroyed = true
		sdl.Do(func() {
			s.UIFont.Close()
			s.infoSurface.Free()
			s.infoTexture.Destroy()
		})
	}
}
