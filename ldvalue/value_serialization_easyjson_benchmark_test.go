// +build launchdarkly_easyjson

package ldvalue

import (
	"testing"

	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

func BenchmarkSerializeComplexValueEasyJSON(b *testing.B) {
	value := makeComplexValue()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer := jwriter.Writer{}
		value.MarshalEasyJSON(&writer)
	}
}

func BenchmarkDeserializeComplexValueEasyJSON(b *testing.B) {
	data, _ := makeComplexValue().MarshalJSON()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lexer := jlexer.Lexer{Data: data}
		benchmarkValueResult.UnmarshalEasyJSON(&lexer)
	}
}
