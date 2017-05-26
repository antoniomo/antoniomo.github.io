package main

import (
	"math/rand"
	"testing"

	pborman "github.com/pborman/uuid"
	satori "github.com/satori/go.uuid"
)

func BenchmarkSatoriV4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		satori.NewV4()
	}
}

func BenchmarkPbormanV4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pborman.NewRandom()
	}
}

func BenchmarkPbormanV4MathRand(b *testing.B) {
	rsource := rand.New(rand.NewSource(1)) // Unsafe concurrent use!
	pborman.SetRand(rsource)
	for i := 0; i < b.N; i++ {
		pborman.NewRandom()
	}
}
