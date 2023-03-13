package main

/*
#cgo CFLAGS: -Wall

void openGLDebugCallback(int, int, int, int, int, char*, void*);
*/
import "C"

import (
	"fmt"
	"runtime"
	"strings"
	"time"
	"unsafe"

	gl "github.com/chsc/gogl/gl43"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"

	"github.com/dt-rush/sameriver/v3"
)

type PointLight struct {
	Position mgl.Vec3
	Color    mgl.Vec3
	AttCoeff gl.Float
}

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
	gl.Enable(gl.CULL_FACE)
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

func LoadPointLights() (pointLights []PointLight) {
	pointLights = []PointLight{
		PointLight{
			Position: mgl.Vec3{100, 5, 0},
			Color:    mgl.Vec3{1, 1, 1},
			AttCoeff: 0.1,
		},
	}
	return pointLights
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

var flameAccum = sameriver.NewTimeAccumulator(500)

func main() {

	window, context := SDLInit()
	defer sdl.Quit()
	defer window.Destroy()
	defer sdl.GLDeleteContext(context)

	glInit()

	// GUESS WHAT
	program := createprogram()
	// USE PROGRAM
	gl.UseProgram(program)
	assertGLErr()

	// set up point lights
	pointLights := LoadPointLights()

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

	uniPointLights := gl.GetUniformLocation(program, gl.GLString("pointLights"))
	fmt.Printf("uniPointLights location: %d\n", uniPointLights)
	gl.Uniform1i(uniPointLights, 0) // 0 matches TEXTURE0 above
	// numpointlights is needed to iterate
	uniNumPointLights := gl.GetUniformLocation(program, gl.GLString("numPointLights"))
	fmt.Printf("uniNumPointLights location: %d\n", uniNumPointLights)
	gl.Uniform1i(uniNumPointLights, gl.Int(len(pointLights)))

	// PROJECTION,VIEW UNIFORMS
	uniProjection := gl.GetUniformLocation(program, gl.GLString("projection"))
	fmt.Printf("projection uniform location: %d\n", uniProjection)
	uniView := gl.GetUniformLocation(program, gl.GLString("view"))
	fmt.Printf("view uniform location: %d\n", uniView)

	projection := mgl.Perspective(mgl.DegToRad(45.0), float32(winWidth)/winHeight, 0.1, 500.0)
	eye := mgl.Vec3{0, 20, 0}
	zoomOut := float32(0.6)
	eyeBack := mgl.Scale3D(zoomOut, zoomOut, zoomOut).Mul4x1(mgl.Vec4{eye[0], eye[1], eye[2], 1})
	eye = mgl.Vec3{eyeBack[0], eyeBack[1], eyeBack[2]}
	target := mgl.Vec3{0, 0, 0}
	up := mgl.Vec3{0, 1, 0}
	view := mgl.LookAtV(eye, target, up)

	gl.UniformMatrix4fv(uniProjection, 1, gl.FALSE, (*gl.Float)(unsafe.Pointer(&projection[0])))
	gl.UniformMatrix4fv(uniView, 1, gl.FALSE, (*gl.Float)(unsafe.Pointer(&view[0])))

	var event sdl.Event
	var running bool
	running = true

	var dt_prepare_avg *float32 = nil
	var dt_draw_avg *float32 = nil
	handleKBEvent := func(ke *sdl.KeyboardEvent) {
		if ke.Type == sdl.KEYDOWN {
			if ke.Keysym.Sym == sdl.K_q {
				running = false
			}
		}
	}

	for running {
		// t in seconds as float to 3 places
		// t := float32(time.Since(t0).Milliseconds()) / 1e3
		// dt_ms := float32(time.Since(dt0).Nanoseconds()) / 1e6
		for event = sdl.PollEvent(); event != nil; event =
			sdl.PollEvent() {
			switch e := event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.MouseMotionEvent:
				// nothing
			case *sdl.KeyboardEvent:
				handleKBEvent(e)
			}
		}

		pointLightsFlat := flattenPointLights(pointLights)

		// buffer data
		t2 := time.Now()
		gl.BindBuffer(gl.TEXTURE_BUFFER, pointLightBuffer)
		gl.BufferData(gl.TEXTURE_BUFFER, gl.Sizeiptr(len(pointLightsFlat)*floatSz), gl.Pointer(&pointLightsFlat[0]), gl.STATIC_DRAW)
		// draw
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		// gl.FrontFace(gl.CW)

		window.GLSwap()
		dt_draw := float32(time.Since(t2).Nanoseconds()) / 1e6
		if dt_draw_avg == nil {
			dt_draw_avg = &dt_draw
		} else {
			*dt_draw_avg = (*dt_draw_avg + dt_draw) / 2.0
		}
		time.Sleep(50 * time.Millisecond)
	}
	fmt.Printf("avg prepare ms: %f\n", *dt_prepare_avg)
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

#define BYPASS 0
#define ORIGINAL_MIX 0.7
#define PIXEL_SIZE 1
#define ambient 0.08
#define constAtt 1.0

in VS_FS_INTERFACE {
    vec3 pos;
    vec3 norm;
} pass;

uniform int isPointLightCube;
uniform int numPointLights;
uniform samplerBuffer pointLights;

out vec4 outColor;

void main()
{
	if (isPointLightCube == 1) {
		outColor = vec4(vec3(1), 1);
		return;
	}
	vec3 totalLight = vec3(ambient);
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

   outColor = totalLight;	
}
`
)
