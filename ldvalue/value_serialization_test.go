package ldvalue

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/launchdarkly/go-jsonstream/v2/jreader"
	"github.com/launchdarkly/go-jsonstream/v2/jwriter"

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
		})
	}
}

func TestMarshalRaw(t *testing.T) {
	// This is separate from the MarshalUnmarshal test because you never get a Raw when you unmarshal.
	s := `{"a":1}`
	value := Raw(json.RawMessage(s))

	bytes, err := json.Marshal(value)
	assert.NoError(t, err)
	assert.Equal(t, s, string(bytes))

	w := jwriter.NewWriter()
	value.WriteToJSONWriter(&w)
	assert.NoError(t, w.Error())
	assert.Equal(t, s, string(w.Bytes()))
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
