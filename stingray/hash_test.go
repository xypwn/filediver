package stingray_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xypwn/filediver/stingray"
)

func BenchmarkMurmurHash64a(b *testing.B) {
	s := "hello, world!"
	var h stingray.Hash
	for b.Loop() {
		h = stingray.Sum(s)
	}
	if h.Value != 0xd18abe154a2a9637 {
		b.Fatal("invalid hash result")
	}
}

func TestMurmurHash64a(t *testing.T) {
	require := require.New(t)
	require.Equal(uint64(0xd18abe154a2a9637), stingray.Sum("hello, world!").Value)
	require.Equal(uint64(0x2f9dea7bd39b9f7d), stingray.Sum("hello, world!!!!").Value) // length divisible by 8
}
