package main

import (
	"math"
)

type Vec2D struct {
	X float64
	Y float64
}

func (v Vec2D) ToPoint() Vec2D {
	return Vec2D{v.X, v.Y}
}

func (v1 Vec2D) Add(v2 Vec2D) Vec2D {
	return Vec2D{v1.X + v2.X, v1.Y + v2.Y}
}

func (v1 Vec2D) Sub(v2 Vec2D) Vec2D {
	return Vec2D{v1.X - v2.X, v1.Y - v2.Y}
}

func (v1 Vec2D) Distance(v2 Vec2D) (dx, dy, d float64) {
	dx = float64(v2.X - v1.X)
	dy = float64(v2.Y - v1.Y)
	d = math.Sqrt(dx*dx + dy*dy)
	return dx, dy, d
}

func (v1 Vec2D) ScalarCross(v2 Vec2D) float64 {
	return v1.X*v2.Y - v1.Y*v2.X
}

func (v1 Vec2D) Dot(v2 Vec2D) float64 {
	return v1.X*v2.X + v1.Y*v2.Y
}

func (v Vec2D) Magnitude() float64 {
	return math.Sqrt(v.X*v.X + v.Y*v.Y)
}

func (v1 Vec2D) Project(v2 Vec2D) float64 {
	return v1.Dot(v2.Unit())
}

func (v Vec2D) PerpendicularUnit() Vec2D {
	m := v.Magnitude()
	return Vec2D{v.Y / m, -v.X / m}
}

func (v Vec2D) Scale(r float64) Vec2D {
	return Vec2D{r * v.X, r * v.Y}
}

func (v Vec2D) Unit() Vec2D {
	return v.Scale(1 / v.Magnitude())
}

func (v Vec2D) Truncate(val float64) Vec2D {
	m := v.Magnitude()
	if m > val {
		return v.Scale(val / m)
	} else {
		return v
	}
}

func (v Vec2D) XComponent() Vec2D {
	return Vec2D{v.X, 0}
}

func (v Vec2D) YComponent() Vec2D {
	return Vec2D{0, v.Y}
}

func (v1 Vec2D) AngleBetween(v2 Vec2D) float64 {
	if v1.Magnitude() == 0 || v2.Magnitude() == 0 {
		return 0.0
	}
	d := v1.Dot(v2) / (v1.Magnitude() * v2.Magnitude())
	if d >= 1.0 {
		return 0.0
	} else if d <= -1.0 {
		return math.Pi
	} else {
		return math.Acos(d)
	}
}
