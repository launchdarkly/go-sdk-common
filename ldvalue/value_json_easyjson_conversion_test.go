//go:build launchdarkly_easyjson

package ldvalue

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/mailru/easyjson"
	"github.com/stretchr/testify/assert"
)

func TestValueEasyJsonMarshalUnmarshal(t *testing.T) {
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
			j, err := easyjson.Marshal(item.value)
			assert.NoError(t, err)
			assert.Equal(t, item.json, string(j))

			assert.Equal(t, item.json, item.value.String())
			assert.Equal(t, item.json, item.value.JSONString())
			assert.Equal(t, json.RawMessage(item.json), item.value.AsRaw())

			var v Value
			err = easyjson.Unmarshal([]byte(item.json), &v)
			assert.NoError(t, err)
			assert.Equal(t, item.value, v)
		})
	}
}

func TestEasyJsonUnmarshalErrorConditions(t *testing.T) {
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
	} {
		t.Run(string(data), func(t *testing.T) {
			assert.Error(t, easyjson.Unmarshal(data, &v))
		})
	}
}

func TestEasyJsonMarshalRaw(t *testing.T) {
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

			j, err := easyjson.Marshal(value)
			assert.NoError(t, err)
			assert.Equal(t, params.output, string(j))
		})
	}
}

func TestValueArrayEasyJsonMarshalUnmarshal(t *testing.T) {
	items := []struct {
		valueArray ValueArray
		json       string
	}{
		{ValueArray{}, nullAsJSON},
		{ValueArrayOf(), `[]`},
		{ValueArrayOf(String("a"), String("b")), `["a", "b"]`},
	}
	for _, item := range items {
		t.Run(fmt.Sprintf("json %v", item.json), func(t *testing.T) {
			j, err := easyjson.Marshal(item.valueArray)
			assert.NoError(t, err)
			assert.JSONEq(t, item.json, string(j))

			assert.JSONEq(t, item.json, item.valueArray.String())
			assert.JSONEq(t, item.json, item.valueArray.JSONString())

			var a ValueArray
			err = easyjson.Unmarshal([]byte(item.json), &a)
			assert.NoError(t, err)
			assert.Equal(t, item.valueArray, a)
		})
	}
}

func TestValueMapEasyJsonMarshalUnmarshal(t *testing.T) {
	items := []struct {
		valueMap ValueMap
		json     string
	}{
		{ValueMap{}, nullAsJSON},
		{ValueMapBuild().Build(), `{}`},
		{ValueMapBuild().Set("a", Int(1)).Set("b", Int(2)).Build(), `{"a": 1, "b": 2}`},
	}
	for _, item := range items {
		t.Run(fmt.Sprintf("json %v", item.json), func(t *testing.T) {
			j, err := easyjson.Marshal(item.valueMap)
			assert.NoError(t, err)
			assert.JSONEq(t, item.json, string(j))

			assert.JSONEq(t, item.json, item.valueMap.String())
			assert.JSONEq(t, item.json, item.valueMap.JSONString())

			var m ValueMap
			err = easyjson.Unmarshal([]byte(item.json), &m)
			assert.NoError(t, err)
			assert.Equal(t, item.valueMap, m)
		})
	}
}
