package ldvalue

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/launchdarkly/go-jsonstream/v3/jreader"
	"github.com/launchdarkly/go-jsonstream/v3/jwriter"

	"github.com/stretchr/testify/assert"
)

type structWithOptionalBools struct {
	B1 OptionalBool `json:"b1"`
	B2 OptionalBool `json:"b2"`
	B3 OptionalBool `json:"b3"`
}

type structWithOptionalInts struct {
	N1 OptionalInt `json:"n1"`
	N2 OptionalInt `json:"n2"`
	N3 OptionalInt `json:"n3"`
}

type structWithOptionalStrings struct {
	S1 OptionalString `json:"s1"`
	S2 OptionalString `json:"s2"`
	S3 OptionalString `json:"s3"`
}

func TestOptionalBoolJSONMarshalling(t *testing.T) {
	testWithMarshaler := func(t *testing.T, marshal func(OptionalBool) ([]byte, error)) {
		bytes, err := marshal(OptionalBool{})
		assert.NoError(t, err)
		assert.Equal(t, nullAsJSON, string(bytes))

		bytes, err = marshal(NewOptionalBool(true))
		assert.NoError(t, err)
		assert.Equal(t, `true`, string(bytes))

		bytes, err = marshal(NewOptionalBool(false))
		assert.NoError(t, err)
		assert.Equal(t, `false`, string(bytes))
	}

	t.Run("with json.Marshal", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalBool) ([]byte, error) {
			return json.Marshal(o)
		})

		swos := structWithOptionalBools{B1: NewOptionalBool(true), B2: NewOptionalBool(false)}
		bytes, err := json.Marshal(swos)
		assert.NoError(t, err)
		assert.Equal(t, `{"b1":true,"b2":false,"b3":null}`, string(bytes))
	})

	t.Run("with WriteToJSONWriter", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalBool) ([]byte, error) {
			w := jwriter.NewWriter()
			o.WriteToJSONWriter(&w)
			return w.Bytes(), w.Error()
		})
	})
}

func TestOptionalIntJSONMarshalling(t *testing.T) {
	testWithMarshaler := func(t *testing.T, marshal func(OptionalInt) ([]byte, error)) {
		bytes, err := marshal(OptionalInt{})
		assert.NoError(t, err)
		assert.Equal(t, nullAsJSON, string(bytes))

		bytes, err = marshal(NewOptionalInt(3))
		assert.NoError(t, err)
		assert.Equal(t, `3`, string(bytes))
	}

	t.Run("with json.Marshal", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalInt) ([]byte, error) {
			return json.Marshal(o)
		})

		swos := structWithOptionalInts{N1: NewOptionalInt(3)}
		bytes, err := json.Marshal(swos)
		assert.NoError(t, err)
		assert.Equal(t, `{"n1":3,"n2":null,"n3":null}`, string(bytes))
	})

	t.Run("with WriteToJSONWriter", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalInt) ([]byte, error) {
			w := jwriter.NewWriter()
			o.WriteToJSONWriter(&w)
			return w.Bytes(), w.Error()
		})
	})
}

func TestOptionalStringJSONMarshalling(t *testing.T) {
	testWithMarshaler := func(t *testing.T, marshal func(OptionalString) ([]byte, error)) {
		bytes, err := marshal(OptionalString{})
		assert.NoError(t, err)
		assert.Equal(t, nullAsJSON, string(bytes))

		bytes, err = marshal(NewOptionalString(`a "good" string`))
		assert.NoError(t, err)
		assert.Equal(t, `"a \"good\" string"`, string(bytes))
	}

	t.Run("with json.Marshal", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalString) ([]byte, error) {
			return json.Marshal(o)
		})

		swos := structWithOptionalStrings{S1: NewOptionalString("yes")}
		bytes, err := json.Marshal(swos)
		assert.NoError(t, err)
		assert.Equal(t, `{"s1":"yes","s2":null,"s3":null}`, string(bytes))
	})

	t.Run("with WriteToJSONWriter", func(t *testing.T) {
		testWithMarshaler(t, func(o OptionalString) ([]byte, error) {
			w := jwriter.NewWriter()
			o.WriteToJSONWriter(&w)
			return w.Bytes(), w.Error()
		})
	})
}

