package main

import (
	"github.com/beefsack/go-astar"
	"math"
)

func (c *WorldMapCell) PathNeighbors() []astar.Pather {
	neighbors := make([]astar.Pather, 0)
	for dy := -1; dy <= 1; dy++ {
		if c.pos.Y+dy < 0 ||
			c.pos.Y+dy > WORLD_CELLHEIGHT-1 {
			continue
		}
		for dx := -1; dx <= 1; dx++ {
			if c.pos.X+dx < 0 ||
				c.pos.X+dx > WORLD_CELLWIDTH-1 {
				continue
			}
			neighbors = append(neighbors,
				&c.m.cells[c.pos.Y+dy][c.pos.X+dx])
		}
	}
	return neighbors
}

func (c *WorldMapCell) PathNeighborCost(to astar.Pather) float64 {
	dx := math.Abs(float64(c.pos.X - to.(*WorldMapCell).pos.X))
	dy := math.Abs(float64(c.pos.Y - to.(*WorldMapCell).pos.Y))
	distance := math.Sqrt(dx*dx + dy*dy)
	return distance * float64(TERRAIN_COSTS[to.(*WorldMapCell).kind])
}

func (c *WorldMapCell) PathEstimatedCost(to astar.Pather) float64 {
	dx := math.Abs(float64(c.pos.X - to.(*WorldMapCell).pos.X))
	dy := math.Abs(float64(c.pos.Y - to.(*WorldMapCell).pos.Y))
	return dx + dy
}
