package main

import (
	"github.com/beefsack/go-astar"
	"time"
)

func (w *World) ComputePath() float64 {
	return w.ComputeEntityPathHandRolled()
}

func (w *World) ComputeEntityPathHandRolled() float64 {
	var t_ms float64
	if w.e != nil && w.e.moveTarget != nil {
		t0 := time.Now()
		path := w.c2.Path(
			w.m.CellAt(w.e.pos).pos,
			w.m.CellAt(*w.e.moveTarget).pos)
		t_ms = float64(time.Since(t0).Nanoseconds()) / float64(1e6)
		if len(path) > 0 {
			w.e.path = path
		}
	}
	w.c.Clear()
	return t_ms
}

func (w *World) ComputeEntityPathUnrolled() float64 {
	var t_ms float64
	if w.e != nil && w.e.moveTarget != nil {
		t0 := time.Now()
		path, _, found := w.c.Path(
			w.m.CellAt(w.e.pos),
			w.m.CellAt(*w.e.moveTarget))
		t_ms = float64(time.Since(t0).Nanoseconds()) / float64(1e6)
		if found {
			w.e.path = path
		}
	}
	w.c.Clear()
	return t_ms
}

func (w *World) ComputeEntityPath() float64 {
	var t_ms float64
	if w.e != nil && w.e.moveTarget != nil {
		t0 := time.Now()
		path, _, found := astar.Path(
			w.m.CellAt(w.e.pos),
			w.m.CellAt(*w.e.moveTarget))
		t_ms = float64(time.Since(t0).Nanoseconds()) / float64(1e6)
		if found {
			cellsPath := make([]Position, len(path))
			for i, pather := range path {
				cellsPath[i] = pather.(*WorldMapCell).pos
			}
			w.e.path = cellsPath
		}
	}
	return t_ms
}

func (w *World) MoveEntity() {
	if w.e != nil && w.e.moveTarget != nil && w.e.path != nil && len(w.e.path) > 1 {
		last_ix := len(w.e.path) - 1
		var target Position
		var validTarget = false
		for !validTarget {
			target = w.e.path[last_ix]
			reached := (w.e.pos == target)
			if reached {
				w.e.path = w.e.path[:last_ix]
				last_ix--
			} else {
				validTarget = true
			}
		}
		w.e.pos = w.e.path[last_ix]
		w.e.path = w.e.path[:last_ix]
		if len(w.e.path) == 0 {
			w.e.moveTarget = nil
			w.e.path = nil
		}
	}

}
