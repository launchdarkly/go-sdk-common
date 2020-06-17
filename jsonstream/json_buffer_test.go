package jsonstream

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type valueTest struct {
	name     string
	encoding string
	action   func(j *JSONBuffer)
}

var allValueTests = []valueTest{
	{"null", "null", func(j *JSONBuffer) { j.WriteNull() }},
	{"bool true", "true", func(j *JSONBuffer) { j.WriteBool(true) }},
	{"bool true", "false", func(j *JSONBuffer) { j.WriteBool(false) }},
	{"int", "3", func(j *JSONBuffer) { j.WriteInt(3) }},
	{"zero int", "0", func(j *JSONBuffer) { j.WriteInt(0) }},
	{"negative int", "-3", func(j *JSONBuffer) { j.WriteInt(-3) }},
	{"large int", "123456", func(j *JSONBuffer) { j.WriteInt(123456) }},
	{"negative large int", "-123456", func(j *JSONBuffer) { j.WriteInt(-123456) }},
	{"uint64", "3", func(j *JSONBuffer) { j.WriteUint64(3) }},
	{"zero uint64", "0", func(j *JSONBuffer) { j.WriteUint64(0) }},
	{"large uint64", "12345678901234567890", func(j *JSONBuffer) { j.WriteUint64(12345678901234567890) }},
	{"float", "3.5", func(j *JSONBuffer) { j.WriteFloat64(3.5) }},
	{"zero float", "0", func(j *JSONBuffer) { j.WriteFloat64(0) }},
	{"integer as float", "3", func(j *JSONBuffer) { j.WriteFloat64(float64(3)) }},
	{"negative float", "-3.5", func(j *JSONBuffer) { j.WriteFloat64(-3.5) }},
	{"string", `"abc"`, func(j *JSONBuffer) { j.WriteString("abc") }},
	{"empty string", `""`, func(j *JSONBuffer) { j.WriteString("") }},
	{"array", `[]`, func(j *JSONBuffer) {
		j.BeginArray()
		j.EndArray()
	}},
	{"object", `{}`, func(j *JSONBuffer) {
		j.BeginObject()
		j.EndObject()
	}},
	{"raw", `{"a":1}`, func(j *JSONBuffer) { j.WriteRaw(json.RawMessage(`{"a":1}`)) }},
}

type escapeTest struct {
	ch rune
	s  string
}

var allEscapeTests = []escapeTest{
	{'\b', `\b`}, {'\t', `\t`}, {'\n', `\n`}, {'\f', `\f`}, {'\r', `\r`}, {'"', `\"`}, {'\\', `\\`},
}

func testEncoding(t *testing.T, name string, expected string, action func(*JSONBuffer)) {
	t.Run(name, func(t *testing.T) {
		j := NewJSONBuffer()
		action(j)
		bytes, err := j.Get()
		if err == nil {
			assert.Equal(t, expected, string(bytes))
		} else {
			assert.NoError(t, err, "data written to buffer so far: %s", j.GetPartial())
		}
	})
}

func TestWriteSimpleValues(t *testing.T) {
	for _, test := range allValueTests {
		testEncoding(t, test.name, test.encoding, test.action)
	}
}

