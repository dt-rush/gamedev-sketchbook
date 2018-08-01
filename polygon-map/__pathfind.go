package main

func (c *PathCalculator) Path(from, to *Point2D) (
	path []Point2D, distance float64, found bool) {

	// find the starting vertex and the ending vertex, being those that are,
	// from the start and end respectively, next closest to the other.
	// example:
	//
	//     FROM           v1      v2
	//
	//          v4       v5          v6
	//                 v7
	//
	//         v9                v10
	//                 v12
	//
	//
	//                        TO
	//
	//in the above, the starting would be v4, the ending would be v12
	//
	// we can find the nearest vertices for a given vertex, in order to
	// search through them, by using the method described here:
	//
	// https://stackoverflow.com/a/14325990/9715599
	//
	// thus, we're basically searching the "

	heap.Init(c.q)
	var fromNode *PathNode = c.nm[.... seomthing ....]
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
			p := []Point2D{}
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
				// if we're here, this is a valid neighbor Point2D to investigate
				var neighbor = &c.wm.cells[current.cell.pos.Y+iy][current.cell.pos.X+ix]
				var neighborNode *PathNode = c.nm[neighbor.pos.Y][neighbor.pos.X]
				if neighborNode == nil {
					neighborNode = &PathNode{cell: neighbor}
					c.nm[neighbor.pos.Y][neighbor.pos.X] = neighborNode
				}
				dx := current.cell.pos.X - neighbor.pos.X
				dy := current.cell.pos.Y - neighbor.pos.Y
				distance := math.Sqrt(float64(dx*dx + dy*dy))
				costToNeighbor := distance * current.cell.CostToTransitionTo(neighbor)
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
