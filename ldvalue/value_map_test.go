package ldvalue

import (
	"fmt"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilValueMap(t *testing.T) {
	m := ValueMap{}
	assert.False(t, m.IsDefined())
	assert.Equal(t, 0, m.Count())
	assert.Equal(t, Null(), m.Get("item"))
}

func TestCopyValueMap(t *testing.T) {
	m := map[string]Value{"item0": Int(1), "item1": Int(2)}
	vm1 := CopyValueMap(m)
	assert.Equal(t, m, vm1.data)
	shouldNotBeSameMap(t, m, vm1.data)

	vm2 := CopyValueMap(nil)
	assert.Nil(t, vm2.data)
	assert.False(t, vm2.IsDefined())

	vm3 := CopyValueMap(map[string]Value{})
	assert.NotNil(t, vm3.data)
	assert.Equal(t, 0, len(vm3.data))
	assert.True(t, vm3.IsDefined())
}

func TestCopyArbitraryValueMap(t *testing.T) {
	m := map[string]interface{}{"item0": "a", "item1": "b"}
	vm1 := CopyArbitraryValueMap(m)
	assert.Equal(t, map[string]Value{"item0": String("a"), "item1": String("b")}, vm1.data)

	vm2 := CopyArbitraryValueMap(nil)
	assert.Nil(t, vm2.data)
}

func TestValueMapBuild(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	b := ValueMapBuild().Set("item0", item0).Set("item1", item1)
	valueMap := b.Build()

	assert.Equal(t, 2, valueMap.Count())
	keys := valueMap.Keys(nil)
	sort.Strings(keys)
	assert.Equal(t, []string{"item0", "item1"}, keys)

	item0x := Bool(true)
	b.Set("item0", item0x)
	valueMapAfterModifyingBuilder := b.Build()
	assert.Equal(t, item0x, valueMapAfterModifyingBuilder.Get("item0"))
	assert.Equal(t, item0, valueMap.Get("item0")) // verifies builder's copy-on-write behavior

	assert.True(t, valueMap.IsDefined())

	m2 := ValueMapBuildWithCapacity(3).Set("item0", item0).Set("item1", item1).Build()
	assert.Equal(t, valueMap, m2)
}

func TestValueMapBuildFromMap(t *testing.T) {
	m0 := ValueMapBuild().Set("item0", Int(1)).Set("item1", Int(2)).Build()

	m1 := ValueMapBuildFromMap(m0).Build()
	assert.Equal(t, m0, m1)
	shouldBeSameMap(t, m0.data, m1.data)

	m2 := ValueMapBuild().Set("item2", Int(3))
	m2.SetAllFromValueMap(m0)
	assert.Equal(t, ValueMapBuild().Set("item0", Int(1)).Set("item1", Int(2)).Set("item2", Int(3)).Build(), m2.Build())

	// test copy-on-write behavior
	m3 := ValueMapBuild().Set("item0", Int(1)).Build()
	m4 := ValueMapBuildFromMap(m3).Set("item1", Int(2)).Build()
	assert.Equal(t, ValueMapBuild().Set("item0", Int(1)).Set("item1", Int(2)).Build(), m4)
	shouldNotBeSameMap(t, m3.data, m4.data)
	m5 := ValueMapBuild().SetAllFromValueMap(m3).Set("item1", Int(2)).Build()
	assert.NotEqual(t, m3, m5)
	shouldNotBeSameMap(t, m3.data, m5.data)
	assert.Equal(t, m4, m5)
}

func TestValueMapBuilderRemove(t *testing.T) {
	m0 := ValueMapBuild().Set("item0", Int(1)).Set("item1", Int(2)).Remove("item0").Build()
	assert.Equal(t, ValueMapBuild().Set("item1", Int(2)).Build(), m0)

	m1 := ValueMapBuildFromMap(m0).Remove("item1").Build()
	assert.Equal(t, ValueMapBuild().Build(), m1)
	assert.NotEqual(t, m0, m1)
	shouldNotBeSameMap(t, m0.data, m1.data)
}

func TestValueMapBuilderHasKey(t *testing.T) {
	var b ValueMapBuilder
	assert.False(t, b.HasKey("key1"))
	assert.False(t, b.HasKey("key2"))

	b.Set("key1", Int(1))
	assert.True(t, b.HasKey("key1"))
	assert.False(t, b.HasKey("key2"))

	b.Set("key2", Int(2))
	assert.True(t, b.HasKey("key1"))
	assert.True(t, b.HasKey("key2"))

	b.Remove("key1")
	assert.False(t, b.HasKey("key1"))
	assert.True(t, b.HasKey("key2"))
}

func TestValueMapBuilderSafety(t *testing.T) {
	// empty instance is safe to use
	var emptyInstance ValueMapBuilder
	emptyInstance.Set("key", Int(1))
	assert.Equal(t, ValueMapBuild().Set("key", Int(1)).Build(), emptyInstance.Build())

	// nil pointer is safe to use
	var nilPtr *ValueMapBuilder
	assert.Nil(t, nilPtr.Set("key", Int(1)))
	assert.Nil(t, nilPtr.SetAllFromValueMap(ValueMap{}))
	assert.Nil(t, nilPtr.Remove("key1"))
	assert.Equal(t, ValueMap{}, nilPtr.Build())
}

func TestValueMapGetByKey(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	m := ValueMapBuild().Set("item0", item0).Set("item1", item1).Build()

	assert.Equal(t, item0, m.Get("item0"))
	assert.Equal(t, item1, m.Get("item1"))
	assert.Equal(t, Null(), m.Get("bad-key"))

	item, ok := m.TryGet("item0")
	assert.True(t, ok)
	assert.Equal(t, item0, item)
	item, ok = m.TryGet("bad-key")
	assert.False(t, ok)
	assert.Equal(t, Null(), item)
}

func TestConvertValueMapToArbitraryValues(t *testing.T) {
	m := ValueMapBuild().Set("x", ArrayOf(Int(2))).Build()
	expected := map[string]interface{}{"x": []interface{}{float64(2)}}
	assert.Equal(t, expected, m.AsArbitraryValueMap())
}

func TestConvertValueMapFromArbitraryValuesAndBackAgain(t *testing.T) {
	mapValue0 := map[string]interface{}{"x": []interface{}{"b"}}
	m := CopyArbitraryValueMap(mapValue0)
	mapValue1 := m.AsArbitraryValueMap()
	assert.Equal(t, mapValue0, mapValue1)
	// Verify that the map was deep-copied
	mapValue0["x"].([]interface{})[0] = "c"
	assert.NotEqual(t, mapValue0, mapValue1)
}

func TestValueMapEqual(t *testing.T) {
	valueFns := []func() ValueMap{
		func() ValueMap { return ValueMap{} },
		func() ValueMap { return ValueMapBuild().Build() },
		func() ValueMap { return ValueMapBuild().Set("a", Int(1)).Build() },
		func() ValueMap { return ValueMapBuild().Set("a", Int(2)).Build() },
		func() ValueMap { return ValueMapBuild().Set("a", Int(1)).Set("b", Int(1)).Build() },
	}
	for i, fn0 := range valueFns {
		v0 := fn0()
		for j, fn1 := range valueFns {
			v1 := fn1()
			if i == j {
				assert.True(t, v0.Equal(v1), "%s should equal %s", v0, v1)
				assert.True(t, v0.Equal(v1), "%s should equal %s conversely", v1, v0)
			} else {
				assert.False(t, v0.Equal(v1), "%s should not equal %s", v0, v1)
				assert.False(t, v1.Equal(v0), "%s should not equal %s", v1, v0)
			}
		}
	}
}

func TestValueMapKeys(t *testing.T) {
	assert.Nil(t, ValueMap{}.Keys(nil))
	assert.Nil(t, ValueMapBuild().Build().Keys(nil))

	m := ValueMapBuild().Set("a", Int(1)).Build()

	assert.Equal(t, []string{"a"}, m.Keys(nil))
	slice1 := []string{"x", "y", "z"}
	assert.Equal(t, []string{"a"}, m.Keys(slice1))
	assert.Equal(t, []string{"a", "y", "z"}, slice1) // proves slice was reused
	slice2 := []string{}
	assert.Equal(t, []string{"a"}, m.Keys(slice2))

	m1 := ValueMapBuild().Set("a", Int(1)).Set("b", Int(2)).Set("c", Int(3)).Build()
	keys := m1.Keys(nil)
	sort.Strings(keys)
	assert.Equal(t, []string{"a", "b", "c"}, keys)
}

func TestValueMapAsValue(t *testing.T) {
	assert.Equal(t, Null(), ValueMap{}.AsValue())

	m := ValueMapBuild().Set("a", Int(1)).Set("b", Int(2)).Build()
	v := m.AsValue()
	assert.Equal(t, ObjectBuild().Set("a", Int(1)).Set("b", Int(2)).Build(), v)
	shouldBeSameMap(t, m.data, v.objectValue.data)
}

func TestValueMapAsMap(t *testing.T) {
	assert.Nil(t, ValueMap{}.AsMap())

	m := ValueMapBuild().Set("a", Bool(false)).Set("b", Bool(true)).Build()
	mm := m.AsMap()
	assert.Equal(t, map[string]Value{"a": Bool(false), "b": Bool(true)}, mm)
	shouldNotBeSameMap(t, m.data, mm)
}

func TestValueMapAsArbitraryValueMap(t *testing.T) {
	assert.Nil(t, ValueMap{}.AsArbitraryValueMap())

	m := ValueMapBuild().Set("a", Bool(false)).Set("b", Bool(true)).Build()
	mm := m.AsArbitraryValueMap()
	assert.Equal(t, map[string]interface{}{"a": false, "b": true}, mm)
}

func TestTransformValueMap(t *testing.T) {
	fnNoChanges := func(key string, value Value) (string, Value, bool) {
		return key, value, true
	}
	fnAbsoluteValuesAndNoOddNumbers := func(key string, value Value) (string, Value, bool) {
		if value.IntValue()%2 == 1 {
			return key, value, false // first return value should be ignored since second one is false
		}
		if value.IntValue() < 0 {
			return key, Int(-value.IntValue()), true
		}
		return key, value, true
	}
	fnTransformUsingKey := func(key string, value Value) (string, Value, bool) {
		return key, String(fmt.Sprintf("%s=%s", key, value.JSONString())), true
	}

	m1 := ValueMapBuild().Set("a", Int(2)).Set("b", Int(4)).Set("c", Int(6)).Build()
	m1a := m1.Transform(fnNoChanges)
	m1b := m1.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// Should have no changes...
	assert.Equal(t, m1, m1a)
	assert.Equal(t, m1, m1b)
	// ...and should be wrapping the *same* map, not a copy
	m1.data["a"] = Int(0)
	assert.Equal(t, m1, m1a)
	assert.Equal(t, m1, m1b)

	m2 := ValueMapBuild().Set("a", Int(2)).Set("b", Int(4)).Set("c", Int(1)).Set("d", Int(-6)).Build()
	m2a := m2.Transform(fnNoChanges)
	m2b := m2.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// m2a should have no changes, and should be wrapping the same map
	assert.Equal(t, m2, m2a)
	m2.data["a"] = Int(0)
	assert.Equal(t, m2, m2a)
	// m2b should have a transformed map
	assert.Equal(t, ValueMapBuild().Set("a", Int(2)).Set("b", Int(4)).Set("d", Int(6)).Build(), m2b)

	// Edge case where the only element is dropped
	m3 := ValueMapBuild().Set("a", Int(1)).Build()
	assert.Equal(t, ValueMapBuild().Build(), m3.Transform(fnAbsoluteValuesAndNoOddNumbers))

	// Transformation function that uses the key parameter
	m4 := ValueMapBuild().Set("a", Int(2)).Set("b", Int(4)).Build()
	assert.Equal(t, ValueMapBuild().Set("a", String("a=2")).Set("b", String("b=4")).Build(),
		m4.Transform(fnTransformUsingKey))

	// Case where we guarantee that the first element we iterated through is *not* modified - map
	// iteration order is nondeterministic, we just want to verify that we've hit all code paths
	n := 0
	fnTransformUsingKeyButNotFirst := func(key string, value Value) (string, Value, bool) {
		n++
		if n > 1 {
			return fnTransformUsingKey(key, value)
		}
		return key, value, true
	}
	m5 := m4.Transform(fnTransformUsingKeyButNotFirst)
	assert.NotEqual(t, m4, m5)
	assert.Equal(t, m4.Count(), m5.Count())
	if m5.Get("a").IsNumber() {
		assert.Equal(t, Int(2), m5.Get("a"))
		assert.Equal(t, String("b=4"), m5.Get("b"))
	} else {
		assert.Equal(t, String("a=2"), m5.Get("a"))
		assert.Equal(t, Int(4), m5.Get("b"))
	}

	shouldNotCallThis := func(key string, value Value) (string, Value, bool) {
		assert.Fail(t, "should not have called function")
		return key, value, true
	}
	assert.Equal(t, ValueMap{}, ValueMap{}.Transform(shouldNotCallThis))
	assert.Equal(t, ValueMapBuild().Build(), ValueMapBuild().Build().Transform(shouldNotCallThis))
}

func shouldBeSameMap(t *testing.T, m0 map[string]Value, m1 map[string]Value) {
	m0["temp-property"] = Bool(true)
	assert.Equal(t, m0, m1, "ValueMaps should be sharing same map but it was copied instead")
	delete(m0, "temp-property")
}

func shouldNotBeSameMap(t *testing.T, m0 map[string]Value, m1 map[string]Value) {
	m0["temp-property"] = Bool(true)
	assert.NotEqual(t, m0, m1, "ValueMaps should not be sharing same map but they are")
	delete(m0, "temp-property")
}

func TestCopyArbitraryMapOfType(t *testing.T) {
	var mNil map[string]string
	vm := CopyArbitraryValue(mNil)
	assert.Equal(t, ObjectType, vm.Type())

	mEmpty := map[string]string{}
	vm = CopyArbitraryValue(mEmpty)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{}, vm.objectValue.data)

	mStr := map[string]string{"a": "1", "b": "2", "c": "3"}
	vm = CopyArbitraryValue(mStr)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": String("1"), "b": String("2"), "c": String("3")}, vm.objectValue.data)

	mBool := map[string]bool{"a": true, "b": false, "c": true}
	vm = CopyArbitraryValue(mBool)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Bool(true), "b": Bool(false), "c": Bool(true)}, vm.objectValue.data)

	mInt := map[string]int{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mInt)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mInt8 := map[string]int8{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mInt8)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mInt16 := map[string]int16{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mInt16)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mInt32 := map[string]int32{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mInt32)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mInt64 := map[string]int64{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mInt64)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mUint := map[string]uint{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mUint)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mUint8 := map[string]uint8{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mUint8)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mUint16 := map[string]uint16{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mUint16)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mUint32 := map[string]uint32{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mUint32)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mUint64 := map[string]uint64{"a": 1, "b": 2, "c": 3}
	vm = CopyArbitraryValue(mUint64)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Int(1), "b": Int(2), "c": Int(3)}, vm.objectValue.data)

	mFloat32 := map[string]float32{"a": 1.0, "b": 2.0, "c": 3.0}
	vm = CopyArbitraryValue(mFloat32)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Float64(1.0), "b": Float64(2.0), "c": Float64(3.0)}, vm.objectValue.data)

	mFloat64 := map[string]float64{"a": 1.1, "b": 2.2, "c": 3.3}
	vm = CopyArbitraryValue(mFloat64)
	assert.Equal(t, ObjectType, vm.Type())
	assert.Equal(t, map[string]Value{"a": Float64(1.1), "b": Float64(2.2), "c": Float64(3.3)}, vm.objectValue.data)
}
