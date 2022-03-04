package ldcontext

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-sdk-common.v3/ldvalue"

	"github.com/stretchr/testify/assert"
)

func makeContextUnmarshalUnimportantVariantsParams() []contextSerializationParams {
	return []contextSerializationParams{
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "name": null}`},

		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {}}`},

		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "transient": false}`},

		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {"secondary": null}}`},

		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {"unknownPropIsIgnored": false}}`},

		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {"redactedAttributes": ["name"]}}`},
		// redactedAttributes is only a thing in the event output format, not the regular format
	}
}

func makeContextUnmarshalFromOldUserSchemaParams() []contextSerializationParams {
	ret := []contextSerializationParams{
		{New("key1"), `{"key": "key1"}`},

		{NewBuilder("").setAllowEmptyKey(true).Build(), `{"key": ""}`}, // allowed only in old-style user JSON

		{NewBuilder("key2").Name("my-name").Build(),
			`{"key": "key2", "name": "my-name"}`},
		{NewBuilder("key2").Build(),
			`{"key": "key2", "name": null}`},

		{NewBuilder("key3").Secondary("value").Build(),
			`{"key": "key3", "secondary": "value"}`},
		{NewBuilder("key3").Build(),
			`{"key": "key3", "secondary": null}`},

		{NewBuilder("key4").Transient(true).Build(),
			`{"key": "key4", "anonymous": true}`},
		{NewBuilder("key4").Build(),
			`{"key": "key4", "anonymous": false}`},
		{NewBuilder("key4").Build(),
			`{"key": "key4", "anonymous": null}`},

		{NewBuilder("key6").SetBool("attr1", true).Build(),
			`{"key": "key6", "custom": {"attr1": true}}`},
		{NewBuilder("key6").SetBool("attr1", false).Build(),
			`{"key": "key6", "custom": {"attr1": false}}`},
		{NewBuilder("key6").SetInt("attr1", 123).Build(),
			`{"key": "key6", "custom": {"attr1": 123}}`},
		{NewBuilder("key6").SetFloat64("attr1", 1.5).Build(),
			`{"key": "key6", "custom": {"attr1": 1.5}}`},
		{NewBuilder("key6").SetString("attr1", "xyz").Build(),
			`{"key": "key6", "custom": {"attr1": "xyz"}}`},
		{NewBuilder("key6").SetValue("attr1", ldvalue.ArrayOf(ldvalue.Int(10), ldvalue.Int(20))).Build(),
			`{"key": "key6", "custom": {"attr1": [10, 20]}}`},
		{NewBuilder("key6").SetValue("attr1", ldvalue.ObjectBuild().Set("a", ldvalue.Int(1)).Build()).Build(),
			`{"key": "key6", "custom": {"attr1": {"a": 1}}}`},

		{NewBuilder("key6").Name("x").Private("name", "email").Build(),
			`{"key": "key6", "name": "x", "privateAttributeNames":["name", "email"]}`},
		{NewBuilder("key6").Name("x").Private("name", "email").Build(),
			`{"key": "key6", "name": "x", "privateAttributeNames":["name", "email"]}`},
		{NewBuilder("key6").Name("x").Build(),
			`{"key": "key6", "name": "x", "privateAttributeNames":[]}`},
		{NewBuilder("key6").Name("x").Build(),
			`{"key": "key6", "name": "x", "privateAttributeNames":null}`},

		{NewBuilder("key7").Name("x").Build(),
			`{"key": "key7", "unknownTopLevelPropIsIgnored": {"a": 1}, "name": "x"}`},

		{NewBuilder("key8").Name("x").Build(),
			`{"key": "key8", "name": "x", "privateAttrs": ["name"]}`},
		// privateAttrs is only a thing in the event output format
	}
	for _, attr := range []string{"firstName", "lastName", "email", "country", "avatar", "ip"} {
		ret = append(ret,
			contextSerializationParams{
				NewBuilder("user-key").SetString(attr, "x").Build(),
				fmt.Sprintf(`{"key": "user-key", %q: "x"}`, attr),
			},
			contextSerializationParams{
				NewBuilder("user-key").Build(),
				fmt.Sprintf(`{"key": "user-key", %q: null}`, attr),
			},
		)
	}
	for _, customValue := range []ldvalue.Value{
		ldvalue.Bool(true),
		ldvalue.Bool(false),
		ldvalue.Int(123),
		ldvalue.Float64(1.5),
		ldvalue.String("xyz"),
		ldvalue.ArrayOf(ldvalue.Int(10), ldvalue.Int(20)),
		ldvalue.ObjectBuild().Set("a", ldvalue.Int(1)).Build(),
	} {
		ret = append(ret,
			contextSerializationParams{
				NewBuilder("user-key").SetValue("my-attr", customValue).Build(),
				fmt.Sprintf(`{"key": "user-key", "custom": {"my-attr": %s}}`, customValue.JSONString()),
			},
		)
	}
	return ret
}

func makeAllContextUnmarshalingParams() []contextSerializationParams {
	params := makeContextMarshalingAndUnmarshalingParams()
	params = append(params, makeContextUnmarshalUnimportantVariantsParams()...)
	params = append(params, makeContextUnmarshalFromOldUserSchemaParams()...)
	return params
}

