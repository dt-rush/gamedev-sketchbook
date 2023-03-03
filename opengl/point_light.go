package main

import (
	gl "github.com/chsc/gogl/gl43"
	mgl "github.com/go-gl/mathgl/mgl32"
)

type PointLight struct {
	Position mgl.Vec3
	Color    mgl.Vec3
	AttCoeff gl.Float
}