func TestOptionalBoolJSONUnmarshalling(t *testing.T) {
	testWithUnmarshaler := func(t *testing.T, unmarshal func([]byte, *OptionalBool) error) {
		var o OptionalBool
		err := unmarshal([]byte(nullAsJSON), &o)
		assert.NoError(t, err)
		assert.False(t, o.IsDefined())

		err = unmarshal([]byte(`true`), &o)
		assert.NoError(t, err)
		assert.Equal(t, NewOptionalBool(true), o)

		err = unmarshal([]byte(`false`), &o)
		assert.NoError(t, err)
		assert.Equal(t, NewOptionalBool(false), o)

		err = unmarshal([]byte(`3`), &o)
		assert.Error(t, err)
		assert.IsType(t, &json.UnmarshalTypeError{}, err)

		err = unmarshal([]byte(`x`), &o)
		assert.Error(t, err)
		assert.IsType(t, &json.SyntaxError{}, err)
	}

	t.Run("with json.Unmarshal", func(t *testing.T) {
		testWithUnmarshaler(t, func(data []byte, o *OptionalBool) error {
			return json.Unmarshal(data, o)
		})

		var swos structWithOptionalBools
		err := json.Unmarshal([]byte(`{"b1":true,"b2":false,"b3":null}`), &swos)
		assert.NoError(t, err)
		assert.Equal(t, NewOptionalBool(true), swos.B1)
		assert.Equal(t, NewOptionalBool(false), swos.B2)
		assert.Equal(t, OptionalBool{}, swos.B3)
	})

	t.Run("with ReadFromJSONReader", func(t *testing.T) {
		testWithUnmarshaler(t, func(data []byte, o *OptionalBool) error {
			r := jreader.NewReader(data)
			o.ReadFromJSONReader(&r)
			return jreader.ToJSONError(r.Error(), o)
		})
	})
}

func TestOptionalIntJSONUnmarshalling(t *testing.T) {
	var o OptionalInt
	err := json.Unmarshal([]byte(nullAsJSON), &o)
	assert.NoError(t, err)
	assert.False(t, o.IsDefined())

	err = json.Unmarshal([]byte(`3`), &o)
	assert.NoError(t, err)
	assert.Equal(t, NewOptionalInt(3), o)

	err = json.Unmarshal([]byte(`true`), &o)
	assert.Error(t, err)
	assert.IsType(t, &json.UnmarshalTypeError{}, err)

	err = json.Unmarshal([]byte(`x`), &o)
	assert.Error(t, err)
	assert.IsType(t, &json.SyntaxError{}, err)

	var swos structWithOptionalInts
	err = json.Unmarshal([]byte(`{"n1":3,"n3":null}`), &swos)
	assert.NoError(t, err)
	assert.Equal(t, NewOptionalInt(3), swos.N1)
	assert.Equal(t, OptionalInt{}, swos.N2)
	assert.Equal(t, OptionalInt{}, swos.N3)
}

func TestOptionalStringJSONUnmarshalling(t *testing.T) {
	var o OptionalString
	err := json.Unmarshal([]byte(nullAsJSON), &o)
	assert.NoError(t, err)
	assert.False(t, o.IsDefined())

	err = json.Unmarshal([]byte(`"a \"good\" string"`), &o)
	assert.NoError(t, err)
	assert.True(t, o.IsDefined())
	assert.Equal(t, `a "good" string`, o.StringValue())

	err = json.Unmarshal([]byte("3"), &o)
	assert.Error(t, err)
	assert.IsType(t, &json.UnmarshalTypeError{}, err)

	var swos structWithOptionalStrings
	err = json.Unmarshal([]byte(`{"s1":"yes","s3":null}`), &swos)
	assert.NoError(t, err)
	assert.Equal(t, NewOptionalString("yes"), swos.S1)
	assert.Equal(t, OptionalString{}, swos.S2)
	assert.Equal(t, OptionalString{}, swos.S3)
}

func TestOptionalsJSONString(t *testing.T) {
	for _, v := range []JSONStringer{
		OptionalBool{},
		NewOptionalBool(true),
		NewOptionalBool(false),
		OptionalInt{},
		NewOptionalInt(1),
		OptionalString{},
		NewOptionalString(""),
		NewOptionalString("a"),
	} {
		t.Run(fmt.Sprintf("%+v", v), func(t *testing.T) {
			jsonBytes, _ := json.Marshal(v)
			assert.Equal(t, string(jsonBytes), v.JSONString())
		})
	}
}
