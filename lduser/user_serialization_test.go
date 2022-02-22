package lduser

import (
	"encoding/json"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func parseJSONAsMap(t *testing.T, bytes []byte) map[string]interface{} {
	var props map[string]interface{}
	err := json.Unmarshal(bytes, &props)
	require.NoError(t, err)
	return props
}

func TestUserStringIsJSONRepresentation(t *testing.T) {
	user := newUserBuilderWithAllPropertiesSet("some-key").Build()
	bytes, err := json.Marshal(user)
	require.NoError(t, err)
	assert.Equal(t, parseJSONAsMap(t, bytes), parseJSONAsMap(t, []byte(user.String())))
}

func TestJSONMarshal(t *testing.T) {
	doUserMarshalingTests(t, json.Marshal)
}

func doUserMarshalingTests(t *testing.T, marshalFn func(interface{}) ([]byte, error)) {
	getJSONAsMap := func(t *testing.T, thing interface{}) map[string]interface{} {
		bytes, err := marshalFn(thing)
		require.NoError(t, err)
		return parseJSONAsMap(t, bytes)
	}

	t.Run("compact representation with only key", func(t *testing.T) {
		user := NewUser("some-key")
		bytes, err := marshalFn(user)
		require.NoError(t, err)
		assert.Equal(t, `{"key":"some-key"}`, string(bytes))
	})

	t.Run("string attributes", func(t *testing.T) {
		for a, setter := range optionalStringSetters {
			t.Run(string(a), func(t *testing.T) {
				builder := NewUserBuilder("some-key")
				setter(builder, "value")
				user := builder.Build()

				props := getJSONAsMap(t, user)
				assert.Equal(t, "some-key", props["key"])
				assert.Equal(t, "value", props[string(a)])

				for a1, _ := range optionalStringSetters {
					if a1 != a {
						unwantedValue, found := props[string(a1)]
						assert.False(t, found)
						assert.Nil(t, unwantedValue)
					}
				}
			})
		}
	})

	t.Run("anonymous attribute", func(t *testing.T) {
		props1 := getJSONAsMap(t, NewUserBuilder("some-key").Build())
		assert.Nil(t, props1["anonymous"])

		props2 := getJSONAsMap(t, NewUserBuilder("some-key").Anonymous(true).Build())
		assert.Equal(t, true, props2["anonymous"])

		props3 := getJSONAsMap(t, NewUserBuilder("some-key").Anonymous(false).Build())
		assert.Equal(t, false, props3["anonymous"])
	})

	t.Run("custom attribute", func(t *testing.T) {
		props1 := getJSONAsMap(t, NewUserBuilder("some-key").Build())
		assert.Nil(t, props1["custom"])

		props2 := getJSONAsMap(t, NewUserBuilder("some-key").Custom("thing", ldvalue.String("value")).Build())
		assert.Equal(t, map[string]interface{}{"thing": "value"}, props2["custom"])
	})

	t.Run("private attributes", func(t *testing.T) {
		props1 := getJSONAsMap(t, NewUserBuilder("some-key").Name("value").Build())
		assert.Equal(t, "value", props1["name"])
		assert.Nil(t, props1["privateAttributeNames"])

		props2 := getJSONAsMap(t, NewUserBuilder("some-key").Name("value").AsPrivateAttribute().Build())
		assert.Equal(t, "value", props2["name"])
		assert.Equal(t, []interface{}{"name"}, props2["privateAttributeNames"])
	})
}

func TestJSONUnmarshal(t *testing.T) {
	doUserUnmarshalingTests(t, json.Unmarshal)
}

func doUserUnmarshalingTests(t *testing.T, unmarshalFn func([]byte, interface{}) error) {
	unmarshalUser := func(t *testing.T, jsonProps map[string]interface{}) User {
		bytes, err := json.Marshal(jsonProps)
		require.NoError(t, err)
		var user User
		err = unmarshalFn(bytes, &user)
		require.NoError(t, err)
		return user
	}

	t.Run("string attributes", func(t *testing.T) {
		for a, _ := range optionalStringSetters {
			t.Run(string(a), func(t *testing.T) {
				props := map[string]interface{}{"key": "some-key"}
				props[string(a)] = "value"
				user := unmarshalUser(t, props)

				assert.Equal(t, "some-key", user.GetKey())
				assert.Equal(t, ldvalue.String("value"), user.GetAttribute(a))

				for a1, _ := range optionalStringSetters {
					if a1 != a {
						assert.Equal(t, ldvalue.Null(), user.GetAttribute(a1))
					}
				}
			})
		}
	})

	t.Run("anonymous attribute", func(t *testing.T) {
		user1 := unmarshalUser(t, map[string]interface{}{"key": "some-key"})
		assert.Equal(t, ldvalue.Null(), user1.GetAttribute(AnonymousAttribute))

		user2 := unmarshalUser(t, map[string]interface{}{"key": "some-key", "anonymous": true})
		assert.Equal(t, ldvalue.Bool(true), user2.GetAttribute(AnonymousAttribute))

		user3 := unmarshalUser(t, map[string]interface{}{"key": "some-key", "anonymous": false})
		assert.Equal(t, ldvalue.Bool(false), user3.GetAttribute(AnonymousAttribute))
	})

	t.Run("custom attribute", func(t *testing.T) {
		user1 := unmarshalUser(t, map[string]interface{}{"key": "some-key"})
		assert.Equal(t, ldvalue.Null(), user1.GetAttribute(UserAttribute("thing")))

		user2 := unmarshalUser(t, map[string]interface{}{
			"key":    "some-key",
			"custom": map[string]interface{}{"thing": "value"},
		})
		assert.Equal(t, ldvalue.String("value"), user2.GetAttribute(UserAttribute("thing")))
	})

	t.Run("private attributes", func(t *testing.T) {
		user1 := unmarshalUser(t, map[string]interface{}{"key": "some-key", "name": "value"})
		assert.False(t, user1.IsPrivateAttribute("name"))

		user2 := unmarshalUser(t, map[string]interface{}{
			"key":                   "some-key",
			"name":                  "value",
			"privateAttributeNames": []string{"name"},
		})
		assert.True(t, user2.IsPrivateAttribute("name"))
	})

	t.Run("nulls are allowed for arrays and objects", func(t *testing.T) {
		user := unmarshalUser(t, map[string]interface{}{"key": "some-key", "custom": nil, "privateAttributeNames": nil})
		assert.Equal(t, ldvalue.ValueMap{}, user.custom)
		assert.Nil(t, user.privateAttributes)
	})

	t.Run("data with wrong field type", func(t *testing.T) {
		var user User
		err := unmarshalFn([]byte(`{"key":[1,2,3]}`), &user)
		assert.Error(t, err)
	})

	t.Run("missing key is invalid", func(t *testing.T) {
		var user User
		err := unmarshalFn([]byte(`{"name":"n"}`), &user)
		assert.Equal(t, ErrMissingKey(), err)
	})

	t.Run("null key is invalid", func(t *testing.T) {
		var user User
		err := unmarshalFn([]byte(`{"key":null,"name":"n"}`), &user)
		assert.Equal(t, ErrMissingKey(), err)
	})

	t.Run("empty string key is valid", func(t *testing.T) {
		var user User
		err := unmarshalFn([]byte(`{"key":"","name":"n"}`), &user)
		assert.NoError(t, err)
		assert.Equal(t, "", user.GetKey())
		assert.Equal(t, "n", user.GetName().StringValue())
	})

	t.Run("unknown attributes are ignored", func(t *testing.T) {
		user := unmarshalUser(t, map[string]interface{}{
			"key":  "some-key",
			"a":    3,
			"b":    nil,
			"c":    []interface{}{},
			"name": "value",
		})
		assert.Equal(t, "some-key", user.GetKey())
		assert.Equal(t, "value", user.GetName().StringValue())
	})

	t.Run("malformed data", func(t *testing.T) {
		var user User
		for _, data := range [][]byte{
			nil,
			[]byte{},
			[]byte("null"),
			[]byte("true"),
			[]byte("[]"),
			[]byte("{"),
			[]byte("{true"),
			[]byte(`{"key":true}`),
			[]byte(`{"name":true}`),
			[]byte(`{"privateAttributeNames":{}`),
			[]byte(`{"privateAttributeNames":[true]`),
		} {
			t.Run(string(data), func(t *testing.T) {
				err := unmarshalFn(data, &user)
				assert.Error(t, err)
				assert.NotEqual(t, ErrMissingKey(), err)
			})
		}
	})
}

func TestMissingKeyHasErrorMessage(t *testing.T) {
	assert.NotEqual(t, "", ErrMissingKey().Error())
}
