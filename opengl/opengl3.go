package main

import (
	"fmt"
	"math"
	"runtime"
	"strings"
	"time"
	"unsafe"

	gl "github.com/chsc/gogl/gl43"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
)

var f gl.Float
var floatSz int = int(unsafe.Sizeof(f))
var i gl.Uint
var intSz int = int(unsafe.Sizeof(i))

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

	// DEBUG callback
	debugCallback := func(source uint32, gltype uint32, id uint32, severity uint32, length int32, message string, userParam unsafe.Pointer) {
		fmt.Printf("OpenGL Error: %v\n", message)
	}
	gl.Enable(gl.DEBUG_OUTPUT)
	gl.DebugMessageCallback(gl.Pointer(unsafe.Pointer(&debugCallback)), gl.Pointer(nil))

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

func printLiteralMat4(name string, m mgl.Mat4) {
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

func buildModels(positions, rotations, scales []mgl.Vec3) []mgl.Mat4 {
	models := make([]mgl.Mat4, len(positions))
	for i := 0; i < len(positions); i++ {
		t := mgl.Translate3D(positions[i][0], positions[i][1], positions[i][2])
		r := mgl.Ident4()
		r = r.Mul4(mgl.HomogRotate3D(rotations[i][0], mgl.Vec3{1, 0, 0}))
		r = r.Mul4(mgl.HomogRotate3D(rotations[i][1], mgl.Vec3{0, 1, 0}))
		r = r.Mul4(mgl.HomogRotate3D(rotations[i][2], mgl.Vec3{0, 0, 1}))
		s := mgl.Scale3D(scales[i][0], scales[i][1], scales[i][2])
		models[i] = t.Mul4(r.Mul4(s))
	}
	return models
}

const INSTANCE_DIM = 3
const N_INSTANCES = INSTANCE_DIM * INSTANCE_DIM

func main() {

	window, context := SDLInit()
	defer sdl.Quit()
	defer window.Destroy()
	defer sdl.GLDeleteContext(context)

	glInit()

	// load model
	verts, indices := LoadObjVerts("amitabha_small.obj")

	// set up transforms for each instance
	positions := make([]mgl.Vec3, N_INSTANCES)
	rotations := make([]mgl.Vec3, N_INSTANCES)
	scales := make([]mgl.Vec3, N_INSTANCES)
	for i := 0; i < INSTANCE_DIM; i++ {
		for j := 0; j < INSTANCE_DIM; j++ {
			ix := INSTANCE_DIM*i + j
			positions[ix] = mgl.Vec3{
				2 * float32(-INSTANCE_DIM/2+i),
				0,
				2 * float32(-INSTANCE_DIM/2+j)}
			rotations[ix] = mgl.Vec3{0, 0, 0}
			s := float32(0.333)
			scales[ix] = mgl.Vec3{s, s, s}
		}
	}

	fmt.Printf("%d bytes of vert data\n", floatSz*len(verts))
	fmt.Printf("%d bytes of index data\n", intSz*len(indices))

	// VERTEX BUFFER
	var vertexBuffer gl.Uint
	gl.GenBuffers(1, &vertexBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(verts)*floatSz), gl.Pointer(&verts[0]), gl.STATIC_DRAW)

	// INDEX BUFFER
	var indexBuffer gl.Uint
	gl.GenBuffers(1, &indexBuffer)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, indexBuffer)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, gl.Sizeiptr(len(indices)*intSz), gl.Pointer(&indices[0]), gl.STATIC_DRAW)

	// INSTANCE MODEL BUFFER
	// (note: BufferData() for this occurs in the loop after we update the models)
	var modelBuffer gl.Uint
	gl.GenBuffers(1, &modelBuffer)

	/*
		// COLOUR BUFFER
		var colourbuffer gl.Uint
		gl.GenBuffers(1, &colourbuffer)
		gl.BindBuffer(gl.ARRAY_BUFFER, colourbuffer)
		gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(colors)*floatSz), gl.Pointer(&colors[0]), gl.STATIC_DRAW)
	*/

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

	/*
		// VERTEX ARRAY HOOK COLOURS
		colorLoc := gl.Uint(gl.GetAttribLocation(program, gl.GLString("color")))
		gl.EnableVertexAttribArray(colorLoc)
		gl.BindBuffer(gl.ARRAY_BUFFER, colourbuffer)
		gl.VertexAttribPointer(colorLoc, 3, gl.FLOAT, gl.FALSE, 0, nil)
	*/

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

	// ALIVENESS UNIFORM (used in fragment shader to saturate colour)
	uniAlivenessString := gl.GLString("aliveness")
	uniAliveness := gl.GetUniformLocation(program, uniAlivenessString)

	// PROJECTION,VIEW UNIFORMS
	uniProjectionString := gl.GLString("projection")
	uniProjection := gl.GetUniformLocation(program, uniProjectionString)
	fmt.Printf("projection uniform location: %d\n", uniProjection)
	uniViewString := gl.GLString("view")
	uniView := gl.GetUniformLocation(program, uniViewString)
	fmt.Printf("view uniform location: %d\n", uniView)

	projection := mgl.Perspective(mgl.DegToRad(45.0), float32(winWidth)/winHeight, 0.1, 500.0)
	eye := mgl.Vec3{0, INSTANCE_DIM + 3, 3 + 2*INSTANCE_DIM}
	fmt.Println(eye)
	zoomOut := float32(0.6)
	eyeBack := mgl.Scale3D(zoomOut, zoomOut, zoomOut).Mul4x1(mgl.Vec4{eye[0], eye[1], eye[2], 1})
	eye = mgl.Vec3{eyeBack[0], eyeBack[1], eyeBack[2]}
	target := mgl.Vec3{0, 0, 0}
	up := mgl.Vec3{0, 1, 0}
	view := mgl.LookAtV(eye, target, up)

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

	// used to wait for GPU if it's not done yet
	syncFence := gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)

	var xCenter float32
	var yCenter float32
	var aliveness float32
	for running {
		dt_ms := float32(time.Since(t0).Nanoseconds()) / 1e6
		for event = sdl.PollEvent(); event != nil; event =
			sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				// fmt.Printf("[%dms]MouseMotion\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Which, t.X, t.Y, t.XRel, t.YRel)
				xCenter = -1.0 + 2*float32(e.X)/float32(winWidth)
				yCenter = -1.0 + 2*float32(e.Y)/float32(winHeight)
				// distance from midpoint of screen
				r := math.Sqrt(float64((xCenter * xCenter) + (yCenter * yCenter)))
				// normal distribution
				sd := 0.1
				aliveness = float32(math.Exp(-(r*r)/(2*sd))/(sd*math.Sqrt(2*math.Pi))) / 3
				// if you reset mouseY to zero whenever there hasn't been a mouseevent in x ms for very small x, you
				// get a detector for smooth continuous motion that snaps to zero - great as a mechanic for a meditation
				// mode
			case *sdl.KeyboardEvent:
				handleKBEvent(e)
			}
		}

		// modify/prepare models
		for i, m := range models {
			rot := mgl.HomogRotate3DY(0.05 * float32(i) / float32(len(models)))
			models[i] = rot.Mul4(m)
		}
		gl.BindBuffer(gl.ARRAY_BUFFER, modelBuffer)
		t1 := time.Now()
		for i := 0; i < N_INSTANCES; i++ {
			x := i % N_INSTANCES
			// rotations[i] = rotations[i].Add(mgl.Vec3{0, 0.01, 0})
			rotations[i] = mgl.Vec3{0, 2 * math.Pi * xCenter, 0}
			// yAmplitude := float32(math.Log(N_INSTANCES+1)) * aliveness
			yAmplitude := float32(0.0)
			positions[i] = mgl.Vec3{positions[i][0], yAmplitude * float32(0.1*math.Sin(float64(2*math.Pi*(float32(x)/(INSTANCE_DIM/3.0)+dt_ms/1000.0)))), positions[i][2]}
		}
		models = buildModels(positions, rotations, scales)
		modelsFlat := make([]float32, N_INSTANCES*4*4)
		for m := 0; m < N_INSTANCES; m++ {
			copy(modelsFlat[(m*4*4):], models[m][:])
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
		gl.Uniform1f(uniAliveness, gl.Float(aliveness))
		bufferDone := false
		for !bufferDone {
			bufferWaitResult := gl.ClientWaitSync(syncFence, gl.SYNC_FLUSH_COMMANDS_BIT, 5000)
			switch bufferWaitResult {
			case gl.ALREADY_SIGNALED:
				bufferDone = true
			case gl.CONDITION_SATISFIED:
				bufferDone = true
			case gl.TIMEOUT_EXPIRED:
				fmt.Println("buffer data waitsync reached timeout. retrying.")
			case gl.WAIT_FAILED:
				panic("gl.WAIT_FAILED in ClientWaitSync for BufferData. Should never happen!")
			}
		}
		dt_buffer := float32(time.Since(t2).Nanoseconds()) / 1e6
		if dt_buffer_avg == nil {
			dt_buffer_avg = &dt_buffer
		} else {
			*dt_buffer_avg = (*dt_buffer_avg + dt_buffer) / 2.0
		}

		// draw
		t3 := time.Now()

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// gl.FrontFace(gl.CW)
		gl.DrawElementsInstanced(gl.TRIANGLES, gl.Sizei(len(indices)), gl.UNSIGNED_INT, gl.Pointer(&indices[0]), N_INSTANCES)

		drawDone := false
		for !drawDone {
			drawWaitResult := gl.ClientWaitSync(syncFence, gl.SYNC_FLUSH_COMMANDS_BIT, 5000)
			switch drawWaitResult {
			case gl.ALREADY_SIGNALED:
				drawDone = true
			case gl.CONDITION_SATISFIED:
				drawDone = true
			case gl.TIMEOUT_EXPIRED:
				fmt.Println("draw data waitsync reached timeout. retrying.")
			case gl.WAIT_FAILED:
				panic("gl.WAIT_FAILED in ClientWaitSync for Draw. Should never happen!")
			}
		}
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

#define BYPASS 2
#define PIXEL_SIZE 3.0
out vec4 outColor;
in vec3 fragmentColor;
uniform float aliveness;

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
		outColor = vec4(mix(onOff, fragmentColor, clamp(aliveness/5, 0, 1)), 1.0);
	} else if (BYPASS == 1) {
		outColor = vec4(vec3(brightness), 1.0);
	} else if (BYPASS == 2) {
		outColor = vec4(1.0);
	}
}
`
)
