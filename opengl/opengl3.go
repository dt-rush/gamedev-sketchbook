package main

import (
	"fmt"
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
			colors[i+1] = 0.1
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

func main() {
	window, context := SDLInit()
	defer sdl.Quit()
	defer window.Destroy()
	defer sdl.GLDeleteContext(context)

	glInit()

	verts, colors := makeCube()
	models := make([]mgl32.Mat4, 3*3)
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			translation := mgl32.Translate3D(float32(-4+3*i), 0, float32(-4+3*j))
			rotation := mgl32.HomogRotate3D(0, mgl32.Vec3{1, 0, 0})
			s := float32(1.0)
			scale := mgl32.Scale3D(s, s, s)
			model := translation.Mul4(rotation.Mul4(scale))
			models[3*i+j] = model
		}
	}
	modelsFlat := make([]float32, 3*3*4*4)
	for m := 0; m < 3*3; m++ {
		copy(modelsFlat[(m*16):], models[m][:])
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
	gl.BindBuffer(gl.ARRAY_BUFFER, modelBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(modelsFlat)*floatSz), gl.Pointer(&modelsFlat[0]), gl.STATIC_DRAW)

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
	eye := mgl32.Vec3{6, 6, 8}
	target := mgl32.Vec3{0, 0, 0}
	up := mgl32.Vec3{0, 1, 0}
	view := mgl32.LookAtV(eye, target, up)

	fmt.Println()
	printLiteralMat4("projection", projection)
	printLiteralMat4("view", view)
	printLiteralMat4("model", models[0])

	fmt.Println()
	fmt.Printf("view: %v\n", view)
	fmt.Printf("model[0]: %v\n", models[0])

	gl.UniformMatrix4fv(uniProjection, 1, gl.FALSE, (*gl.Float)(unsafe.Pointer(&projection[0])))
	gl.UniformMatrix4fv(uniView, 1, gl.FALSE, (*gl.Float)(unsafe.Pointer(&view[0])))

	var event sdl.Event
	var running bool
	running = true
	for running {
		for event = sdl.PollEvent(); event != nil; event =
			sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				// fmt.Printf("[%dms]MouseMotion\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Which, t.X, t.Y, t.XRel, t.YRel)
			}
		}
		drawgl(verts, colors)
		window.GLSwap()
		time.Sleep(50 * time.Millisecond)
	}
}

func drawgl(verts, colors []gl.Float) {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.DrawArraysInstanced(gl.TRIANGLES, gl.Int(0), gl.Sizei(len(verts)*4), 3*3)
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

#define PIXEL_SIZE 1.5
out vec4 outColor;
in vec3 fragmentColor;
void main()
{
	vec2 xy = gl_FragCoord.xy;
	float brightness = (fragmentColor.r + fragmentColor.g + fragmentColor.b) / 3;
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

		outColor = vec4(vec3(result), 1.0);
		outColor = vec4(vec3(brightness), 1.0);
		// outColor = vec4(1.0);
}
`
)
