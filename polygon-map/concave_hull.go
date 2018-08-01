// taken from https://github.com/furstenheim/ConcaveHull/
package main

/**
Golang implementation of https://github.com/skipperkongen/jts-algorithm-pack/blob/master/src/org/geodelivery/jap/concavehull/SnapHull.java
which is a Java port of st_concavehull from Postgis 2.0
*/

import (
	"github.com/furstenheim/SimpleRTree"
	"github.com/furstenheim/go-convex-hull-2d"
	"github.com/paulmach/go.geo"
	"github.com/paulmach/go.geo/reducers"
	"math"
	"sort"
	"sync"
)

type convexHullFlatPoints FlatPoints

type lexSorter FlatPoints

func (s lexSorter) Less(i, j int) bool {
	if s[2*i] < s[2*j] {
		return true
	}
	if s[2*i] > s[2*j] {
		return false
	}

	if s[2*i+1] < s[2*j+1] {
		return true
	}

	if s[2*i+1] > s[2*j+1] {
		return false
	}
	return true
}

func (s lexSorter) Len() int {
	return len(s) / 2
}

func (s lexSorter) Swap(i, j int) {
	s[2*i], s[2*i+1], s[2*j], s[2*j+1] = s[2*j], s[2*j+1], s[2*i], s[2*i+1]
}

const DEFAULT_SEGLENGTH = 0.001

type concaver struct {
	rtree     *SimpleRTree.SimpleRTree
	seglength float64
}

func Compute(points FlatPoints) (concaveHull FlatPoints) {
	sort.Sort(lexSorter(points))
	return ComputeFromSorted(points)
}

// Compute concave hull from sorted points. Points are expected to be sorted lexicographically by (x,y)
func ComputeFromSorted(points FlatPoints) (concaveHull FlatPoints) {
	// Create a copy so that convex hull and index can modify the array in different ways
	pointsCopy := make(FlatPoints, 0, len(points))
	pointsCopy = append(pointsCopy, points...)
	rtree := SimpleRTree.New()
	var wg sync.WaitGroup
	wg.Add(2)
	// Convex hull
	go func() {
		points = go_convex_hull_2d.NewFromSortedArray(points).(FlatPoints)
		wg.Done()
	}()

	func() {
		rtree.LoadSortedArray(SimpleRTree.FlatPoints(pointsCopy))
		wg.Done()
	}()
	wg.Wait()
	var c concaver
	c.seglength = DEFAULT_SEGLENGTH
	c.rtree = rtree
	return c.computeFromSorted(points)
}

func (c *concaver) computeFromSorted(convexHull FlatPoints) (concaveHull FlatPoints) {
	// degerated case
	if convexHull.Len() < 3 {
		return convexHull
	}
	concaveHull = make([]float64, 0, 2*convexHull.Len())
	x0, y0 := convexHull.Take(0)
	concaveHull = append(concaveHull, x0, y0)
	for i := 0; i < convexHull.Len(); i++ {
		x1, y1 := convexHull.Take(i)
		var x2, y2 float64
		if i == convexHull.Len()-1 {
			x2, y2 = convexHull.Take(0)
		} else {
			x2, y2 = convexHull.Take(i + 1)
		}
		sideSplit := c.segmentize(x1, y1, x2, y2)
		concaveHull = append(concaveHull, sideSplit...)
	}
	path := reducers.DouglasPeucker(geo.NewPathFromFlatXYData(concaveHull), c.seglength)
	// reused allocated array
	concaveHull = concaveHull[0:0]
	reducedPoints := path.Points()

	for _, p := range reducedPoints {
		concaveHull = append(concaveHull, p.Lng(), p.Lat())
	}
	return concaveHull
}

// Split side in small edges, for each edge find closest point. Remove duplicates
func (c *concaver) segmentize(x1, y1, x2, y2 float64) (points []float64) {
	dist := math.Sqrt((x1-x2)*(x1-x2) + (y1-y2)*(y1-y2))
	nSegments := math.Ceil(dist / c.seglength)
	factor := 1 / nSegments
	flatPoints := make([]float64, 0, int(2*nSegments))
	vX := factor * (x2 - x1)
	vY := factor * (y2 - y1)

	closestPoints := make(map[int][2]float64)
	closestPoints[0] = [2]float64{x1, y1}
	closestPoints[int(nSegments)] = [2]float64{x2, y2}

	if nSegments > 1 {
		stack := make([]searchItem, 0)
		stack = append(stack, searchItem{left: 0, right: int(nSegments), lastLeft: 0, lastRight: int(nSegments)})
		for len(stack) > 0 {
			var item searchItem
			item, stack = stack[len(stack)-1], stack[:len(stack)-1]
			index := (item.left + item.right) / 2
			currentX := x1 + vX*float64(index)
			currentY := y1 + vY*float64(index)
			x, y, _, _ := c.rtree.FindNearestPoint(currentX, currentY)
			isNewLeft := x != closestPoints[item.lastLeft][0] || y != closestPoints[item.lastLeft][1]
			isNewRight := x != closestPoints[item.lastRight][0] || y != closestPoints[item.lastRight][1]

			// we don't know the point
			if isNewLeft && isNewRight {
				closestPoints[index] = [2]float64{x, y}
				if index-item.left > 1 {
					stack = append(stack, searchItem{left: item.left, right: index, lastLeft: item.lastLeft, lastRight: index})
				}
				if item.right-index > 1 {
					stack = append(stack, searchItem{left: index, right: item.right, lastLeft: index, lastRight: item.lastRight})
				}
			} else if isNewLeft {
				if index-item.left > 1 {
					stack = append(stack, searchItem{left: item.left, right: index, lastLeft: item.lastLeft, lastRight: item.lastRight})
				}
			} else if isNewRight {
				// don't add point to closest points, but we need to keep looking on the right side
				if item.right-index > 1 {
					stack = append(stack, searchItem{left: index, right: item.right, lastLeft: item.lastLeft, lastRight: item.lastRight})
				}
			}
		}
	}
	// always add last point of the segment
	for i := 1; i <= int(nSegments); i++ {
		point, ok := closestPoints[i]
		if ok {
			flatPoints = append(flatPoints, point[0], point[1])
		}
	}
	return flatPoints
}

type searchItem struct {
	left, right, lastLeft, lastRight int
}

type FlatPoints []float64

func (fp FlatPoints) Len() int {
	return len(fp) / 2
}

func (fp FlatPoints) Slice(i, j int) go_convex_hull_2d.Interface {
	return fp[2*i : 2*j]
}

func (fp FlatPoints) Swap(i, j int) {
	fp[2*i], fp[2*i+1], fp[2*j], fp[2*j+1] = fp[2*j], fp[2*j+1], fp[2*i], fp[2*i+1]
}

func (fp FlatPoints) Take(i int) (x1, y1 float64) {
	return fp[2*i], fp[2*i+1]
}
