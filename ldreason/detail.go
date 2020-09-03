package ldreason

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// EvaluationDetail is an object returned by LDClient.VariationDetail, combining the result of a
// flag evaluation with an explanation of how it was calculated.
type EvaluationDetail struct {
	// Value is the result of the flag evaluation. This will be either one of the flag's variations or
	// the default value that was passed to the Variation method.
	Value ldvalue.Value
	// VariationIndex is the index of the returned value within the flag's list of variations, e.g.
	// 0 for the first variation. This is an ldvalue.OptionalInt rather than an int, because there is
	// no variation index if the application default value was returned because the flag could not be
	// evaluated.
	VariationIndex ldvalue.OptionalInt
	// Reason is an EvaluationReason object describing the main factor that influenced the flag
	// evaluation value.
	Reason EvaluationReason
}

// IsDefaultValue returns true if the result of the evaluation was the application default value.
// This means that an error prevented the flag from being evaluated; the Reason field should contain
// an error value such as NewEvalReasonError(EvalErrorFlagNotFound).
func (d EvaluationDetail) IsDefaultValue() bool {
	return !d.VariationIndex.IsDefined()
}

// NewEvaluationDetail constructs an EvaluationDeteail, specifying all fields.
func NewEvaluationDetail(
	value ldvalue.Value,
	variationIndex ldvalue.OptionalInt,
	reason EvaluationReason,
) EvaluationDetail {
	return EvaluationDetail{Value: value, VariationIndex: variationIndex, Reason: reason}
}

// NewEvaluationDetailForError constructs an EvaluationDetail for an error condition.
func NewEvaluationDetailForError(errorKind EvalErrorKind, defaultValue ldvalue.Value) EvaluationDetail {
	return EvaluationDetail{Value: defaultValue, Reason: NewEvalReasonError(errorKind)}
}
