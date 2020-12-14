package ldvalue

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayOf(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	value := ArrayOf(item0, item1)

	assert.Equal(t, ArrayType, value.Type())
	assert.Equal(t, 2, value.Count())

	assert.True(t, value.IsDefined())
	assert.False(t, value.IsNull())
	assert.False(t, value.IsNumber())
	assert.False(t, value.IsInt())

	assert.False(t, value.BoolValue())
	assert.Equal(t, 0, value.IntValue())
	assert.Equal(t, float64(0), value.Float64Value())
	assert.Equal(t, "", value.StringValue())
	assert.Equal(t, OptionalString{}, value.AsOptionalString())
}

func TestArrayBuild(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	builder := ArrayBuild().Add(item0).Add(item1)
	value := builder.Build()

	assert.Equal(t, ArrayType, value.Type())
	assert.Equal(t, 2, value.Count())
	assert.Equal(t, ArrayOf(item0, item1), value)

	item2 := Bool(true)
	builder.Add(item2)
	valueAfterModifyingBuilder := builder.Build()

	assert.Equal(t, ArrayType, valueAfterModifyingBuilder.Type())
	assert.Equal(t, 3, valueAfterModifyingBuilder.Count())
	assert.Equal(t, item2, valueAfterModifyingBuilder.GetByIndex(2))

	assert.Equal(t, 2, value.Count()) // verifies builder's copy-on-write behavior

	assert.Equal(t, ArrayOf(), ArrayBuild().Build())
}

func TestArrayGetByIndex(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	value := ArrayOf(item0, item1)

	assert.Equal(t, item0, value.GetByIndex(0))
	assert.Equal(t, item1, value.GetByIndex(1))
	assert.Equal(t, Null(), value.GetByIndex(-1))
	assert.Equal(t, Null(), value.GetByIndex(2))

	item, ok := value.TryGetByIndex(0)
	assert.True(t, ok)
	assert.Equal(t, item0, item)
	item, ok = value.TryGetByIndex(2)
	assert.False(t, ok)
	assert.Equal(t, Null(), item)
}

func TestUsingArrayMethodsForNonArrayValue(t *testing.T) {
	values := []Value{
		Null(),
		Bool(true),
		Int(1),
		Float64(2.5),
		String(""),
		Raw(json.RawMessage("1")),
	}
	for _, v := range values {
		t.Run(v.String(), func(t *testing.T) {
			assert.Equal(t, 0, v.Count())
			assert.Equal(t, Null(), v.GetByIndex(0))
			_, ok := v.TryGetByIndex(0)
			assert.False(t, ok)
		})
	}
}

func TestObjectBuild(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	b := ObjectBuild().Set("item0", item0).Set("item1", item1)
	value := b.Build()

	assert.Equal(t, ObjectType, value.Type())
	assert.Equal(t, 2, value.Count())
	keys := value.Keys()
	sort.Strings(keys)
	assert.Equal(t, []string{"item0", "item1"}, keys)

	item0x := Bool(true)
	b.Set("item0", item0x)
	valueAfterModifyingBuilder := b.Build()
	assert.Equal(t, item0x, valueAfterModifyingBuilder.GetByKey("item0"))
	assert.Equal(t, item0, value.GetByKey("item0")) // verifies builder's copy-on-write behavior

	assert.True(t, value.IsDefined())
	assert.False(t, value.IsNull())
	assert.False(t, value.IsNumber())
	assert.False(t, value.IsInt())

	assert.False(t, value.BoolValue())
	assert.Equal(t, 0, value.IntValue())
	assert.Equal(t, float64(0), value.Float64Value())
	assert.Equal(t, "", value.StringValue())
	assert.Equal(t, OptionalString{}, value.AsOptionalString())
}

func TestObjectGetByKey(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	value := ObjectBuild().Set("item0", item0).Set("item1", item1).Build()

	assert.Equal(t, item0, value.GetByKey("item0"))
	assert.Equal(t, item1, value.GetByKey("item1"))
	assert.Equal(t, Null(), value.GetByKey("bad-key"))

	item, ok := value.TryGetByKey("item0")
	assert.True(t, ok)
	assert.Equal(t, item0, item)
	item, ok = value.TryGetByKey("bad-key")
	assert.False(t, ok)
	assert.Equal(t, Null(), item)
}

func TestUsingObjectMethodsForNonObjectValue(t *testing.T) {
	values := []Value{
		Null(),
		Bool(true),
		Int(1),
		Float64(2.5),
		String(""),
		Raw(json.RawMessage("1")),
	}
	for _, v := range values {
		t.Run(v.String(), func(t *testing.T) {
			assert.Nil(t, v.Keys())
			assert.Equal(t, Null(), v.GetByKey("x"))
			_, ok := v.TryGetByKey("x")
			assert.False(t, ok)
		})
	}
}

