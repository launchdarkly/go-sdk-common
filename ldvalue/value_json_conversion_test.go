package ldvalue

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/launchdarkly/go-jsonstream/v3/jreader"
	"github.com/launchdarkly/go-jsonstream/v3/jwriter"

	"github.com/stretchr/testify/assert"
)

func TestValueJSONMarshalUnmarshal(t *testing.T) {
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

			w := jwriter.NewWriter()
			value.WriteToJSONWriter(&w)
			assert.NoError(t, w.Error())
			assert.Equal(t, params.output, string(w.Bytes()))
		})
	}
}

func TestValueUnmarshalErrorConditions(t *testing.T) {
	var v Value
	for _, data := range [][]byte{
		nil,
		{},
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

func TestValueArrayJSONMarshalUnmarshal(t *testing.T) {
	items := []struct {
		valueArray ValueArray
		json       string
	}{
		{ValueArray{}, nullAsJSON},
		{ValueArrayBuild().Build(), `[]`},
		{ValueArrayOf(String("a"), String("b")), `["a","b"]`},
	}
	for _, item := range items {
		t.Run(fmt.Sprintf("json %v", item.json), func(t *testing.T) {
			j, err := json.Marshal(item.valueArray)
			assert.NoError(t, err)
			assert.Equal(t, item.json, string(j))

			assert.Equal(t, item.json, item.valueArray.String())
			assert.Equal(t, item.json, item.valueArray.JSONString())

			var a ValueArray
			err = json.Unmarshal([]byte(item.json), &a)
			assert.NoError(t, err)
			assert.Equal(t, item.valueArray, a)

			r := jreader.NewReader([]byte(item.json))
			a = ValueArray{}
			a.ReadFromJSONReader(&r)
			assert.NoError(t, r.Error())
			assert.Equal(t, item.valueArray, a)

			w := jwriter.NewWriter()
			item.valueArray.WriteToJSONWriter(&w)
			assert.NoError(t, w.Error())
			assert.Equal(t, item.json, string(w.Bytes()))
		})
	}

	for _, badJSON := range []string{"true", "1", `"x"`, "{}"} {
		err := json.Unmarshal([]byte(badJSON), &ValueArray{})
		assert.Error(t, err)
		assert.IsType(t, &json.UnmarshalTypeError{}, err)
	}
}

func TestValueMapJSONMarshalUnmarshal(t *testing.T) {
	items := []struct {
		valueMap ValueMap
		json     string
	}{
		{ValueMap{}, nullAsJSON},
		{ValueMapBuild().Build(), `{}`},
		{ValueMapBuild().Set("a", Bool(true)).Build(), `{"a":true}`},
	}
	for _, item := range items {
		t.Run(fmt.Sprintf("json %v", item.json), func(t *testing.T) {
			j, err := json.Marshal(item.valueMap)
			assert.NoError(t, err)
			assert.Equal(t, item.json, string(j))

			assert.Equal(t, item.json, item.valueMap.String())
			assert.Equal(t, item.json, item.valueMap.JSONString())

			var m ValueMap
			err = json.Unmarshal([]byte(item.json), &m)
			assert.NoError(t, err)
			assert.Equal(t, item.valueMap, m)

			r := jreader.NewReader([]byte(item.json))
			m = ValueMap{}
			m.ReadFromJSONReader(&r)
			assert.NoError(t, r.Error())
			assert.Equal(t, item.valueMap, m)

			w := jwriter.NewWriter()
			item.valueMap.WriteToJSONWriter(&w)
			assert.NoError(t, w.Error())
			assert.Equal(t, item.json, string(w.Bytes()))
		})
	}

	for _, badJSON := range []string{"true", "1", `"x"`, "[]"} {
		err := json.Unmarshal([]byte(badJSON), &ValueMap{})
		assert.Error(t, err)
		assert.IsType(t, &json.UnmarshalTypeError{}, err)
	}
}
