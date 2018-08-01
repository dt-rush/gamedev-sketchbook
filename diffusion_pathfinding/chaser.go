package main

import (
	"math"
)

type Chaser struct {
	w *World

	pos        Vec2D
	vel        Vec2D
	steer      Vec2D
	moveTarget *Vec2D
}

func NewChaser(pos Vec2D, w *World) *Chaser {
	return &Chaser{
		w:   w,
		pos: pos,
	}
}

func (c *Chaser) GreaterNeighbor() *Vec2D {
	cpos := c.w.dm.ToGridSpace(c.pos)
	max := 0.0
	next := cpos
	if c.w.dm.d[cpos.X][cpos.Y] == 1.0 {
		return nil
	}
	for _, ix := range neighborIXs {
		x, y := cpos.X+ix[0], cpos.Y+ix[1]
		if !c.w.dm.InGrid(x, y) {
			continue
		}
		isDiagonal := ix[0]*ix[1] != 0
		if isDiagonal && (c.w.dm.CellHasObstacle(cpos.X+ix[0], cpos.Y) ||
			c.w.dm.CellHasObstacle(cpos.X, cpos.Y+ix[1])) {
			continue
		}
		d := c.w.dm.d[x][y]
		if d > max {
			max = d
			next = Position{x, y}
		}
	}
	p := c.w.dm.ToWorldSpace(next)
	return &p
}

func (c *Chaser) Update() {
	c.UpdateVel()
	c.Move()
}

func (c *Chaser) UpdateVel() {
	if c.moveTarget == nil {
		c.moveTarget = c.GreaterNeighbor()
	}
	if c.moveTarget != nil &&
		c.pos.Sub(*c.moveTarget).Magnitude() < GRIDCELL_WORLD_W/4 {
		c.moveTarget = c.GreaterNeighbor()
	}
	if c.moveTarget == nil {
		return
	}
	toward := c.moveTarget.Sub(c.pos)
	c.steer = toward.Unit()
	angle := c.vel.AngleBetween(c.steer)
	maxVel := (MOVESPEED / 2) * (1 - 0.9*(angle/math.Pi))
	c.vel = c.vel.Add(c.steer).Truncate(maxVel)
}

func (c *Chaser) Move() {
	vel := c.vel
	vX := c.vel.XComponent()
	vY := c.vel.YComponent()
	for _, o := range c.w.obstacles {
		orec := Rect2D{
			float64(GRIDCELL_WORLD_W * o.X),
			float64(GRIDCELL_WORLD_H * o.Y),
			GRIDCELL_WORLD_W,
			GRIDCELL_WORLD_H}
		erec := Rect2D{
			c.pos.X - POINTSZ/2 - 2, c.pos.Y - POINTSZ/2 - 2,
			POINTSZ + 4, POINTSZ + 4}
		erecX := erec.Add(vX)
		if orec.Overlaps(erecX) {
			dxL := orec.X - (c.pos.X + POINTSZ/2)
			dxR := c.pos.X - (orec.X + orec.W + POINTSZ/2)
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
			dyD := orec.Y - (c.pos.Y + POINTSZ/2)
			dyU := c.pos.Y - (orec.Y + orec.H + POINTSZ/2)
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
	c.pos = c.pos.Add(vel)
}
