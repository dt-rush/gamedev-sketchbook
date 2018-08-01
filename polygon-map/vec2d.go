package main

import (
	"math"
)

type Vec2D struct {
	X float64
	Y float64
}

func (v Vec2D) ToPoint() Point2D {
	return Point2D{
		int(v.X),
		int(v.Y)}
}

func VecFromPoints(p1 Point2D, p2 Point2D) Vec2D {
	return Vec2D{float64(p2.X - p1.X), float64(p2.Y - p1.Y)}
}

func (v1 Vec2D) Sub(v2 Vec2D) Vec2D {
	return Vec2D{v1.X - v2.X, v1.Y - v2.Y}
}

func (v1 Vec2D) ScalarCross(v2 Vec2D) float64 {
	return v1.X*v2.Y - v1.Y*v2.X
}

func (v Vec2D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v Vec2D) PerpendicularUnit() Vec2D {
	m := v.Magnitude()
	return Vec2D{v.Y / m, -v.X / m}
}
