package ldcontext

import (
	"fmt"
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/internal/sharedtest"
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

// If a benchmark's name ends in NoAlloc, our CI will enforce that it does not cause any heap allocations.

func BenchmarkContextMinimumAllocationSize(b *testing.B) {
	// This just measures the minimum number of bytes taken up by a Context.
	for i := 0; i < b.N; i++ {
		benchmarkContextPtr = makeEmptyContextPointer()
	}
}

func makeEmptyContextPointer() *Context {
	return &Context{}
}

func BenchmarkContextGetCustomAttrNoAlloc(b *testing.B) {
	for _, n := range []int{sharedtest.SmallNumberOfCustomAttributes, sharedtest.LargeNumberOfCustomAttributes} {
		b.Run(fmt.Sprintf("with %d attributes", n), func(b *testing.B) {
			builder := NewBuilder("key")
			attrs := sharedtest.MakeCustomAttributeNamesAndValues(n)
			for _, a := range attrs {
				builder.SetValue(a.Name, a.Value)
			}
			c := builder.Build()
			lastAttr := attrs[n-1].Name
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				benchmarkValue = c.GetValue(lastAttr)
			}
		})
	}
}

func BenchmarkContextGetCustomAttrNestedPropertyNoAlloc(b *testing.B) {
	targetValue := ldvalue.String("17 Highbrow Street")
	objectValue := ldvalue.ObjectBuild().Set("street", ldvalue.ObjectBuild().Set("line1", targetValue).Build()).Build()
	c := makeBasicBuilder().SetValue("address", objectValue)
	attrRef := ldattr.NewRef("/address/street/line1")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkValue = c.Build().GetValueForRef(attrRef)
	}
}
