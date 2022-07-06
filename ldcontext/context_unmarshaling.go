package ldcontext

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"
	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/launchdarkly/go-jsonstream/v2/jreader"
)

// See internalAttributeNameIfPossible().
var internCommonAttributeNamesMap = makeInternCommonAttributeNamesMap() //nolint:gochecknoglobals

func makeInternCommonAttributeNamesMap() map[string]string {
	ret := make(map[string]string)
	for _, a := range []string{"email", "firstName", "lastName", "country", "ip", "avatar"} {
		ret[a] = a
	}
	return ret
}

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
	r := jreader.NewReader(data)
	return ContextSerialization.UnmarshalFromJSONReader(&r, c)
}

func unmarshalFromJSONReader(r *jreader.Reader, c *Context, usingEventFormat bool) {
	// Do a first pass where we just check for the "kind" property, because that determines what
	// schema we use to parse everything else.
	kind, hasKind, err := parseKindOnly(r)
	if err != nil {
		r.AddError(err)
		return
	}
	switch {
	case !hasKind:
		err = unmarshalOldUserSchema(c, r, usingEventFormat)
	case kind == MultiKind:
		err = unmarshalMultiKind(c, r, usingEventFormat)
	default:
		err = unmarshalSingleKind(c, r, "", usingEventFormat)
	}
	if err != nil {
		r.AddError(err)
	}
}

