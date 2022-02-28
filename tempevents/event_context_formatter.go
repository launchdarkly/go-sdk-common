// Package tempevents contains code under development that will be moved into go-sdk-events. It is
// currently in go-sdk-common-private because it relies on ldcontext code that is not yet ready.
package tempevents

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldcontext"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

// EventContextFormatter provides the special JSON serialization format that is used when including Context
// data in analytics events. In this format, some attribute values may be redacted based on the SDK's
// events configuration and/or the per-Context setting of ldcontext.Builder.Private().
type EventContextFormatter struct {
	allAttributesPrivate bool
	privateAttributes    map[string]*privateAttrLookupNode
}

type privateAttrLookupNode struct {
	attribute *ldcontext.AttrRef
	children  map[string]*privateAttrLookupNode
}

// EventContextFormatterOptions contains optional parameters for creating an EventContextFormatter.
type EventContextFormatterOptions struct {
	// AllAttributesPrivate is true if all optional attributes (that is, not including things like "key")
	// should be considered private. If this is true, then the PrivateAttributes field and the
	// per-context private attributes are ignored.
	AllAttributesPrivate bool

	// PrivateAttributes is a list of attribute references (either simple names, or slash-delimited
	// paths) that should be considered private.
	PrivateAttributes []ldcontext.AttrRef
}

// NewEventContextFormatter creates an EventContextFormatter.
//
// An instance of this type is owned by the eventOutputFormatter that is responsible for writing all
// JSON event data. It is created at SDK initialization time based on the SDK configuration.
func NewEventContextFormatter(options EventContextFormatterOptions) *EventContextFormatter {
	ret := &EventContextFormatter{allAttributesPrivate: options.AllAttributesPrivate}
	if len(options.PrivateAttributes) != 0 {
		// Reformat the list of private attributes into a map structure that will allow
		// for faster lookups.
		ret.privateAttributes = makePrivateAttrLookupData(options.PrivateAttributes)
	}
	return ret
}

func makePrivateAttrLookupData(attrRefList []ldcontext.AttrRef) map[string]*privateAttrLookupNode {
	// This function transforms a list of AttrRefs into a data structure that allows for more efficient
	// implementation of EventContextFormatter.checkGloballyPrivate().
	//
	// For instance, if the original AttrRefs were "/name", "/address/street", and "/address/city",
	// it would produce the following map:
	//
	// "name": {
	//   attribute: NewAttrRef("/name"),
	// },
	// "address": {
	//   children: {
	//     "street": {
	//       attribute: NewAttrRef("/address/street/"),
	//     },
	//     "city": {
	//       attribute: NewAttrRef("/address/city/"),
	//     },
	//   },
	// }
	ret := make(map[string]*privateAttrLookupNode)
	for _, a := range attrRefList {
		parentMap := &ret
		for i := 0; i < a.Depth(); i++ {
			name, _ := a.Component(i)
			if *parentMap == nil {
				*parentMap = make(map[string]*privateAttrLookupNode)
			}
			nextNode := (*parentMap)[name]
			if nextNode == nil {
				nextNode = &privateAttrLookupNode{}
				if i == a.Depth()-1 {
					aa := a
					nextNode.attribute = &aa
				}
				(*parentMap)[name] = nextNode
			}
			parentMap = &nextNode.children
		}
	}
	return ret
}

// WriteContext serializes a Context in the format appropriate for an analytics event, redacting
// private attributes if necessary.
func (f *EventContextFormatter) WriteContext(w *jwriter.Writer, c *ldcontext.Context) {
	if c.Err() != nil {
		w.AddError(c.Err())
		return
	}
	if c.Multiple() {
		f.writeContextInternalMulti(w, c)
	} else {
		f.writeContextInternalSingle(w, c, true)
	}
}

