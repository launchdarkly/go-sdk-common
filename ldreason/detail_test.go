package ldreason

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
)

func TestDetailConstructor(t *testing.T) {
	detail := NewEvaluationDetail(ldvalue.Bool(true), 1, NewEvalReasonFallthrough())
	assert.Equal(t, ldvalue.Bool(true), detail.Value)
	assert.Equal(t, ldvalue.NewOptionalInt(1), detail.VariationIndex)
	assert.Equal(t, NewEvalReasonFallthrough(), detail.Reason)
	assert.False(t, detail.IsDefaultValue())
}

func TestDetailErrorConstructor(t *testing.T) {
	detail := NewEvaluationDetailForError(EvalErrorFlagNotFound, ldvalue.Bool(false))
	assert.Equal(t, ldvalue.Bool(false), detail.Value)
	assert.Equal(t, ldvalue.OptionalInt{}, detail.VariationIndex)
	assert.Equal(t, NewEvalReasonError(EvalErrorFlagNotFound), detail.Reason)
	assert.True(t, detail.IsDefaultValue())
}
