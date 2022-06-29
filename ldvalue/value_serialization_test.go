package ldvalue

import (
	"encoding/json"
	"fmt"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream"

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"

	"github.com/stretchr/testify/assert"
)

func TestJsonMarshalUnmarshal(t *testing.T) {
	items := []struct {
		value Value
		json  string
	}{
		{Null(), nullAsJSON},
		{Bool(true), "true"},
		{Bool(false), "false"},
		{Int(1), "1"},
		{Float64(1), "1"},
		{Float64(2.5), "2.5"},
		{String("x"), `"x"`},
		{ArrayOf(), `[]`},
		{ArrayBuild().Add(Bool(true)).Add(String("x")).Build(), `[true,"x"]`},
		{ObjectBuild().Build(), `{}`},
		{ObjectBuild().Set("a", Bool(true)).Build(), `{"a":true}`},
	}
	for _, item := range items {
		t.Run(fmt.Sprintf("type %s, json %v", item.value.Type(), item.json), func(t *testing.T) {
			j, err := json.Marshal(item.value)
			assert.NoError(t, err)
			assert.Equal(t, item.json, string(j))

			assert.Equal(t, item.json, item.value.String())
			assert.Equal(t, item.json, item.value.JSONString())
			assert.Equal(t, json.RawMessage(item.json), item.value.AsRaw())

			var v Value
			err = json.Unmarshal([]byte(item.json), &v)
			assert.NoError(t, err)
			assert.Equal(t, item.value, v)

			assert.Equal(t, item.value, Parse([]byte(item.json)))

			r := jreader.NewReader([]byte(item.json))
			var v1 Value
			v1.ReadFromJSONReader(&r)
			assert.NoError(t, r.Error())
			assert.Equal(t, item.value, v1)

			w := jwriter.NewWriter()
			item.value.WriteToJSONWriter(&w)
			bytes := w.Bytes()
			assert.NoError(t, w.Error())
			assert.Equal(t, item.json, string(bytes))

			var buf jsonstream.JSONBuffer // deprecated API
			item.value.WriteToJSONBuffer(&buf)
			bytes, err = buf.Get()
			assert.NoError(t, err)
			assert.Equal(t, item.json, string(bytes))
		})
	}
}

func TestMarshalRaw(t *testing.T) {
	// This is separate from the MarshalUnmarshal test because you never get a Raw when you unmarshal.
	for _, params := range []struct {
		desc   string
		input  json.RawMessage
		output string
	}{
		{"valid JSON", json.RawMessage(`{"a":1}`), `{"a":1}`},
		{"zero-length", json.RawMessage{}, `null`},
		{"nil", json.RawMessage(nil), `null`},
	} {
		t.Run(params.desc, func(t *testing.T) {
			value := Raw(params.input)

			bytes, err := json.Marshal(value)
			assert.NoError(t, err)
			assert.Equal(t, params.output, string(bytes))

			var buf jsonstream.JSONBuffer
			value.WriteToJSONBuffer(&buf)
			bytes, err = buf.Get()
			assert.NoError(t, err)
			assert.Equal(t, params.output, string(bytes))
		})
	}
}

func TestUnmarshalErrorConditions(t *testing.T) {
	var v Value
	for _, data := range [][]byte{
		nil,
		[]byte{},
		[]byte("what"),
		[]byte("["),
		[]byte("[what"),
		[]byte("{"),
		[]byte("{what"),
		[]byte(`{"no":what`),
		[]byte(`1,`),
	} {
		assert.Error(t, json.Unmarshal(data, &v))
		assert.Equal(t, Null(), Parse(data))
	}
}
