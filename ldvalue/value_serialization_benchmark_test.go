package ldvalue

import (
	"fmt"
	"testing"
)

func makeComplexValue() Value {
	longArray := ArrayBuild()
	for i := 0; i < 50; i++ {
		longArray.Add(String(fmt.Sprintf("value%d", i)))
	}
	return ObjectBuild().
		Set("prop1", String("simple string")).
		Set("prop2", String("string\twith\"escapes\"")).
		Set("prop3", longArray.Build()).
		Set("prop4", ObjectBuild().Set("sub1", Int(1)).Set("sub2", Int(2)).Build()).
		Build()
}

func BenchmarkSerializeComplexValue(b *testing.B) {
	value := makeComplexValue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkJSONResult, benchmarkErrResult = value.MarshalJSON()
	}
}

func BenchmarkDeserializeComplexValue(b *testing.B) {
	data, _ := makeComplexValue().MarshalJSON()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkValueResult.UnmarshalJSON(data)
	}
}
