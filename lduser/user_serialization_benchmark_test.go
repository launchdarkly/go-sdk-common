package lduser

import (
	"encoding/json"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/internal/sharedtest"

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

func doMarshalBenchmark(b *testing.B, marshalFn func(*User) ([]byte, error)) {
	for _, p := range makeBenchmarkMarshalTestParams() {
		b.Run(p.name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if benchmarkJSONResult, benchmarkErr = marshalFn(&p.user); benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func doUnmarshalBenchmark(b *testing.B, unmarshalFn func(*User, []byte) error) {
	for _, p := range sharedtest.MakeOldUserUnmarshalingTestParams() {
		b.Run(p.Name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				if benchmarkErr = unmarshalFn(&benchmarkUserResult, p.Data); benchmarkErr != nil {
					b.Fatal(benchmarkErr)
				}
			}
		})
	}
}

func jsonMarshalTestFn(u *User) ([]byte, error) {
	return json.Marshal(u)
}

func jsonStreamMarshalTestFn(u *User) ([]byte, error) {
	w := jwriter.NewWriter()
	u.WriteToJSONWriter(&w)
	return w.Bytes(), w.Error()
}

func jsonUnmarshalTestFn(u *User, data []byte) error {
	return json.Unmarshal(data, u)
}

func jsonStreamUnmarshalTestFn(u *User, data []byte) error {
	r := jreader.NewReader(data)
	u.ReadFromJSONReader(&r)
	return r.Error()
}

func BenchmarkJSONMarshal(b *testing.B) {
	doMarshalBenchmark(b, jsonMarshalTestFn)
}

func BenchmarkJSONStreamMarshal(b *testing.B) {
	doMarshalBenchmark(b, jsonStreamMarshalTestFn)
}

func BenchmarkJSONUnmarshal(b *testing.B) {
	doUnmarshalBenchmark(b, jsonUnmarshalTestFn)
}

func BenchmarkJSONStreamUnmarshal(b *testing.B) {
	doUnmarshalBenchmark(b, jsonStreamUnmarshalTestFn)
}
