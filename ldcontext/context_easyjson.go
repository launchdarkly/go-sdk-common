//go:build launchdarkly_easyjson
// +build launchdarkly_easyjson

package ldcontext

import (
	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"

	"github.com/mailru/easyjson/jlexer"
	ej_jwriter "github.com/mailru/easyjson/jwriter"
)

// This conditionally-compiled file provides a custom marshal function for the Context type in
// EasyJSON.
//
// EasyJSON's code generator does recognize the same MarshalJSON and UnmarshalJSON methods used by
// encoding/json, and will call them if present. But this mechanism is inefficient: when marshaling
// it requires the allocation of intermediate byte slices, and when unmarshaling it causes the
// JSON object to be parsed twice. It is preferable to have our marshal/unmarshal methods write to
// and read from the EasyJSON Writer/Lexer directly.
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

func (c *Context) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	// We have to repeat part of the first-pass parsing logic here instead of just calling
	// c.ReadFromJSONReader, because doing the latter would cause the wrapped Lexer's state to be
	// modified by the first pass.
	copyOfLexer := *lexer
	wrappedReader0 := jreader.NewReaderFromEasyJSONLexer(&copyOfLexer)
	kind, hasKind, err := parseKindOnly(&wrappedReader0)
	if err != nil {
		lexer.AddError(err)
		return
	}
	wrappedReader1 := jreader.NewReaderFromEasyJSONLexer(lexer)
	switch {
	case !hasKind:
		err = unmarshalOldUserSchema(c, &wrappedReader1)
	case kind == MultiKind:
		err = unmarshalMultiKind(c, &wrappedReader1)
	default:
		err = unmarshalSingleKind(c, &wrappedReader1, "")
	}
	if err != nil {
		lexer.AddError(err)
	}
}
