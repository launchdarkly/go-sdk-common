//go:build launchdarkly_easyjson

package ldreason

import (
	"testing"

	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
)

func TestReasonEasyJSONSerializationAndDeserialization(t *testing.T) {
	for _, param := range makeSerializationTestParams() {
		t.Run(param.expectedJSON, func(t *testing.T) {
			actual, err := easyjson.Marshal(param.reason)
			assert.NoError(t, err)
			assert.JSONEq(t, param.expectedJSON, string(actual))

			if param.reason.IsDefined() {
				var r1 EvaluationReason
				err = easyjson.Unmarshal(actual, &r1)
				assert.NoError(t, err)
				assert.Equal(t, param.reason, r1)
			}
		})
	}
}
