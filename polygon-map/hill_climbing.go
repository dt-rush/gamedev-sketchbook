package main

import (
	"fmt"
)

func (wm *WorldMap) nextHighest(
	p Point2D, scan int) *Point2D {

	var ascendPoint Point2D
	highest := -1e9
	step := 1
	for iy := -scan * step; iy <= scan*step; iy += step {
		if iy == 0 {
			continue
		}
		py := int(float64(p.Y+iy) / PSCALE)
		if !(py > 0 && py < PH-1) {
			continue
		}
		for ix := -scan * step; ix <= scan*step; ix += step {
			if ix == 0 {
				continue
			}
			px := int(float64(p.X+ix) / PSCALE)
			if !(px > 0 && px < PW-1) {
				continue
			}
			// look at the value here are if higher than max found among
			// considered points, keep track of it
			pval := wm.perlin[py][px]
			if pval > highest {
				highest = pval
				ascendPoint = Point2D{p.X + ix, p.Y + iy}
			}
		}
	}
	if highest == -1e9 {
		return nil
	}
	fmt.Printf("%v->%v\n", p, ascendPoint)
	return &ascendPoint
}

func (wm *WorldMap) highestAwayFromSource(
	p Point2D, source Point2D, scan int) *Point2D {

	var ascendPoint Point2D
	highest := -1e9
	step := 1
	for iy := -scan * step; iy <= scan*step; iy += step {
		py := int(float64(p.Y+iy) / PSCALE)
		if !(py > 0 && py < PH-1) {
			continue
		}
		for ix := -scan * step; ix <= scan*step; ix += step {
			px := int(float64(p.X+ix) / PSCALE)
			if !(px > 0 && px < PW-1) {
				continue
			}
			// if the point is closer to the center of the lake than
			// where we currently are, don't consider it
			if (source.X-p.X)*(source.X-p.X)+
				(source.Y-p.Y)*(source.Y-p.Y) >
				(source.X-(p.X+ix))*(source.X-(p.X+ix))+
					(source.Y-(p.Y+iy))*(source.Y-(p.Y+iy)) {
				continue
			}
			// look at the value here are if higher than max found among
			// considered points, keep track of it
			pval := wm.perlin[py][px]
			if pval > highest {
				highest = pval
				ascendPoint = Point2D{p.X + ix, p.Y + iy}
			}
		}
	}
	if highest == -1e9 {
		return nil
	}
	return &ascendPoint
}
