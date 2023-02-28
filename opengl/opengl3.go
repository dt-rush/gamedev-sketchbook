package main

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"
	"unsafe"

	gl "github.com/chsc/gogl/gl43"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
)

func makeCube() (verts, colors []gl.Float) {
	verts = []gl.Float{
		-1.0, -1.0, -1.0,
		-1.0, -1.0, 1.0,
		-1.0, 1.0, 1.0,

		1.0, 1.0, -1.0,
		-1.0, -1.0, -1.0,
		-1.0, 1.0, -1.0,

		1.0, -1.0, 1.0,
		-1.0, -1.0, -1.0,
		1.0, -1.0, -1.0,

		1.0, 1.0, -1.0,
		1.0, -1.0, -1.0,
		-1.0, -1.0, -1.0,

		-1.0, -1.0, -1.0,
		-1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0,

		1.0, -1.0, 1.0,
		-1.0, -1.0, 1.0,
		-1.0, -1.0, -1.0,

		-1.0, 1.0, 1.0,
		-1.0, -1.0, 1.0,
		1.0, -1.0, 1.0,

		1.0, 1.0, 1.0,
		1.0, -1.0, -1.0,
		1.0, 1.0, -1.0,

		1.0, -1.0, -1.0,
		1.0, 1.0, 1.0,
		1.0, -1.0, 1.0,

		1.0, 1.0, 1.0,
		1.0, 1.0, -1.0,
		-1.0, 1.0, -1.0,

		1.0, 1.0, 1.0,
		-1.0, 1.0, -1.0,
		-1.0, 1.0, 1.0,

		1.0, 1.0, 1.0,
		-1.0, 1.0, 1.0,
		1.0, -1.0, 1.0,
	}
	colors = make([]gl.Float, len(verts))
	for i := 0; i < len(verts); i += 3 {
		y := verts[i+1]
		if y == -1.0 {
			colors[i] = 1.0
			colors[i+1] = 1.0
			colors[i+2] = 1.0
		} else {
			colors[i] = 0.1
			colors[i+1] = 1.0
			colors[i+2] = 0.1
		}
	}
	return verts, colors
}

func logShaderCompileError(s gl.Uint) {
	var logLength gl.Int
	gl.GetShaderiv(s, gl.INFO_LOG_LENGTH, &logLength)

	log := strings.Repeat("\x00", int(logLength+1))
	gl.GetShaderInfoLog(s, gl.Sizei(logLength), nil, gl.GLString(log))
	// in engine this would be logError
	fmt.Printf("failed to compile shader: %v\n", log)
}

func compileShader(kind gl.Enum, src string) gl.Uint {
	s := gl.CreateShader(kind)
	s_source := gl.GLString(src)
	gl.ShaderSource(s, 1, &s_source, nil)
	gl.CompileShader(s)
	var status gl.Int
	gl.GetShaderiv(s, gl.COMPILE_STATUS, &status)
	fmt.Printf("Compiled Vertex Shader: %v\n", status)
	if status == gl.FALSE {
		logShaderCompileError(s)
	}
	return s
}

func createprogram() gl.Uint {
	// CREATE PROGRAM
	program := gl.CreateProgram()

	// shaders
	vs := compileShader(gl.VERTEX_SHADER, vertexShaderSource)
	fs := compileShader(gl.FRAGMENT_SHADER, fragmentShaderSource)

	gl.AttachShader(program, vs)
	gl.AttachShader(program, fs)

	fragoutstring := gl.GLString("outColor")
	defer gl.GLStringFree(fragoutstring)
	gl.BindFragDataLocation(program, gl.Uint(0), fragoutstring)

	gl.LinkProgram(program)
	var linkstatus gl.Int
	gl.GetProgramiv(program, gl.LINK_STATUS, &linkstatus)
	fmt.Printf("Program Link: %v\n", linkstatus)

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)

	return program
}

func glInit() {
	gl.Init()
	gl.Viewport(0, 0, gl.Sizei(winWidth), gl.Sizei(winHeight))
	// OPENGL FLAGS
	gl.ClearColor(0.0, 0.1, 0.0, 1.0)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
}

func SDLInit() (window *sdl.Window, context sdl.GLContext) {
	var err error
	runtime.LockOSThread()
	if err = sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		panic(err)
	}
	window, err = sdl.CreateWindow(winTitle, sdl.WINDOWPOS_UNDEFINED,
		sdl.WINDOWPOS_UNDEFINED,
		winWidth, winHeight, sdl.WINDOW_OPENGL)
	if err != nil {
		panic(err)
	}
	context, err = window.GLCreateContext()
	if err != nil {
		panic(err)
	}

	return window, context
}

