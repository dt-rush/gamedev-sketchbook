package main

import (
// "math"
)

// G:			reference to the grid of cells we're pathing over
// OH			NodeHeap ("Open Heap") used to pop off the nodes with the lowest
//					F during search (open nodes)
// N: 			incremented each time we calculate (used to avoid having to
//					clear values in various arrays)
//
// WhichList:	2d array shadowing used to keep track of which list, open or
//					closed, the node is on
// From:		link to the prior node in path
// G:	 		path cost
// H:	 		heuristic
// F:			G + H
// HeapIX:		keeps track of the heap index of the element at this position
type PathComputer struct {
	WM *WorldMap
	OH *NodeHeap
	N  int
	// these 2D arrays store info about each node
	WhichList [][]int
	From      [][]Position
	G         [][]int
	H         [][]int
	F         [][]int
	HeapIX    [][]int
}

// special value used for the "From" of the start node
var NOWHERE = Position{-1, -1}

func NewPathComputer(wm *WorldMap) *PathComputer {

	// make 2D array rows
	// NOTE: in array-speak, the "rows" are columns. It's just nicer to put
	// X as the first coordinate instead of Y
	whichList := make([][]int, WORLD_CELLWIDTH)
	from := make([][]Position, WORLD_CELLWIDTH)
	g := make([][]int, WORLD_CELLWIDTH)
	h := make([][]int, WORLD_CELLWIDTH)
	f := make([][]int, WORLD_CELLWIDTH)
	heapIX := make([][]int, WORLD_CELLWIDTH)
	// make 2D array columns
	for x := 0; x < WORLD_CELLWIDTH; x++ {
		whichList[x] = make([]int, WORLD_CELLHEIGHT)
		from[x] = make([]Position, WORLD_CELLHEIGHT)
		g[x] = make([]int, WORLD_CELLHEIGHT)
		h[x] = make([]int, WORLD_CELLHEIGHT)
		f[x] = make([]int, WORLD_CELLHEIGHT)
		heapIX[x] = make([]int, WORLD_CELLHEIGHT)
	}
	// make node heap
	pc := &PathComputer{
		WM:        wm,
		N:         0,
		WhichList: whichList,
		From:      from,
		G:         g,
		H:         h,
		F:         f,
		HeapIX:    heapIX,
	}
	oh := NewNodeHeap(pc)
	pc.OH = oh
	return pc
}

// neighbor x, y offsets
//
//                                       X
//      --------------------------------->
//     |
//     |    -1,  1     0,  1     1,  1
//     |
//     |    -1,  0               1,  0
//     |
//     |    -1, -1     0, -1     1, -1
//     |
//  Y  v
//
//
var neighborIXs = [][2]int{
	[2]int{-1, 1},
	[2]int{1, 1},
	[2]int{-1, -1},
	[2]int{1, -1},
	[2]int{0, 1},
	[2]int{-1, 0},
	[2]int{1, 0},
	[2]int{0, -1},
}

func (pc *PathComputer) Heuristic(p1 Position, p2 Position) int {
	dx := p1.X - p2.X
	dy := p1.Y - p2.Y
	result := 10 * (dx + dy)
	if result < 0 {
		result *= -1
	}
	return result
	// return int(10 * math.Sqrt(float64(dx*dx+dy*dy)))
}

func (pc *PathComputer) Path(start Position, end Position) (path []Position) {
	// clear the heap which contains leftover nodes from the last calculation
	pc.OH.Clear()
	// increment N (easier than clearing arrays)
	// we increment by 2 since we use WhichList == pc.N for OPEN and
	// WhichList == pc.N + 1 for CLOSED
	pc.N += 2

	// add first node to open heap (whichlist == pc.N)
	pc.WhichList[start.X][start.Y] = pc.N
	// store a special value for the "From" of the first node
	pc.From[start.X][start.Y] = NOWHERE
	pc.G[start.X][start.Y] = 0
	pc.H[start.X][start.Y] = pc.Heuristic(start, end)
	pc.WhichList[start.X][start.Y] = pc.N
	pc.OH.Add(start)
	// while open heap has elements...
	for pc.OH.Len() > 0 {
		// pop from open heap
		cur, err := pc.OH.Pop()
		// if err, we have exhausted all squares on open heap and found no path
		// return empty list
		if err != nil {
			return []Position{}
		}
		// set as CLOSED (pc.N + 1)
		pc.WhichList[cur.X][cur.Y] = pc.N + 1
		// if the current cell is the end, we're here. build the return list
		if cur.X == end.X && cur.Y == end.Y {
			path = make([]Position, 0)
			for cur != NOWHERE {
				path = append(path, Position{cur.X, cur.Y})
				cur = pc.From[cur.X][cur.Y]
			}
			// return the path to the user
			return path
		}
		// else, we have yet to complete the path. So:
		// for each neighbor
		for _, neighborIX := range neighborIXs {
			// get the coordinates of the cell we will check the cost to
			// by applying an offset to cur's coordinates
			x := cur.X + neighborIX[0]
			y := cur.Y + neighborIX[1]
			// continue loop to next neighbor early if not in grid
			if !pc.WM.InGrid(x, y) {
				continue
			}
			// dist is an integer expression of the distance from
			// cur to the neighbor cell we're looking at here.
			// if either x or y offset is 0, we're moving straight,
			// so put 10. Otherwise we're moving diagonal, so put 14
			// (these are 1 and sqrt(2), but made into integers for speed)
			var dist int
			if neighborIX[0]*neighborIX[1] == 0 {
				dist = 10
			} else {
				dist = 14
			}
			// multiply distance by terrain cost
			terrainCost := TERRAIN_COSTS[pc.WM.cells[x][y].kind]
			dist = dist * terrainCost

			// compute g, h, f for the current cell
			g := pc.G[cur.X][cur.Y] + dist
			h := pc.Heuristic(Position{x, y}, end)

			// don't consider this neighbor if the neighbor is in the closed
			// list *and* our g is greater or equal to its g score (we already
			// have a better way to get to it)
			closed := pc.WhichList[x][y] == pc.N+1
			if closed && g >= pc.G[x][y] {
				continue
			}

			// if not on open heap, add it with "From" == cur
			open := pc.WhichList[x][y] == pc.N
			if !open {
				// set From, G, H
				pc.From[x][y] = cur
				pc.G[x][y] = g
				pc.H[x][y] = h
				// set whichlist == OPEN
				pc.WhichList[x][y] = pc.N
				// push to open heap
				pc.OH.Add(Position{x, y})
			} else {
				// if it *is* on the open heap already, check to see if
				// this is a better path to that square
				gAlready := pc.G[x][y]
				if g < gAlready {
					// if the open node could be reached better by
					// this path, set the g to the new lower g, set the
					// "From" reference to cur and fix up the heap because
					// we've changed the value of one of its elements
					pc.From[x][y] = cur
					pc.OH.Modify(pc.HeapIX[x][y], g)
				}
			}
		}
	}
	return path
}
