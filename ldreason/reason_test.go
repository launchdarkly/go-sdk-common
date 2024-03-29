package ldreason

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/launchdarkly/go-jsonstream/v3/jreader"
	"github.com/launchdarkly/go-jsonstream/v3/jwriter"

	"github.com/stretchr/testify/assert"
)

func TestReasonIsDefined(t *testing.T) {
	assert.False(t, EvaluationReason{}.IsDefined())
	assert.True(t, NewEvalReasonOff().IsDefined())
	assert.True(t, NewEvalReasonFallthrough().IsDefined())
	assert.True(t, NewEvalReasonFallthroughExperiment(true).IsDefined())
	assert.True(t, NewEvalReasonTargetMatch().IsDefined())
	assert.True(t, NewEvalReasonRuleMatch(0, "").IsDefined())
	assert.True(t, NewEvalReasonRuleMatchExperiment(0, "", true).IsDefined())
	assert.True(t, NewEvalReasonPrerequisiteFailed("").IsDefined())
	assert.True(t, NewEvalReasonError(EvalErrorFlagNotFound).IsDefined())
}

func TestReasonKind(t *testing.T) {
	assert.Equal(t, EvalReasonOff, NewEvalReasonOff().GetKind())
	assert.Equal(t, EvalReasonFallthrough, NewEvalReasonFallthrough().GetKind())
	assert.Equal(t, EvalReasonFallthrough, NewEvalReasonFallthroughExperiment(true).GetKind())
	assert.Equal(t, EvalReasonTargetMatch, NewEvalReasonTargetMatch().GetKind())
	assert.Equal(t, EvalReasonRuleMatch, NewEvalReasonRuleMatch(0, "").GetKind())
	assert.Equal(t, EvalReasonRuleMatch, NewEvalReasonRuleMatchExperiment(0, "", true).GetKind())
	assert.Equal(t, EvalReasonPrerequisiteFailed, NewEvalReasonPrerequisiteFailed("").GetKind())
	assert.Equal(t, EvalReasonError, NewEvalReasonError(EvalErrorFlagNotFound).GetKind())
}

func TestReasonRuleProperties(t *testing.T) {
	r := NewEvalReasonRuleMatch(1, "id")
	assert.Equal(t, 1, r.GetRuleIndex())
	assert.Equal(t, "id", r.GetRuleID())

	r = NewEvalReasonRuleMatchExperiment(1, "id", true)
	assert.Equal(t, 1, r.GetRuleIndex())
	assert.Equal(t, "id", r.GetRuleID())

	for _, r := range []EvaluationReason{
		NewEvalReasonOff(), NewEvalReasonFallthrough(), NewEvalReasonFallthroughExperiment(true), NewEvalReasonTargetMatch(),
		NewEvalReasonPrerequisiteFailed(""), NewEvalReasonError(EvalErrorFlagNotFound),
	} {
		t.Run(string(r.GetKind()), func(t *testing.T) {
			assert.Equal(t, -1, r.GetRuleIndex())
			assert.Equal(t, "", r.GetRuleID())
		})
	}
}

func TestReasonExperimentProperties(t *testing.T) {
	r := NewEvalReasonFallthroughExperiment(true)
	assert.Equal(t, true, r.IsInExperiment())

	r = NewEvalReasonRuleMatchExperiment(1, "id", true)
	assert.Equal(t, true, r.IsInExperiment())

	for _, r := range []EvaluationReason{
		NewEvalReasonOff(), NewEvalReasonFallthrough(), NewEvalReasonFallthroughExperiment(false), NewEvalReasonTargetMatch(),
		NewEvalReasonRuleMatch(1, "id"),
		NewEvalReasonRuleMatchExperiment(1, "id", false), NewEvalReasonPrerequisiteFailed(""), NewEvalReasonError(EvalErrorFlagNotFound),
	} {
		t.Run(string(r.GetKind()), func(t *testing.T) {
			assert.Equal(t, false, r.IsInExperiment())
		})
	}
}

