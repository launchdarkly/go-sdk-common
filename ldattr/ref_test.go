package ldattr

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/ldvalue"

	"github.com/stretchr/testify/assert"
)

func TestRefInvalid(t *testing.T) {
	for _, p := range []struct {
		input         string
		expectedError error
	}{
		{"", errAttributeEmpty},
		{"/", errAttributeEmpty},
		{"//", errAttributeExtraSlash},
		{"/a//b", errAttributeExtraSlash},
		{"/a/b/", errAttributeExtraSlash},
	} {
		t.Run(fmt.Sprintf("input string %q", p.input), func(t *testing.T) {
			a := NewRef(p.input)
			assert.True(t, a.IsDefined())
			assert.Equal(t, p.expectedError, a.Err())
			assert.Equal(t, p.input, a.String())
			assert.Equal(t, 0, a.Depth())
		})
	}

	t.Run("uninitialized", func(t *testing.T) {
		var a Ref
		assert.False(t, a.IsDefined())
		assert.Equal(t, errAttributeEmpty, a.Err())
		assert.Equal(t, "", a.String())
		assert.Equal(t, 0, a.Depth())
	})
}

func TestRefWithNoLeadingSlash(t *testing.T) {
	for _, s := range []string{
		"key",
		"kind",
		"name",
		"name/with/slashes",
		"name~0~1with-what-looks-like-escape-sequences",
	} {
		t.Run(fmt.Sprintf("input string %q", s), func(t *testing.T) {
			a := NewRef(s)
			assert.True(t, a.IsDefined())
			assert.NoError(t, a.Err())
			assert.Equal(t, s, a.String())
			assert.Equal(t, 1, a.Depth())
			name, _ := a.Component(0)
			assert.Equal(t, s, name)
		})
	}
}

func TestRefSimpleWithLeadingSlash(t *testing.T) {
	for _, params := range []struct {
		input string
		path  string
	}{
		{"/key", "key"},
		{"/kind", "kind"},
		{"/name", "name"},
		{"/custom", "custom"},
		{"/0", "0"},
		{"/name~1with~1slashes~0and~0tildes~2~x~~", "name/with/slashes~and~tildes~2~x~~"},
	} {
		t.Run(fmt.Sprintf("input string %q", params.input), func(t *testing.T) {
			a := NewRef(params.input)
			assert.True(t, a.IsDefined())
			assert.NoError(t, a.Err())
			assert.Equal(t, params.input, a.String())
			assert.Equal(t, 1, a.Depth())
			name, _ := a.Component(0)
			assert.Equal(t, params.path, name)
		})
	}
}

func TestNewNameRef(t *testing.T) {
	a0 := NewNameRef("name")
	assert.Equal(t, NewRef("name"), a0)

	a1 := NewNameRef("a/b")
	assert.Equal(t, NewRef("a/b"), a1)

	a2 := NewNameRef("/a/b~c")
	assert.Equal(t, NewRef("/~1a~1b~0c"), a2)
	assert.Equal(t, 1, a2.Depth())

	a3 := NewNameRef("/")
	assert.Equal(t, NewRef("/~1"), a3)
	assert.Equal(t, 1, a3.Depth())

	a4 := NewNameRef("")
	assert.Equal(t, errAttributeEmpty, a4.Err())
}

func TestRefComponents(t *testing.T) {
	undefined := -99
	for _, params := range []struct {
		input         string
		depth         int
		index         int
		expectedName  string
		expectedIndex int
	}{
		{"", 0, 0, "", undefined},
		{"key", 1, 0, "key", undefined},
		{"/key", 1, 0, "key", undefined},
		{"/a/b", 2, 0, "a", undefined},
		{"/a/b", 2, 1, "b", undefined},
		{"/a~1b/c", 2, 0, "a/b", undefined},
		{"/a/10/20/30x", 4, 1, "10", 10},
		{"/a/10/20/30x", 4, 2, "20", 20},
		{"/a/10/20/30x", 4, 3, "30x", undefined},

		// invalid arguments don't cause an error, they just return empty values
		{"", 0, -1, "", undefined},
		{"key", 1, -1, "", undefined},
		{"key", 1, 1, "", undefined},
		{"/key", 1, -1, "", undefined},
		{"/key", 1, 1, "", undefined},
		{"/a/b", 2, -1, "", undefined},
		{"/a/b", 2, 2, "", undefined},
	} {
		t.Run(fmt.Sprintf("input string %q, index %d", params.input, params.index), func(t *testing.T) {
			a := NewRef(params.input)
			assert.Equal(t, params.depth, a.Depth())
			name, index := a.Component(params.index)
			assert.Equal(t, params.expectedName, name)
			if params.expectedIndex == undefined {
				assert.Equal(t, ldvalue.OptionalInt{}, index)
			} else {
				assert.Equal(t, ldvalue.NewOptionalInt(params.expectedIndex), index)
			}
		})
	}
}

func TestRefMarshalJSON(t *testing.T) {
	for _, p := range []struct {
		ref  Ref
		json string
	}{
		{Ref{}, `null`},
		{NewRef("a"), `"a"`},
		{NewRef("/a/b"), `"/a/b"`},
		{NewRef("////invalid"), `"////invalid"`},
	} {
		t.Run(p.json, func(t *testing.T) {
			bytes, err := json.Marshal(p.ref)
			assert.NoError(t, err)
			assert.Equal(t, p.json, string(bytes))
		})
	}
}

func TestRefUnmarshalJSON(t *testing.T) {
	for _, p := range []struct {
		json    string
		ref     Ref
		success bool
	}{
		{`null`, Ref{}, true},
		{`"a"`, NewRef("a"), true},
		{`"/a/b"`, NewRef("/a/b"), true},
		{`"////invalid"`, NewRef("////invalid"), true},
		{`true`, Ref{}, false},
		{`2`, Ref{}, false},
		{`[]`, Ref{}, false},
		{`{}`, Ref{}, false},
		{`.`, Ref{}, false},
		{``, Ref{}, false},
	} {
		t.Run(p.json, func(t *testing.T) {
			var ref Ref
			err := json.Unmarshal([]byte(p.json), &ref)
			assert.Equal(t, p.ref, ref)
			if p.success {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
			}
		})
	}
}
