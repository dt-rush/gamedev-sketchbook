package main

import (
	"github.com/aquilax/go-perlin"
)

func PerlinNoiseInt2D(w int, h int, scale float64,
	alpha float64, beta float64, n int, seed int64) [][]float64 {

	p := perlin.NewPerlin(alpha, beta, n, seed)
	output := make([][]float64, h)
	for y := 0; y < h; y++ {
		output[y] = make([]float64, w)
		for x := 0; x < w; x++ {
			output[y][x] = 1 + p.Noise2D(float64(y)/scale, float64(x)/scale)
		}
	}
	return output
}

func OpPerlins(a [][]float64, b [][]float64,
	op func(a float64, b float64) float64) [][]float64 {
	// NOTE: this will blow up or quietly do weird stuff
	// if they aren't the same dimensions
	sum := make([][]float64, len(a))
	for y := 0; y < len(a); y++ {
		sum[y] = make([]float64, len(a[y]))
		for x := 0; x < len(a); x++ {
			sum[y][x] = op(a[y][x], b[y][x])
		}
	}
	return sum
}

func SmoothPerlin(p [][]float64) [][]float64 {
	smoothed := make([][]float64, len(p))
	for y := 0; y < PH; y++ {
		smoothed[y] = make([]float64, PW)
		for x := 0; x < PW; x++ {
			smoothed[y][x] = AvgOfNeighbors(x, y, p)
		}
	}
	return smoothed
}

func AvgOfNeighbors(x int, y int, p [][]float64) float64 {
	kernel := [][]float64{
		[]float64{1 / 16.0, 1 / 8.0, 1 / 16.0},
		[]float64{1 / 8.0, 1 / 4.0, 1 / 8.0},
		[]float64{1 / 16.0, 1 / 8.0, 1 / 16.0},
	}
	var sum float64
	var wsum float64
	for iy := -1; iy <= 1; iy++ {
		if y+iy < 0 || y+iy > PH-1 {
			continue
		}
		for ix := 0; ix <= 1; ix++ {
			if x+ix < 0 || x+ix > PW-1 {
				continue
			}
			w := kernel[1+iy][1+ix]
			sum += w * p[y+iy][x+ix]
			wsum += w
		}
	}
	return sum / wsum
}
