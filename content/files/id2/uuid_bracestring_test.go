package main

import (
	"testing"

	pborman "github.com/pborman/uuid"
	satori "github.com/satori/go.uuid"
)

var (
	UUIDstring = "{" + pborman.New() + "}"
)

func withBraceParse(s string) pborman.UUID {
	if s[0] == '{' {
		s = s[1 : len(s)-1]
	}
	return pborman.Parse(s)
}

func BenchmarkSatoriFromBraceString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		satori.FromString(UUIDstring)
	}
}

func BenchmarkPbormanFromBraceString(b *testing.B) {
	for i := 0; i < b.N; i++ {
		withBraceParse(UUIDstring)
	}
}