func (f *EventContextFormatter) writeContextInternalSingle(w *jwriter.Writer, c *ldcontext.Context, includeKind bool) {
	obj := w.Object()
	if includeKind {
		obj.Name(ldcontext.AttrNameKind).String(string(c.Kind()))
	}

	obj.Name(ldcontext.AttrNameKey).String(c.Key())

	optionalAttrNames := make([]string, 0, 20) // arbitrary capacity, expanded if necessary by GetOptionalAttributeNames
	redactedAttrs := make([]string, 0, 20)

	optionalAttrNames = c.GetOptionalAttributeNames(optionalAttrNames)

	for _, key := range optionalAttrNames {
		if value, ok := c.GetValue(key); ok {
			if f.allAttributesPrivate {
				// If allAttributesPrivate is true, then there's no complex filtering or recursing to be done: all of
				// these values are by definition private, so just add their names to the redacted list.
				escapedAttrName := ldcontext.NewAttrRefForName(key).String()
				redactedAttrs = append(redactedAttrs, escapedAttrName)
				continue
			}
			path := make([]string, 0, 10)
			f.writeFilteredAttribute(w, c, &obj, path, key, value, &redactedAttrs)
		}
	}

	if c.Transient() || c.Secondary().IsDefined() || len(redactedAttrs) != 0 {
		metaJSON := obj.Name("_meta").Object()
		if c.Transient() {
			metaJSON.Name(ldcontext.AttrNameTransient).Bool(true)
		}
		if s, defined := c.Secondary().Get(); defined {
			metaJSON.Name(ldcontext.AttrNameSecondary).String(s)
		}
		if len(redactedAttrs) != 0 {
			privateAttrsJSON := metaJSON.Name("privateAttrs").Array()
			for _, a := range redactedAttrs {
				privateAttrsJSON.String(a)
			}
			privateAttrsJSON.End()
		}
		metaJSON.End()
	}

	obj.End()
}

func (f *EventContextFormatter) writeContextInternalMulti(w *jwriter.Writer, c *ldcontext.Context) {
	obj := w.Object()
	obj.Name(ldcontext.AttrNameKind).String(string(ldcontext.MultiKind))

	for i := 0; i < c.MultiKindCount(); i++ {
		mc, _ := c.MultiKindByIndex(i)
		obj.Name(string(mc.Kind()))
		f.writeContextInternalSingle(w, &mc, false)
	}

	obj.End()
}

// writeFilteredAttribute checks whether a given value should be considered private, and then
// either writes the attribute to the output JSON object if it is *not* private, or adds the
// corresponding attribute reference to the redactedAttrs list if it is private.
//
// The parentPath parameter indicates where we are in the context data structure. If it is empty,
// we are at the top level and "key" is an attribute name. If it is not empty, we are recursing
// into the properties of an attribute value that is a JSON object: for instance, if parentPath
// is ["billing", "address"] and key is "street", then the top-level attribute is "billing" and
// has a value in the form {"address": {"street": ...}} and we are now deciding whether to
// write the "street" property. See maybeRedact() for the logic involved in that decision.
//
// If allAttributesPrivate is true, this method is never called.
func (f *EventContextFormatter) writeFilteredAttribute(
	w *jwriter.Writer,
	c *ldcontext.Context,
	parentObj *jwriter.ObjectState,
	parentPath []string,
	key string,
	value ldvalue.Value,
	redactedAttrs *[]string,
) {
	path := append(parentPath, key) //nolint:gocritic // purposely not assigning to same slice

	isRedacted, nestedPropertiesAreRedacted := f.maybeRedact(c, path, value.Type(), redactedAttrs)

	if value.Type() != ldvalue.ObjectType {
		// For all value types except object, the question is only "is there a private attribute
		// reference that directly points to this property", since there are no nested properties.
		if !isRedacted {
			parentObj.Name(key)
			value.WriteToJSONWriter(w)
		}
		return
	}

	// If the value is an object, then there are three possible outcomes: 1. this value is
	// completely redacted, so drop it and do not recurse; 2. the value is not redacted, and
	// and neither are any subproperties within it, so output the whole thing as-is; 3. the
	// value itself is not redacted, but some subproperties within it are, so we'll need to
	// recurse through it and filter as we go.
	if isRedacted {
		return // outcome 1
	}
	parentObj.Name(key)
	if !nestedPropertiesAreRedacted {
		value.WriteToJSONWriter(w) // writes the whole value unchanged
		return                     // outcome 2
	}
	subObj := w.Object() // writes the opening brace for the output object
	value.Enumerate(func(index int, subKey string, subValue ldvalue.Value) bool {
		// recurse to write or not write each property - outcome 3
		f.writeFilteredAttribute(w, c, &subObj, path, subKey, subValue, redactedAttrs)
		return true // true here just means "keep enumerating properties"
	})
	subObj.End() // writes the closing brace for the output object
}

