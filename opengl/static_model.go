package main

import (
	"fmt"
	"time"

	gl "github.com/chsc/gogl/gl43"
	mgl "github.com/go-gl/mathgl/mgl32"
)

type StaticModel struct {
	name string
	// the vertex array holding vertexbuffer and modelbuffer
	vao gl.Uint
	// the vao's buffers
	vertexBuffer gl.Uint
	modelBuffer  gl.Uint

	// vertexbuffer backing
	vertNormArr []gl.Float

	// modelbuffer backing
	nInstances                   int
	positions, rotations, scales []mgl.Vec3
	// TODO: rename transform?
	models     []mgl.Mat4
	modelsFlat []float32

	// wait for commands to finish
	syncFence gl.Sync

	// timing statistics
	buffer_avg_ms *float64
	draw_avg_ms   *float64
}

func NewStaticModel(filename string, n int) StaticModel {
	m := StaticModel{
		name:       filename,
		nInstances: n,
		positions:  make([]mgl.Vec3, n),
		rotations:  make([]mgl.Vec3, n),
		scales:     make([]mgl.Vec3, n),
	}
	m.buildModels()
	m.flatten()
	m.vertNormArr = LoadObj(filename)
	fmt.Printf("%d bytes of vert,norm data in %s\n", floatSz*len(m.vertNormArr), m.name)
	m.vertexBuffer, m.modelBuffer = m.genBuffers()
	m.syncFence = gl.FenceSync(gl.SYNC_GPU_COMMANDS_COMPLETE, 0)
	return m
}

func (m *StaticModel) genBuffers() (gl.Uint, gl.Uint) {
	var vertexBuffer, modelBuffer gl.Uint
	// VERTEX,NORM BUFFER
	gl.GenBuffers(1, &vertexBuffer)
	fmt.Printf("model[%s].vertexBuffer = %d\n", m.name, vertexBuffer)
	// INSTANCE MODEL BUFFER
	gl.GenBuffers(1, &modelBuffer)
	fmt.Printf("model[%s].modelBuffer = %d\n", m.name, modelBuffer)
	return vertexBuffer, modelBuffer
}

func (m *StaticModel) genVAO(program gl.Uint) {
	// VERTEX ARRAY
	gl.GenVertexArrays(1, &m.vao)
	fmt.Printf("model[%s].vao = %d\n", m.name, m.vao)
	gl.BindVertexArray(m.vao)
	// vertex buffer
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vertexBuffer)
	//     vert attrib
	vertLoc := gl.Uint(gl.GetAttribLocation(program, gl.GLString("vert")))
	fmt.Printf("vert attrib loc: %d\n", vertLoc)
	gl.EnableVertexAttribArray(vertLoc)
	// gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.VertexAttribPointer(vertLoc, 3, gl.FLOAT, gl.FALSE, gl.Sizei(6*floatSz), nil)
	//     norm attrib
	normLoc := gl.Uint(gl.GetAttribLocation(program, gl.GLString("norm")))
	fmt.Printf("norm attrib loc: %d\n", normLoc)
	gl.EnableVertexAttribArray(normLoc)
	// gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.VertexAttribPointer(normLoc, 3, gl.FLOAT, gl.FALSE, gl.Sizei(6*floatSz), gl.Pointer(uintptr(3*floatSz)))

	// models instanced attrib (divisor)
	gl.BindBuffer(gl.ARRAY_BUFFER, m.modelBuffer)
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
	// gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.VertexAttribDivisor(loc1, 1)
	gl.VertexAttribDivisor(loc2, 1)
	gl.VertexAttribDivisor(loc3, 1)
	gl.VertexAttribDivisor(loc4, 1)
}

func (m *StaticModel) buildModels() {
	if m.models == nil {
		m.models = make([]mgl.Mat4, m.nInstances)
	}
	for i := 0; i < len(m.positions); i++ {
		t := mgl.Translate3D(m.positions[i][0], m.positions[i][1], m.positions[i][2])
		r := mgl.Ident4()
		r = r.Mul4(mgl.HomogRotate3D(m.rotations[i][0], mgl.Vec3{1, 0, 0}))
		r = r.Mul4(mgl.HomogRotate3D(m.rotations[i][1], mgl.Vec3{0, 1, 0}))
		r = r.Mul4(mgl.HomogRotate3D(m.rotations[i][2], mgl.Vec3{0, 0, 1}))
		s := mgl.Scale3D(m.scales[i][0], m.scales[i][1], m.scales[i][2])
		m.models[i] = t.Mul4(r.Mul4(s))
	}
}

func (m *StaticModel) flatten() {
	m.modelsFlat = make([]float32, len(m.models)*4*4)
	for i := 0; i < len(m.models); i++ {
		copy(m.modelsFlat[(i*4*4):], m.models[i][:])
	}
}

func (m *StaticModel) bufferData() {
	t0 := time.Now()
	// send vertex data of obj
	gl.BindBuffer(gl.ARRAY_BUFFER, m.vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(m.vertNormArr)*floatSz), gl.Pointer(&m.vertNormArr[0]), gl.STATIC_DRAW)
	// send model mat4 data
	m.buildModels()
	m.flatten()
	gl.BindBuffer(gl.ARRAY_BUFFER, m.modelBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, gl.Sizeiptr(len(m.modelsFlat)*floatSz), gl.Pointer(&m.modelsFlat[0]), gl.STATIC_DRAW)
	bufferDone := false
	for !bufferDone {
		bufferWaitResult := gl.ClientWaitSync(m.syncFence, gl.SYNC_FLUSH_COMMANDS_BIT, 5000)
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
	dt_buffer := float64(time.Since(t0).Nanoseconds()) / 1e6
	if m.buffer_avg_ms == nil {
		m.buffer_avg_ms = &dt_buffer
	} else {
		*m.buffer_avg_ms = (*m.buffer_avg_ms + dt_buffer) / 2.0
	}
}

func (m *StaticModel) Draw() {
	t0 := time.Now()
	gl.BindVertexArray(m.vao)
	m.bufferData()
	gl.DrawArraysInstanced(gl.TRIANGLES, gl.Int(0), gl.Sizei(len(m.vertNormArr)*floatSz), gl.Sizei(m.nInstances))
	drawDone := false
	for !drawDone {
		drawWaitResult := gl.ClientWaitSync(m.syncFence, gl.SYNC_FLUSH_COMMANDS_BIT, 5000)
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
	dt_draw := float64(time.Since(t0).Nanoseconds()) / 1e6
	if m.draw_avg_ms == nil {
		m.draw_avg_ms = &dt_draw
	} else {
		*m.draw_avg_ms = (*m.draw_avg_ms + dt_draw) / 2.0
	}
}

func (m *StaticModel) DumpStats() map[string]float64 {
	stats := make(map[string]float64)
	stats["draw_avg_ms"] = *m.draw_avg_ms
	stats["buffer_avg_ms"] = *m.buffer_avg_ms
	return stats
}