func TestReasonPrerequisiteFailedProperties(t *testing.T) {
	r := NewEvalReasonPrerequisiteFailed("key")
	assert.Equal(t, "key", r.GetPrerequisiteKey())

	for _, r := range []EvaluationReason{
		NewEvalReasonOff(), NewEvalReasonFallthrough(), NewEvalReasonFallthroughExperiment(true), NewEvalReasonTargetMatch(),
		NewEvalReasonRuleMatch(0, "id"), NewEvalReasonRuleMatchExperiment(0, "id", true), NewEvalReasonError(EvalErrorFlagNotFound),
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
		NewEvalReasonOff(), NewEvalReasonFallthrough(), NewEvalReasonFallthroughExperiment(true), NewEvalReasonTargetMatch(),
		NewEvalReasonRuleMatch(0, "id"), NewEvalReasonRuleMatchExperiment(0, "id", true), NewEvalReasonPrerequisiteFailed("key"),
	} {
		t.Run(string(r.GetKind()), func(t *testing.T) {
			assert.Equal(t, EvalErrorKind(""), r.GetErrorKind())
		})
	}
}

func TestReasonUnboundedSegmentsStatus(t *testing.T) {
	for _, r := range []EvaluationReason{
		NewEvalReasonOff(), NewEvalReasonFallthrough(), NewEvalReasonTargetMatch(),
		NewEvalReasonRuleMatch(0, "id"), NewEvalReasonPrerequisiteFailed("key"),
		NewEvalReasonError(EvalErrorFlagNotFound),
	} {
		t.Run(string(r.GetKind()), func(t *testing.T) {
			assert.Equal(t, BigSegmentsStatus(""), r.GetBigSegmentsStatus())
			r1 := NewEvalReasonFromReasonWithBigSegmentsStatus(r, BigSegmentsHealthy)
			assert.Equal(t, BigSegmentsHealthy, r1.GetBigSegmentsStatus())
		})
	}
}

type serializationTestParams struct {
	reason       EvaluationReason
	stringRep    string
	expectedJSON string
}

func TestReasonSerializationAndDeserialization(t *testing.T) {
	baseParams := []serializationTestParams{
		{EvaluationReason{}, "", "null"},
		{NewEvalReasonOff(), "OFF", `{"kind":"OFF"}`},
		{NewEvalReasonFallthrough(), "FALLTHROUGH", `{"kind":"FALLTHROUGH"}`},
		{NewEvalReasonFallthroughExperiment(true), "FALLTHROUGH", `{"kind":"FALLTHROUGH","inExperiment":true}`},
		{NewEvalReasonFallthroughExperiment(false), "FALLTHROUGH", `{"kind":"FALLTHROUGH"}`},
		{NewEvalReasonTargetMatch(), "TARGET_MATCH", `{"kind":"TARGET_MATCH"}`},
		{NewEvalReasonRuleMatch(1, "x"), "RULE_MATCH(1,x)", `{"kind":"RULE_MATCH","ruleIndex":1,"ruleId":"x"}`},
		{NewEvalReasonRuleMatchExperiment(1, "x", true), "RULE_MATCH(1,x)", `{"kind":"RULE_MATCH","ruleIndex":1,"ruleId":"x","inExperiment":true}`},
		{NewEvalReasonRuleMatchExperiment(1, "x", false), "RULE_MATCH(1,x)", `{"kind":"RULE_MATCH","ruleIndex":1,"ruleId":"x"}`},
		{NewEvalReasonPrerequisiteFailed("x"), "PREREQUISITE_FAILED(x)", `{"kind":"PREREQUISITE_FAILED","prerequisiteKey":"x"}`},
		{NewEvalReasonError(EvalErrorWrongType), "ERROR(WRONG_TYPE)", `{"kind":"ERROR","errorKind":"WRONG_TYPE"}`},
	}
	params := baseParams
	for _, param := range baseParams {
		if param.reason.IsDefined() {
			params = append(params, serializationTestParams{
				reason:    NewEvalReasonFromReasonWithBigSegmentsStatus(param.reason, BigSegmentsHealthy),
				stringRep: param.stringRep,
				expectedJSON: strings.TrimSuffix(param.expectedJSON, "}") +
					`,"bigSegmentsStatus":"HEALTHY"}`,
			})
		}
	}

	for _, param := range params {
		t.Run(param.expectedJSON, func(t *testing.T) {
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
		})
	}

	var r EvaluationReason
	err := json.Unmarshal([]byte(`{"kind":[1]}`), &r)
	assert.Error(t, err)
}
