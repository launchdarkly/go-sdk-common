// +build  launchdarkly_easyjson

// Code generated by easyjson for marshaling/unmarshaling. DO NOT EDIT.

package lduser

import (
	json "encoding/json"
	easyjson "github.com/mailru/easyjson"
	jlexer "github.com/mailru/easyjson/jlexer"
	jwriter "github.com/mailru/easyjson/jwriter"
)

// suppress unused package warning
var (
	_ *json.RawMessage
	_ *jlexer.Lexer
	_ *jwriter.Writer
	_ easyjson.Marshaler
)

func easyjson9ae05462DecodeGopkgInLaunchdarklyGoSdkCommonV2Lduser(in *jlexer.Lexer, out *userEquivalentStruct) {
	isTopLevel := in.IsStart()
	if in.IsNull() {
		if isTopLevel {
			in.Consumed()
		}
		in.Skip()
		return
	}
	in.Delim('{')
	for !in.IsDelim('}') {
		key := in.UnsafeFieldName(false)
		in.WantColon()
		if in.IsNull() {
			in.Skip()
			in.WantComma()
			continue
		}
		switch key {
		case "key":
			(out.Key).UnmarshalEasyJSON(in)
		case "secondary":
			(out.Secondary).UnmarshalEasyJSON(in)
		case "ip":
			(out.Ip).UnmarshalEasyJSON(in)
		case "country":
			(out.Country).UnmarshalEasyJSON(in)
		case "email":
			(out.Email).UnmarshalEasyJSON(in)
		case "firstName":
			(out.FirstName).UnmarshalEasyJSON(in)
		case "lastName":
			(out.LastName).UnmarshalEasyJSON(in)
		case "avatar":
			(out.Avatar).UnmarshalEasyJSON(in)
		case "name":
			(out.Name).UnmarshalEasyJSON(in)
		case "anonymous":
			(out.Anonymous).UnmarshalEasyJSON(in)
		case "custom":
			(out.Custom).UnmarshalEasyJSON(in)
		case "privateAttributeNames":
			if in.IsNull() {
				in.Skip()
				out.PrivateAttributeNames = nil
			} else {
				in.Delim('[')
				if out.PrivateAttributeNames == nil {
					if !in.IsDelim(']') {
						out.PrivateAttributeNames = make([]string, 0, 4)
					} else {
						out.PrivateAttributeNames = []string{}
					}
				} else {
					out.PrivateAttributeNames = (out.PrivateAttributeNames)[:0]
				}
				for !in.IsDelim(']') {
					var v1 string
					v1 = string(in.String())
					out.PrivateAttributeNames = append(out.PrivateAttributeNames, v1)
					in.WantComma()
				}
				in.Delim(']')
			}
		default:
			in.SkipRecursive()
		}
		in.WantComma()
	}
	in.Delim('}')
	if isTopLevel {
		in.Consumed()
	}
}
func easyjson9ae05462EncodeGopkgInLaunchdarklyGoSdkCommonV2Lduser(out *jwriter.Writer, in userEquivalentStruct) {
	out.RawByte('{')
	first := true
	_ = first
	{
		const prefix string = ",\"key\":"
		out.RawString(prefix[1:])
		(in.Key).MarshalEasyJSON(out)
	}
	if (in.Secondary).IsDefined() {
		const prefix string = ",\"secondary\":"
		out.RawString(prefix)
		(in.Secondary).MarshalEasyJSON(out)
	}
	if (in.Ip).IsDefined() {
		const prefix string = ",\"ip\":"
		out.RawString(prefix)
		(in.Ip).MarshalEasyJSON(out)
	}
	if (in.Country).IsDefined() {
		const prefix string = ",\"country\":"
		out.RawString(prefix)
		(in.Country).MarshalEasyJSON(out)
	}
	if (in.Email).IsDefined() {
		const prefix string = ",\"email\":"
		out.RawString(prefix)
		(in.Email).MarshalEasyJSON(out)
	}
	if (in.FirstName).IsDefined() {
		const prefix string = ",\"firstName\":"
		out.RawString(prefix)
		(in.FirstName).MarshalEasyJSON(out)
	}
	if (in.LastName).IsDefined() {
		const prefix string = ",\"lastName\":"
		out.RawString(prefix)
		(in.LastName).MarshalEasyJSON(out)
	}
	if (in.Avatar).IsDefined() {
		const prefix string = ",\"avatar\":"
		out.RawString(prefix)
		(in.Avatar).MarshalEasyJSON(out)
	}
	if (in.Name).IsDefined() {
		const prefix string = ",\"name\":"
		out.RawString(prefix)
		(in.Name).MarshalEasyJSON(out)
	}
	if (in.Anonymous).IsDefined() {
		const prefix string = ",\"anonymous\":"
		out.RawString(prefix)
		(in.Anonymous).MarshalEasyJSON(out)
	}
	if (in.Custom).IsDefined() {
		const prefix string = ",\"custom\":"
		out.RawString(prefix)
		(in.Custom).MarshalEasyJSON(out)
	}
	if len(in.PrivateAttributeNames) != 0 {
		const prefix string = ",\"privateAttributeNames\":"
		out.RawString(prefix)
		{
			out.RawByte('[')
			for v2, v3 := range in.PrivateAttributeNames {
				if v2 > 0 {
					out.RawByte(',')
				}
				out.String(string(v3))
			}
			out.RawByte(']')
		}
	}
	out.RawByte('}')
}

// MarshalJSON supports json.Marshaler interface
func (v userEquivalentStruct) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	easyjson9ae05462EncodeGopkgInLaunchdarklyGoSdkCommonV2Lduser(&w, v)
	return w.Buffer.BuildBytes(), w.Error
}

// MarshalEasyJSON supports easyjson.Marshaler interface
func (v userEquivalentStruct) MarshalEasyJSON(w *jwriter.Writer) {
	easyjson9ae05462EncodeGopkgInLaunchdarklyGoSdkCommonV2Lduser(w, v)
}

// UnmarshalJSON supports json.Unmarshaler interface
func (v *userEquivalentStruct) UnmarshalJSON(data []byte) error {
	r := jlexer.Lexer{Data: data}
	easyjson9ae05462DecodeGopkgInLaunchdarklyGoSdkCommonV2Lduser(&r, v)
	return r.Error()
}

// UnmarshalEasyJSON supports easyjson.Unmarshaler interface
func (v *userEquivalentStruct) UnmarshalEasyJSON(l *jlexer.Lexer) {
	easyjson9ae05462DecodeGopkgInLaunchdarklyGoSdkCommonV2Lduser(l, v)
}
