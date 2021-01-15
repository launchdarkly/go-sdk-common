// +build launchdarkly_easyjson

package lduser

import (
	"testing"

	easyjson "github.com/mailru/easyjson"
)

func TestEasyJSONMarshal(t *testing.T) {
	doUserMarshalingTests(t, func(value interface{}) ([]byte, error) {
		return easyjson.Marshal(value.(easyjson.Marshaler))
	})
}

func TestEasyJSONUnmarshal(t *testing.T) {
	doUserUnmarshalingTests(t, func(data []byte, target interface{}) error {
		return easyjson.Unmarshal(data, target.(easyjson.Unmarshaler))
	})
}
