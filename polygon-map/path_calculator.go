package main

// import "container/heap"
// import "math"

type PathCalculator struct {
	wm *WorldMap
	q  *PathNodePQueue
	nm []*PathNode
}

func NewPathCalculator(wm *WorldMap) *PathCalculator {
	// note: we estimate the needed capacity for the
	// path node p-queue by multiplying the number of
	// lakes by 16, assuming there are 16 vertices per
	// lake. The slice will grow if there's more
	// anyway, so at most this will affect the very
	// first computation
	c := PathCalculator{
		wm: wm,
		q:  NewPathNodePQueue(16 * len(wm.Lakes)),
		nm: make([]*PathNode, len(wm.Vertices))}
	return &c
}

func (c *PathCalculator) Clear() {
	c.q.Clear()
	for i := 0; i < len(c.nm); i++ {
		c.nm[i] = nil
	}
}

// TODO: placeholder. We have a few subproblems to solve first, such as:
// - rock generation (4 to 7 points arranged a certain mean distance
//    from a center)
// - testing if a line interesects the convex hull of a set of points
// - building the line-of-sight neighbours of all vertexes given that certain
//    of them are part of convex hulls which block line-of-sight
func (c *PathCalculator) Path(from, to *Point2D) (
	path []Point2D, distance float64, found bool) {
	return nil, 0, false
}