func TestWriteStringFormatting(t *testing.T) {
	for _, test := range allEscapeTests {
		expected1 := fmt.Sprintf("%sabcd", test.s)
		testEncoding(t, expected1, `"`+expected1+`"`, func(j *JSONBuffer) {
			j.WriteString(fmt.Sprintf("%sabcd", string(test.ch)))
		})

		expected2 := fmt.Sprintf("ab%scd", test.s)
		testEncoding(t, expected2, `"`+expected2+`"`, func(j *JSONBuffer) {
			j.WriteString(fmt.Sprintf("ab%scd", string(test.ch)))
		})

		expected3 := fmt.Sprintf("abcd%s", test.s)
		testEncoding(t, expected3, `"`+expected3+`"`, func(j *JSONBuffer) {
			j.WriteString(fmt.Sprintf("abcd%s", string(test.ch)))
		})

		expected4 := fmt.Sprintf("ab%scd%sef", test.s, test.s)
		testEncoding(t, expected4, `"`+expected4+`"`, func(j *JSONBuffer) {
			j.WriteString(fmt.Sprintf("ab%scd%sef", string(test.ch), string(test.ch)))
		})

		expected5 := fmt.Sprintf("ab%s%scd", test.s, test.s)
		testEncoding(t, expected5, `"`+expected5+`"`, func(j *JSONBuffer) {
			j.WriteString(fmt.Sprintf("ab%s%scd", string(test.ch), string(test.ch)))
		})
	}

	// Multi-byte characters do not get special handling - JSON allows them to be escaped as hex sequences,
	// but does not require it.
	emojiStr := "🦜🦄😂🦹🏻‍♀️🦹‍♂️🧶😻 yes"
	testEncoding(t, "multi-byte characters", `"`+emojiStr+`"`, func(j *JSONBuffer) {
		j.WriteString(emojiStr)
	})
}

func TestWriteArray(t *testing.T) {
	testEncoding(t, "empty", "[]", func(j *JSONBuffer) {
		j.BeginArray()
		j.EndArray()
	})

	testEncoding(t, "single", "[1]", func(j *JSONBuffer) {
		j.BeginArray()
		j.WriteInt(1)
		j.EndArray()
	})

	testEncoding(t, "multiple", "[1,2]", func(j *JSONBuffer) {
		j.BeginArray()
		j.WriteInt(1)
		j.WriteInt(2)
		j.EndArray()
	})

	testEncoding(t, "nested", "[1,[2,3],4]", func(j *JSONBuffer) {
		j.BeginArray()
		j.WriteInt(1)
		j.BeginArray()
		j.WriteInt(2)
		j.WriteInt(3)
		j.EndArray()
		j.WriteInt(4)
		j.EndArray()
	})
}

func TestWriteObject(t *testing.T) {
	testEncoding(t, "empty", "{}", func(j *JSONBuffer) {
		j.BeginObject()
		j.EndObject()
	})

	testEncoding(t, "single", `{"a":1}`, func(j *JSONBuffer) {
		j.BeginObject()
		j.WriteName("a")
		j.WriteInt(1)
		j.EndObject()
	})

	testEncoding(t, "multiple", `{"a":1,"b":2}`, func(j *JSONBuffer) {
		j.BeginObject()
		j.WriteName("a")
		j.WriteInt(1)
		j.WriteName("b")
		j.WriteInt(2)
		j.EndObject()
	})

	testEncoding(t, "nested", `{"a":{"b":1,"c":2},"d":3}`, func(j *JSONBuffer) {
		j.BeginObject()
		j.WriteName("a")
		j.BeginObject()
		j.WriteName("b")
		j.WriteInt(1)
		j.WriteName("c")
		j.WriteInt(2)
		j.EndObject()
		j.WriteName("d")
		j.WriteInt(3)
		j.EndObject()
	})
}

func TestWriteMultipleValues(t *testing.T) {
	testEncoding(t, "default separator", "1\n2\n3", func(j *JSONBuffer) {
		j.WriteInt(1)
		j.WriteInt(2)
		j.WriteInt(3)
	})

	testEncoding(t, "nil is same as default separator", "1\n2\n3", func(j *JSONBuffer) {
		j.SetSeparator(nil)
		j.WriteInt(1)
		j.WriteInt(2)
		j.WriteInt(3)
	})

	testEncoding(t, "custom separator", "1! 2! 3", func(j *JSONBuffer) {
		j.SetSeparator([]byte("! "))
		j.WriteInt(1)
		j.WriteInt(2)
		j.WriteInt(3)
	})
}

