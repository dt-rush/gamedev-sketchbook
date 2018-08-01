package main

import (
	"math"
	"time"
)

func (w *World) ComputePath() float64 {
	return w.ComputeEntityPathUnrolled()
}

func (w *World) ComputeEntityPathUnrolled() float64 {
	var t_ms float64
	if w.e != nil && w.e.moveTarget != nil {
		t0 := time.Now()
		path, _, found := w.c.Path(&w.e.pos, w.e.moveTarget)
		t_ms = float64(time.Since(t0).Nanoseconds()) / float64(1e6)
		if found {
			w.e.path = path
		}
	}
	w.c.Clear()
	return t_ms
}

func (w *World) MoveEntity() {
	if w.e != nil && w.e.moveTarget != nil && w.e.path != nil {
		if len(w.e.path) == 0 {
			w.e.moveTarget = nil
			w.e.path = nil
		}
		last_ix := len(w.e.path) - 1
		var target Point2D
		var dx, dy, d float64
		var validTarget = false
		for !validTarget {
			target = w.e.path[last_ix]
			dx = float64(target.X - w.e.pos.X)
			dy = float64(target.Y - w.e.pos.Y)
			d = math.Sqrt(dx*dx + dy*dy)
			reached := (d <= 2)
			if reached {
				w.e.path = w.e.path[:last_ix]
				last_ix--
			} else {
				validTarget = true
			}
		}
		w.e.pos.X += int(math.Max(1, dx/d))
		w.e.pos.Y += int(math.Max(1, dy/d))
	}
}
