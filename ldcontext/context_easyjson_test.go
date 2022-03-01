//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package ldcontext

import (
	"testing"

	"github.com/mailru/easyjson"
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
