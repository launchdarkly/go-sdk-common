package ldvalue

import (
	"math/rand"
	"testing"
)

func BenchmarkReflectCopyMapStringSmall(b *testing.B) {
	input := generateRandomMap(10)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkReflectCopyMapStringLarge(b *testing.B) {
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

func BenchmarkReflectCopySliceStringSmall(b *testing.B) {
	input := generateStringSlice(10)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkReflectCopySliceStringMedium(b *testing.B) {
	input := generateStringSlice(100)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkReflectCopySliceStringLarge(b *testing.B) {
	input := generateStringSlice(1_000)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkReflectCopySliceIntSmall(b *testing.B) {
	input := generateIntSlice(10)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkReflectCopySliceIntMedium(b *testing.B) {
	input := generateIntSlice(100)

	for i := 0; i < b.N; i++ {
		_ = CopyArbitraryValue(input)
	}
}

func BenchmarkReflectCopySliceIntLarge(b *testing.B) {
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
