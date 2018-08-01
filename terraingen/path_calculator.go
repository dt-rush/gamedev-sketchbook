package main

import "container/heap"
import "math"

type PathCalculator struct {
	wm *WorldMap
	q  *PathNodePQueue
	nm [WORLD_CELLHEIGHT][WORLD_CELLWIDTH]*PathNode
}

func NewPathCalculator(wm *WorldMap) *PathCalculator {
	c := PathCalculator{
		wm: wm,
		q:  NewPathNodePQueue(WORLD_CELLWIDTH * WORLD_CELLHEIGHT)}
	return &c
}

func (c *PathCalculator) Clear() {
	c.q.Clear()
	for y := 0; y < WORLD_CELLHEIGHT; y++ {
		for x := 0; x < WORLD_CELLHEIGHT; x++ {
			c.nm[y][x] = nil
		}
	}
}

func (c *PathCalculator) Path(from, to *WorldMapCell) (
	path []Position, distance float64, found bool) {

	heap.Init(c.q)
	var fromNode *PathNode = c.nm[from.pos.Y][from.pos.X]
	if fromNode == nil {
		fromNode = &PathNode{cell: from}
		c.nm[from.pos.Y][from.pos.X] = fromNode
	}
	fromNode.open = true
	heap.Push(c.q, fromNode)
	for {
		if c.q.Len() == 0 {
			// There's no path, return found false.
			return
		}
		current := heap.Pop(c.q).(*PathNode)
		current.open = false
		current.closed = true

		if current.cell.pos.X == to.pos.X &&
			current.cell.pos.Y == to.pos.Y {
			// Found a path to the goal.
			p := []Position{}
			curr := current
			for curr != nil {
				p = append(p, curr.cell.pos)
				curr = curr.parent
			}
			return p, current.cost, true
		}
		for iy := -1; iy <= 1; iy++ {
			if current.cell.pos.Y+iy < 0 ||
				current.cell.pos.Y+iy > WORLD_CELLHEIGHT-1 {
				continue
			}
			for ix := -1; ix <= 1; ix++ {
				if current.cell.pos.X+ix < 0 ||
					current.cell.pos.X+ix > WORLD_CELLWIDTH-1 {
					continue
				}
				// if we're here, this is a valid neighbor position to investigate
				var neighbor = &c.wm.cells[current.cell.pos.Y+iy][current.cell.pos.X+ix]
				var neighborNode *PathNode = c.nm[neighbor.pos.Y][neighbor.pos.X]
				if neighborNode == nil {
					neighborNode = &PathNode{cell: neighbor}
					c.nm[neighbor.pos.Y][neighbor.pos.X] = neighborNode
				}
				dx := current.cell.pos.X - neighbor.pos.X
				dy := current.cell.pos.Y - neighbor.pos.Y
				distance := math.Sqrt(float64(dx*dx + dy*dy))
				terrainCost := float64(TERRAIN_COSTS[neighbor.kind])
				costToNeighbor := distance * terrainCost
				cost := current.cost + costToNeighbor
				if cost < neighborNode.cost {
					if neighborNode.open {
						heap.Remove(c.q, neighborNode.index)
					}
					neighborNode.open = false
					neighborNode.closed = false
				}
				if !neighborNode.open && !neighborNode.closed {
					neighborNode.cost = cost
					neighborNode.open = true
					heuristic := (math.Abs(float64(neighbor.pos.X-to.pos.X)) +
						math.Abs(float64(neighbor.pos.Y-to.pos.Y)))
					neighborNode.rank = cost + heuristic
					neighborNode.parent = current
					heap.Push(c.q, neighborNode)
				}
			}
		}
	}

}
