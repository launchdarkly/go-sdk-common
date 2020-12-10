package ldreason

import (
	"encoding/json"
	"testing"

	"github.com/launchdarkly/go-jsonstream/jreader"
	"github.com/launchdarkly/go-jsonstream/jwriter"
	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"

	"github.com/stretchr/testify/assert"
)

func TestReasonIsDefined(t *testing.T) {
	assert.False(t, EvaluationReason{}.IsDefined())
	assert.True(t, NewEvalReasonOff().IsDefined())
	assert.True(t, NewEvalReasonFallthrough().IsDefined())
	assert.True(t, NewEvalReasonTargetMatch().IsDefined())
	assert.True(t, NewEvalReasonRuleMatch(0, "").IsDefined())
	assert.True(t, NewEvalReasonPrerequisiteFailed("").IsDefined())
	assert.True(t, NewEvalReasonError(EvalErrorFlagNotFound).IsDefined())
}

func TestReasonKind(t *testing.T) {
	assert.Equal(t, EvalReasonOff, NewEvalReasonOff().GetKind())
	assert.Equal(t, EvalReasonFallthrough, NewEvalReasonFallthrough().GetKind())
	assert.Equal(t, EvalReasonTargetMatch, NewEvalReasonTargetMatch().GetKind())
	assert.Equal(t, EvalReasonRuleMatch, NewEvalReasonRuleMatch(0, "").GetKind())
	assert.Equal(t, EvalReasonPrerequisiteFailed, NewEvalReasonPrerequisiteFailed("").GetKind())
	assert.Equal(t, EvalReasonError, NewEvalReasonError(EvalErrorFlagNotFound).GetKind())
}

func TestReasonRuleProperties(t *testing.T) {
	r := NewEvalReasonRuleMatch(1, "id")
	assert.Equal(t, 1, r.GetRuleIndex())
	assert.Equal(t, "id", r.GetRuleID())

	for _, r := range []EvaluationReason{
		NewEvalReasonOff(), NewEvalReasonFallthrough(), NewEvalReasonTargetMatch(),
		NewEvalReasonPrerequisiteFailed(""), NewEvalReasonError(EvalErrorFlagNotFound),
	} {
		t.Run(string(r.GetKind()), func(t *testing.T) {
			assert.Equal(t, -1, r.GetRuleIndex())
			assert.Equal(t, "", r.GetRuleID())
		})
	}
}

func TestReasonPrerequisiteFailedProperties(t *testing.T) {
	r := NewEvalReasonPrerequisiteFailed("key")
	assert.Equal(t, "key", r.GetPrerequisiteKey())

	for _, r := range []EvaluationReason{
		NewEvalReasonOff(), NewEvalReasonFallthrough(), NewEvalReasonTargetMatch(),
		NewEvalReasonRuleMatch(0, "id"), NewEvalReasonError(EvalErrorFlagNotFound),
	} {
		t.Run(string(r.GetKind()), func(t *testing.T) {
			assert.Equal(t, "", r.GetPrerequisiteKey())
		})
	}
}

func TestReasonErrorProperties(t *testing.T) {
	r := NewEvalReasonError(EvalErrorFlagNotFound)
	assert.Equal(t, EvalErrorFlagNotFound, r.GetErrorKind())

	for _, r := range []EvaluationReason{
		NewEvalReasonOff(), NewEvalReasonFallthrough(), NewEvalReasonTargetMatch(),
		NewEvalReasonRuleMatch(0, "id"), NewEvalReasonPrerequisiteFailed("key"),
	} {
		t.Run(string(r.GetKind()), func(t *testing.T) {
			assert.Equal(t, EvalErrorKind(""), r.GetErrorKind())
		})
	}
}

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

		var r2 EvaluationReason
		reader := jreader.NewReader(actual)
		r2.ReadFromJSONReader(&reader)
		assert.NoError(t, reader.Error())
		assert.Equal(t, param.reason, r2)

		w := jwriter.NewWriter()
		param.reason.WriteToJSONWriter(&w)
		assert.NoError(t, w.Error())
		bytes := w.Bytes()
		assert.JSONEq(t, param.expectedJSON, string(bytes))

		var buf jsonstream.JSONBuffer
		param.reason.WriteToJSONBuffer(&buf)
		bytes, err = buf.Get()
		assert.NoError(t, err)
		assert.JSONEq(t, param.expectedJSON, string(bytes))
	}

	var r EvaluationReason
	err := json.Unmarshal([]byte(`{"kind":[1]}`), &r)
	assert.Error(t, err)
}
