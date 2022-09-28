package ldattr

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/launchdarkly/go-sdk-common/v3/lderrors"

	"github.com/stretchr/testify/assert"
)

func TestRefInvalid(t *testing.T) {
	for _, p := range []struct {
		input         string
		expectedError error
	}{
		{"", lderrors.ErrAttributeEmpty{}},
		{"/", lderrors.ErrAttributeEmpty{}},
		{"//", lderrors.ErrAttributeExtraSlash{}},
		{"/a//b", lderrors.ErrAttributeExtraSlash{}},
		{"/a/b/", lderrors.ErrAttributeExtraSlash{}},
		{"/a~x", lderrors.ErrAttributeInvalidEscape{}},
		{"/a~", lderrors.ErrAttributeInvalidEscape{}},
		{"/a/b~x", lderrors.ErrAttributeInvalidEscape{}},
		{"/a/b~", lderrors.ErrAttributeInvalidEscape{}},
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
		assert.Equal(t, lderrors.ErrAttributeEmpty{}, a.Err())
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
			assert.Equal(t, s, a.Component(0))
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
		{"/name~1with~1slashes~0and~0tildes", "name/with/slashes~and~tildes"},
	} {
		t.Run(fmt.Sprintf("input string %q", params.input), func(t *testing.T) {
			a := NewRef(params.input)
			assert.True(t, a.IsDefined())
			assert.NoError(t, a.Err())
			assert.Equal(t, params.input, a.String())
			assert.Equal(t, 1, a.Depth())
			assert.Equal(t, params.path, a.Component(0))
		})
	}
}

func TestNewLiteralRef(t *testing.T) {
	a0 := NewLiteralRef("name")
	assert.Equal(t, NewRef("name"), a0)

	a1 := NewLiteralRef("a/b")
	assert.Equal(t, NewRef("a/b"), a1)

	a2 := NewLiteralRef("/a/b~c")
	assert.Equal(t, NewRef("/~1a~1b~0c"), a2)
	assert.Equal(t, 1, a2.Depth())

	a3 := NewLiteralRef("/")
	assert.Equal(t, NewRef("/~1"), a3)
	assert.Equal(t, 1, a3.Depth())

	a4 := NewLiteralRef("")
	assert.Equal(t, lderrors.ErrAttributeEmpty{}, a4.Err())
}

func TestRefComponents(t *testing.T) {
	for _, params := range []struct {
		input        string
		depth        int
		index        int
		expectedName string
	}{
		{"", 0, 0, ""},
		{"key", 1, 0, "key"},
		{"/key", 1, 0, "key"},
		{"/a/b", 2, 0, "a"},
		{"/a/b", 2, 1, "b"},
		{"/a~1b/c", 2, 0, "a/b"},
		{"/a~0b/c", 2, 0, "a~b"},
		{"/a/10/20/30x", 4, 1, "10"},
		{"/a/10/20/30x", 4, 2, "20"},
		{"/a/10/20/30x", 4, 3, "30x"},

		// invalid arguments don't cause an error, they just return empty values
		{"", 0, -1, ""},
		{"key", 1, -1, ""},
		{"key", 1, 1, ""},
		{"/key", 1, -1, ""},
		{"/key", 1, 1, ""},
		{"/a/b", 2, -1, ""},
		{"/a/b", 2, 2, ""},
	} {
		t.Run(fmt.Sprintf("input string %q, index %d", params.input, params.index), func(t *testing.T) {
			a := NewRef(params.input)
			assert.Equal(t, params.depth, a.Depth())
			name := a.Component(params.index)
			assert.Equal(t, params.expectedName, name)
		})
	}
}

func TestRefEqual(t *testing.T) {
	refs := []Ref{
		{},
		NewRef(""),
		NewRef("a"),
		NewRef("b"),
		NewRef("/a/b"),
		NewRef("/a/c"),
		NewRef("///"),
	}
	for i, a := range refs {
		sameValue := a
		assert.True(t, sameValue.Equal(a))
		for j, differentValue := range refs {
			if j == i {
				continue
			}
			assert.False(t, differentValue.Equal(a))
		}
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
