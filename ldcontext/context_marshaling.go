package ldcontext

import (
	"github.com/launchdarkly/go-sdk-common/v3/ldattr"

	"github.com/launchdarkly/go-jsonstream/v2/jwriter"
)

// MarshalJSON provides JSON serialization for Context when using json.MarshalJSON.
//
// LaunchDarkly's JSON schema for contexts is standardized across SDKs. There are two output formats,
// depending on whether it is a single-kind context or a multi-kind context. Unlike the unmarshaler,
// the marshaler never uses the old-style user context schema from older SDKs.
//
// If the Context is invalid (that is, it returns a non-nil Error()) then marshaling fails with the
// same error.
func (c Context) MarshalJSON() ([]byte, error) {
	w := jwriter.NewWriter()
	ContextSerialization.MarshalToJSONWriter(&w, &c)
	return w.Bytes(), w.Error()
}

func (c *Context) writeToJSONWriterInternalSingle(w *jwriter.Writer, withinKind Kind) {
	obj := w.Object()
	if withinKind == "" {
		obj.Name(ldattr.KindAttr).String(string(c.kind))
	}

	obj.Name(ldattr.KeyAttr).String(c.key)
	if c.name.IsDefined() {
		obj.Name(ldattr.NameAttr).String(c.name.StringValue())
	}
	keys := make([]string, 0, 50) // arbitrary size to preallocate on stack
	for _, k := range c.attributes.Keys(keys) {
		obj.Name(k)
		c.attributes.Get(k).WriteToJSONWriter(w)
	}
	if c.transient {
		obj.Name(ldattr.TransientAttr).Bool(true)
	}

	if c.secondary.IsDefined() || len(c.privateAttrs) != 0 {
		metaJSON := obj.Name(jsonPropMeta).Object()
		if c.secondary.IsDefined() {
			metaJSON.Name(jsonPropSecondary).String(c.secondary.StringValue())
		}
		if len(c.privateAttrs) != 0 {
			privateAttrsJSON := metaJSON.Name(jsonPropPrivate).Array()
			for _, a := range c.privateAttrs {
				privateAttrsJSON.String(a.String())
			}
			privateAttrsJSON.End()
		}
		metaJSON.End()
	}

	obj.End()
}

func (c Context) writeToJSONWriterInternalMulti(w *jwriter.Writer) {
	obj := w.Object()
	obj.Name(ldattr.KindAttr).String(string(MultiKind))

	for _, mc := range c.multiContexts {
		obj.Name(string(mc.Kind()))
		mc.writeToJSONWriterInternalSingle(w, mc.Kind())
	}

	obj.End()
}