func makeEventOutputFormatUnmarshalingParams() []contextSerializationParams {
	var params []contextSerializationParams
	for _, p := range makeAllContextUnmarshalingParams() {
		// The regular input data includes some contexts with _meta.privateAttributes or privateAttributeNames--
		// which in the regular format will get parsed, but in the event format they are ignored. It also
		// includes some contexts with _meta.redactedAttributes or privateAttrs-- which in the regular format
		// would be ignored, but are parsed in the event format.
		if strings.Contains(p.json, `"redactedAttributes"`) || strings.Contains(p.json, `"privateAttrs"`) {
			continue
		}
		p.context.privateAttrs = nil
		params = append(params, p)
	}
	params = append(params,
		contextSerializationParams{
			NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {"redactedAttributes": []}}`,
		},
		contextSerializationParams{
			NewBuilder("my-key").PreviouslyRedacted([]string{"a", "b"}).Build(),
			`{"kind": "user", "key": "my-key", "_meta": {"redactedAttributes": ["a", "b"]}}`,
		},
		contextSerializationParams{
			NewBuilder("my-key").Build(),
			`{"key": "my-key", "privateAttrs": []}`, // old-style user
		},
		contextSerializationParams{
			NewBuilder("my-key").PreviouslyRedacted([]string{"a", "b"}).Build(),
			`{"key": "my-key", "privateAttrs": ["a", "b"]}`, // old-style user
		},
	)
	return params
}

func contextUnmarshalingTests(t *testing.T, unmarshalFn func(*Context, []byte) error) {
	t.Run("valid data", func(t *testing.T) {
		for _, p := range makeAllContextUnmarshalingParams() {
			t.Run(p.json, func(t *testing.T) {
				var c Context
				err := unmarshalFn(&c, []byte(p.json))
				assert.NoError(t, err)
				assert.Equal(t, p.context, c)
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		for _, badJSON := range []string{
			`null`,
			`false`,
			`1`,
			`"x"`,
			`[]`,
			`{}`,

			// wrong type for top-level property
			`{"kind": null}`,
			`{"kind": true}`,
			`{"kind": "org", "key": null}`,
			`{"kind": "org", "key": true}`,
			`{"kind": "multi", "org": null}`,
			`{"kind": "multi", "org": true}`,
			`{"kind": "org", "key": "my-key", "name": true}`,
			`{"kind": "org", "key": "my-key", "transient": "yes"}}`,
			`{"kind": "org", "key": "my-key", "transient": null}}`,

			`{"kind": "org"}`,             // missing key
			`{"kind": "user", "key": ""}`, // empty key not allowed in new-style context
			`{"kind": "kind"}`,            // illegal kind

			// wrong type within _meta
			`{"kind": "org", "key": "my-key", "_meta": true}}`,
			`{"kind": "org", "key": "my-key", "_meta": {"secondary": true}}}`,
			`{"kind": "org", "key": "my-key", "_meta": {"privateAttributes": true}}}`,

			`{"kind": "multi"}`,                                           // multi kind with no kinds
			`{"kind": "multi", "user": {"key": ""}}`,                      // multi kind where subcontext fails validation
			`{"kind": "multi", "user": {"key": true}}`,                    // multi kind where subcontext is malformed
			`{"kind": "multi", "org": {"key": "x"}, "org": {"key": "y"}}`, // multi kind with repeated kind

			// wrong types in old user schema
			`{"key": null}`,
			`{"key": true}`,
			`{"key": "my-key", "secondary": true}`,
			`{"key": "my-key", "anonymous": "x"}`,
			`{"key": "my-key", "name": true}`,
			`{"key": "my-key", "firstName": true}`,
			`{"key": "my-key", "lastName": true}`,
			`{"key": "my-key", "email": true}`,
			`{"key": "my-key", "country": true}`,
			`{"key": "my-key", "avatar": true}`,
			`{"key": "my-key", "ip": true}`,
			`{"key": "my-key", "custom": true}`,
			`{"key": "my-key", "privateAttributeNames": true}`,
		} {
			t.Run(badJSON, func(t *testing.T) {
				var c Context
				err := unmarshalFn(&c, []byte(badJSON))
				assert.Error(t, err)
			})
		}
	})
}

func jsonUnmarshalTestFn(c *Context, data []byte) error {
	return json.Unmarshal(data, c)
}

func jsonStreamUnmarshalTestFn(c *Context, data []byte) error {
	r := jreader.NewReader(data)
	ContextSerialization{}.UnmarshalFromJSONReader(&r, c)
	return r.Error()
}

func TestContextJSONUnmarshal(t *testing.T) {
	contextUnmarshalingTests(t, jsonUnmarshalTestFn)
}

func TestContextReadFromJSONReader(t *testing.T) {
	contextUnmarshalingTests(t, jsonStreamUnmarshalTestFn)
}

func TestContextUnmarshalEventOutputFormat(t *testing.T) {
	for _, p := range makeEventOutputFormatUnmarshalingParams() {
		t.Run(p.json, func(t *testing.T) {
			r := jreader.NewReader([]byte(p.json))
			var c Context
			ContextSerialization{}.UnmarshalFromJSONReaderEventOutputFormat(&r, &c)
			assert.NoError(t, r.Error())
			assert.Equal(t, p.context, c)
		})
	}
}