// maybeRedact is called by writeFilteredAttribute to decide whether or not a given value (or,
// possibly, properties within it) should be considered private, based on the private attribute
// references in either 1. the EventContextFormatter configuration or 2. this specific Context.
//
// If the value should be private, then the first return value is true, and also the attribute
// reference is added to redactedAttrs.
//
// The second return value indicates whether there are any private attribute references
// designating properties *within* this value. That is, if attrPath is ["address"], and the
// configuration says that "/address/street" is private, then the second return value will be
// true, which tells us that we can't just dump the value of the "address" object directly into
// the output but will need to filter its properties.
//
// Note that even though an AttrRef can contain numeric path components to represent an array
// element lookup, for the purposes of flag evaluations (like "/animals/0" which conceptually
// represents context.animals[0]), those will not work as private attribute references since
// we do not recurse to redact anything within an array value. A reference like "/animals/0"
// would only work if context.animals were an object with a property named "0".
//
// If allAttributesPrivate is true, this method is never called.
func (f *EventContextFormatter) maybeRedact(
	c *ldcontext.Context,
	attrPath []string,
	valueType ldvalue.ValueType,
	redactedAttrs *[]string,
) (bool, bool) {
	// First check against the EventContextFormatter configuration.
	redactedAttrRef, nestedPropertiesAreRedacted := f.checkGlobalPrivateAttrRefs(attrPath)
	if redactedAttrRef != nil {
		*redactedAttrs = append(*redactedAttrs, redactedAttrRef.String())
		return true, false
		// true, false = "this attribute itself is redacted, never mind its children"
	}

	shouldCheckForNestedProperties := valueType == ldvalue.ObjectType

	// Now check the per-Context configuration. Unlike the EventContextFormatter configuration, this
	// does not have a lookup map, just a list of AttrRefs.
	for i := 0; i < c.PrivateAttributeCount(); i++ {
		a, _ := c.PrivateAttributeByIndex(i)
		depth := a.Depth()
		if depth < len(attrPath) {
			// If the attribute reference is shorter than the current path, then it can't possibly be a match,
			// because if it had matched the first part of our path, we wouldn't have recursed this far.
			continue
		}
		if !shouldCheckForNestedProperties && depth > len(attrPath) {
			continue
		}
		match := true
		for j := 0; j < len(attrPath); j++ {
			name, _ := a.Component(j)
			if name != attrPath[j] {
				match = false
				break
			}
		}
		if match {
			if depth == len(attrPath) {
				*redactedAttrs = append(*redactedAttrs, a.String())
				return true, false
				// true, false = "this attribute itself is redacted, never mind its children"
			}
			nestedPropertiesAreRedacted = true
		}
	}
	return false, nestedPropertiesAreRedacted // false = "this attribute itself is not redacted"
}

// Checks whether the given attribute or subproperty matches any AttrRef that was designated as
// private in the SDK options given to newEventContextFormatter.
//
// If attrPath has just one element, it is the name of a top-level attribute. If it has multiple
// elements, it is a path to a property within a custom object attribute: for instance, if you
// represented the overall context as a JSON object, the attrPath ["billing", "address", "street"]
// would refer to the street property within something like {"billing": {"address": {"street": "x"}}}.
//
// The first return value is nil if the attribute does not need to be redacted; otherwise it is the
// specific attribute reference that was matched.
//
// The second return value is true if and only if there's at least one configured private
// attribute reference for *children* of attrPath (and there is not one for attrPath itself, since if
// there was, we would not bother recursing to write the children). See comments on writeFilteredAttribute.
func (f EventContextFormatter) checkGlobalPrivateAttrRefs(attrPath []string) (
	redactedAttrRef *ldcontext.AttrRef, nestedPropertiesAreRedacted bool,
) {
	redactedAttrRef = nil
	nestedPropertiesAreRedacted = false
	lookup := f.privateAttributes
	if lookup == nil {
		return
	}
	for i, pathComponent := range attrPath {
		nextNode := lookup[pathComponent]
		if nextNode == nil {
			break
		}
		if i == len(attrPath)-1 {
			if nextNode.attribute != nil {
				redactedAttrRef = nextNode.attribute
				return
			}
			nestedPropertiesAreRedacted = true
			return
		} else if nextNode.children != nil {
			lookup = nextNode.children
			continue
		}
	}
	return
}
