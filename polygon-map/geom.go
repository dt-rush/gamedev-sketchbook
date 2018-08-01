package main

type Polygon []Point2D

// taken from:
// github.com/soniakeys/raycast
func rayIntersectsSegment(p, a, b Point2D) bool {
	return (a.Y > p.Y) != (b.Y > p.Y) &&
		p.X < (b.X-a.X)*(p.Y-a.Y)/(b.Y-a.Y)+a.X
}

// taken from:
// github.com/soniakeys/raycast
func Point2DInPolygon(pt Point2D, pg Polygon) bool {
	if len(pg) < 3 {
		return false
	}
	a := pg[0]
	in := rayIntersectsSegment(pt, pg[len(pg)-1], a)
	for _, b := range pg[1:] {
		if rayIntersectsSegment(pt, a, b) {
			in = !in
		}
		a = b
	}
	return in
}
