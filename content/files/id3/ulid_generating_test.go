package main

import (
	crand "crypto/rand"
	mrand "math/rand"
	"testing"

	imdarioULID "github.com/imdario/go-ulid"
	oklogULID "github.com/oklog/ulid"
	pbormanUUID "github.com/pborman/uuid"
)

func BenchmarkPbormanUUIDV4(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pbormanUUID.NewRandom()
	}
}

func BenchmarkImdarioULID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		imdarioULID.New()
	}
}

func BenchmarkOklogULID(b *testing.B) {
	for i := 0; i < b.N; i++ {
		oklogULID.MustNew(oklogULID.Now(), crand.Reader)
	}
}

func BenchmarkPbormanUUIDV4MathRand(b *testing.B) {
	rsource := mrand.New(mrand.NewSource(1)) // Unsafe concurrent use!
	pbormanUUID.SetRand(rsource)
	for i := 0; i < b.N; i++ {
		pbormanUUID.NewRandom()
	}
}

func BenchmarkOklogULIDMathRand(b *testing.B) {
	rsource := mrand.New(mrand.NewSource(1)) // Unsafe concurrent use!
	for i := 0; i < b.N; i++ {
		oklogULID.MustNew(oklogULID.Now(), rsource)
	}
}
