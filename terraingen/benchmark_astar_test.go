package main

import (
	"math/rand"
	"testing"
	"time"
)

func BenchmarkAstar(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	w := NewWorld()
	N := 1024 * 16
	positions := make([]PositionPair, N)
	for i, _ := range positions {
		positions[i] = PositionPair{
			Position{
				rand.Intn(WORLD_CELLWIDTH),
				rand.Intn(WORLD_CELLHEIGHT)},
			Position{
				rand.Intn(WORLD_CELLWIDTH),
				rand.Intn(WORLD_CELLHEIGHT)}}
	}
	b.ResetTimer()
	for i := 0; i < 1024*16; i++ {
		w.e = &Entity{
			pos:        positions[i].p1,
			moveTarget: &positions[i].p2}
		w.ComputeEntityPath()
	}
}

func BenchmarkAstarUnrolled(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	w := NewWorld()
	N := 1024 * 16
	positions := make([]PositionPair, N)
	for i, _ := range positions {
		positions[i] = PositionPair{
			Position{
				rand.Intn(WORLD_CELLWIDTH),
				rand.Intn(WORLD_CELLHEIGHT)},
			Position{
				rand.Intn(WORLD_CELLWIDTH),
				rand.Intn(WORLD_CELLHEIGHT)}}
	}
	b.ResetTimer()
	for i := 0; i < 1024*16; i++ {
		w.e = &Entity{
			pos:        positions[i].p1,
			moveTarget: &positions[i].p2}
		w.ComputeEntityPathUnrolled()
	}
}

func BenchmarkAstarHandRolled(b *testing.B) {
	rand.Seed(time.Now().UnixNano())
	w := NewWorld()
	N := 1024 * 16
	positions := make([]PositionPair, N)
	for i, _ := range positions {
		positions[i] = PositionPair{
			Position{
				rand.Intn(WORLD_CELLWIDTH),
				rand.Intn(WORLD_CELLHEIGHT)},
			Position{
				rand.Intn(WORLD_CELLWIDTH),
				rand.Intn(WORLD_CELLHEIGHT)}}
	}
	b.ResetTimer()
	for i := 0; i < 1024*16; i++ {
		w.e = &Entity{
			pos:        positions[i].p1,
			moveTarget: &positions[i].p2}
		w.ComputeEntityPathHandRolled()
	}
}