func printLiteralMat4(name string, m mgl32.Mat4) {
	for i := 0; i < 4; i++ {
		fmt.Printf("vec4 %s%d = vec4(", name, i)
		for j := 0; j < 4; j++ {
			fmt.Printf("%f", m[i*4+j])
			if j != 3 {
				fmt.Printf(",")
			}
		}
		fmt.Printf(");\n")
	}
	fmt.Printf("mat4 %s = mat4(", name)
	for i := 0; i < 4; i++ {
		fmt.Printf("%s%d", name, i)
		if i != 3 {
			fmt.Printf(",")
		}
	}
	fmt.Printf(");\n")
}

func buildModels(positions, rotations, scales []mgl32.Vec3) []mgl32.Mat4 {
	models := make([]mgl32.Mat4, len(positions))
	for i := 0; i < len(positions); i++ {
		t := mgl32.Translate3D(positions[i][0], positions[i][1], positions[i][2])
		r := mgl32.Ident4()
		r = r.Mul4(mgl32.HomogRotate3D(rotations[i][0], mgl32.Vec3{1, 0, 0}))
		r = r.Mul4(mgl32.HomogRotate3D(rotations[i][1], mgl32.Vec3{0, 1, 0}))
		r = r.Mul4(mgl32.HomogRotate3D(rotations[i][2], mgl32.Vec3{0, 0, 1}))
		s := mgl32.Scale3D(scales[i][0], scales[i][1], scales[i][2])
		models[i] = t.Mul4(r.Mul4(s))
	}
	return models
}

const CUBE_DIM = 7
const N_CUBES = CUBE_DIM * CUBE_DIM

