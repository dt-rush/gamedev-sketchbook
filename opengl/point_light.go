package main

import (
	mgl "github.com/go-gl/mathgl/mgl32"
)

type PointLight struct {
	Position mgl.Vec3
	Color    mgl.Vec3
}
