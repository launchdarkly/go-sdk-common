//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package ldcontext

import (
	"testing"

	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"
	"github.com/stretchr/testify/assert"
)

func easyJSONMarshalTestFn(c *Context) ([]byte, error) {
	return easyjson.Marshal(c)
}

func easyJSONUnmarshalTestFn(c *Context, data []byte) error {
	return easyjson.Unmarshal(data, c)
}

func TestContextEasyJSONMarshal(t *testing.T) {
	contextMarshalingTests(t, easyJSONMarshalTestFn)
}

func TestContextEasyJSONUnmarshal(t *testing.T) {
	contextUnmarshalingTests(t, easyJSONUnmarshalTestFn)
}

func TestContextEasyJSONUnmarshalEventOutputFormat(t *testing.T) {
	for _, p := range makeEventOutputFormatUnmarshalingParams() {
		t.Run(p.json, func(t *testing.T) {
			in := jlexer.Lexer{Data: []byte(p.json)}
			var c Context
			ContextSerialization{}.UnmarshalFromEasyJSONLexerEventOutputFormat(&in, &c)
			assert.NoError(t, in.Error())
			assert.Equal(t, p.context, c)
		})
	}
}
