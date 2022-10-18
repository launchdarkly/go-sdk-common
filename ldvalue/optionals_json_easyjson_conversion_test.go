//go:build launchdarkly_easyjson

package ldvalue

import (
	"testing"

	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
)

func TestOptionalBoolEasyJSONMarshalling(t *testing.T) {
	bytes, err := easyjson.Marshal(OptionalBool{})
	assert.NoError(t, err)
	assert.Equal(t, nullAsJSON, string(bytes))

	bytes, err = easyjson.Marshal(NewOptionalBool(true))
	assert.NoError(t, err)
	assert.Equal(t, `true`, string(bytes))

	bytes, err = easyjson.Marshal(NewOptionalBool(false))
	assert.NoError(t, err)
	assert.Equal(t, `false`, string(bytes))
}

func TestOptionalIntEasyJSONMarshalling(t *testing.T) {
	bytes, err := easyjson.Marshal(OptionalInt{})
	assert.NoError(t, err)
	assert.Equal(t, nullAsJSON, string(bytes))

	bytes, err = easyjson.Marshal(NewOptionalInt(3))
	assert.NoError(t, err)
	assert.Equal(t, `3`, string(bytes))
}

func TestOptionalStringEasyJSONMarshalling(t *testing.T) {
	bytes, err := easyjson.Marshal(OptionalString{})
	assert.NoError(t, err)
	assert.Equal(t, nullAsJSON, string(bytes))

	bytes, err = easyjson.Marshal(NewOptionalString(`a "good" string`))
	assert.NoError(t, err)
	assert.Equal(t, `"a \"good\" string"`, string(bytes))
}

func TestOptionalBoolEasyJSONUnmarshalling(t *testing.T) {
	var o OptionalBool
	err := easyjson.Unmarshal([]byte(nullAsJSON), &o)
	assert.NoError(t, err)
	assert.False(t, o.IsDefined())

	err = easyjson.Unmarshal([]byte(`true`), &o)
	assert.NoError(t, err)
	assert.Equal(t, NewOptionalBool(true), o)

	err = easyjson.Unmarshal([]byte(`false`), &o)
	assert.NoError(t, err)
	assert.Equal(t, NewOptionalBool(false), o)

	err = easyjson.Unmarshal([]byte(`3`), &o)
	assert.Error(t, err)

	err = easyjson.Unmarshal([]byte(`x`), &o)
	assert.Error(t, err)
}

func TestOptionalIntEasyJSONUnmarshalling(t *testing.T) {
	var o OptionalInt
	err := easyjson.Unmarshal([]byte(nullAsJSON), &o)
	assert.NoError(t, err)
	assert.False(t, o.IsDefined())

	err = easyjson.Unmarshal([]byte(`3`), &o)
	assert.NoError(t, err)
	assert.Equal(t, NewOptionalInt(3), o)

	err = easyjson.Unmarshal([]byte(`true`), &o)
	assert.Error(t, err)

	err = easyjson.Unmarshal([]byte(`x`), &o)
	assert.Error(t, err)
}

func TestOptionalStringEasyJSONUnmarshalling(t *testing.T) {
	var o OptionalString
	err := easyjson.Unmarshal([]byte(nullAsJSON), &o)
	assert.NoError(t, err)
	assert.False(t, o.IsDefined())

	err = easyjson.Unmarshal([]byte(`"a \"good\" string"`), &o)
	assert.NoError(t, err)
	assert.True(t, o.IsDefined())
	assert.Equal(t, `a "good" string`, o.StringValue())

	err = easyjson.Unmarshal([]byte("3"), &o)
	assert.Error(t, err)
}
