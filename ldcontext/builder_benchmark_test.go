package ldcontext

import (
	"fmt"
	"testing"

	"github.com/launchdarkly/go-sdk-common/v4/internal/sharedtest"
)

// If a benchmark's name ends in NoAlloc, our CI will enforce that it does not cause any heap allocations.

func BenchmarkNewNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkContext = New("key")
	}
}

func BenchmarkBuildFromLocalBuilderNoCustomAttrsNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		var b Builder
		b.Key("key")
		b.Name("x")
		benchmarkContext = b.Build()
	}
}

func BenchmarkBuildWithNoCustomAttrs(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkContext = NewBuilder("key").Name("x").Build()
	}
}

func BenchmarkBuildWithCustomAttributes(b *testing.B) {
	for _, n := range []int{
		sharedtest.SmallNumberOfCustomAttributes,
		sharedtest.MiddleNumberOfCustomAttributes,
		sharedtest.LargeNumberOfCustomAttributes,
	} {
		b.Run(fmt.Sprintf("with %d attributes", n), func(b *testing.B) {
			attrs := sharedtest.MakeCustomAttributeNamesAndValues(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				builder := NewBuilder("key")
				for _, a := range attrs {
					builder.SetValue(a.Name, a.Value)
				}
				benchmarkContext = builder.Build()
			}
		})
	}
}

func BenchmarkBuildWithCustomAttributesWithCapacity(b *testing.B) {
	for _, n := range []int{
		sharedtest.SmallNumberOfCustomAttributes,
		sharedtest.MiddleNumberOfCustomAttributes,
		sharedtest.LargeNumberOfCustomAttributes,
	} {
		b.Run(fmt.Sprintf("with %d attributes", n), func(b *testing.B) {
			attrs := sharedtest.MakeCustomAttributeNamesAndValues(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				builder := NewBuilderWithCapacity("key", n)
				for _, a := range attrs {
					builder.SetValue(a.Name, a.Value)
				}
				benchmarkContext = builder.Build()
			}
		})
	}
}

func BenchmarkBuildWithPrivate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkContext = NewBuilder("key").Name("name").Private("name").Build()
	}
}