func TestConvertComplexTypesFromArbitraryValue(t *testing.T) {
	t.Run("[]interface{}", func(t *testing.T) {
		v := CopyArbitraryValue([]interface{}{2, []interface{}{"x"}})
		assert.Equal(t, ArrayOf(Int(2), ArrayOf(String("x"))), v)
	})
	t.Run("[]Value", func(t *testing.T) {
		v := CopyArbitraryValue([]Value{Int(2), ArrayOf(String("x"))})
		assert.Equal(t, ArrayOf(Int(2), ArrayOf(String("x"))), v)
	})
	t.Run("map[string]interface{}", func(t *testing.T) {
		v := CopyArbitraryValue(map[string]interface{}{"x": []interface{}{2}})
		assert.Equal(t, ObjectBuild().Set("x", ArrayOf(Int(2))).Build(), v)
	})
	t.Run("map[string]Value", func(t *testing.T) {
		v := CopyArbitraryValue(map[string]Value{"x": ArrayOf(Int(2))})
		assert.Equal(t, ObjectBuild().Set("x", ArrayOf(Int(2))).Build(), v)
	})
	t.Run("arbitrary struct", func(t *testing.T) {
		s := struct {
			X int `json:"x"`
		}{2}
		v := CopyArbitraryValue(s)
		assert.Equal(t, ObjectBuild().Set("x", Int(2)).Build(), v)
	})
}

func TestConvertComplexTypesToArbitraryValue(t *testing.T) {
	t.Run("array", func(t *testing.T) {
		v := ArrayOf(Int(2), ArrayOf(String("x")))
		expected := []interface{}{float64(2), []interface{}{"x"}}
		assert.Equal(t, expected, v.AsArbitraryValue())
	})
	t.Run("object", func(t *testing.T) {
		v := ObjectBuild().Set("x", ArrayOf(Int(2))).Build()
		expected := map[string]interface{}{"x": []interface{}{float64(2)}}
		assert.Equal(t, expected, v.AsArbitraryValue())
	})
}

func TestConvertComplexTypesFromArbitraryValueAndBackAgain(t *testing.T) {
	t.Run("map[string]interface{}", func(t *testing.T) {
		mapValue0 := map[string]interface{}{"x": []interface{}{"b"}}
		v := CopyArbitraryValue(mapValue0)
		mapValue1 := v.AsArbitraryValue()
		assert.Equal(t, mapValue0, mapValue1)
		// Verify that the map was deep-copied
		mapValue0["x"].([]interface{})[0] = "c"
		assert.NotEqual(t, mapValue0, mapValue1)
	})
	t.Run("[]interface{}", func(t *testing.T) {
		sliceValue0 := []interface{}{[]interface{}{"b"}}
		v := CopyArbitraryValue(sliceValue0)
		sliceValue1 := v.AsArbitraryValue()
		assert.Equal(t, sliceValue0, sliceValue1)
		// Verify that the slice was deep-copied
		sliceValue0[0].([]interface{})[0] = "c"
		assert.NotEqual(t, sliceValue0, sliceValue1)
	})
}

func TestEqualComplexTypes(t *testing.T) {
	valueFns := []func() Value{
		func() Value { return Null() },
		func() Value { return Bool(false) },
		func() Value { return ArrayOf() },
		func() Value { return ArrayOf(Int(1)) },
		func() Value { return ArrayOf(Int(2)) },
		func() Value { return ArrayOf(Int(1), ArrayOf(String("a"))) },
		func() Value { return ArrayOf(Int(1), ArrayOf(String("a"), String("b"))) },
		func() Value { return ObjectBuild().Build() },
		func() Value { return ObjectBuild().Set("a", Int(1)).Build() },
		func() Value { return ObjectBuild().Set("a", Int(2)).Build() },
		func() Value { return ObjectBuild().Set("a", Int(1)).Set("b", Int(1)).Build() },
	}
	for i, fn0 := range valueFns {
		v0 := fn0()
		for j, fn1 := range valueFns {
			v1 := fn1()
			if i == j {
				valuesShouldBeEqual(t, v0, v1)
			} else {
				valuesShouldNotBeEqual(t, v0, v1)
			}
		}
	}
}

type enumerateParams struct {
	index int
	key   string
	value Value
}

func recordEnumerateCalls(value Value, stopFn func(enumerateParams) bool) []enumerateParams {
	ret := []enumerateParams{}
	value.Enumerate(func(index int, key string, v Value) bool {
		params := enumerateParams{index, key, v}
		ret = append(ret, params)
		if stopFn != nil && stopFn(params) {
			return false
		}
		return true
	})
	return ret
}

