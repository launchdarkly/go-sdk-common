package lduser

import (
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"
)

var (
	benchmarkUserBuilder                        = NewUserBuilder("key")
	benchmarkUserBuilderWithAllScalarAttributes = NewUserBuilder("key").
							Secondary("s").
							IP("i").
							Country("c").
							Email("e").
							FirstName("f").
							LastName("l").
							Avatar("a").
							Name("n").
							Anonymous(true)

	benchmarkUserResult  User
	benchmarkUserPointer *User

	benchmarkValue ldvalue.Value

	benchmarkJSONResult []byte

	benchmarkErr error
)

type benchmarkMarshalTestParams struct {
	name string
	user User
}

func makeBenchmarkMarshalTestParams() []benchmarkMarshalTestParams {
	return []benchmarkMarshalTestParams{
		{"user with key only", makeBenchmarkUserWithKeyOnly()},
		{"user with few attrs", makeBenchmarkUserWithFewAttributes()},
		{"user with all attrs", makeBenchmarkUserWithAllAttributes()},
	}
}

func makeBenchmarkUserWithKeyOnly() User {
	return NewUser("user-key")
}

func makeBenchmarkUserWithFewAttributes() User {
	return NewUserBuilder("user-key").
		Name("Name").
		Email("test@example.com").
		SetAttribute(UserAttribute("attr"), ldvalue.String("value")).
		Build()
}

func makeBenchmarkUserWithAllAttributes() User {
	return NewUserBuilder("user-key").
		Secondary("secondary-value").
		Name("Name").
		IP("ip-value").
		Country("us").
		Email("test@example.com").
		FirstName("First").
		LastName("Last").
		Avatar("avatar-value").
		Anonymous(true).
		SetAttribute(UserAttribute("attr"), ldvalue.String("value")).
		Build()
}
