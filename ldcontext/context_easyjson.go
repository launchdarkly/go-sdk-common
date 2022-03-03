//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package ldcontext

import (
	"gopkg.in/launchdarkly/go-sdk-common.v3/ldvalue"

	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"

	"github.com/mailru/easyjson/jlexer"
	ej_jwriter "github.com/mailru/easyjson/jwriter"
)

// This conditionally-compiled file provides custom marshal and unmarshal functions for the Context
// type in EasyJSON.
//
// EasyJSON's code generator does recognize the same MarshalJSON and UnmarshalJSON methods used by
// encoding/json, and will call them if present. But this mechanism is inefficient: when marshaling
// it requires the allocation of intermediate byte slices, and when unmarshaling it causes the
// JSON object to be parsed twice. It is preferable to have our marshal/unmarshal methods write to
// and read from the EasyJSON Writer/Lexer directly.
//
// Unmarshaling is the most performance-critical code path, because the client-side endpoints of
// the LD back-end use this implementation to get the context parameters for every request. So,
// rather than using an adapter to delegate jsonstream operations to EasyJSON, as we do for many
// other types-- which is preferred if performance is a bit less critical, because then we only
// have to write the logic once-- the Context unmarshaler is fully reimplemented here with direct
// calls to EasyJSON lexer methods. This allows us to take full advantage of EasyJSON optimizations
// that are available in our service code but may not be available in customer application code,
// such as the use of the unsafe package for direct []byte-to-string conversion.
//
// This means that if we make changes to the schema or the unmarshaling logic, we will need to
// update both context_unmarshaling.go and context_easyjson.go. Our unit tests run the same test
// data against both implementations to verify that they are in sync.
//
// For more information, see: https://gopkg.in/launchdarkly/go-jsonstream.v1

func (c Context) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	if c.err != nil {
		writer.Error = c.err
		return
	}
	wrappedWriter := jwriter.NewWriterFromEasyJSONWriter(writer)
	c.WriteToJSONWriter(&wrappedWriter)
}

func (c *Context) UnmarshalEasyJSON(in *jlexer.Lexer) {
	if in.IsNull() {
		in.Delim('{') // to trigger an "expected an object, got null" error
		return
	}

	// Do a first pass where we just check for the "kind" property, because that determines what
	// schema we use to parse everything else.
	kind, hasKind, err := parseKindOnlyEasyJSON(in)
	if err != nil {
		in.AddError(err)
		return
	}

	switch {
	case !hasKind:
		unmarshalOldUserSchemaEasyJSON(c, in)
	case kind == MultiKind:
		unmarshalMultiKindEasyJSON(c, in)
	default:
		unmarshalSingleKindEasyJSON(c, in, "")
	}
}

func unmarshalSingleKindEasyJSON(c *Context, in *jlexer.Lexer, knownKind Kind) {
	c.defined = true
	if knownKind != "" {
		c.kind = Kind(knownKind)
	}
	hasKey := false
	in.Delim('{')
	for !in.IsDelim('}') {
		// Because the field name will often be a literal that we won't be retaining, we don't want the overhead
		// of allocating a string for it every time. So we call UnsafeBytes(), which still reads a JSON string
		// like String(), but returns the data as a subslice of the existing byte slice if possible-- allocating
		// a new byte slice only in the unlikely case that there were escape sequences. Go's switch statement is
		// optimized so that doing "switch string(key)" does *not* allocate a string, but just uses the bytes.
		key := in.UnsafeBytes()
		in.WantColon()
		switch string(key) {
		case AttrNameKind:
			c.kind = Kind(in.String())
		case AttrNameKey:
			c.key = in.String()
			hasKey = true
		case AttrNameName:
			c.name = readOptStringEasyJSON(in)
		case AttrNameTransient:
			c.transient = in.Bool()
		case jsonPropMeta:
			in.Delim('{')
			for !in.IsDelim('}') {
				key := in.UnsafeBytes() // see comment above
				in.WantColon()
				switch string(key) {
				case jsonPropSecondary:
					c.secondary = readOptStringEasyJSON(in)
				case jsonPropPrivate:
					if in.IsNull() {
						in.Skip()
					} else {
						in.Delim('[')
						for !in.IsDelim(']') {
							c.privateAttrs = append(c.privateAttrs, NewAttrRef(in.String()))
							in.WantComma()
						}
						in.Delim(']')
					}
				default:
					// Unrecognized property names within _meta are ignored. Calling SkipRecursive makes the Lexer
					// consume and discard the property value so we can advance to the next object property.
					in.SkipRecursive()
				}
				in.WantComma()
			}
			in.Delim('}')
		default:
			if in.IsNull() {
				in.Skip()
			} else {
				var v ldvalue.Value
				v.UnmarshalEasyJSON(in)
				if c.attributes == nil {
					c.attributes = make(map[string]ldvalue.Value)
				}
				c.attributes[internAttributeNameIfPossible(key)] = v
			}
		}
		in.WantComma()
	}
	in.Delim('}')
	if in.Error() != nil {
		return
	}
	if !hasKey {
		in.AddError(errJSONKeyMissing)
		return
	}
	c.kind, c.err = validateSingleKind(c.kind)
	if c.err != nil {
		in.AddError(c.err)
		return
	}
	if c.key == "" {
		c.err = errContextKeyEmpty
		in.AddError(c.err)
	} else {
		c.fullyQualifiedKey = makeFullyQualifiedKeySingleKind(c.kind, c.key, true)
	}
}

