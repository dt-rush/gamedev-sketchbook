package main

// a golang transcription of https://github.com/jsmolka/hull

import (
	"fmt"
	"math"
	"sort"
)

// distance
func dist(a [2]float64, b [2]float64) float64 {
	dx := b[0] - a[0]
	dy := b[1] - a[1]
	return math.Sqrt(dx*dx + dy*dy)
}

// k nearest neighbors
func knn(ps [][2]float64, p [2]float64, k int) [][2]float64 {
	cpy := make([][2]float64, len(ps))
	copy(cpy, ps)
	sort.Slice(cpy,
		func(i int, j int) bool {
			return dist(cpy[i], p) < dist(cpy[j], p)
		})
	if len(cpy) < 3 {
		return cpy
	}
	return cpy[0:k]
}

// checks if lines p1 -> p2, p3 -> p4 intersect
func intersects(p1 [2]float64, p2 [2]float64, p3 [2]float64, p4 [2]float64) bool {
	p0_x, p0_y := p1[0], p1[1]
	p1_x, p1_y := p2[0], p2[1]
	p2_x, p2_y := p3[0], p3[1]
	p3_x, p3_y := p4[0], p4[1]

	s10_x := p1_x - p0_x
	s10_y := p1_y - p0_y
	s32_x := p3_x - p2_x
	s32_y := p3_y - p2_y

	denom := s10_x*s32_y - s32_x*s10_y
	if denom == 0 {
		return false
	}

	denom_positive := denom < 0
	s02_x := p0_x - p2_x
	s02_y := p0_y - p2_y
	s_numer := s10_x*s02_y - s10_y*s02_x
	if s_numer < 0 == denom_positive {
		return false
	}

	t_numer := s32_x*s02_y - s32_y*s02_x
	if t_numer < 0 == denom_positive {
		return false
	}

	if (s_numer > denom) == denom_positive ||
		(t_numer > denom) == denom_positive {
		return false
	}

	t := t_numer / denom
	x := p0_x + (t * s10_x)
	y := p0_y + (t * s10_y)
	xy := [2]float64{x, y}

	if xy == p1 || xy == p2 || xy == p3 || xy == p4 {
		return false
	}

	return true
}

// calculate the angle between two pints and a previous angle
func angle(p1 [2]float64, p2 [2]float64, previousAngle float64) float64 {
	return math.Mod(
		(math.Atan2(p1[1]-p2[1], p1[0]-p2[0])-previousAngle),
		(math.Pi*2)) - math.Pi
}

// checks if point is in polygon
func pointInPolygon(point [2]float64, polygon [][2]float64) bool {
	size := len(polygon)
	for i := 0; i < size; i++ {
		min_ := math.Min(polygon[i][0], polygon[(i+1)%size][0])
		max_ := math.Max(polygon[i][0], polygon[(i+1)%size][0])
		if min_ < point[0] && point[0] <= max_ {
			p := polygon[i][1] - polygon[(i+1)%size][1]
			q := polygon[i][0] - polygon[(i+1)%size][0]
			point_y := (point[0]-polygon[i][0])*p/q + polygon[i][1]
			if point_y < point[1] {
				return true
			}
		}
	}
	return false
}

// Calculates the concave hull for given points
// Input is a list of 2D points [(x, y), ...]
// k defines the number of of considered neighbours
func concave(points [][2]float64, k int) [][2]float64 {
	// make sure k >= 3
	if k < 3 {
		k = 3
	}
	// remove duplicates
	dataset := make([][2]float64, 0)
	for i, p := range points {
		unique := true
		for j, p2 := range points {
			if i == j {
				continue
			}
			if p == p2 {
				unique = false
			}
		}
		if unique {
			dataset = append(dataset, p)
		}
	}
	// check edge cases
	if len(dataset) <= 3 {
		return dataset
	}
	// make sure k neighbors can be found
	if len(dataset)-1 < k {
		k = len(dataset) - 1
	}
	// set up state
	firstPoint := dataset[0]
	// python has more clean way to express the below:
	// first_point = min(dataset, key=lambda x: x[1])
	for _, p := range dataset {
		if p[0] < firstPoint[0] {
			firstPoint = p
		}
	}
	currentPoint := firstPoint
	// initialize hull with first point
	hull := [][2]float64{firstPoint}
	fmt.Println(hull)
	removePoint(&dataset, firstPoint)
	previousAngle := 0.0
	for (currentPoint != firstPoint || len(hull) == 1) && len(dataset) > 0 {
		if len(hull) == 3 {
			// add first point again
			dataset = append(dataset, firstPoint)
		}
		knPoints := knn(dataset, currentPoint, k)
		cPoints := make([][2]float64, len(knPoints))
		copy(cPoints, knPoints)
		sort.Slice(cPoints, func(i int, j int) bool {
			ai := -angle(cPoints[i], currentPoint, previousAngle)
			aj := -angle(cPoints[j], currentPoint, previousAngle)
			return ai < aj
		})
		its := true
		i := -1
		for its && i < len(cPoints)-1 {
			i += 1
			var lastPoint int
			if cPoints[i] == firstPoint {
				lastPoint = 1
			} else {
				lastPoint = 0
			}
			j := 1
			its = false
			for !its && j < len(hull)-lastPoint {
				its = intersects(
					hull[len(hull)-1],
					cPoints[i],
					hull[len(hull)-j-1],
					hull[len(hull)-j])
				j += 1
			}
		}
		// all points intersect
		// try again with higher number of neighbors
		if its {
			return concave(points, k+1)
		}
		previousAngle = angle(cPoints[i], currentPoint, 0)
		currentPoint = cPoints[i]
		hull = append(hull, currentPoint) // valid candidate was found
		removePoint(&dataset, currentPoint)
	}
	return hull
}

// utility function to remove element from slice
// assumes element is in slice
func removePoint(ps *[][2]float64, p [2]float64) {
	for i, p2 := range *ps {
		if p2 == p {
			(*ps) = append((*ps)[:i], (*ps)[i+1:]...)
			return
		}
	}
}
