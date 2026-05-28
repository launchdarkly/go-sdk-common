package ldcontext

import (
	"testing"
)

// BenchmarkJSONUnmarshal uses json.Unmarshal; BenchmarkJSONStreamUnmarshal uses the jsonstream API via
// ReadFromJSONReader. They both end up calling the same underlying logic, but json.Unmarshal has some
// extra indirection.

func doUnmarshalBenchmark(b *testing.B, unmarshalFn func(*Context, []byte) error) {
	for _, p := range makeBenchmarkUnmarshalTestParams() {
		b.Run(p.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if benchmarkErr = unmarshalFn(&benchmarkContext, p.Data); benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	doUnmarshalBenchmark(b, jsonUnmarshalTestFn)
}

func BenchmarkJSONStreamUnmarshal(b *testing.B) {
	doUnmarshalBenchmark(b, jsonStreamUnmarshalTestFn)
}
