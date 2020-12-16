// +build launchdarkly_easyjson

package lduser

import (
	"testing"

	"github.com/mailru/easyjson/jlexer"
	"github.com/mailru/easyjson/jwriter"
)

func BenchmarkUserSerializationWithAllAttributesEasyJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		writer := jwriter.Writer{}
		benchmarkSimpleUser.MarshalEasyJSON(&writer)
	}
}

func BenchmarkUserDeserializationWithAllAttributesEasyJSON(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lexer := jlexer.Lexer{Data: benchmarkUserWithAllAttributesJSON}
		benchmarkUserResult.UnmarshalEasyJSON(&lexer)
	}
}
