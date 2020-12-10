package ldreason

import (
	"fmt"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream" //nolint:staticcheck // using a deprecated API
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

// EvalReasonKind defines the possible values of the Kind property of EvaluationReason.
type EvalReasonKind string

const (
	// EvalReasonOff indicates that the flag was off and therefore returned its configured off value.
	EvalReasonOff EvalReasonKind = "OFF"
	// EvalReasonTargetMatch indicates that the user key was specifically targeted for this flag.
	EvalReasonTargetMatch EvalReasonKind = "TARGET_MATCH"
	// EvalReasonRuleMatch indicates that the user matched one of the flag's rules.
	EvalReasonRuleMatch EvalReasonKind = "RULE_MATCH"
	// EvalReasonPrerequisiteFailed indicates that the flag was considered off because it had at
	// least one prerequisite flag that either was off or did not return the desired variation.
	EvalReasonPrerequisiteFailed EvalReasonKind = "PREREQUISITE_FAILED"
	// EvalReasonFallthrough indicates that the flag was on but the user did not match any targets
	// or rules.
	EvalReasonFallthrough EvalReasonKind = "FALLTHROUGH"
	// EvalReasonError indicates that the flag could not be evaluated, e.g. because it does not
	// exist or due to an unexpected error. In this case the result value will be the default value
	// that the caller passed to the client.
	EvalReasonError EvalReasonKind = "ERROR"
)

// EvalErrorKind defines the possible values of the ErrorKind property of EvaluationReason.
type EvalErrorKind string

const (
	// EvalErrorClientNotReady indicates that the caller tried to evaluate a flag before the client
	// had successfully initialized.
	EvalErrorClientNotReady EvalErrorKind = "CLIENT_NOT_READY"
	// EvalErrorFlagNotFound indicates that the caller provided a flag key that did not match any
	// known flag.
	EvalErrorFlagNotFound EvalErrorKind = "FLAG_NOT_FOUND"
	// EvalErrorMalformedFlag indicates that there was an internal inconsistency in the flag data,
	// e.g. a rule specified a nonexistent variation.
	EvalErrorMalformedFlag EvalErrorKind = "MALFORMED_FLAG"
	// EvalErrorUserNotSpecified indicates that the caller passed a user without a key for the user
	// parameter.
	EvalErrorUserNotSpecified EvalErrorKind = "USER_NOT_SPECIFIED"
	// EvalErrorWrongType indicates that the result value was not of the requested type, e.g. you
	// called BoolVariationDetail but the value was an integer.
	EvalErrorWrongType EvalErrorKind = "WRONG_TYPE"
	// EvalErrorException indicates that an unexpected error stopped flag evaluation; check the
	// log for details.
	EvalErrorException EvalErrorKind = "EXCEPTION"
)

// EvaluationReason describes the reason that a flag evaluation producted a particular value.
//
// This struct is immutable; its properties can be accessed only via getter methods.
type EvaluationReason struct {
	kind            EvalReasonKind
	ruleIndex       ldvalue.OptionalInt
	ruleID          string
	prerequisiteKey string
	errorKind       EvalErrorKind
}

// IsDefined returns true if this EvaluationReason has a non-empty GetKind(). It is false for a
// zero value of EvaluationReason{}.
func (r EvaluationReason) IsDefined() bool {
	return r.kind != ""
}

// String returns a concise string representation of the reason. Examples: "OFF", "ERROR(WRONG_TYPE)".
func (r EvaluationReason) String() string {
	switch r.kind {
	case EvalReasonRuleMatch:
		return fmt.Sprintf("%s(%d,%s)", r.kind, r.ruleIndex.OrElse(0), r.ruleID)
	case EvalReasonPrerequisiteFailed:
		return fmt.Sprintf("%s(%s)", r.kind, r.prerequisiteKey)
	case EvalReasonError:
		return fmt.Sprintf("%s(%s)", r.kind, r.errorKind)
	default:
		return string(r.GetKind())
	}
}

// GetKind describes the general category of the reason.
func (r EvaluationReason) GetKind() EvalReasonKind {
	return r.kind
}

// GetRuleIndex provides the index of the rule that was matched (0 being the first), if
// the Kind is EvalReasonRuleMatch. Otherwise it returns -1.
func (r EvaluationReason) GetRuleIndex() int {
	return r.ruleIndex.OrElse(-1)
}

// GetRuleID provides the unique identifier of the rule that was matched, if the Kind is
// EvalReasonRuleMatch. Otherwise it returns an empty string. Unlike the rule index, this
// identifier will not change if other rules are added or deleted.
func (r EvaluationReason) GetRuleID() string {
	return r.ruleID
}

// GetPrerequisiteKey provides the flag key of the prerequisite that failed, if the Kind
// is EvalReasonPrerequisiteFailed. Otherwise it returns an empty string.
func (r EvaluationReason) GetPrerequisiteKey() string {
	return r.prerequisiteKey
}

// GetErrorKind describes the general category of the error, if the Kind is EvalReasonError.
// Otherwise it returns an empty string.
func (r EvaluationReason) GetErrorKind() EvalErrorKind {
	return r.errorKind
}

// NewEvalReasonOff returns an EvaluationReason whose Kind is EvalReasonOff.
func NewEvalReasonOff() EvaluationReason {
	return EvaluationReason{kind: EvalReasonOff}
}

// NewEvalReasonFallthrough returns an EvaluationReason whose Kind is EvalReasonFallthrough.
func NewEvalReasonFallthrough() EvaluationReason {
	return EvaluationReason{kind: EvalReasonFallthrough}
}

// NewEvalReasonTargetMatch returns an EvaluationReason whose Kind is EvalReasonTargetMatch.
func NewEvalReasonTargetMatch() EvaluationReason {
	return EvaluationReason{kind: EvalReasonTargetMatch}
}

// NewEvalReasonRuleMatch returns an EvaluationReason whose Kind is EvalReasonRuleMatch.
func NewEvalReasonRuleMatch(ruleIndex int, ruleID string) EvaluationReason {
	return EvaluationReason{kind: EvalReasonRuleMatch,
		ruleIndex: ldvalue.NewOptionalInt(ruleIndex), ruleID: ruleID}
}

// NewEvalReasonPrerequisiteFailed returns an EvaluationReason whose Kind is EvalReasonPrerequisiteFailed.
func NewEvalReasonPrerequisiteFailed(prereqKey string) EvaluationReason {
	return EvaluationReason{kind: EvalReasonPrerequisiteFailed, prerequisiteKey: prereqKey}
}

// NewEvalReasonError returns an EvaluationReason whose Kind is EvalReasonError.
func NewEvalReasonError(errorKind EvalErrorKind) EvaluationReason {
	return EvaluationReason{kind: EvalReasonError, errorKind: errorKind}
}

// MarshalJSON implements custom JSON serialization for EvaluationReason.
func (r EvaluationReason) MarshalJSON() ([]byte, error) {
	return jwriter.MarshalJSONWithWriter(r)
}

// UnmarshalJSON implements custom JSON deserialization for EvaluationReason.
func (r *EvaluationReason) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, r)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Unmarshal.
// See https://github.com/launchdarkly/go-jsonstream for more details.
func (r *EvaluationReason) ReadFromJSONReader(reader *jreader.Reader) {
	var ret EvaluationReason
	for obj := reader.ObjectOrNull(); obj.Next(); {
		switch string(obj.Name()) {
		case "kind":
			ret.kind = EvalReasonKind(reader.String())
		case "ruleId":
			ret.ruleID = reader.String()
		case "ruleIndex":
			ret.ruleIndex = ldvalue.NewOptionalInt(reader.Int())
		case "errorKind":
			ret.errorKind = EvalErrorKind(reader.String())
		case "prerequisiteKey":
			ret.prerequisiteKey = reader.String()
		}
	}
	if reader.Error() == nil {
		*r = ret
	}
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Marshal.
// See https://github.com/launchdarkly/go-jsonstream for more details.
func (r EvaluationReason) WriteToJSONWriter(w *jwriter.Writer) {
	if r.kind == "" {
		w.Null()
		return
	}
	obj := w.Object()
	obj.Property("kind")
	w.String(string(r.kind))
	if r.ruleIndex.IsDefined() {
		obj.Property("ruleIndex")
		w.Int(r.ruleIndex.OrElse(0))
		if r.ruleID != "" {
			obj.Property("ruleId")
			w.String(r.ruleID)
		}
	}
	if r.kind == EvalReasonPrerequisiteFailed {
		obj.Property("prerequisiteKey")
		w.String(r.prerequisiteKey)
	}
	if r.kind == EvalReasonError {
		obj.Property("errorKind")
		w.String(string(r.errorKind))
	}
	obj.End()
}

// WriteToJSONBuffer provides JSON serialization for use with the deprecated jsonstream API.
//
// Deprecated: this method is provided for backward compatibility. The LaunchDarkly SDK no longer
// uses this API; instead it uses the newer https://github.com/launchdarkly/go-jsonstream.
func (r EvaluationReason) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	jsonstream.WriteToJSONBufferThroughWriter(r, j)
}