func TestEnumerateSimpleTypes(t *testing.T) {
	values := []Value{Bool(true), Int(1), String("x")}
	for _, value := range values {
		calls := recordEnumerateCalls(value, nil)
		assert.Equal(t, []enumerateParams{enumerateParams{0, "", value}}, calls)
	}
	assert.Equal(t, []enumerateParams{}, recordEnumerateCalls(Null(), nil))
}

func TestEnumerateArray(t *testing.T) {
	assert.Equal(t, []enumerateParams{}, recordEnumerateCalls(ArrayOf(), nil))

	assert.Equal(t, []enumerateParams{
		enumerateParams{0, "", Int(1)},
		enumerateParams{1, "", String("a")},
	}, recordEnumerateCalls(ArrayOf(Int(1), String("a")), nil))

	assert.Equal(t, []enumerateParams{
		enumerateParams{0, "", Int(1)},
	}, recordEnumerateCalls(ArrayOf(Int(1), String("a")), func(p enumerateParams) bool {
		return p.index == 0
	}))
}

func TestEnumerateObject(t *testing.T) {
	assert.Equal(t, []enumerateParams{}, recordEnumerateCalls(ObjectBuild().Build(), nil))

	o1 := ObjectBuild().Set("a", Int(1)).Set("b", Int(2)).Build()
	e1 := recordEnumerateCalls(o1, nil)
	sort.Slice(e1, func(i, j int) bool { return e1[i].key < e1[j].key })
	assert.Equal(t, []enumerateParams{
		enumerateParams{0, "a", Int(1)},
		enumerateParams{0, "b", Int(2)},
	}, e1)

	e2 := recordEnumerateCalls(o1, func(p enumerateParams) bool { return true })
	assert.Len(t, e2, 1)
}

func TestTransformSimpleTypes(t *testing.T) {
	fnDropAllValues := func(index int, key string, value Value) (Value, bool) {
		return String("should ignore this value since the second value is false"), false
	}
	fnWrapInArray := func(index int, key string, value Value) (Value, bool) {
		return ArrayOf(value), true
	}
	values := []Value{Bool(true), Int(1), String("x")}
	for _, value := range values {
		assert.Equal(t, Null(), value.Transform(fnDropAllValues))
		assert.Equal(t, ArrayOf(value), value.Transform(fnWrapInArray))
	}
	assert.Equal(t, Null(), Null().Transform(fnDropAllValues))
	assert.Equal(t, Null(), Null().Transform(fnWrapInArray))
}

func TestTransformArray(t *testing.T) {
	// The implementation of this delegates to ValueMap.Transform, so we're just verifying a
	// basic set of inputs and outputs rather than every possible behavior.
	fnAddOne := func(index int, key string, value Value) (Value, bool) {
		return Int(value.IntValue() + 1), true
	}
	a1 := ArrayOf(Int(2), Int(4))
	a1a := a1.Transform(fnAddOne)
	assert.Equal(t, ArrayOf(Int(3), Int(5)), a1a)
}

func TestTransformObject(t *testing.T) {
	// The implementation of this delegates to ValueMap.Transform, so we're just verifying a
	// basic set of inputs and outputs rather than every possible behavior.
	fnAddOne := func(index int, key string, value Value) (Value, bool) {
		return Int(value.IntValue() + 1), true
	}
	o1 := ObjectBuild().Set("a", Int(2)).Set("b", Int(4)).Build()
	o1a := o1.Transform(fnAddOne)
	assert.Equal(t, ObjectBuild().Set("a", Int(3)).Set("b", Int(5)).Build(), o1a)
}

func TestAsValueArray(t *testing.T) {
	value := ArrayOf(String("a"))
	a := value.AsValueArray()
	assert.Equal(t, ValueArrayOf(String("a")), a)
	shouldBeSameSlice(t, a.data, value.arrayValue.data)

	assert.Equal(t, ValueArray{}, Null().AsValueArray())
	assert.Equal(t, ValueArray{}, Bool(true).AsValueArray())
	assert.Equal(t, ValueArray{}, Int(1).AsValueArray())
	assert.Equal(t, ValueArray{}, String("x").AsValueArray())
	assert.Equal(t, ValueArray{}, ObjectBuild().Build().AsValueArray())
}

func TestAsValueMap(t *testing.T) {
	value := ObjectBuild().Set("a", Int(1)).Build()
	m := value.AsValueMap()
	assert.Equal(t, ValueMapBuild().Set("a", Int(1)).Build(), m)
	shouldBeSameMap(t, m.data, value.objectValue.data)

	assert.Equal(t, ValueMap{}, Null().AsValueMap())
	assert.Equal(t, ValueMap{}, Bool(true).AsValueMap())
	assert.Equal(t, ValueMap{}, Int(1).AsValueMap())
	assert.Equal(t, ValueMap{}, String("x").AsValueMap())
	assert.Equal(t, ValueMap{}, ArrayOf().AsValueMap())
}
