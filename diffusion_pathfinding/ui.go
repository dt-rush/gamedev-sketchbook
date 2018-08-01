package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

type UI struct {
	// renderer
	r *sdl.Renderer
	// font
	f *ttf.Font
	// screen texture
	st *sdl.Texture
	// message
	msgs map[int]string
}

func NewUI(r *sdl.Renderer, f *ttf.Font) *UI {
	st, err := r.CreateTexture(
		sdl.PIXELFORMAT_RGBA8888,
		sdl.TEXTUREACCESS_TARGET,
		WINDOW_WIDTH,
		WINDOW_HEIGHT)
	st.SetBlendMode(sdl.BLENDMODE_BLEND)
	if err != nil {
		panic(err)
	}
	return &UI{
		r:    r,
		f:    f,
		st:   st,
		msgs: make(map[int]string),
	}
}

func (ui *UI) UpdateMsg(i int, msg string) {

	ui.msgs[i] = msg

	ui.r.SetRenderTarget(ui.st)
	defer ui.r.SetRenderTarget(nil)

	ui.r.SetDrawColor(0, 0, 0, 0)
	ui.r.Clear()

	ui.renderMsgsToST(sdl.Color{255, 255, 255, 255})
}

// render message to screen texture
func (ui *UI) renderMsgsToST(color sdl.Color) {

	for i, msg := range ui.msgs {
		var surface *sdl.Surface
		var texture *sdl.Texture
		var err error
		surface, err = ui.f.RenderUTF8Blended(msg, color)
		if err != nil {
			panic(err)
		}
		texture, err = ui.r.CreateTextureFromSurface(
			surface)
		if err != nil {
			panic(err)
		}
		// this copies our texture to the *target* texture
		// (screen_texture)
		ui.r.SetDrawColor(
			color.R,
			color.G,
			color.B,
			color.A)
		w, h, err := ui.f.SizeUTF8(msg)
		if err == nil {
			dst := &sdl.Rect{
				10,
				int32(FONTSZ + i*FONTSZ),
				int32(w),
				int32(h)}
			ui.r.Copy(texture, nil, dst)
		}

		// free the resources allocated above
		surface.Free()
		texture.Destroy()
	}

}