func unmarshalMultiKindEasyJSON(c *Context, in *jlexer.Lexer) {
	var b MultiBuilder
	in.Delim('{')
	for !in.IsDelim('}') {
		name := in.String()
		in.WantColon()
		if name == AttrNameKind {
			in.SkipRecursive()
		} else {
			var subContext Context
			unmarshalSingleKindEasyJSON(&subContext, in, Kind(name))
			b.Add(subContext)
		}
		in.WantComma()
	}
	in.Delim('}')
	if in.Error() == nil {
		*c = b.Build()
		if err := c.Err(); err != nil {
			in.AddError(err)
		}
	}
}

func unmarshalOldUserSchemaEasyJSON(c *Context, in *jlexer.Lexer) {
	c.defined = true
	c.kind = DefaultKind
	hasKey := false
	in.Delim('{')
	for !in.IsDelim('}') {
		// See comment about UnsafeBytes in unmarshalSingleKindEasyJSON.
		key := in.UnsafeBytes()
		in.WantColon()
		switch string(key) {
		case AttrNameKey:
			c.key = in.String()
			hasKey = true
		case AttrNameName:
			c.name = readOptStringEasyJSON(in)
		case jsonPropSecondary:
			c.secondary = readOptStringEasyJSON(in)
		case jsonPropOldUserAnonymous:
			if in.IsNull() {
				in.Skip()
				c.transient = false
			} else {
				c.transient = in.Bool()
			}
		case jsonPropOldUserCustom:
			in.Delim('{')
			for !in.IsDelim('}') {
				name := in.String()
				in.WantColon()
				if in.IsNull() {
					in.Skip()
				} else {
					var value ldvalue.Value
					value.UnmarshalEasyJSON(in)
					if c.attributes == nil {
						c.attributes = make(map[string]ldvalue.Value)
					}
					c.attributes[name] = value
				}
				in.WantComma()
			}
			in.Delim('}')
		case jsonPropPrivate:
			if in.IsNull() {
				in.Skip()
			} else {
				in.Delim('[')
				for !in.IsDelim(']') {
					c.privateAttrs = append(c.privateAttrs, NewAttrRefForName(in.String()))
					in.WantComma()
				}
				in.Delim(']')
			}
		case "firstName", "lastName", "email", "country", "avatar", "ip":
			if in.IsNull() {
				in.Skip()
			} else {
				name := internAttributeNameIfPossible(key)
				if c.attributes == nil {
					c.attributes = make(map[string]ldvalue.Value)
				}
				c.attributes[name] = ldvalue.String(in.String())
			}
		default:
			// In the old user schema, unrecognized top-level property names are ignored. Calling SkipRecursive
			// makes the Lexer consume and discard the property value so we can advance to the next object property.
			in.SkipRecursive()
		}
		in.WantComma()
	}
	if in.Error() != nil {
		return
	}
	if !hasKey {
		in.AddError(errJSONKeyMissing)
		return
	}
	c.fullyQualifiedKey = c.key
}

func parseKindOnlyEasyJSON(originalLexer *jlexer.Lexer) (Kind, bool, error) {
	// Make an exact copy of the original lexer so that changes in its state will not
	// affect the original lexer; both point to the same []byte array, but each has its
	// own "current position" and "next token" fields.
	in := *originalLexer
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if key == AttrNameKind {
			kind := in.String()
			return Kind(kind), true, in.Error()
		}
		in.SkipRecursive()
		in.WantComma()
	}
	in.Delim('}')
	return "", false, in.Error()
}

func readOptStringEasyJSON(in *jlexer.Lexer) ldvalue.OptionalString {
	if in.IsNull() {
		in.Skip()
		return ldvalue.OptionalString{}
	} else {
		return ldvalue.NewOptionalString(in.String())
	}
}
