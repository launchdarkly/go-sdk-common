package ldcontext

import (
	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

// Note: other ContextSerialization methods are in the conditionally-compiled file
// context_easyjson.go.

// ContextSerialization is a container for JSON marshaling and unmarshaling methods that are
// not normally used directly by applications. These methods are exported because they are used
// in LaunchDarkly service code and the Relay Proxy.
type ContextSerialization struct{}

// UnmarshalFromJSONReader unmarshals a Context with the jsonstream Reader API.
func (s ContextSerialization) UnmarshalFromJSONReader(r *jreader.Reader, c *Context) {
	unmarshalFromJSONReader(r, c, false)
}

// UnmarshalFromJSONReaderEventOutputFormat unmarshals a Context with the jsonstream Reader API,
// using the alternate JSON schema that is used for contexts in analytics event data, where the
// property _meta.redactedAttributes is translated into Builder.PreviouslyRedacted(). This can be
// used by the Relay Proxy or other LaunchDarkly services when reading events sent by SDKs.
//
// The marshaler for this schema is not implemented in go-sdk-common; all event output is
// produced by the go-sdk-events package.
func (s ContextSerialization) UnmarshalFromJSONReaderEventOutputFormat(r *jreader.Reader, c *Context) {
	unmarshalFromJSONReader(r, c, true)
}

// MarshalToJSONWriter marshals a Context with the jsonstream Writer API.
func (s ContextSerialization) MarshalToJSONWriter(w *jwriter.Writer, c *Context) {
	if c.err != nil {
		w.AddError(c.err)
		return
	}
	if c.multiContexts == nil {
		c.writeToJSONWriterInternalSingle(w, "")
	} else {
		c.writeToJSONWriterInternalMulti(w)
	}
}
