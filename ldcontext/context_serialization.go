package ldcontext

import (
	"github.com/launchdarkly/go-jsonstream/v2/jreader"
	"github.com/launchdarkly/go-jsonstream/v2/jwriter"
)

// Note: other ContextSerialization methods are in the conditionally-compiled file
// context_easyjson.go.

// ContextSerializationMethods contains JSON marshaling and unmarshaling methods that are not
// normally used directly by applications. These methods are exported because they are used in
// LaunchDarkly service code and the Relay Proxy.
type ContextSerializationMethods struct{}

// ContextSerialization is the global entry point for ContextSerializationMethods.
var ContextSerialization ContextSerializationMethods //nolint:gochecknoglobals

// UnmarshalFromJSONReader unmarshals a Context with the jsonstream Reader API.
//
// In case of failure, the error is both returned from the method and stored as a failure state in
// the Reader.
func (s ContextSerializationMethods) UnmarshalFromJSONReader(r *jreader.Reader, c *Context) error {
	unmarshalFromJSONReader(r, c)
	return r.Error()
}

// UnmarshalWithKindAndKeyOnly is a special unmarshaling mode where all properties except kind and
// key are discarded. It works for both single-kind and multi-kind contexts. This is more efficient
// than the regular unmarshaling logic in situations where contexts need to be indexed by Key or
// FullyQualifiedKey.
//
// Because most properties are discarded immediately without checking their value, some error
// conditions (for instance, a "name" property whose value is not a string) will not be detected by
// this method. It will fail only if validation related to the kind or key fails.
func (s ContextSerializationMethods) UnmarshalWithKindAndKeyOnly(r *jreader.Reader, c *Context) error {
	unmarshalWithKindAndKeyOnly(r, c)
	return r.Error()
}

// MarshalToJSONWriter marshals a Context with the jsonstream Writer API.
func (s ContextSerializationMethods) MarshalToJSONWriter(w *jwriter.Writer, c *Context) {
	if err := c.Err(); err != nil {
		w.AddError(err)
		return
	}
	if c.multiContexts == nil {
		c.writeToJSONWriterInternalSingle(w, "")
	} else {
		c.writeToJSONWriterInternalMulti(w)
	}
}
