package lduser

import (
	"fmt"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/internal/sharedtest"
)

func BenchmarkUserMinimumAllocationSize(b *testing.B) {
	// This just measures the minimum number of bytes taken up by a User.
	for i := 0; i < b.N; i++ {
		benchmarkUserPointer = makeEmptyUserPointer()
	}
}

func makeEmptyUserPointer() *User {
	return &User{}
}

func BenchmarkUserGetCustomAttrNoAlloc(b *testing.B) {
	for _, n := range []int{sharedtest.SmallNumberOfCustomAttributes, sharedtest.LargeNumberOfCustomAttributes} {
		b.Run(fmt.Sprintf("with %d attributes", n), func(b *testing.B) {
			builder := NewUserBuilder("key")
			attrs := sharedtest.MakeCustomAttributeNamesAndValues(n)
			for _, a := range attrs {
				builder.SetAttribute(UserAttribute(a.Name), a.Value)
			}
			user := builder.Build()
			lastAttr := UserAttribute(attrs[n-1].Name)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkValue = user.GetAttribute(lastAttr)
			}
		})
	}
}