func parseKindOnly(originalReader *jreader.Reader) (Kind, bool, error) {
	// Make an exact copy of the original Reader so that changes in its state will not
	// affect the original Reader; both point to the same []byte array, but each has its
	// own "current position" and "next token" fields.
	r := *originalReader
	for obj := r.Object(); obj.Next(); {
		if string(obj.Name()) == ldattr.KindAttr {
			kind := r.String()
			if r.Error() == nil && kind == "" {
				return "", false, errContextKindEmpty
			}
			return Kind(kind), true, r.Error()
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

func unmarshalSingleKind(c *Context, r *jreader.Reader, knownKind Kind, usingEventFormat bool) error {
	var b Builder
	if knownKind != "" {
		b.Kind(knownKind)
	}
	hasKey := false
	for obj := r.Object(); obj.Next(); {
		switch string(obj.Name()) {
		case ldattr.KindAttr:
			b.Kind(Kind(r.String()))
		case ldattr.KeyAttr:
			b.Key(r.String())
			hasKey = true
		case ldattr.NameAttr:
			b.OptName(readOptString(r))
		case ldattr.AnonymousAttr:
			b.Anonymous(r.Bool())
		case jsonPropMeta:
			for metaObj := r.ObjectOrNull(); metaObj.Next(); {
				switch string(metaObj.Name()) {
				case jsonPropSecondary:
					b.OptSecondary(readOptString(r))
				case jsonPropPrivate:
					if usingEventFormat {
						_ = r.SkipValue()
						continue
					}
					readPrivateAttributes(r, &b, false)
				case jsonPropRedacted:
					if !usingEventFormat {
						_ = r.SkipValue()
						continue
					}
					readPrivateAttributes(r, &b, false)
				default:
					// Unrecognized property names within _meta are ignored. Calling SkipValue makes the Reader
					// consume and discard the property value so we can advance to the next object property.
					_ = r.SkipValue()
				}
			}
		default:
			var v ldvalue.Value
			v.ReadFromJSONReader(r)
			b.SetValue(internAttributeNameIfPossible(obj.Name()), v)
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

func unmarshalMultiKind(c *Context, r *jreader.Reader, usingEventFormat bool) error {
	var b MultiBuilder
	for obj := r.Object(); obj.Next(); {
		name := string(obj.Name())
		if name == ldattr.KindAttr {
			_ = r.SkipValue()
			continue
		}
		var subContext Context
		if err := unmarshalSingleKind(&subContext, r, Kind(name), usingEventFormat); err != nil {
			return err
		}
		b.Add(subContext)
	}
	*c = b.Build()
	return c.Err()
}

func unmarshalOldUserSchema(c *Context, r *jreader.Reader, usingEventFormat bool) error {
	var b Builder
	b.setAllowEmptyKey(true)
	hasKey := false
	for obj := r.Object(); obj.Next(); {
		switch string(obj.Name()) {
		case ldattr.KeyAttr:
			b.Key(r.String())
			hasKey = true
		case ldattr.NameAttr:
			b.OptName(readOptString(r))
		case jsonPropSecondary:
			b.OptSecondary(readOptString(r))
		case ldattr.AnonymousAttr:
			value, _ := r.BoolOrNull()
			b.Anonymous(value)
		case jsonPropOldUserCustom:
			for customObj := r.ObjectOrNull(); customObj.Next(); {
				name := string(customObj.Name())
				var value ldvalue.Value
				value.ReadFromJSONReader(r)
				b.SetValue(name, value)
			}
		case jsonPropOldUserPrivate:
			if usingEventFormat {
				_ = r.SkipValue()
				continue
			}
			readPrivateAttributes(r, &b, true)
			// The "true" here means to interpret the strings as literal attribute names, since the
			// attribute reference path syntax was not used in the old user schema.
		case jsonPropOldUserRedacted:
			if !usingEventFormat {
				_ = r.SkipValue()
				continue
			}
			readPrivateAttributes(r, &b, true)
		case "firstName", "lastName", "email", "country", "avatar", "ip":
			if s := readOptString(r); s.IsDefined() {
				b.SetString(internAttributeNameIfPossible(obj.Name()), s.StringValue())
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

// internAttributeNameIfPossible takes a byte slice representing a property name, and returns an existing
// string if we already have a string literal equal to that name; otherwise it converts the bytes to a string.
//
// The reason for this logic is that LaunchDarkly-enabled applications will generally send the same attribute
// names over and over again, and we can guess what many of them will be. The old user model had standard
// top-level properties with predefined names like "email", which now are mostly considered custom attributes
// that are stored as map entries instead of struct fields. In a high-traffic environment where many contexts
// are being deserialized, i.e. the LD client-side service endpoints, if we are servicing 1000 requests that
// each have users with "firstName" and "lastName" attributes, it's desirable to reuse those strings rather
// than allocating a new string each time; the overall memory usage may be negligible but the allocation and
// GC overhead still adds up.
//
// Recent versions of Go have an optimization for looking up string(x) as a string key in a map if x is a
// byte slice, so that it does *not* have to allocate a string instance just to do this.
func internAttributeNameIfPossible(nameBytes []byte) string {
	if internedName, ok := internCommonAttributeNamesMap[string(nameBytes)]; ok {
		return internedName
	}
	return string(nameBytes)
}

func unmarshalWithKindAndKeyOnly(r *jreader.Reader, c *Context) {
	kind, hasKind, err := parseKindOnly(r)
	if err != nil {
		r.AddError(err)
		return
	}
	switch {
	case !hasKind:
		err = unmarshalWithKindAndKeyOnlyOldUser(r, c)
	case kind == MultiKind:
		var mb MultiBuilder
		for obj := r.Object(); obj.Next(); {
			switch string(obj.Name()) {
			case ldattr.KindAttr:
				_ = r.SkipValue()
			default:
				kind := Kind(obj.Name())
				var mc Context
				if err = unmarshalWithKindAndKeyOnlySingleKind(r, &mc, kind); err != nil {
					break
				}
				mb.Add(mc)
			}
		}
		*c = mb.Build()
		err = c.Err()
	default:
		err = unmarshalWithKindAndKeyOnlySingleKind(r, c, "")
	}
	if err != nil {
		r.AddError(err)
	}
}

func unmarshalWithKindAndKeyOnlySingleKind(r *jreader.Reader, c *Context, kind Kind) error {
	var key string
	hasKey := false
	for obj := r.Object(); obj.Next(); {
		switch string(obj.Name()) {
		case ldattr.KindAttr:
			kind = Kind(r.String())
		case ldattr.KeyAttr:
			key = r.String()
			hasKey = true
		default:
			_ = r.SkipValue()
		}
	}
	if !hasKey {
		r.AddError(errJSONKeyMissing)
		return errJSONKeyMissing
	}
	*c = NewWithKind(kind, key)
	return c.Err()
}

func unmarshalWithKindAndKeyOnlyOldUser(r *jreader.Reader, c *Context) error {
	for obj := r.Object(); obj.Next(); {
		switch string(obj.Name()) {
		case ldattr.KeyAttr:
			key := r.String()
			var b Builder
			*c = b.setAllowEmptyKey(true).Key(key).Build()
			return c.Err()
		default:
			_ = r.SkipValue()
		}
	}
	r.AddError(errJSONKeyMissing)
	return errJSONKeyMissing
}

func readPrivateAttributes(r *jreader.Reader, b *Builder, asLiterals bool) {
	for privateArr := r.ArrayOrNull(); privateArr.Next(); {
		b.PrivateRef(refOrLiteralRef(r.String(), asLiterals))
	}
}

func refOrLiteralRef(s string, asLiteral bool) ldattr.Ref {
	if asLiteral {
		return ldattr.NewLiteralRef(s)
	}
	return ldattr.NewRef(s)
}
