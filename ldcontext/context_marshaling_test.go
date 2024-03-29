package ldcontext

import (
	"encoding/json"
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/lderrors"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"
	"github.com/launchdarkly/go-test-helpers/v3/jsonhelpers"

	"github.com/launchdarkly/go-jsonstream/v3/jwriter"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type contextSerializationParams struct {
	context Context
	json    string
}

func makeContextMarshalingAndUnmarshalingParams() []contextSerializationParams {
	return []contextSerializationParams{
		{NewWithKind("org", "key1"), `{"kind": "org", "key": "key1"}`},

		{New("key1b"), `{"kind": "user", "key": "key1b"}`},

		{NewBuilder("key1c").Kind("org").Build(),
			`{"kind": "org", "key": "key1c"}`},

		{NewBuilder("key2").Name("my-name").Build(),
			`{"kind": "user", "key": "key2", "name": "my-name"}`},

		{NewBuilder("key4").Anonymous(true).Build(),
			`{"kind": "user", "key": "key4", "anonymous": true}`},
		{NewBuilder("key5").Anonymous(false).Build(),
			`{"kind": "user", "key": "key5"}`},

		{NewBuilder("key6").SetBool("attr1", true).Build(),
			`{"kind": "user", "key": "key6", "attr1": true}`},
		{NewBuilder("key6").SetBool("attr1", false).Build(),
			`{"kind": "user", "key": "key6", "attr1": false}`},
		{NewBuilder("key6").SetInt("attr1", 123).Build(),
			`{"kind": "user", "key": "key6", "attr1": 123}`},
		{NewBuilder("key6").SetFloat64("attr1", 1.5).Build(),
			`{"kind": "user", "key": "key6", "attr1": 1.5}`},
		{NewBuilder("key6").SetString("attr1", "xyz").Build(),
			`{"kind": "user", "key": "key6", "attr1": "xyz"}`},
		{NewBuilder("key6").SetValue("attr1", ldvalue.ArrayOf(ldvalue.Int(10), ldvalue.Int(20))).Build(),
			`{"kind": "user", "key": "key6", "attr1": [10, 20]}`},
		{NewBuilder("key6").SetValue("attr1", ldvalue.ObjectBuild().Set("a", ldvalue.Int(1)).Build()).Build(),
			`{"kind": "user", "key": "key6", "attr1": {"a": 1}}`},

		{NewBuilder("key7").Private("a").PrivateRef(ldattr.NewRef("/b/c")).Build(),
			`{"kind": "user", "key": "key7", "_meta": {"privateAttributes": ["a", "/b/c"]}}`},

		{NewMulti(NewWithKind("org", "my-org-key"), New("my-user-key")),
			`{"kind": "multi", "org": {"key": "my-org-key"}, "user": {"key": "my-user-key"}}`},
	}
}

func makeContextMarshalingEventOutputFormatParams() []contextSerializationParams {
	var ret []contextSerializationParams
	for _, p := range makeContextMarshalingAndUnmarshalingParams() {
		transformed := p
		transformed.json = translateRegularContextJSONToEventOutputJSONAndViceVersa(transformed.json)
		ret = append(ret, transformed)
	}
	return ret
}

func contextMarshalingTests(t *testing.T, marshalFn func(*Context) ([]byte, error)) {
	for _, params := range makeContextMarshalingAndUnmarshalingParams() {
		t.Run(params.json, func(t *testing.T) {
			bytes, err := marshalFn(&params.context)
			assert.NoError(t, err)
			jsonhelpers.AssertEqual(t, params.json, bytes)
		})
	}

	t.Run("invalid context", func(t *testing.T) {
		c := New("")
		_, err := marshalFn(&c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), lderrors.ErrContextKeyEmpty{}.Error())
		// We compare the error string, rather than checking for equality to errContextKeyEmpty itself, because
		// the JSON marshaller may decorate the error in its own type with additional text.
	})

	t.Run("uninitialized context", func(t *testing.T) {
		var c Context
		_, err := marshalFn(&c)
		require.Error(t, err)
		assert.Contains(t, err.Error(), lderrors.ErrContextUninitialized{}.Error())
	})
}

func jsonMarshalTestFn(c *Context) ([]byte, error) {
	return json.Marshal(c)
}

func jsonStreamMarshalTestFn(c *Context) ([]byte, error) {
	w := jwriter.NewWriter()
	ContextSerialization.MarshalToJSONWriter(&w, c)
	return w.Bytes(), w.Error()
}

func TestContextJSONMarshal(t *testing.T) {
	contextMarshalingTests(t, jsonMarshalTestFn)
}

func TestContextWriteToJSONWriter(t *testing.T) {
	contextMarshalingTests(t, jsonStreamMarshalTestFn)
}

func TestContextMarshalEventOutputFormat(t *testing.T) {
	for _, p := range makeContextMarshalingEventOutputFormatParams() {
		t.Run(p.json, func(t *testing.T) {
			w := jwriter.NewWriter()
			ec := EventOutputContext{Context: p.context}
			ContextSerialization.MarshalToJSONWriterEventOutput(&w, &ec)
			assert.NoError(t, w.Error())
			jsonhelpers.AssertEqual(t, p.json, w.Bytes())
		})
	}

	t.Run("invalid context", func(t *testing.T) {
		c := New("")
		ec := EventOutputContext{Context: c}
		w := jwriter.NewWriter()
		ContextSerialization.MarshalToJSONWriterEventOutput(&w, &ec)
		require.Error(t, w.Error())
		assert.Contains(t, w.Error().Error(), lderrors.ErrContextKeyEmpty{}.Error())
	})

	t.Run("uninitialized context", func(t *testing.T) {
		var c Context
		ec := EventOutputContext{Context: c}
		w := jwriter.NewWriter()
		ContextSerialization.MarshalToJSONWriterEventOutput(&w, &ec)
		require.Error(t, w.Error())
		assert.Contains(t, w.Error().Error(), lderrors.ErrContextUninitialized{}.Error())
	})
}

func TestContextJSONStringIsEquivalentToJSONMarshal(t *testing.T) {
	for _, params := range makeContextMarshalingAndUnmarshalingParams() {
		t.Run(params.json, func(t *testing.T) {
			bytes, err := json.Marshal(&params.context)
			assert.NoError(t, err)
			jsonhelpers.AssertEqual(t, string(bytes), params.context.JSONString())
		})
	}
}
