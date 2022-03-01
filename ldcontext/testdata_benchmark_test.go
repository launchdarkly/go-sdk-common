package ldcontext

import (
	"encoding/json"

	"gopkg.in/launchdarkly/go-sdk-common.v2/internal/sharedtest"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

var (
	benchmarkContext    Context
	benchmarkContextPtr *Context
	benchmarkJSONResult []byte
	benchmarkValue      ldvalue.Value
	benchmarkErr        error
)

type keyAndLDValue struct {
	key   string
	value ldvalue.Value
}

type benchmarkMarshalTestParams struct {
	name    string
	context Context
}

func makeBenchmarkMarshalTestParams() []benchmarkMarshalTestParams {
	return []benchmarkMarshalTestParams{
		{"context with key only", makeBenchmarkContextWithKeyOnly()},
		{"context with few attrs", makeBenchmarkContextWithFewAttributes()},
		{"context with all attrs", makeBenchmarkContextWithAllAttributes()},
	}
}

func makeBenchmarkUnmarshalTestParams() []sharedtest.UnmarshalingTestParams {
	ret := []sharedtest.UnmarshalingTestParams{
		{"context with key only", makeBenchmarkContextWithKeyOnlyJSON()},
		{"context with few attrs", makeBenchmarkContextWithFewAttributesJSON()},
		{"context with all attrs", makeBenchmarkContextWithAllAttributesJSON()},
	}
	return append(ret, sharedtest.MakeOldUserUnmarshalingTestParams()...)
}

func makeBenchmarkContextWithKeyOnly() Context {
	return New("user-key")
}

func makeBenchmarkContextWithKeyOnlyJSON() []byte {
	data, _ := json.Marshal(makeBenchmarkContextWithKeyOnly())
	return data
}

func makeBenchmarkContextWithFewAttributes() Context {
	return NewBuilder("user-key").
		SetString("name", "Name").
		SetString("email", "test@example.com").
		SetString("attr", "value").
		Build()
}

func makeBenchmarkContextWithFewAttributesJSON() []byte {
	data, _ := json.Marshal(makeBenchmarkContextWithFewAttributes())
	return data
}

func makeBenchmarkContextWithAllAttributes() Context {
	return NewBuilder("user-key").
		Secondary("secondary-value").
		SetString("name", "Name").
		SetString("ip", "ip-value").
		SetString("country", "us").
		SetString("email", "test@example.com").
		SetString("firstName", "First").
		SetString("lastName", "Last").
		SetString("avatar", "avatar-value").
		Transient(true).
		SetString("attr", "value").
		Build()
}

func makeBenchmarkContextWithAllAttributesJSON() []byte {
	data, _ := json.Marshal(makeBenchmarkContextWithAllAttributes())
	return data
}
