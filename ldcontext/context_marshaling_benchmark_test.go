package ldcontext

import (
	"testing"
)

// BenchmarkJSONMarshal uses json.Marshal; BenchmarkJSONStreamMarshal uses the jsonstream API via
// WriteToJSONWriter. They both end up calling the same underlying logic, but json.Marshal has some
// extra indirection.
//
// Marshaling via EasyJSON is covered in the conditionally-compiled file context_easyjson_benchmark_test.go.

func doMarshalBenchmark(b *testing.B, marshalFn func(*Context) ([]byte, error)) {
	for _, p := range makeBenchmarkMarshalTestParams() {
		b.Run(p.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if benchmarkJSONResult, benchmarkErr = marshalFn(&p.context); benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	doMarshalBenchmark(b, jsonMarshalTestFn)
}

func BenchmarkJSONStreamMarshal(b *testing.B) {
	doMarshalBenchmark(b, jsonStreamMarshalTestFn)
}
