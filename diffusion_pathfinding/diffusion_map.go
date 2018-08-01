package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"time"
)

type DiffusionMap struct {
	// values of the diffusion field
	d [GRID_CELL_DIMENSION][GRID_CELL_DIMENSION]float64
	// ticker used to time updates to the map (since it can be expensive)
	tick *time.Ticker
	// whether an obstacle exists at that point in the field
	os [GRID_CELL_DIMENSION][GRID_CELL_DIMENSION]bool
	// renderer reference
	r *sdl.Renderer
	// screen texture
	st *sdl.Texture
}

func NewDiffusionMap(
	r *sdl.Renderer, obstacles *[]Position, tick time.Duration) *DiffusionMap {
	st, err := r.CreateTexture(
		sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_TARGET,
		WINDOW_WIDTH,
		WINDOW_HEIGHT)
	st.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}

	dm := DiffusionMap{
		tick: time.NewTicker(tick),
		r:    r,
		st:   st}

	for _, o := range *obstacles {
		dm.AddObstacle(o)
	}

	return &dm
}

func (m *DiffusionMap) ClearObstacles() {
	for x := 0; x < GRID_CELL_DIMENSION; x++ {
		for y := 0; y < GRID_CELL_DIMENSION; y++ {
			m.os[x][y] = false
		}
	}
}

func (m *DiffusionMap) AddObstacle(o Position) {
	m.os[o.X][o.Y] = true
}

func (m *DiffusionMap) UpdateTexture() {
	m.r.SetRenderTarget(m.st)
	defer m.r.SetRenderTarget(nil)

	m.r.SetDrawColor(0, 0, 0, 0)
	m.r.Clear()
	for x := 0; x < GRID_CELL_DIMENSION; x++ {
		for y := 0; y < GRID_CELL_DIMENSION; y++ {
			val := uint8(255 * m.d[x][y])
			var c sdl.Color
			if m.CellHasObstacle(x, y) {
				c = sdl.Color{R: val, G: 0, B: 0}
			} else {
				c = sdl.Color{R: val, G: val, B: val}
			}
			drawRect(m.r,
				Rect2D{
					float64(x * GRIDCELL_WORLD_W),
					float64(y * GRIDCELL_WORLD_H),
					GRIDCELL_WORLD_W,
					GRIDCELL_WORLD_H},
				c,
			)
			drawPoint(m.r,
				Vec2D{float64(x*GRIDCELL_WORLD_W + GRIDCELL_WORLD_W/2),
					float64(y*GRIDCELL_WORLD_H + GRIDCELL_WORLD_H/2)},
				sdl.Color{R: 255, G: 255, B: 255}, 1)

		}
	}
}

func (m *DiffusionMap) CellHasObstacle(x int, y int) bool {
	return m.os[x][y]
}

func (m *DiffusionMap) ToGridSpace(p Vec2D) Position {
	x := int(p.X / GRIDCELL_WORLD_W)
	y := int(p.Y / GRIDCELL_WORLD_H)
	if x > GRID_CELL_DIMENSION-1 {
		x = GRID_CELL_DIMENSION - 1
	}
	if x < 0 {
		x = 0
	}
	if y > GRID_CELL_DIMENSION-1 {
		y = GRID_CELL_DIMENSION - 1
	}
	if y < 0 {
		y = 0
	}
	return Position{x, y}
}

func (m *DiffusionMap) ToWorldSpace(p Position) Vec2D {
	return Vec2D{
		float64(p.X*GRIDCELL_WORLD_W + GRIDCELL_WORLD_W/2),
		float64(p.Y*GRIDCELL_WORLD_H + GRIDCELL_WORLD_H/2)}
}

func (m *DiffusionMap) InGrid(x int, y int) bool {
	return x >= 0 && x < GRID_CELL_DIMENSION &&
		y >= 0 && y < GRID_CELL_DIMENSION
}

func (m *DiffusionMap) Diffuse(pos Vec2D) {

	init := m.ToGridSpace(pos)
	m.d[init.X][init.Y] = 1.0

	var avgOfNeighbors = func(x int, y int) float64 {
		neumann := [][2]int{
			[2]int{-1, 0},
			[2]int{1, 0},
			[2]int{0, -1},
			[2]int{0, 1},
		}
		sum := 0.0
		n := 0.0
		for _, neu := range neumann {
			ox := x + neu[0]
			oy := y + neu[1]
			if m.InGrid(ox, oy) {
				sum += m.d[ox][oy]
				n++
			}
		}
		return sum / n
	}

	for x := 0; x < GRID_CELL_DIMENSION; x++ {
		for y := 0; y < GRID_CELL_DIMENSION; y++ {
			if y == init.Y && x == init.X {
				// don't diffuse the init point
				continue
			}
			if m.os[x][y] {
				// obstacles have value zero
				m.d[x][y] = 0.0
				continue
			}
			m.d[x][y] = avgOfNeighbors(x, y)
			m.d[x][y] *= 0.998
		}
	}

	m.UpdateTexture()
}
