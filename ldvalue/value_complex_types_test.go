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
