package ldvalue

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeMarshallableValues() []Value {
	return []Value{
		Null(),
		Bool(true),
		Bool(false),
		Int(1),
		Float64(2.5),
		String("x"),
		ArrayBuild().Add(Bool(true)).Add(String("x")).Build(),
		ObjectBuild().Set("a", Bool(true)).Build(),
	}
}

func makeMarshallableValueArrays() []ValueArray {
	return []ValueArray{
		{},
		ValueArrayOf(),
		ValueArrayOf(String("a")),
	}
}

func makeMarshallableValueMaps() []ValueMap {
	return []ValueMap{
		{},
		ValueMapBuild().Build(),
		ValueMapBuild().Set("a", Int(1)).Build(),
	}
}

func TestValueAsStringerIsSameAsJSONString(t *testing.T) {
	for _, value := range makeMarshallableValues() {
		t.Run(value.JSONString(), func(t *testing.T) {
			assert.Equal(t, value.JSONString(), value.String())
		})
	}
}

func TestValueArrayAsStringerIsSameAsJSONString(t *testing.T) {
	for _, value := range makeMarshallableValueArrays() {
		t.Run(value.JSONString(), func(t *testing.T) {
			assert.Equal(t, value.JSONString(), value.String())
		})
	}
}

func TestValueMapAsStringerIsSameAsJSONString(t *testing.T) {
	for _, value := range makeMarshallableValueMaps() {
		t.Run(value.JSONString(), func(t *testing.T) {
			assert.Equal(t, value.JSONString(), value.String())
		})
	}
}

func TestValueMarshalTextIsSameAsJSONString(t *testing.T) {
	for _, value := range makeMarshallableValues() {
		t.Run(value.JSONString(), func(t *testing.T) {
			bytes, err := value.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, value.JSONString(), string(bytes))
		})
	}
}

func TestValueUnmarshalText(t *testing.T) {
	for _, expected := range makeMarshallableValues() {
		t.Run(expected.JSONString(), func(t *testing.T) {
			var actual Value
			require.NoError(t, actual.UnmarshalText([]byte(expected.JSONString())))
			assert.Equal(t, expected, actual)
		})
	}

	t.Run("non-JSON inputs are treated as strings", func(t *testing.T) {
		for _, s := range []string{``, `t`, `{hello`} {
			expected := String(s)
			var actual Value
			require.NoError(t, actual.UnmarshalText([]byte(s)))
			assert.Equal(t, expected, actual)
		}
	})
}

func TestValueArrayMarshalTextIsSameAsJSONMarshal(t *testing.T) {
	for _, expected := range makeMarshallableValueArrays() {
		t.Run(expected.JSONString(), func(t *testing.T) {
			bytes, err := expected.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, expected.JSONString(), string(bytes))
		})
	}
}

func TestValueArrayUnmarshalTextIsSameAsJSONUnmarshal(t *testing.T) {
	for _, expected := range makeMarshallableValueArrays() {
		t.Run(expected.JSONString(), func(t *testing.T) {
			var actual ValueArray
			require.NoError(t, actual.UnmarshalText([]byte(expected.JSONString())))
			assert.Equal(t, expected, actual)
		})
	}
}

func TestValueMapMarshalTextIsSameAsJSONMarshal(t *testing.T) {
	for _, expected := range makeMarshallableValueMaps() {
		t.Run(expected.JSONString(), func(t *testing.T) {
			bytes, err := expected.MarshalText()
			require.NoError(t, err)
			assert.Equal(t, expected.JSONString(), string(bytes))
		})
	}
}

func TestValueMapUnmarshalTextIsSameAsJSONUnmarshal(t *testing.T) {
	for _, expected := range makeMarshallableValueMaps() {
		t.Run(expected.JSONString(), func(t *testing.T) {
			var actual ValueMap
			require.NoError(t, actual.UnmarshalText([]byte(expected.JSONString())))
			assert.Equal(t, expected, actual)
		})
	}
}
