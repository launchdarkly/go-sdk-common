//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package lduser

import (
	"testing"

	easyjson "github.com/mailru/easyjson"
)

func easyJSONMarshalTestFn(u *User) ([]byte, error) {
	return easyjson.Marshal(u)
}

func easyJSONUnmarshalTestFn(u *User, data []byte) error {
	return easyjson.Unmarshal(data, u)
}

func BenchmarkEasyJSONMarshal(b *testing.B) {
	doMarshalBenchmark(b, easyJSONMarshalTestFn)
}

func BenchmarkEasyJSONUnmarshal(b *testing.B) {
	doUnmarshalBenchmark(b, easyJSONUnmarshalTestFn)
}