func main() {
	window, context := SDLInit()
	defer sdl.Quit()
	defer window.Destroy()
	defer sdl.GLDeleteContext(context)

	glInit()

	verts, colors := makeCube()

	positions := make([]mgl32.Vec3, N_CUBES)
	rotations := make([]mgl32.Vec3, N_CUBES)
	scales := make([]mgl32.Vec3, N_CUBES)
	for i := 0; i < CUBE_DIM; i++ {
		for j := 0; j < CUBE_DIM; j++ {
			ix := CUBE_DIM*i + j
			positions[ix] = mgl32.Vec3{
				float32(-CUBE_DIM/2 + i),
				0,
				float32(-CUBE_DIM/2 + j)}
			rotations[ix] = mgl32.Vec3{0, 0, 0}
			s := float32(0.333)
			scales[ix] = mgl32.Vec3{s, s, s}
		}
	}

	var f gl.Float
	floatSz := int(unsafe.Sizeof(f))

	// VERTEX BUFFER
	var vertexBuffer gl.Uint
	gl.GenBuffers(1, &vertexBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(verts)*floatSz), gl.Pointer(&verts[0]), gl.STATIC_DRAW)

	// INSTANCE MODEL BUFFER
	var modelBuffer gl.Uint
	gl.GenBuffers(1, &modelBuffer)

	// COLOUR BUFFER
	var colourbuffer gl.Uint
	gl.GenBuffers(1, &colourbuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, colourbuffer)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(colors)*floatSz), gl.Pointer(&colors[0]), gl.STATIC_DRAW)

	// GUESS WHAT
	program := createprogram()

	// VERTEX ARRAY
	var VertexArrayID gl.Uint
	gl.GenVertexArrays(1, &VertexArrayID)
	gl.BindVertexArray(VertexArrayID)
	vertLoc := gl.Uint(gl.GetAttribLocation(program, gl.GLString("vert")))
	gl.EnableVertexAttribArray(vertLoc)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.VertexAttribPointer(vertLoc, 3, gl.FLOAT, gl.FALSE, 0, nil)

	// VERTEX ARRAY HOOK COLOURS
	colorLoc := gl.Uint(gl.GetAttribLocation(program, gl.GLString("color")))
	gl.EnableVertexAttribArray(colorLoc)
	gl.BindBuffer(gl.ARRAY_BUFFER, colourbuffer)
	gl.VertexAttribPointer(colorLoc, 3, gl.FLOAT, gl.FALSE, 0, nil)

	// VERTEX ARRAY HOOK MODELS
	gl.BindBuffer(gl.ARRAY_BUFFER, modelBuffer)
	modelLoc := gl.Uint(gl.GetAttribLocation(program, gl.GLString("model")))
	loc1 := modelLoc + 0
	loc2 := modelLoc + 1
	loc3 := modelLoc + 2
	loc4 := modelLoc + 3
	gl.EnableVertexAttribArray(loc1)
	gl.EnableVertexAttribArray(loc2)
	gl.EnableVertexAttribArray(loc3)
	gl.EnableVertexAttribArray(loc4)
	gl.VertexAttribPointer(loc1, 4, gl.FLOAT, gl.FALSE, gl.Sizei(floatSz*4*4), gl.Pointer(uintptr(0)))
	gl.VertexAttribPointer(loc2, 4, gl.FLOAT, gl.FALSE, gl.Sizei(floatSz*4*4), gl.Pointer(uintptr(4*floatSz)))
	gl.VertexAttribPointer(loc3, 4, gl.FLOAT, gl.FALSE, gl.Sizei(floatSz*4*4), gl.Pointer(uintptr(8*floatSz)))
	gl.VertexAttribPointer(loc4, 4, gl.FLOAT, gl.FALSE, gl.Sizei(floatSz*4*4), gl.Pointer(uintptr(12*floatSz)))
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.VertexAttribDivisor(loc1, 1)
	gl.VertexAttribDivisor(loc2, 1)
	gl.VertexAttribDivisor(loc3, 1)
	gl.VertexAttribDivisor(loc4, 1)

	// USE PROGRAM
	gl.UseProgram(program)

	// PROJECTION,VIEW UNIFORMS
	uniProjectionString := gl.GLString("projection")
	uniProjection := gl.GetUniformLocation(program, uniProjectionString)
	fmt.Printf("projection uniform location: %d\n", uniProjection)
	uniViewString := gl.GLString("view")
	uniView := gl.GetUniformLocation(program, uniViewString)
	fmt.Printf("view uniform location: %d\n", uniView)

	projection := mgl32.Perspective(mgl32.DegToRad(45.0), float32(winWidth)/winHeight, 0.1, 100.0)
	eye := mgl32.Vec3{0, CUBE_DIM, 2 * CUBE_DIM}
	zoomOut := float32(0.6)
	eyeBack := mgl32.Scale3D(zoomOut, zoomOut, zoomOut).Mul4x1(mgl32.Vec4{eye[0], eye[1], eye[2], 1})
	eye = mgl32.Vec3{eyeBack[0], eyeBack[1], eyeBack[2]}
	target := mgl32.Vec3{0, 0, 0}
	up := mgl32.Vec3{0, 1, 0}
	view := mgl32.LookAtV(eye, target, up)

	fmt.Println()
	printLiteralMat4("projection", projection)
	printLiteralMat4("view", view)
	models := buildModels(positions, rotations, scales)
	printLiteralMat4("model", models[0])

	fmt.Println()
	fmt.Printf("view: %v\n", view)
	fmt.Printf("model[0]: %v\n", models[0])

	gl.UniformMatrix4fv(uniProjection, 1, gl.FALSE, (*gl.Float)(unsafe.Pointer(&projection[0])))
	gl.UniformMatrix4fv(uniView, 1, gl.FALSE, (*gl.Float)(unsafe.Pointer(&view[0])))

	var event sdl.Event
	var running bool
	running = true
	t0 := time.Now()
	var dt_prepare_avg *float32 = nil
	var dt_buffer_avg *float32 = nil
	var dt_draw_avg *float32 = nil
	handleKBEvent := func(ke *sdl.KeyboardEvent) {
		if ke.Type == sdl.KEYDOWN {
			if ke.Keysym.Sym == sdl.K_q {
				running = false
			}
		}
	}
	for running {
		dt_ms := float32(time.Since(t0).Nanoseconds()) / 1e6
		for event = sdl.PollEvent(); event != nil; event =
			sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				// fmt.Printf("[%dms]MouseMotion\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Which, t.X, t.Y, t.XRel, t.YRel)
			case *sdl.KeyboardEvent:
				handleKBEvent(e)
			}
		}

		// modify/prepare models
		for i, m := range models {
			rot := mgl32.HomogRotate3DY(0.05 * float32(i) / float32(len(models)))
			models[i] = rot.Mul4(m)
		}
		gl.BindBuffer(gl.ARRAY_BUFFER, modelBuffer)
		t1 := time.Now()
		for i := 0; i < N_CUBES; i++ {
			x := i % N_CUBES
			rotations[i] = rotations[i].Add(mgl32.Vec3{0, 0.02, 0})
			positions[i] = mgl32.Vec3{positions[i][0], float32(0.2 * math.Sin(float64(2*math.Pi*(float32(x)/6+dt_ms/3000.0)))), positions[i][2]}
		}
		models = buildModels(positions, rotations, scales)
		modelsFlat := make([]float32, N_CUBES*4*4)
		for m := 0; m < N_CUBES; m++ {
			copy(modelsFlat[(m*16):], models[m][:])
		}
		dt_prepare := float32(time.Since(t1).Nanoseconds()) / 1e6
		if dt_prepare_avg == nil {
			dt_prepare_avg = &dt_prepare
		} else {
			*dt_prepare_avg = (*dt_prepare_avg + dt_prepare) / 2.0
		}

		// buffer data
		t2 := time.Now()
		gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(modelsFlat)*floatSz), gl.Pointer(&modelsFlat[0]), gl.STATIC_DRAW)
		dt_buffer := float32(time.Since(t2).Nanoseconds()) / 1e6
		if dt_buffer_avg == nil {
			dt_buffer_avg = &dt_buffer
		} else {
			*dt_buffer_avg = (*dt_buffer_avg + dt_buffer) / 2.0
		}

		// draw
		t3 := time.Now()
		drawgl(verts, colors)
		window.GLSwap()
		dt_draw := float32(time.Since(t3).Nanoseconds()) / 1e6
		if dt_draw_avg == nil {
			dt_draw_avg = &dt_draw
		} else {
			*dt_draw_avg = (*dt_draw_avg + dt_draw) / 2.0
		}
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("avg prepare ms: %f\n", *dt_prepare_avg)
	fmt.Printf("avg buffer ms: %f\n", *dt_buffer_avg)
	fmt.Printf("avg draw ms: %f\n", *dt_draw_avg)
}

