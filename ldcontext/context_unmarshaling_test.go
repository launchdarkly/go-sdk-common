package ldcontext

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/launchdarkly/go-jsonstream/v3/jreader"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func makeContextUnmarshalUnimportantVariantsParams() []contextSerializationParams {
	// These are test cases that only apply to unmarshaling, because marshaling will never produce this specific JSON.
	return []contextSerializationParams{
		// explicit null is same as unset for optional string attrs
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "name": null}`},

		// explicit null is same as unset for custom attrs
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "customAttr": null}`},

		// explicit false is same as unset for anonymous
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "anonymous": false}`},

		// _meta: {} is same as no _meta
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {}}`},

		// _meta: null is same as no _meta
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": null}`},

		// privateAttributes: [] is same as no privateAttributes
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {"privateAttributes": null}}`},

		// privateAttributes: null is same as no privateAttributes
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {"privateAttributes": null}}`},

		// unrecognized properties within _meta are ignored
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {"unknownProp": false}}`},

		// redactedAttributes is only a thing in the event output format, not the regular format
		{NewBuilder("my-key").Build(),
			`{"kind": "user", "key": "my-key", "_meta": {"redactedAttributes": ["name"]}}`},
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

		{contextWithSecondary(New("key3"), "value"),
			`{"key": "key3", "secondary": "value"}`},
		{New("key3"),
			`{"key": "key3", "secondary": null}`},

		{NewBuilder("key4").Anonymous(true).Build(),
			`{"key": "key4", "anonymous": true}`},
		{NewBuilder("key4").Build(),
			`{"key": "key4", "anonymous": false}`},
		{NewBuilder("key4").Build(),
			`{"key": "key4", "anonymous": null}`},

		{NewBuilder("key6").Build(),
			`{"key": "key6", "custom": {}}`},
		{NewBuilder("key6").Build(),
			`{"key": "key6", "custom": null}`},
		{NewBuilder("key6").Build(),
			`{"key": "key6", "custom": {"attr1": null}}`},
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

		{NewBuilder("key9").Name("x").SetInt("a", 1).Anonymous(true).Build(),
			`{"key": "key9", "name": "x", "anonymous": true, "custom": ` +
				`{"kind": ".", "key": ".", "name": ".", "anonymous": false, "_meta": true, "a": 1}}`},
		// don't let custom attributes overwrite top-level properties with the same reserved names
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

func makeAllContextUnmarshalingEventOutputFormatParams() []contextSerializationParams {
	var ret []contextSerializationParams
	for _, p := range makeAllContextUnmarshalingParams() {
		transformed := p
		transformed.json = translateRegularContextJSONToEventOutputJSONAndViceVersa(transformed.json)
		ret = append(ret, transformed)
	}
	return ret
}

func makeContextUnmarshalingErrorInputs() []string {
	return []string{
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
		`{"kind": "org", "key": "my-key", "anonymous": "yes"}}`,
		`{"kind": "org", "key": "my-key", "anonymous": null}}`,

		`{"kind": "org"}`,             // missing key
		`{"kind": "user", "key": ""}`, // empty key not allowed in new-style context
		`{"kind": "", "key": "x"}`,    // empty kind not allowed in new-style context
		`{"kind": "ørg", "key": "x"}`, // illegal kind

		// wrong type within _meta
		`{"kind": "org", "key": "my-key", "_meta": true}}`,
		`{"kind": "org", "key": "my-key", "_meta": {"privateAttributes": true}}}`,

		`{"kind": "multi"}`,                                                       // multi kind with no kinds
		`{"kind": "multi", "user": {"key": ""}}`,                                  // multi kind where subcontext fails validation
		`{"kind": "multi", "user": {"key": true}}`,                                // multi kind where subcontext is malformed
		`{"kind": "multi", "org": {"key": "x"}, "org": {"key": "y"}}`,             // multi kind with repeated kind
		`{"kind": "multi", "multi": {"user": {"key": "a"}, "org": {"key": "x"}}}`, // multi within multi

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

		// missing key in old user schema
		`{"name": "x"}`,
	}
}

func makeContextUnmarshalingEventOutputFormatErrorInputs() []string {
	var ret []string
	for _, s := range makeContextUnmarshalingErrorInputs() {
		ret = append(ret, translateRegularContextJSONToEventOutputJSONAndViceVersa(s))
	}
	return ret
}

func translateRegularContextJSONToEventOutputJSONAndViceVersa(s string) string {
	// In the regular test inputs, redactedAttributes and privateAttrs are property names we do *not*
	// expect to see because they are for the event output format, so the test inputs that reference
	// them have a "this should be ignored" expectation.
	s = strings.ReplaceAll(s, `"redactedAttributes"`, `<temp1>`)
	s = strings.ReplaceAll(s, `"privateAttrs"`, `<temp2>`) // old schema

	s = strings.ReplaceAll(s, `"privateAttributes"`, `"redactedAttributes"`)
	s = strings.ReplaceAll(s, `"privateAttributeNames"`, `"privateAttrs"`) // old schema

	// ...so, in the event output format, we use privateAttributes and privateAttributeNames for the
	// "this should be ignored" cases.
	s = strings.ReplaceAll(s, `<temp1>`, `"privateAttributes"`)
	s = strings.ReplaceAll(s, `<temp2>`, `"privateAttributeNames"`)

	return s
}

func contextUnmarshalingTests(
	t *testing.T,
	unmarshalSingleContextFn func(*Context, []byte) error,
	unmarshalArrayFn func(*[]Context, []byte) error,
) {
	t.Run("valid data", func(t *testing.T) {
		for _, p := range makeAllContextUnmarshalingParams() {
			t.Run(p.json, func(t *testing.T) {
				var c Context
				err := unmarshalSingleContextFn(&c, []byte(p.json))
				assert.NoError(t, err)
				assert.Equal(t, p.context, c)
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		for _, badJSON := range makeContextUnmarshalingErrorInputs() {
			t.Run(badJSON, func(t *testing.T) {
				var c Context
				err := unmarshalSingleContextFn(&c, []byte(badJSON))
				assert.Error(t, err)
			})
		}
	})

	if unmarshalArrayFn != nil {
		t.Run("within an array", func(t *testing.T) {
			// This test shows that in our streaming implementations, the unmarshaler leaves the input
			// stream in the proper state at the end of the object, so it will correctly parse a
			// subsequent value.
			invariantSecondContext := New("simple")
			invariantSecondContextJSON := `{"kind": "user", "key": "simple"}`
			for _, p := range makeAllContextUnmarshalingParams() {
				t.Run(p.json, func(t *testing.T) {
					input := `[ ` + p.json + `, ` + invariantSecondContextJSON + ` ]`
					var cs []Context
					err := unmarshalArrayFn(&cs, []byte(input))
					assert.NoError(t, err)
					if assert.Len(t, cs, 2) {
						assert.Equal(t, p.context, cs[0])
						assert.Equal(t, invariantSecondContext, cs[1])
					}
				})
			}
		})
	}

	t.Run("unmarshaling twice", func(t *testing.T) {
		// This test verifies that if you take an uninitialized Context{}, and unmarshal it from JSON, and
		// then unmarshal the same instance again, you get the same result (i.e. fields are overwritten
		// rather than being appended).
		for _, p := range makeAllContextUnmarshalingParams() {
			t.Run(p.json, func(t *testing.T) {
				var c Context
				err := unmarshalSingleContextFn(&c, []byte(p.json))
				// Here we're using an if, rather than an assert, because we already asserted these conditions
				// in the "valid data" test. If something is really wrong such that it's returning an error or
				// the Equal test fails, we want that to show up more specifically in that test, rather than
				// producing confounding results in this one.
				if err == nil && c.Equal(p.context) {
					c1 := c
					require.NoError(t, unmarshalSingleContextFn(&c, []byte(p.json)))
					assert.Equal(t, c1, c)
				}
			})
		}
	})
}

func jsonUnmarshalTestFn(c *Context, data []byte) error {
	return json.Unmarshal(data, c)
}

func jsonStreamUnmarshalTestFn(c *Context, data []byte) error {
	r := jreader.NewReader(data)
	ContextSerialization.UnmarshalFromJSONReader(&r, c)
	return r.Error()
}

func jsonStreamUnmarshalArrayTestFn(cs *[]Context, data []byte) error {
	r := jreader.NewReader(data)
	for arr := r.Array(); arr.Next(); {
		var c Context
		ContextSerialization.UnmarshalFromJSONReader(&r, &c)
		*cs = append(*cs, c)
	}
	return r.Error()
}

func TestContextJSONUnmarshal(t *testing.T) {
	contextUnmarshalingTests(t, jsonUnmarshalTestFn, nil)
}

func TestContextReadFromJSONReader(t *testing.T) {
	contextUnmarshalingTests(t, jsonStreamUnmarshalTestFn, jsonStreamUnmarshalArrayTestFn)
}

func TestContextUnmarshalEventOutputFormat(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		for _, p := range makeAllContextUnmarshalingEventOutputFormatParams() {
			t.Run(p.json, func(t *testing.T) {
				r := jreader.NewReader([]byte(p.json))
				var c EventOutputContext
				ContextSerialization.UnmarshalFromJSONReaderEventOutput(&r, &c)
				assert.NoError(t, r.Error())
				assert.Equal(t, p.context, c.Context)
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		for _, badJSON := range makeContextUnmarshalingEventOutputFormatErrorInputs() {
			t.Run(badJSON, func(t *testing.T) {
				r := jreader.NewReader([]byte(badJSON))
				var c EventOutputContext
				ContextSerialization.UnmarshalFromJSONReaderEventOutput(&r, &c)
				assert.Error(t, r.Error())
			})
		}
	})
}

func TestContextReadKindAndKeyOnly(t *testing.T) {
	t.Run("valid data", func(t *testing.T) {
		for _, p := range makeAllContextUnmarshalingParams() {
			t.Run(p.json, func(t *testing.T) {
				var fullContext, minimalContext Context

				r1 := jreader.NewReader([]byte(p.json))
				err := ContextSerialization.UnmarshalFromJSONReader(&r1, &fullContext)
				require.NoError(t, err)
				require.NoError(t, r1.Error())
				r2 := jreader.NewReader([]byte(p.json))
				err = ContextSerialization.UnmarshalWithKindAndKeyOnly(&r2, &minimalContext)
				require.NoError(t, err)
				require.NoError(t, r2.Error())

				allFull, allMini := fullContext.GetAllIndividualContexts(nil), minimalContext.GetAllIndividualContexts(nil)
				if assert.Len(t, allMini, len(allFull)) {
					for i, fc := range allFull {
						mc := allMini[i]
						assert.Equal(t, fc.Kind(), mc.Kind())
						assert.Equal(t, fc.Key(), mc.Key())
					}
				}
				assert.Equal(t, fullContext.FullyQualifiedKey(), minimalContext.FullyQualifiedKey())
			})
		}
	})

	t.Run("error cases", func(t *testing.T) {
		for _, badJSON := range []string{
			// This is a deliberately shorter list than the error cases in the regular unmarshaling tests,
			// because this method does not validate every property.
			`null`,
			`false`,
			`1`,
			`"x"`,
			`[]`,
			`{}`,

			// wrong type for kind/key, or individual context was not an object
			`{"kind": null}`,
			`{"kind": true}`,
			`{"kind": "org", "key": null}`,
			`{"kind": "org", "key": true}`,
			`{"kind": "multi", "org": null}`,
			`{"kind": "multi", "org": true}`,

			`{"kind": "org"}`,             // missing key
			`{"kind": "user", "key": ""}`, // empty key not allowed in new-style context
			`{"kind": "", "key": "x"}`,    // empty kind not allowed in new-style context
			`{"kind": "ørg", "key": "x"}`, // illegal kind

			`{"kind": "multi"}`,                                           // multi kind with no kinds
			`{"kind": "multi", "user": {"key": ""}}`,                      // multi kind where subcontext fails validation
			`{"kind": "multi", "user": {"key": true}}`,                    // multi kind where subcontext is malformed
			`{"kind": "multi", "org": {"key": "x"}, "org": {"key": "y"}}`, // multi kind with repeated kind

			// wrong types in old user schema
			`{"key": null}`,
			`{"key": true}`,

			// missing key in old user schema
			`{"name": "x"}`,
		} {
			t.Run(badJSON, func(t *testing.T) {
				var c Context
				r := jreader.NewReader([]byte(badJSON))
				err := ContextSerialization.UnmarshalWithKindAndKeyOnly(&r, &c)
				assert.Error(t, err)
			})
		}
	})
}

func contextWithSecondary(c Context, s string) Context {
	c.secondary = ldvalue.NewOptionalString(s)
	return c
}
