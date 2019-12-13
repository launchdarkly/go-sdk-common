package ldreason

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReasonSerializationAndDeserialization(t *testing.T) {
	params := []struct {
		reason       EvaluationReason
		stringRep    string
		expectedJSON string
	}{
		{EvaluationReason{}, "", "null"},
		{NewEvalReasonOff(), "OFF", `{"kind":"OFF"}`},
		{NewEvalReasonFallthrough(), "FALLTHROUGH", `{"kind":"FALLTHROUGH"}`},
		{NewEvalReasonTargetMatch(), "TARGET_MATCH", `{"kind":"TARGET_MATCH"}`},
		{NewEvalReasonRuleMatch(1, "x"), "RULE_MATCH(1,x)", `{"kind":"RULE_MATCH","ruleIndex":1,"ruleId":"x"}`},
		{NewEvalReasonPrerequisiteFailed("x"), "PREREQUISITE_FAILED(x)", `{"kind":"PREREQUISITE_FAILED","prerequisiteKey":"x"} `},
		{NewEvalReasonError(EvalErrorWrongType), "ERROR(WRONG_TYPE)", `{"kind":"ERROR","errorKind":"WRONG_TYPE"}`},
	}
	for _, param := range params {
		actual, err := json.Marshal(param.reason)
		assert.NoError(t, err)
		assert.JSONEq(t, param.expectedJSON, string(actual))

		var r1 EvaluationReason
		err = json.Unmarshal(actual, &r1)
		assert.NoError(t, err)
		assert.Equal(t, param.reason, r1)

		assert.Equal(t, param.stringRep, param.reason.String())
	}
}
