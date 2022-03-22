//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package ldcontext

import (
	"testing"
)

func BenchmarkEasyJSONMarshal(b *testing.B) {
	doMarshalBenchmark(b, easyJSONMarshalTestFn)
}

func BenchmarkEasyJSONUnmarshal(b *testing.B) {
	doUnmarshalBenchmark(b, easyJSONUnmarshalTestFn)
}