func drawgl(verts, colors []gl.Float) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.DrawArraysInstanced(gl.TRIANGLES, gl.Int(0), gl.Sizei(len(verts)*4), N_CUBES)
}

const (
	winTitle           = "OpenGL Shader"
	winWidth           = 640
	winHeight          = 480
	vertexShaderSource = `
#version 330
layout (location = 0) in vec3 vert;
layout(location = 1) in vec3 color;
layout(location = 2) in mat4 model;
uniform mat4 projection;
uniform mat4 view;
out vec3 fragmentColor;
void main()
{
    gl_Position = projection *  view * model * vec4(vert, 1);
    fragmentColor = color;
}
`
	fragmentShaderSource = `

#version 330

#define BYPASS 0
#define PIXEL_SIZE 1.0
out vec4 outColor;
in vec3 fragmentColor;

void main()
{
	float brightness = (fragmentColor.r + fragmentColor.g + fragmentColor.b) / 3;
	if (BYPASS == 0) {

		vec2 xy = gl_FragCoord.xy;
		vec2 pixel = mod(xy/PIXEL_SIZE, 4.0);

		int x = int(pixel.x);
		int y = int(pixel.y);

		bool result = false;
		if (x == 0 && y == 0) result = brightness > 16.0/17.0;
		else if (x == 2 && y == 2) result = brightness > 15.0/17.0;
		else if (x == 2 && y == 0) result = brightness > 14.0/17.0;
		else if (x == 0 && y == 2) result = brightness > 13.0/17.0;
		else if (x == 1 && y == 1) result = brightness > 12.0/17.0;
		else if (x == 3 && y == 3) result = brightness > 11.0/17.0;
		else if (x == 3 && y == 1) result = brightness > 10.0/17.0;
		else if (x == 1 && y == 3) result = brightness > 09.0/17.0;
		else if (x == 1 && y == 0) result = brightness > 08.0/17.0;
		else if (x == 3 && y == 2) result = brightness > 07.0/17.0;
		else if (x == 3 && y == 0) result = brightness > 06.0/17.0;
		else if (x == 0 && y == 1) result =	brightness > 05.0/17.0;
		else if (x == 1 && y == 2) result = brightness > 04.0/17.0;
		else if (x == 2 && y == 3) result = brightness > 03.0/17.0;
		else if (x == 2 && y == 1) result = brightness > 02.0/17.0;
		else if (x == 0 && y == 3) result = brightness > 01.0/17.0;

		vec3 onOff = vec3(result);
		outColor = vec4(mix(onOff, fragmentColor, 0.3), 1.0);
	} else if (BYPASS == 1) {
		outColor = vec4(vec3(brightness), 1.0);
	} else if (BYPASS == 2) {
		outColor = vec4(1.0);
	}
}
`
)
