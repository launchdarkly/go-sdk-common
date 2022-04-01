//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package ldcontext

import (
	"testing"

	m "github.com/launchdarkly/go-test-helpers/v2/matchers"
	"github.com/mailru/easyjson"
	"github.com/mailru/easyjson/jlexer"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestContextEasyJSONMarshalEventOutputFormat(t *testing.T) {
	for _, p := range makeContextMarshalingEventOutputFormatParams() {
		t.Run(p.json, func(t *testing.T) {
			ec := EventOutputContext{Context: p.context}
			data, err := easyjson.Marshal(ec)
			assert.NoError(t, err)
			m.In(t).Assert(data, m.JSONStrEqual(p.json))
		})
	}

	t.Run("invalid context", func(t *testing.T) {
		c := New("")
		ec := EventOutputContext{Context: c}
		_, err := easyjson.Marshal(ec)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errContextKeyEmpty.Error())
	})

	t.Run("uninitialized context", func(t *testing.T) {
		var c Context
		ec := EventOutputContext{Context: c}
		_, err := easyjson.Marshal(ec)
		require.Error(t, err)
		assert.Contains(t, err.Error(), errContextUninitialized.Error())
	})
}

func TestContextEasyJSONUnmarshalEventOutputFormat(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		for _, p := range makeAllContextUnmarshalingEventOutputFormatParams() {
			t.Run(p.json, func(t *testing.T) {
				var ec EventOutputContext
				err := easyjson.Unmarshal([]byte(p.json), &ec)
				assert.NoError(t, err)
				assert.Equal(t, p.context, ec.Context)
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		for _, badJSON := range makeContextUnmarshalingEventOutputFormatErrorInputs() {
			t.Run(badJSON, func(t *testing.T) {
				var c EventOutputContext
				in := jlexer.Lexer{Data: []byte(badJSON)}
				ContextSerialization.UnmarshalFromEasyJSONLexerEventOutput(&in, &c)
				assert.Error(t, in.Error())
			})
		}
	})
}
