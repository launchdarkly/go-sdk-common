package ldcontext

import (
	"fmt"
	"testing"

	"gopkg.in/launchdarkly/go-sdk-common.v3/ldvalue"

	"github.com/stretchr/testify/assert"
)

func TestAttrRefInvalid(t *testing.T) {
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
			a := NewAttrRef(p.input)
			assert.True(t, a.IsDefined())
			assert.Equal(t, p.expectedError, a.Err())
			assert.Equal(t, p.input, a.String())
			assert.Equal(t, 0, a.Depth())
		})
	}

	t.Run("uninitialized", func(t *testing.T) {
		var a AttrRef
		assert.False(t, a.IsDefined())
		assert.Equal(t, errAttributeEmpty, a.Err())
		assert.Equal(t, "", a.String())
		assert.Equal(t, 0, a.Depth())
	})
}

func TestAttrRefWithNoLeadingSlash(t *testing.T) {
	for _, s := range []string{
		"key",
		"kind",
		"name",
		"name/with/slashes",
		"name~0~1with-what-looks-like-escape-sequences",
	} {
		t.Run(fmt.Sprintf("input string %q", s), func(t *testing.T) {
			a := NewAttrRef(s)
			assert.True(t, a.IsDefined())
			assert.NoError(t, a.Err())
			assert.Equal(t, s, a.String())
			assert.Equal(t, 1, a.Depth())
			name, _ := a.Component(0)
			assert.Equal(t, s, name)
		})
	}
}

func TestAttrRefSimpleWithLeadingSlash(t *testing.T) {
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
			a := NewAttrRef(params.input)
			assert.True(t, a.IsDefined())
			assert.NoError(t, a.Err())
			assert.Equal(t, params.input, a.String())
			assert.Equal(t, 1, a.Depth())
			name, _ := a.Component(0)
			assert.Equal(t, params.path, name)
		})
	}
}

func TestAttrRefForName(t *testing.T) {
	a0 := NewAttrRefForName("name")
	assert.Equal(t, NewAttrRef("name"), a0)

	a1 := NewAttrRefForName("a/b")
	assert.Equal(t, NewAttrRef("a/b"), a1)

	a2 := NewAttrRefForName("/a/b~c")
	assert.Equal(t, NewAttrRef("/~1a~1b~0c"), a2)
	assert.Equal(t, 1, a2.Depth())
}

func TestAttrRefComponents(t *testing.T) {
	undefined := -99
	for _, params := range []struct {
		input         string
		index         int
		expectedName  string
		expectedIndex int
	}{
		{"", 0, "", undefined},
		{"key", 0, "key", undefined},
		{"/key", 0, "key", undefined},
		{"/a/b", 0, "a", undefined},
		{"/a/b", 1, "b", undefined},
		{"/a~1b/c", 0, "a/b", undefined},
		{"/a/10/20/30x", 1, "10", 10},
		{"/a/10/20/30x", 2, "20", 20},
		{"/a/10/20/30x", 3, "30x", undefined},

		// invalid arguments don't cause an error, they just return empty values
		{"", -1, "", undefined},
		{"key", -1, "", undefined},
		{"key", 1, "", undefined},
		{"/key", -1, "", undefined},
		{"/key", 1, "", undefined},
		{"/a/b", -1, "", undefined},
		{"/a/b", 2, "", undefined},
	} {
		t.Run(fmt.Sprintf("input string %q, index %d", params.input, params.index), func(t *testing.T) {
			a := NewAttrRef(params.input)
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