func TestWriteDeeplyNestedStructures(t *testing.T) {
	expected := ""
	var j JSONBuffer
	for i := 0; i < 100; i++ {
		j.BeginObject()
		j.WriteName("x")
		j.BeginArray()
		expected += `{"x":[`
	}
	for i := 0; i < 100; i++ {
		j.EndArray()
		j.EndObject()
		expected += "]}"
	}
	bytes, err := j.Get()
	assert.NoError(t, err)
	assert.Equal(t, expected, string(bytes))
}

func TestUsageErrors(t *testing.T) {
	shouldFail := func(name string, expectedPartialOutput string, action func(j *JSONBuffer)) {
		t.Run(name, func(t *testing.T) {
			j := NewJSONBuffer()
			action(j)
			bytes, err := j.Get()
			assert.Error(t, err)
			assert.Nil(t, bytes, "data written to buffer: %s", bytes)

			assert.Equal(t, err, j.GetError())
			assert.Equal(t, expectedPartialOutput, string(j.GetPartial()))
		})
	}

	shouldFail("BeginArray not closed", `[`, func(j *JSONBuffer) { j.BeginArray() })
	shouldFail("BeginObject not closed", `{`, func(j *JSONBuffer) { j.BeginObject() })

	shouldFail("EndArray before anything", ``, func(j *JSONBuffer) { j.EndArray() })
	shouldFail("EndArray in object before name", `{`, func(j *JSONBuffer) {
		j.BeginObject()
		j.EndArray()
	})
	shouldFail("EndArray in object after name", `{"name":`, func(j *JSONBuffer) {
		j.BeginObject()
		j.WriteName("name")
		j.EndArray()
	})
	shouldFail("EndArray in object after value", `{"name":1`, func(j *JSONBuffer) {
		j.BeginObject()
		j.WriteName("name")
		j.WriteInt(1)
		j.EndArray()
	})

	shouldFail("EndObject before anything", ``, func(j *JSONBuffer) { j.EndObject() })
	shouldFail("EndObject when value is expected", `{"name":`, func(j *JSONBuffer) {
		j.BeginObject()
		j.WriteName("name")
		j.EndObject()
	})
	shouldFail("EndObject in array before value", `[`, func(j *JSONBuffer) {
		j.BeginArray()
		j.EndObject()
	})
	shouldFail("EndObject in array after value", `[1`, func(j *JSONBuffer) {
		j.BeginArray()
		j.WriteInt(1)
		j.EndObject()
	})

	shouldFail("WriteName when value is expected", `[`, func(j *JSONBuffer) {
		j.BeginArray()
		j.WriteName("name")
	})
	shouldFail("WriteName after an error has already occurred", `{`, func(j *JSONBuffer) {
		j.BeginObject()
		j.EndArray()
		j.WriteName("name")
	})

	for _, test := range allValueTests {
		shouldFail(test.name+" when first property name is expected", `{`, func(j *JSONBuffer) {
			j.BeginObject()
			test.action(j)
			j.EndObject()
		})

		shouldFail(test.name+" when subsequent property name is expected", `{"x":true`, func(j *JSONBuffer) {
			j.BeginObject()
			j.WriteName("x")
			j.WriteBool(true)
			test.action(j)
			j.EndObject()
		})

		shouldFail(test.name+" after an error has already occurred", `1`, func(j *JSONBuffer) {
			j.WriteInt(1)
			j.EndArray() // puts buffer in a failed state
			test.action(j)
		})
	}
}

func TestGrowDoesNotAffectOutput(t *testing.T) {
	j := NewJSONBuffer()
	j.WriteInt(1)
	j.Grow(100)
	j.WriteInt(2)
	bytes, err := j.Get()
	assert.NoError(t, err)
	assert.Equal(t, "1\n2", string(bytes))
}

func TestJSONBufferCanBeLocallyAllocated(t *testing.T) {
	var j JSONBuffer
	j.WriteBool(true)
	bytes, err := j.Get()
	assert.NoError(t, err)
	assert.Equal(t, "true", string(bytes))
}
