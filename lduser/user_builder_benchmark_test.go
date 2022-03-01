package lduser

import (
	"fmt"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/internal/sharedtest"
)

func BenchmarkNewUserNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = NewUser("key")
	}
}

func BenchmarkNewAnonymousUserNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = NewAnonymousUser("key")
	}
}

func BenchmarkNewUserFromBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = NewUserBuilder("key").Name("name").Email("email").Build()
	}
}

func BenchmarkBuildWithCustomAttributes(b *testing.B) {
	for _, n := range []int{sharedtest.SmallNumberOfCustomAttributes, sharedtest.LargeNumberOfCustomAttributes} {
		b.Run(fmt.Sprintf("with %d attributes", n), func(b *testing.B) {
			attrs := sharedtest.MakeCustomAttributeNamesAndValues(n)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				builder := NewUserBuilder("key")
				for _, a := range attrs {
					builder.SetAttribute(UserAttribute(a.Name), a.Value)
				}
				benchmarkUserResult = builder.Build()
			}
		})
	}
}

func BenchmarkNewUserFromBuilderWithPrivate(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = NewUserBuilder("key").Name("name").AsPrivateAttribute().Build()
	}
}

func BenchmarkBuildUserWithAllScalarAttributesNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserResult = benchmarkUserBuilderWithAllScalarAttributes.Build()
	}
}

func BenchmarkUserBuilderSetScalarAttributesNoAlloc(b *testing.B) {
	for i := 0; i < b.N; i++ {
		benchmarkUserBuilder.
			Secondary("s").
			IP("i").
			Country("c").
			Email("e").
			FirstName("f").
			LastName("l").
			Avatar("a").
			Name("n").
			Anonymous(true)
	}
}
