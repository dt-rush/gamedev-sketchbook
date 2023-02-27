package main

import (
	"github.com/aquilax/go-perlin"
)

func PerlinNoiseInt2D(w int, h int, scale float64,
	alpha float64, beta float64, n int, seed int64) [][]float64 {

	p := perlin.NewPerlin(alpha, beta, int32(n), seed)
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
