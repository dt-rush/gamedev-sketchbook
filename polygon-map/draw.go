package main

import (
	"github.com/veandco/go-sdl2/gfx"
	"github.com/veandco/go-sdl2/sdl"
)

func drawLake(r *sdl.Renderer, l *Lake) {
	// fill lake
	var c sdl.Color
	if l.highlighted {
		c = sdl.Color{0, 0, 200, 255}
	} else {
		c = sdl.Color{0, 0, 164, 255}
	}
	gfx.FilledPolygonColor(r, l.vx, l.vy, c)
}

func drawLakePoints(r *sdl.Renderer, l *Lake) {
	// draw source of lake
	if DRAW_LAKE_SOURCE {
		r.SetDrawColor(0, 255, 255, 255)
		ssv := worldSpaceToScreenSpace(l.source)
		r.FillRect(&sdl.Rect{int32(ssv.X - 1), int32(ssv.Y - 1), 3, 3})
	}
	// draw vertices of lake
	if DRAW_LAKE_VERTICES {
		for i, v := range l.Vertices {
			if l.highlighted {
				ssv := worldSpaceToScreenSpace(v)
				if l.interpolated[i] {
					r.SetDrawColor(128, 128, 128, 255)
				} else {
					r.SetDrawColor(255, 255, 255, 255)
				}
				r.FillRect(&sdl.Rect{int32(ssv.X - 1), int32(ssv.Y - 1), 3, 3})
			}
		}
	}
}

func (w *World) DrawWorldMap(r *sdl.Renderer) {
	r.SetDrawColor(0, 0, 0, 255)
	r.FillRect(nil)
	for _, l := range w.m.Lakes {
		drawLake(r, l)
	}
	for _, l := range w.m.Lakes {
		if l.highlighted {
			drawLake(r, l)
		}
	}
	for _, l := range w.m.Lakes {
		drawLakePoints(r, l)
	}
	r.Copy(w.m.perlinTexture, nil, nil)
}

func (w *World) drawPerlin(r *sdl.Renderer) {
}

func drawPath(r *sdl.Renderer, p []Point2D) {

}

func drawPoint(r *sdl.Renderer, p *Point2D, c sdl.Color) {
	ssp := worldSpaceToScreenSpace(*p)
	r.SetDrawColor(c.R, c.G, c.B, 255)
	r.FillRect(&sdl.Rect{
		int32(ssp.X - 1),
		int32(ssp.Y - 1),
		3, 3})
}

func (w *World) DrawEntityAndPath(r *sdl.Renderer) {

	if w.e != nil {
		if w.e.path != nil {
			drawPath(r, w.e.path)
		}
		if w.e.moveTarget != nil {
			drawPoint(r, w.e.moveTarget, sdl.Color{R: 0, G: 255, B: 255})
		}
		drawPoint(r, &w.e.pos, sdl.Color{R: 255, G: 0, B: 0})
	}
}

func CreatePerlinTexture(r *sdl.Renderer, perlin [][]float64) *sdl.Texture {
	var rmask uint32
	var gmask uint32
	var bmask uint32
	var amask uint32
	if sdl.BYTEORDER == sdl.LIL_ENDIAN {
		rmask = 0x000000ff
		gmask = 0x0000ff00
		bmask = 0x00ff0000
		amask = 0xff000000
	} else {
		rmask = 0xff000000
		gmask = 0x00ff0000
		bmask = 0x0000ff00
		amask = 0x000000ff
	}
	flags := uint32(0)
	depth := int32(32)
	s, err := sdl.CreateRGBSurface(
		flags,
		WINDOW_WIDTH, WINDOW_HEIGHT,
		depth,
		rmask, gmask, bmask, amask)
	if err != nil {
		panic(err)
	}
	// overlay perlin with opacity
	for y := 0; y < int(WORLD_WIDTH/PSCALE); y++ {
		for x := 0; x < int(WORLD_HEIGHT/PSCALE); x++ {
			pval := uint8(255 * perlin[y][x])
			var color sdl.Color
			if perlin[y][x] < WATER_CUTOFF {
				color = sdl.Color{R: 255, G: 0, B: 0, A: 255}
			} else {
				color = sdl.Color{R: pval, G: pval, B: pval, A: 255}
			}

			ppw := int(WINDOW_WIDTH / (WORLD_WIDTH / PSCALE))
			pph := int(WINDOW_HEIGHT / (WORLD_HEIGHT / PSCALE))
			s.FillRect(&sdl.Rect{
				int32(x * ppw),
				int32(WINDOW_HEIGHT - y*pph),
				int32(ppw),
				int32(pph),
			}, color.Uint32())

		}
	}
	texture, err := r.CreateTextureFromSurface(s)
	if err != nil {
		panic(err)
	}
	texture.SetAlphaMod(0x44)
	return texture
}
