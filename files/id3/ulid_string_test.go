package main

import (
	"crypto/rand"
	"testing"

	imdarioULID "github.com/imdario/go-ulid"
	oklogULID "github.com/oklog/ulid"
	pbormanUUID "github.com/pborman/uuid"
)

var (
	UUID        = pbormanUUID.NewRandom()
	UUIDstring  = UUID.String()
	ULIDimdario = imdarioULID.New()
	ULIDoklog   = oklogULID.MustNew(oklogULID.Now(), rand.Reader)
	ULIDstring  = imdarioULID.New().String()
)

func BenchmarkPbormanUUIDToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = UUID.String()
	}
}

func BenchmarkImdarioULIDToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ULIDimdario.String()
	}
}

func BenchmarkOklogULIDToString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = ULIDoklog.String()
	}
}

func BenchmarkPbormanUUIDFromString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		pbormanUUID.Parse(UUIDstring)
	}
}

func BenchmarkOklogUlidFromString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		oklogULID.MustParse(ULIDstring)
	}
}
