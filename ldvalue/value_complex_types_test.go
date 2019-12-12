package ldvalue

import (
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

	assert.False(t, value.AsBool())
	assert.Equal(t, float64(0), value.AsFloat64())
	assert.Equal(t, "", value.AsString())
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
