package main

/*
#cgo CFLAGS: -Wall

void openGLDebugCallback(int, int, int, int, int, char*, void*);
*/
import "C"

import (
	"fmt"
	"math"
	"math/rand"
	"runtime"
	"strings"
	"time"
	"unsafe"

	gl "github.com/chsc/gogl/gl43"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/dt-rush/sameriver/v3"
)

var f gl.Float
var floatSz int = int(unsafe.Sizeof(f))
var i gl.Uint
var intSz int = int(unsafe.Sizeof(i))

func assertGLErr() {
	err := gl.GetError()
	if err != 0 {
		panic(fmt.Sprintf("GL error: 0x%04x", err))
	}
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
	fmt.Printf("Compiled Shader: %d\n", status)
	if status == gl.FALSE {
		logShaderCompileError(s)
	}
	return s
}

func createprogram() gl.Uint {
	// CREATE PROGRAM
	program := gl.CreateProgram()
	assertGLErr()

	// shaders
	vs := compileShader(gl.VERTEX_SHADER, vertexShaderSource)
	fs := compileShader(gl.FRAGMENT_SHADER, fragmentShaderSource)
	// fs := compileShader(gl.FRAGMENT_SHADER, debugNilFragmentShaderSource)
	assertGLErr()

	gl.AttachShader(program, vs)
	gl.AttachShader(program, fs)
	assertGLErr()

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

//export openGLDebugCallback
func openGLDebugCallback(source C.int, gltype C.int, id C.int, severity C.int, length C.int, message *C.char, userParam unsafe.Pointer) {
	fmt.Printf("OpenGL Error: %v\n", C.GoString(message))
}

func glInit() {
	gl.Init()
	gl.Viewport(0, 0, gl.Sizei(winWidth), gl.Sizei(winHeight))
	// OPENGL FLAGS
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	extensions := gl.GoStringUb(gl.GetString(gl.EXTENSIONS))

	if !strings.Contains(extensions, "GL_ARB_texture_buffer_object") {
		panic("GL_ARGB_texture_buffer_object needs to be enabled")
	} else {
		fmt.Println("GL_ARGB_texture_buffer_object is enabled.")
	}

	// DEBUG callback
	gl.Enable(gl.DEBUG_OUTPUT)
	debugCallbackPtr := (gl.Pointer)(unsafe.Pointer(C.openGLDebugCallback))
	gl.DebugMessageCallback(debugCallbackPtr, gl.Pointer(nil))
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

func flattenModels(models []mgl.Mat4) (flattened []float32) {
	flattened = make([]float32, len(models)*4*4)
	for m := 0; m < len(models); m++ {
		copy(flattened[(m*4*4):], models[m][:])
	}
	return flattened
}

func flattenPointLights(pointLights []PointLight) (flattened []float32) {
	flattened = make([]float32, 0)
	for _, p := range pointLights {
		flattened = append(flattened,
			float32(p.Position[0]),
			float32(p.Position[1]),
			float32(p.Position[2]),
			float32(0.0),
			float32(p.Color[0]),
			float32(p.Color[1]),
			float32(p.Color[2]),
			float32(p.AttCoeff)) // little trick, put the attcoeff in color data vec4, call it RGBI (i intensity aka attenuation coefficient)
	}
	return flattened
}

const INSTANCE_DIM = 1
const N_INSTANCES = INSTANCE_DIM * INSTANCE_DIM

var flameAccum = sameriver.NewTimeAccumulator(500)

func updatePointLights(pointLights []PointLight, dt_ms float32, t float32, xCenter float32) {
	const LIGHT_MODE = "flame"
	var theta float32
	switch LIGHT_MODE {
	case "mouse":
		// move like the sun overhead along x axis
		// x -> [-1, 1]                  // xCenter range on window
		// (x+1)/2 -> [0, 1]
		// 1 - (x+1)/2 -> [1, 0]
		// pi*(1 - (x+1)/2) -> [pi, 0]   // theta
		// sin(theta) -> [ 0...1...0]
		// cos(theta) -> [-1...0...1]
		theta = math.Pi * (1 - (xCenter+1)/2) // mousemovement
		pointLights[0].Position[0] = 5 * float32(math.Cos(float64(theta)))
		pointLights[0].Position[1] = 2*float32(math.Sin(float64(theta))) + 1
		pointLights[0].Position[2] = 5
	case "circle":
		// move in circle
		theta = math.Pi * t
		pointLights[0].Position[0] = 8 * float32(math.Cos(float64(theta)))
		pointLights[0].Position[1] = 3 // const height
		pointLights[0].Position[2] = 8 * float32(math.Sin(float64(theta)))
	case "day":
		theta = math.Pi * t / 2
		pointLights[0].Position[0] = 8 * float32(math.Cos(float64(theta)))
		pointLights[0].Position[1] = 8 * float32(math.Sin(float64(theta)))
		// light completely attenuates when light is below horizon
		if pointLights[0].Position[1] < 0 {
			pointLights[0].AttCoeff = 10
		} else {
			pointLights[0].AttCoeff = 0.01
		}
	case "left":
		pointLights[0].Position[0] = -2
		pointLights[0].Position[1] = 1
		pointLights[0].Position[2] = 0
	case "flame":
		pointLights[0].Color = mgl.Vec3{1.0, 0.0, 0}
		pointLights[0].Position[0] = 0
		pointLights[0].Position[1] = 0
		pointLights[0].Position[2] = 1
		if flameAccum.Tick(float64(dt_ms)) {
			flameAccum = sameriver.NewTimeAccumulator(800 + 2000*rand.Float64())
			pointLights[0].AttCoeff = gl.Float(0.2 + 0.02*(1+math.Cos(math.Pi*rand.Float64()))/2)
			if len(pointLights) > 1 {
				// consider candle as [1]
				pointLights[1].AttCoeff = gl.Float(0.8 + 0.2*(1+math.Cos(math.Pi*rand.Float64()))/2)
			}
		}
	}
}

func main() {

	window, context := SDLInit()
	defer sdl.Quit()
	defer window.Destroy()
	defer sdl.GLDeleteContext(context)

	glInit()

	// set up point lights
	// XYZRGB
	pointLights := []PointLight{
		PointLight{
			Position: mgl.Vec3{100, 5, 0},
			Color:    mgl.Vec3{1, 1, 1},
			AttCoeff: 0.2,
		},
	}
	const ADD_CANDLE = true
	if ADD_CANDLE {
		pointLights = append(pointLights, PointLight{
			Position: mgl.Vec3{0, 0, 1},
			Color:    mgl.Vec3{1, 0.8, 0.2},
			AttCoeff: 1.0,
		})
	}
	const ADD_RG_FRONTLIGHTS = false
	if ADD_RG_FRONTLIGHTS {
		pointLights = append(pointLights, PointLight{
			Position: mgl.Vec3{-10, 0, 5},
			Color:    mgl.Vec3{0.2, 1, 0.2},
			AttCoeff: 0.01,
		})
		pointLights = append(pointLights, PointLight{
			Position: mgl.Vec3{4, 0, 5},
			Color:    mgl.Vec3{1.0, 0.2, 0.2},
			AttCoeff: 0.01,
		})
	}

	// load model
	vertNormArr := LoadObj("amitabha_smaller.obj")
	fmt.Printf("%d bytes of vert,norm data\n", floatSz*len(vertNormArr))

	// set up transforms for each instance
	positions := make([]mgl.Vec3, N_INSTANCES)
	rotations := make([]mgl.Vec3, N_INSTANCES)
	scales := make([]mgl.Vec3, N_INSTANCES)
	for i := 0; i < INSTANCE_DIM; i++ {
		for j := 0; j < INSTANCE_DIM; j++ {
			ix := INSTANCE_DIM*i + j
			positions[ix] = mgl.Vec3{
				1.2 * float32(-INSTANCE_DIM/2+i),
				0,
				1.2 * float32(-INSTANCE_DIM/2+j)}
			rotations[ix] = mgl.Vec3{0, 0, 0}
			s := float32(0.333)
			scales[ix] = mgl.Vec3{s, s, s}
		}
	}

	// VERTEX,NORM BUFFER
	var vertexBuffer gl.Uint
	gl.GenBuffers(1, &vertexBuffer)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(vertNormArr)*floatSz), gl.Pointer(&vertNormArr[0]), gl.STATIC_DRAW)

	// INSTANCE MODEL BUFFER
	// (note: BufferData() for this occurs in the loop after we update the models)
	var modelBuffer gl.Uint
	gl.GenBuffers(1, &modelBuffer)

	// POINT LIGHTS TEXTURE BUFFER (variable size array trick)
	var pointLightBuffer gl.Uint
	gl.GenBuffers(1, &pointLightBuffer)
	gl.BindBuffer(gl.TEXTURE_BUFFER, pointLightBuffer)
	// point light texture object
	var supported gl.Int
	gl.GetInternalformativ(gl.TEXTURE_BUFFER, gl.RGBA32F, gl.INTERNALFORMAT_SUPPORTED, 1, &supported)
	if supported == gl.TRUE {
		fmt.Println("GL_RGBA32F format is supported")
	} else {
		fmt.Println("GL_RGBA32F format is not supported")
	}
	var pointLightTexture gl.Uint
	gl.GenTextures(1, &pointLightTexture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_BUFFER, pointLightTexture)
	gl.TexBuffer(gl.TEXTURE_BUFFER, gl.RGBA32F, pointLightBuffer)
	assertGLErr()

	// GUESS WHAT
	program := createprogram()
	// USE PROGRAM
	gl.UseProgram(program)
	assertGLErr()

	uniPointLights := gl.GetUniformLocation(program, gl.GLString("pointLights"))
	fmt.Printf("uniPointLights location: %d\n", uniPointLights)
	gl.Uniform1i(uniPointLights, 0) // 0 matches TEXTURE0 above
	// numpointlights is needed to iterate
	uniNumPointLights := gl.GetUniformLocation(program, gl.GLString("numPointLights"))
	fmt.Printf("uniNumPointLights location: %d\n", uniNumPointLights)
	gl.Uniform1i(uniNumPointLights, gl.Int(len(pointLights)))

	// VERTEX ARRAY
	var VertexArrayID gl.Uint
	gl.GenVertexArrays(1, &VertexArrayID)
	gl.BindVertexArray(VertexArrayID)
	// vert data
	vertLoc := gl.Uint(gl.GetAttribLocation(program, gl.GLString("vert")))
	fmt.Printf("vert attrib loc: %d\n", vertLoc)
	gl.EnableVertexAttribArray(vertLoc)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.VertexAttribPointer(vertLoc, 3, gl.FLOAT, gl.FALSE, gl.Sizei(6*floatSz), nil)
	// norm data
	normLoc := gl.Uint(gl.GetAttribLocation(program, gl.GLString("norm")))
	fmt.Printf("norm attrib loc: %d\n", normLoc)
	gl.EnableVertexAttribArray(normLoc)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.VertexAttribPointer(normLoc, 3, gl.FLOAT, gl.FALSE, gl.Sizei(6*floatSz), gl.Pointer(uintptr(3*floatSz)))

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

	// PROJECTION,VIEW UNIFORMS
	uniProjection := gl.GetUniformLocation(program, gl.GLString("projection"))
	fmt.Printf("projection uniform location: %d\n", uniProjection)
	uniView := gl.GetUniformLocation(program, gl.GLString("view"))
	fmt.Printf("view uniform location: %d\n", uniView)

	projection := mgl.Perspective(mgl.DegToRad(45.0), float32(winWidth)/winHeight, 0.1, 500.0)
	eye := mgl.Vec3{0, INSTANCE_DIM + 3, 3 + 2*INSTANCE_DIM}
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
	// var yCenter float32
	// var aliveness float32
	for running {
		t := float32(time.Since(t0).Seconds())
		dt_ms := float32(time.Since(t0).Nanoseconds()) / 1e6
		for event = sdl.PollEvent(); event != nil; event =
			sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				// fmt.Printf("[%dms]MouseMotion\tid:%d\tx:%d\ty:%d\txrel:%d\tyrel:%d\n", t.Timestamp, t.Which, t.X, t.Y, t.XRel, t.YRel)
				xCenter = -1.0 + 2*float32(e.X)/float32(winWidth)
				// yCenter = -1.0 + 2*float32(e.Y)/float32(winHeight)
				// distance from midpoint of screen
				// r := math.Sqrt(float64((xCenter * xCenter) + (yCenter * yCenter)))
				// normal distribution
				// sd := 0.1
				// aliveness = float32(math.Exp(-(r*r)/(2*sd))/(sd*math.Sqrt(2*math.Pi))) / 3
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
			rotations[i] = mgl.Vec3{0, math.Pi, 0}
			// rotations[i] = mgl.Vec3{0, 0.1 * t, 0} // auto-rotate
			// rotations[i] = mgl.Vec3{0, math.Pi * xCenter, 0} // mouse rotate
			// yAmplitude := float32(math.Log(N_INSTANCES+1)) * aliveness
			yAmplitude := float32(0.0)
			positions[i] = mgl.Vec3{positions[i][0], yAmplitude * float32(0.1*math.Sin(float64(2*math.Pi*(float32(x)/(INSTANCE_DIM/3.0)+dt_ms/1000.0)))), positions[i][2]}
		}
		models = buildModels(positions, rotations, scales)
		modelsFlat := flattenModels(models)

		// modify/prepare point lights
		updatePointLights(pointLights, dt_ms, t, xCenter)
		pointLightsFlat := flattenPointLights(pointLights)
		dt_prepare := float32(time.Since(t1).Nanoseconds()) / 1e6
		if dt_prepare_avg == nil {
			dt_prepare_avg = &dt_prepare
		} else {
			*dt_prepare_avg = (*dt_prepare_avg + dt_prepare) / 2.0
		}

		// buffer data
		t2 := time.Now()
		gl.BindBuffer(gl.ARRAY_BUFFER, modelBuffer)
		gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(modelsFlat)*floatSz), gl.Pointer(&modelsFlat[0]), gl.STATIC_DRAW)
		gl.BindBuffer(gl.TEXTURE_BUFFER, pointLightBuffer)
		gl.BufferData(gl.TEXTURE_BUFFER, gl.Sizeiptr(len(pointLightsFlat)*floatSz), gl.Pointer(&pointLightsFlat[0]), gl.STATIC_DRAW)
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
		gl.DrawArraysInstanced(gl.TRIANGLES, gl.Int(0), gl.Sizei(len(vertNormArr)*floatSz), N_INSTANCES)

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
	winWidth           = 1024
	winHeight          = 926
	vertexShaderSource = `
#version 430

in vec3 vert;
in vec3 norm;
in mat4 model;

uniform mat4 projection;
uniform mat4 view;

out VS_FS_INTERFACE {
    vec3 pos;
    vec3 norm;
} pass;

void main()
{
    gl_Position = projection *  view * model * vec4(vert, 1);
	pass.pos = vec3(model * vec4(vert, 1.0)); // Transform vertex position to world space
    pass.norm = normalize(mat3(transpose(inverse(model))) * norm); // Transform normal vector to world space
}
`
	fragmentShaderSource = `
#version 430

#define BYPASS 1
#define ORIGINAL_MIX 0.7
#define PIXEL_SIZE 1
#define ambient 0.05
#define constAtt 1.0

in VS_FS_INTERFACE {
    vec3 pos;
    vec3 norm;
} pass;

uniform int numPointLights;
uniform samplerBuffer pointLights;

out vec4 outColor;

void main()
{
	vec3 totalLight = vec3(0);
	for (int i = 0; i < numPointLights; i++) {
		vec3 lightPos = texelFetch(pointLights, 2*i).xyz;
		vec3 lightColor = texelFetch(pointLights, 2*i+1).rgb;
		float attCoeff = texelFetch(pointLights, 2*i+1).w;
		vec3 toLight = lightPos - pass.pos;

		vec3 dir = normalize(toLight);
		float diffuse = max(dot(pass.norm, dir), 0.0);

		float distance = length(toLight);
		float attenuation = 1.0 / (constAtt + attCoeff * distance + attCoeff * distance * distance);
		
		vec3 lightContribution = attenuation * lightColor * diffuse;
		totalLight += lightContribution;
	}

	// compute greyscale brightness
	float brightness = (totalLight.r + totalLight.g + totalLight.b) / 3;

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
		outColor = vec4(mix(onOff, totalLight, ORIGINAL_MIX), 1.0);
	} else if (BYPASS == 1) {
		outColor = vec4(vec3(totalLight), 1.0);
	} else if (BYPASS == 2) {
		outColor = vec4(1.0);
	}

}
`
)
