package ldcontext

import (
	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

// UnmarshalJSON provides JSON deserialization for Context when using json.UnmarshalJSON.
//
// LaunchDarkly's JSON schema for contexts is standardized across SDKs. For unmarshaling, there are
// three supported formats:
//
// (TKTK: consider moving all this content into a non-platform-specific online docs page, since none
// of this is specific to Go)
//
// 1. A single-kind context, identified by a top-level "kind" property that is not "multi".
//
// 2. A multi-kind context, identified by a top-level "kind" property of "multi".
//
// 3. A user context in the format used by older LaunchDarkly SDKs. This has no top-level "kind";
// its kind is assumed to be "user". It follows a different layout in which some predefined
// attribute names are top-level properties, while others are within a "custom" property (or, for
// meta-attributes such as "secondary", are within a "_meta" property). Also, unlike new Contexts,
// old-style users were allowed to have an empty string "" as a key.
//
// Trying to unmarshal any non-struct value, including a JSON null, into a Context will return a
// json.UnmarshalTypeError. If you want to unmarshal optional context data that might be null, pass
// a **Context rather than a *Context to json.Unmarshal.
func (c *Context) UnmarshalJSON(data []byte) error {
	return unmarshalInternal(c, data)
}

func unmarshalInternal(c *Context, data []byte) error {
	// Do a first pass where we just check for the "kind" property, because that determines what
	// schema we use to parse everything else.
	kind, hasKind, err := parseKindOnly(data)
	if err != nil {
		return err
	}
	r := jreader.NewReader(data)
	switch {
	case !hasKind:
		return unmarshalOldUserSchema(c, &r)
	case kind == MultiKind:
		return unmarshalMultiKind(c, &r)
	default:
		return unmarshalSingleKind(c, &r, "")
	}
}

func parseKindOnly(data []byte) (Kind, bool, error) {
	r := jreader.NewReader(data)
	for obj := r.Object(); obj.Next(); {
		if string(obj.Name()) == AttrNameKind {
			return Kind(r.String()), true, r.Error()
			// We can immediately return here and not bother parsing the rest of the JSON object; we'll be
			// creating another Reader that'll start over with the same byte slice for the second pass.
		}
		// If we see any property other than "kind" in this loop, just skip it. Calling SkipValue makes
		// the Reader consume and discard the property value so we can advance to the next object property.
		// Unfortunately, since JSON property ordering is indeterminate, we have no way to know how many
		// properties we might see before we see "kind"-- if we see it at all.
		_ = r.SkipValue()
	}
	return "", false, r.Error()
}

func readOptString(r *jreader.Reader) ldvalue.OptionalString {
	if s, nonNull := r.StringOrNull(); nonNull {
		return ldvalue.NewOptionalString(s)
	}
	return ldvalue.OptionalString{}
}

func unmarshalSingleKind(c *Context, r *jreader.Reader, knownKind Kind) error {
	var b Builder
	if knownKind != "" {
		b.Kind(knownKind)
	}
	hasKey := false
	for obj := r.Object(); obj.Next(); {
		switch string(obj.Name()) {
		case AttrNameKind:
			b.Kind(Kind(r.String()))
		case AttrNameKey:
			b.Key(r.String())
			hasKey = true
		case AttrNameName:
			b.OptName(readOptString(r))
		case jsonPropMeta:
			for metaObj := r.Object(); metaObj.Next(); {
				switch string(metaObj.Name()) {
				case AttrNameSecondary:
					b.OptSecondary(readOptString(r))
				case AttrNameTransient:
					b.Transient(r.Bool())
				case jsonPropPrivate:
					for privateArr := r.ArrayOrNull(); privateArr.Next(); {
						b.PrivateRef(NewAttrRef(r.String()))
					}
				default:
					// Unrecognized property names within _meta are ignored. Calling SkipValue makes the Reader
					// consume and discard the property value so we can advance to the next object property.
					_ = r.SkipValue()
				}
			}
		default:
			var v ldvalue.Value
			v.ReadFromJSONReader(r)
			b.SetValue(string(obj.Name()), v)
		}
	}
	if r.Error() != nil {
		return r.Error()
	}
	if !hasKey {
		return errJSONKeyMissing
	}
	*c = b.Build()
	return c.Err()
}

func unmarshalMultiKind(c *Context, r *jreader.Reader) error {
	var b MultiBuilder
	for obj := r.Object(); obj.Next(); {
		name := string(obj.Name())
		if name == AttrNameKind {
			_ = r.SkipValue()
			continue
		}
		var subContext Context
		if err := unmarshalSingleKind(&subContext, r, Kind(name)); err != nil {
			return err
		}
		b.Add(subContext)
	}
	*c = b.Build()
	return c.Err()
}

func unmarshalOldUserSchema(c *Context, r *jreader.Reader) error {
	var b Builder
	b.setAllowEmptyKey(true)
	hasKey := false
	for obj := r.Object(); obj.Next(); {
		switch string(obj.Name()) {
		case AttrNameKey:
			b.Key(r.String())
			hasKey = true
		case AttrNameName:
			b.OptName(readOptString(r))
		case AttrNameSecondary:
			b.OptSecondary(readOptString(r))
		case jsonPropOldUserAnonymous:
			value, _ := r.BoolOrNull()
			b.Transient(value)
		case jsonPropOldUserCustom:
			for customObj := r.Object(); customObj.Next(); {
				name := string(customObj.Name())
				var value ldvalue.Value
				value.ReadFromJSONReader(r)
				b.SetValue(name, value)
			}
		case jsonPropPrivate:
			for privateArr := r.ArrayOrNull(); privateArr.Next(); {
				b.Private(r.String())
				// Note, we use Private here rather than PrivateRef, because the AttrRef syntax is not used
				// in the old user schema; each string here is by definition a literal attribute name.
			}
		case "firstName", "lastName", "email", "country", "avatar", "ip":
			if s := readOptString(r); s.IsDefined() {
				b.SetString(string(obj.Name()), s.StringValue())
			}
		default:
			// In the old user schema, unrecognized top-level property names are ignored. Calling SkipValue
			// makes the Reader consume and discard the property value so we can advance to the next object property.
			_ = r.SkipValue()
		}
	}
	if r.Error() != nil {
		return r.Error()
	}
	if !hasKey {
		return errJSONKeyMissing
	}
	*c = b.Build()
	return c.Err()
}
