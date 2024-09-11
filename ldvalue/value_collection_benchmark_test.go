package ldvalue

import (
	"math/rand"
	"testing"
)

func BenchmarkCollectionCopyMapStringSmall(b *testing.B) {
	input := generateRandomMap(10)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkCollectionCopyMapStringLarge(b *testing.B) {
	input := generateRandomMap(1_000)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func generateRandomMap(i int) map[string]string {
	m := make(map[string]string, i)

	for j := 0; j < i; j++ {
		m[generateRandomString()] = generateRandomString()
	}

	return m
}

func BenchmarkCollectionCopySliceStringSmall(b *testing.B) {
	input := generateStringSlice(10)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkCollectionCopySliceStringMedium(b *testing.B) {
	input := generateStringSlice(100)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkCollectionCopySliceStringLarge(b *testing.B) {
	input := generateStringSlice(1_000)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkCollectionCopySliceIntSmall(b *testing.B) {
	input := generateIntSlice(10)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkCollectionCopySliceIntMedium(b *testing.B) {
	input := generateIntSlice(100)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkCollectionCopySliceIntLarge(b *testing.B) {
	input := generateIntSlice(1_000)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func generateIntSlice(n int) []int {
	s := make([]int, n)

	for i := 0; i < n; i++ {
		s[i] = rand.Int()
	}

	return s
}

func generateStringSlice(n int) []string {
	s := make([]string, n)

	for i := 0; i < n; i++ {
		s[i] = generateRandomString()
	}

	return s
}

func generateRandomString() string {
	alpha := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, 10)
	for i := range b {
		b[i] = alpha[rand.Intn(len(alpha))]
	}

	return string(b)
}
