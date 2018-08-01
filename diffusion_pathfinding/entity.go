package main

import (
	"math"
)

type Entity struct {
	w *World

	pos        Vec2D
	vel        Vec2D
	steer      Vec2D
	moveTarget *Vec2D
	path       []Vec2D
}

func NewEntity(pos Vec2D, w *World) *Entity {
	return &Entity{
		w:   w,
		pos: pos,
	}
}

func (e *Entity) Update() {
	e.UpdateVel()
	e.Move()
}

func (e *Entity) UpdateVel() {

	if e.moveTarget != nil {
		var nextPathPoint *Vec2D = nil
		for nextPathPoint == nil {
			if len(e.path) == 0 {
				nextPathPoint = e.moveTarget
				break
			}
			nextPathPoint = &e.path[len(e.path)-1]
			if e.pos.Sub(*nextPathPoint).Magnitude() < GRIDCELL_WORLD_W/4 {
				e.path = e.path[:len(e.path)-1]
				nextPathPoint = nil
				continue
			}
		}
		toward := (*nextPathPoint).Sub(e.pos)
		e.steer = toward.Unit()
		angle := e.vel.AngleBetween(e.steer)
		max := MOVESPEED * (1 - sigma4(e.moveTarget.Sub(e.pos).Magnitude(), 16)) *
			(1 - 0.9*(angle/math.Pi))
		e.vel = e.vel.Add(e.steer).Truncate(max)
	}
}

func (e *Entity) Move() {
	vel := e.vel
	vX := e.vel.XComponent()
	vY := e.vel.YComponent()
	for _, o := range e.w.obstacles {
		orec := Rect2D{
			float64(GRIDCELL_WORLD_W * o.X),
			float64(GRIDCELL_WORLD_H * o.Y),
			GRIDCELL_WORLD_W,
			GRIDCELL_WORLD_H}
		erec := Rect2D{
			e.pos.X - ENTITYSZ/2 - 2, e.pos.Y - ENTITYSZ/2 - 2,
			ENTITYSZ + 4, ENTITYSZ + 4}
		erecX := erec.Add(vX)
		if orec.Overlaps(erecX) {
			dxL := orec.X - (e.pos.X + ENTITYSZ/2)
			dxR := e.pos.X - (orec.X + orec.W + ENTITYSZ/2)
			if dxL > 0 {
				// we are to the left
				vX = vX.Truncate(dxL * 0.2)
			} else if dxR > 0 {
				// we are to the right
				vX = vX.Truncate(dxR * 0.2)
			}
		}
		if orec.Overlaps(erec.Add(vX)) {
			vX = Vec2D{0, 0}
		}
		erecY := erec.Add(vY)
		if orec.Overlaps(erecY) {
			dyD := orec.Y - (e.pos.Y + ENTITYSZ/2)
			dyU := e.pos.Y - (orec.Y + orec.H + ENTITYSZ/2)
			if dyD > 0 {
				// we are down
				vY = vY.Truncate(dyD * 0.2)
			} else if dyU > 0 {
				// we are up
				vY = vY.Truncate(dyU * 0.2)
			}
		}
		if orec.Overlaps(erec.Add(vY)) {
			vY = Vec2D{0, 0}
		}
	}
	vel = vX.Add(vY)
	e.pos = e.pos.Add(vel)
}
