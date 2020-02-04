package ldvalue

import (
	"encoding/json"
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnsafeConvertPrimitiveTypesFromArbitraryValue(t *testing.T) {
	assert.Equal(t, Null(), UnsafeUseArbitraryValue(nil))
	assert.Equal(t, Bool(false), UnsafeUseArbitraryValue(false))
	assert.Equal(t, Bool(true), UnsafeUseArbitraryValue(true))
	assert.Equal(t, Int(1), UnsafeUseArbitraryValue(1))
	assert.Equal(t, Float64(2.5), UnsafeUseArbitraryValue(float64(2.5)))
	assert.Equal(t, String("x"), UnsafeUseArbitraryValue("x"))
}

func TestUnsafeConvertPrimitiveTypesToArbitraryValue(t *testing.T) {
	assert.Nil(t, Null().UnsafeArbitraryValue())
	assert.Equal(t, true, Bool(true).UnsafeArbitraryValue())
	assert.Equal(t, false, Bool(false).UnsafeArbitraryValue())
	assert.Equal(t, float64(2), Int(2).UnsafeArbitraryValue())
	assert.Equal(t, "x", String("x").UnsafeArbitraryValue())
}

func TestUnsafeConvertComplexTypesFromArbitraryValue(t *testing.T) {
	t.Run("[]interface{}", func(t *testing.T) {
		originalSlice := []interface{}{2, []interface{}{"x"}}
		value := UnsafeUseArbitraryValue(originalSlice)

		assert.Equal(t, ArrayType, value.Type())
		assert.Equal(t, 2, value.Count())
		assert.Equal(t, Int(2), value.GetByIndex(0))
		assert.Equal(t, ArrayOf(String("x")), value.GetByIndex(1))
		item, ok := value.TryGetByIndex(0)
		assert.True(t, ok)
		assert.Equal(t, Int(2), item)

		assert.Equal(t, Null(), value.GetByIndex(-1))
		assert.Equal(t, Null(), value.GetByIndex(2))
		item, ok = value.TryGetByIndex(2)
		assert.False(t, ok)
		assert.Equal(t, Null(), item)
	})
	t.Run("[]Value", func(t *testing.T) {
		originalSlice := []Value{Int(2), ArrayOf(String("x"))}
		value := UnsafeUseArbitraryValue(originalSlice)
		// this becomes an ordinary array value
		assert.Equal(t, ArrayOf(Int(2), ArrayOf(String("x"))), value)
	})
	t.Run("map[string]interface{}", func(t *testing.T) {
		originalMap := map[string]interface{}{"x": []interface{}{2}}
		value := UnsafeUseArbitraryValue(originalMap)

		assert.Equal(t, ObjectType, value.Type())
		assert.Equal(t, 1, value.Count())
		assert.Equal(t, ArrayOf(Int(2)), value.GetByKey("x"))
		item, ok := value.TryGetByKey("x")
		assert.True(t, ok)
		assert.Equal(t, ArrayOf(Int(2)), item)

		keys := value.Keys()
		sort.Strings(keys)
		assert.Equal(t, []string{"x"}, keys)

		assert.Equal(t, Null(), value.GetByKey("y"))
		item, ok = value.TryGetByKey("y")
		assert.False(t, ok)
		assert.Equal(t, Null(), item)
	})
	t.Run("map[string]Value", func(t *testing.T) {
		originalMap := map[string]Value{"x": ArrayOf(Int(2))}
		value := UnsafeUseArbitraryValue(originalMap)
		// this becomes an ordinary object value
		assert.Equal(t, ObjectBuild().Set("x", ArrayOf(Int(2))).Build(), value)
	})
	t.Run("arbitrary struct", func(t *testing.T) {
		s := struct {
			X int `json:"x"`
		}{2}
		v := CopyArbitraryValue(s)
		// this becomes an ordinary object value
		assert.Equal(t, ObjectBuild().Set("x", Int(2)).Build(), v)
	})
}

func TestUnsafeConvertComplexTypesToSameArbitraryValue(t *testing.T) {
	// This verifies that the unsafe methods do *not* copy slices and maps of interface{}; they
	// preserve the original value. This behavior is needed for Go SDK v4 to avoid unexpected
	// extra overhead of deep-copying when using JsonVariation.
	t.Run("[]interface{}", func(t *testing.T) {
		originalSlice := []interface{}{2, []interface{}{"x"}}
		value := UnsafeUseArbitraryValue(originalSlice)

		assert.Equal(t, ArrayType, value.Type())
		assert.Equal(t, 2, value.Count())

		arbitraryValue := value.UnsafeArbitraryValue()
		resultSlice := arbitraryValue.([]interface{})
		assert.Equal(t, originalSlice, resultSlice)
		originalSlice[0] = 5
		assert.Equal(t, 5, resultSlice[0])
	})
	t.Run("map[string]interface{}", func(t *testing.T) {
		originalMap := map[string]interface{}{"x": []interface{}{2}}
		value := UnsafeUseArbitraryValue(originalMap)

		assert.Equal(t, ObjectType, value.Type())
		assert.Equal(t, 1, value.Count())

		arbitraryValue := value.UnsafeArbitraryValue()
		resultMap := arbitraryValue.(map[string]interface{})
		assert.Equal(t, originalMap, resultMap)
		originalMap["x"] = 5
		assert.Equal(t, 5, resultMap["x"])
	})
}

func TestConvertComplexTypesUnsafelyFromArbitraryValueAndUnsafelyBackAgain(t *testing.T) {
	t.Run("map[string]interface{}", func(t *testing.T) {
		mapValue0 := map[string]interface{}{"x": []interface{}{"b"}}
		v := UnsafeUseArbitraryValue(mapValue0)
		mapValue1 := v.UnsafeArbitraryValue()
		assert.Equal(t, mapValue0, mapValue1)
		// Verify that it's the same map, not deep-copied
		mapValue0["x"].([]interface{})[0] = "c"
		assert.Equal(t, mapValue0, mapValue1)
	})
	t.Run("[]interface{}", func(t *testing.T) {
		sliceValue0 := []interface{}{[]interface{}{"b"}}
		v := UnsafeUseArbitraryValue(sliceValue0)
		sliceValue1 := v.UnsafeArbitraryValue()
		assert.Equal(t, sliceValue0, sliceValue1)
		// Verify that it's the same slice, not deep-copied
		sliceValue0[0].([]interface{})[0] = "c"
		assert.Equal(t, sliceValue0, sliceValue1)
	})
}

func TestConvertComplexTypesUnsafelyFromArbitraryValueAndSafelyBackAgain(t *testing.T) {
	t.Run("map[string]interface{}", func(t *testing.T) {
		mapValue0 := map[string]interface{}{"x": []interface{}{"b"}}
		v := UnsafeUseArbitraryValue(mapValue0)
		mapValue1 := v.AsArbitraryValue()
		assert.Equal(t, mapValue0, mapValue1)
		// Verify that the map was deep-copied
		mapValue0["x"].([]interface{})[0] = "c"
		assert.NotEqual(t, mapValue0, mapValue1)
	})
	t.Run("[]interface{}", func(t *testing.T) {
		sliceValue0 := []interface{}{[]interface{}{"b"}}
		v := UnsafeUseArbitraryValue(sliceValue0)
		sliceValue1 := v.AsArbitraryValue()
		assert.Equal(t, sliceValue0, sliceValue1)
		// Verify that the slice was deep-copied
		sliceValue0[0].([]interface{})[0] = "c"
		assert.NotEqual(t, sliceValue0, sliceValue1)
	})
}

func TestUnsafeValueJsonMarshal(t *testing.T) {
	items := []struct {
		value interface{}
		json  string
	}{
		{nil, "null"},
		{true, "true"},
		{false, "false"},
		{1, "1"},
		{float64(1), "1"},
		{float64(2.5), "2.5"},
		{"x", `"x"`},
		{[]interface{}{true, "x"}, `[true,"x"]`},
		{map[string]interface{}{"a": true}, `{"a":true}`},
		{json.RawMessage("[3]"), "[3]"},
	}
	for _, item := range items {
		t.Run(fmt.Sprintf("value %v, json %v", item.value, item.json), func(t *testing.T) {
			ldValue := UnsafeUseArbitraryValue(item.value)
			j, err := json.Marshal(ldValue)
			assert.NoError(t, err)
			assert.Equal(t, item.json, string(j))

			assert.Equal(t, item.json, ldValue.String())
			assert.Equal(t, item.json, ldValue.JSONString())
		})
	}
}

func TestUnsafeComplexValuesEqualSafeComplexValues(t *testing.T) {
	valueFnGroups := [][]func() Value{
		[]func() Value{
			func() Value { return UnsafeUseArbitraryValue([]interface{}{}) },
			func() Value { return ArrayOf() },
		},
		[]func() Value{
			func() Value { return UnsafeUseArbitraryValue([]interface{}{1}) },
			func() Value { return ArrayOf(Int(1)) },
		},
		[]func() Value{
			func() Value { return UnsafeUseArbitraryValue([]interface{}{1, []interface{}{"a"}}) },
			func() Value { return ArrayOf(Int(1), ArrayOf(String("a"))) },
		},
		[]func() Value{
			func() Value { return UnsafeUseArbitraryValue(map[string]interface{}{}) },
			func() Value { return ObjectBuild().Build() },
		},
		[]func() Value{
			func() Value { return UnsafeUseArbitraryValue(map[string]interface{}{"a": 1}) },
			func() Value { return ObjectBuild().Set("a", Int(1)).Build() },
		},
	}
	for thisGroupIndex, equivalentFns := range valueFnGroups {
		for _, fn0 := range equivalentFns {
			for otherGroupIndex, otherFns := range valueFnGroups {
				for _, fn1 := range otherFns {
					if thisGroupIndex == otherGroupIndex {
						valuesShouldBeEqual(t, fn0(), fn1())
					} else {
						valuesShouldNotBeEqual(t, fn0(), fn1())
					}
				}
			}
		}
	}
}
