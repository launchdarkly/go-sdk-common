package ldvalue

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/launchdarkly/go-jsonstream/v2/jreader"
	"github.com/launchdarkly/go-jsonstream/v2/jwriter"

	"github.com/stretchr/testify/assert"
)

func TestNilValueArray(t *testing.T) {
	a := ValueArray{}
	assert.False(t, a.IsDefined())
	assert.Equal(t, 0, a.Count())
	assert.Equal(t, Null(), a.Get(0))
}

func TestCopyValueArray(t *testing.T) {
	s := []Value{String("a"), String("b")}
	a1 := CopyValueArray(s)
	assert.Equal(t, s, a1.data)
	shouldNotBeSameSlice(t, s, a1.data)

	a2 := CopyValueArray(nil)
	assert.Nil(t, a2.data)

	s3 := []Value{}
	a3 := CopyValueArray(s3)
	assert.Equal(t, []Value{}, a3.data)
}

func TestCopyArbitraryValueArray(t *testing.T) {
	a1 := CopyArbitraryValueArray([]interface{}{"a", "b"})
	assert.Equal(t, []Value{String("a"), String("b")}, a1.data)

	a2 := CopyArbitraryValueArray(nil)
	assert.Nil(t, a2.data)

	a3 := CopyArbitraryValueArray([]interface{}{})
	assert.Equal(t, []Value{}, a3.data)
}

func TestValueArrayOf(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	a1 := ValueArrayOf(item0, item1)

	assert.Equal(t, 2, a1.Count())
	assert.True(t, a1.IsDefined())

	a2 := ValueArrayOf()
	assert.Equal(t, 0, a2.Count())
	assert.True(t, a2.IsDefined())
}

func TestValueArrayBuild(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	builder := ValueArrayBuild().Add(item0).Add(item1)
	a := builder.Build()

	assert.True(t, a.IsDefined())
	assert.Equal(t, 2, a.Count())
	assert.Equal(t, ValueArrayOf(item0, item1), a)

	item2 := Bool(true)
	builder.Add(item2)
	valueAfterModifyingBuilder := builder.Build()

	assert.Equal(t, 3, valueAfterModifyingBuilder.Count())
	assert.Equal(t, item2, valueAfterModifyingBuilder.Get(2))

	assert.Equal(t, 2, a.Count()) // verifies builder's copy-on-write behavior

	assert.Equal(t, ValueArrayOf(), ValueArrayBuild().Build())

	assert.Equal(t, a, ValueArrayBuildWithCapacity(3).Add(item0).Add(item1).Build())
}

func TestValueArrayBuildFromArray(t *testing.T) {
	a0 := ValueArrayOf(String("a"), String("b"))

	a1 := ValueArrayBuildFromArray(a0).Build()
	assert.Equal(t, a0, a1)
	shouldBeSameSlice(t, a0.data, a1.data)

	// test copy-on-write behavior
	a3 := ValueArrayOf(String("a"))
	a4 := ValueArrayBuildFromArray(a3).Add(String("b")).Build()
	assert.Equal(t, ValueArrayOf(String("a"), String("b")), a4)
	shouldNotBeSameSlice(t, a3.data, a4.data)
	a5 := ValueArrayBuild().AddAllFromValueArray(a3).Add(String("b")).Build()
	assert.NotEqual(t, a3, a5)
	shouldNotBeSameSlice(t, a3.data, a5.data)
	assert.Equal(t, a4, a5)
}

func TestValueArrayGetByIndex(t *testing.T) {
	item0 := String("a")
	item1 := Int(1)
	a := ValueArrayOf(item0, item1)

	assert.Equal(t, item0, a.Get(0))
	assert.Equal(t, item1, a.Get(1))
	assert.Equal(t, Null(), a.Get(-1))
	assert.Equal(t, Null(), a.Get(2))

	item, ok := a.TryGet(0)
	assert.True(t, ok)
	assert.Equal(t, item0, item)
	item, ok = a.TryGet(2)
	assert.False(t, ok)
	assert.Equal(t, Null(), item)
}

func TestConvertValueArrayToArbitraryValues(t *testing.T) {
	a := ValueArrayBuild().Add(String("a")).Add(String("b")).Build()
	expected := []interface{}{"a", "b"}
	assert.Equal(t, expected, a.AsArbitraryValueSlice())
}

func TestConvertValueArrayFromArbitraryValuesAndBackAgain(t *testing.T) {
	slice0 := []interface{}{"a", "b"}
	a := CopyArbitraryValueArray(slice0)
	slice1 := a.AsArbitraryValueSlice()
	assert.Equal(t, slice0, slice1)
	// Verify that the slice was deep-copied
	slice0[0] = "c"
	assert.NotEqual(t, slice0, slice1)
}

func TestValueArrayEqual(t *testing.T) {
	valueFns := []func() ValueArray{
		func() ValueArray { return ValueArray{} },
		func() ValueArray { return ValueArrayBuild().Build() },
		func() ValueArray { return ValueArrayBuild().Add(String("a")).Build() },
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

func TestValueArrayAsValue(t *testing.T) {
	assert.Equal(t, Null(), ValueArray{}.AsValue())

	a := ValueArrayOf(String("a"), String("b"))
	v := a.AsValue()
	assert.Equal(t, ArrayOf(String("a"), String("b")), v)
	shouldBeSameSlice(t, a.data, v.arrayValue.data)
}

func TestValueArrayAsSlice(t *testing.T) {
	assert.Nil(t, ValueArray{}.AsSlice())

	a := ValueArrayOf(String("a"), String("b"))
	s := a.AsSlice()
	assert.Equal(t, []Value{String("a"), String("b")}, s)
	shouldNotBeSameSlice(t, a.data, s)
}

func TestValueArrayAsArbitraryValueSlice(t *testing.T) {
	assert.Nil(t, ValueArray{}.AsArbitraryValueSlice())

	a := ValueArrayOf(String("a"), String("b"))
	s := a.AsArbitraryValueSlice()
	assert.Equal(t, []interface{}{"a", "b"}, s)
}

func recordValueArrayEnumerateCalls(a ValueArray, stopFn func(enumerateParams) bool) []enumerateParams {
	ret := []enumerateParams{}
	a.Enumerate(func(index int, v Value) bool {
		params := enumerateParams{index, "", v}
		ret = append(ret, params)
		if stopFn != nil && stopFn(params) {
			return false
		}
		return true
	})
	return ret
}

func TestValueArrayEnumerate(t *testing.T) {
	assert.Equal(t, []enumerateParams{}, recordValueArrayEnumerateCalls(ValueArray{}, nil))
	assert.Equal(t, []enumerateParams{}, recordValueArrayEnumerateCalls(ValueArrayBuild().Build(), nil))

	a1 := ValueArrayOf(String("a"), String("b"))

	e1 := recordValueArrayEnumerateCalls(a1, nil)
	assert.Equal(t, []enumerateParams{
		enumerateParams{0, "", String("a")},
		enumerateParams{1, "", String("b")},
	}, e1)

	e2 := recordValueArrayEnumerateCalls(a1, func(p enumerateParams) bool { return true })
	assert.Len(t, e2, 1)
}

func TestValueArrayTransform(t *testing.T) {
	fnNoChanges := func(index int, value Value) (Value, bool) {
		return value, true
	}
	fnAbsoluteValuesAndNoOddNumbers := func(index int, value Value) (Value, bool) {
		if value.IntValue()%2 == 1 {
			return value, false // first return value should be ignored since second one is false
		}
		if value.IntValue() < 0 {
			return Int(-value.IntValue()), true
		}
		return value, true
	}
	fnTransformUsingIndex := func(index int, value Value) (Value, bool) {
		return String(fmt.Sprintf("%d=%s", index, value.StringValue())), true
	}

	array1 := ValueArrayOf(Int(2), Int(4), Int(6))
	array1a := array1.Transform(fnNoChanges)
	array1b := array1.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// Should have no changes...
	assert.Equal(t, array1, array1a)
	assert.Equal(t, array1, array1b)
	// ...and should be wrapping the *same* slice, not a copy
	shouldBeSameSlice(t, array1.data, array1a.data)
	shouldBeSameSlice(t, array1.data, array1b.data)

	array2 := ValueArrayOf(Int(2), Int(4), Int(1), Int(-6))
	array2a := array2.Transform(fnNoChanges)
	array2b := array2.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// array2a should have no changes, and should be wrapping the same slice
	assert.Equal(t, array2, array2a)
	shouldBeSameSlice(t, array2.data, array2a.data)
	// array2b should have a transformed slice
	assert.Equal(t, ValueArrayOf(Int(2), Int(4), Int(6)), array2b)

	// Same as the array2 tests, except that the first change is a modification, not a deletion
	array3 := ValueArrayOf(Int(2), Int(4), Int(-6), Int(1))
	array3a := array3.Transform(fnNoChanges)
	array3b := array3.Transform(fnAbsoluteValuesAndNoOddNumbers)
	// array3a should have no changes, and should be wrapping the same slice
	assert.Equal(t, array3, array3a)
	shouldBeSameSlice(t, array3.data, array3a.data)
	// array3b should have a transformed slice
	assert.Equal(t, ValueArrayOf(Int(2), Int(4), Int(6)), array3b)

	// Edge case where the very first element is dropped
	array4 := ValueArrayOf(Int(1), Int(2), Int(4))
	array4b := array4.Transform(fnAbsoluteValuesAndNoOddNumbers)
	assert.Equal(t, ValueArrayOf(Int(2), Int(4)), array4b)

	// Edge case where the only element is dropped
	array5 := ValueArrayOf(Int(1))
	assert.Equal(t, ValueArrayOf(), array5.Transform(fnAbsoluteValuesAndNoOddNumbers))

	// Transformation function that uses the index parameter
	array6 := ValueArrayOf(String("a"), String("b"))
	assert.Equal(t, ValueArrayOf(String("0=a"), String("1=b")), array6.Transform(fnTransformUsingIndex))

	shouldNotCallThis := func(index int, value Value) (Value, bool) {
		assert.Fail(t, "should not have called function")
		return value, true
	}
	assert.Equal(t, ValueArray{}, ValueArray{}.Transform(shouldNotCallThis))
	assert.Equal(t, ValueArrayOf(), ValueArrayOf().Transform(shouldNotCallThis))
}

func TestValueArrayJSONMarshalUnmarshal(t *testing.T) {
	items := []struct {
		valueArray ValueArray
		json       string
	}{
		{ValueArray{}, nullAsJSON},
		{ValueArrayBuild().Build(), `[]`},
		{ValueArrayOf(String("a"), String("b")), `["a","b"]`},
	}
	for _, item := range items {
		t.Run(fmt.Sprintf("json %v", item.json), func(t *testing.T) {
			j, err := json.Marshal(item.valueArray)
			assert.NoError(t, err)
			assert.Equal(t, item.json, string(j))

			assert.Equal(t, item.json, item.valueArray.String())
			assert.Equal(t, item.json, item.valueArray.JSONString())

			var a ValueArray
			err = json.Unmarshal([]byte(item.json), &a)
			assert.NoError(t, err)
			assert.Equal(t, item.valueArray, a)

			r := jreader.NewReader([]byte(item.json))
			a = ValueArray{}
			a.ReadFromJSONReader(&r)
			assert.NoError(t, r.Error())
			assert.Equal(t, item.valueArray, a)

			w := jwriter.NewWriter()
			item.valueArray.WriteToJSONWriter(&w)
			assert.NoError(t, w.Error())
			assert.Equal(t, item.json, string(w.Bytes()))
		})
	}

	for _, badJSON := range []string{"true", "1", `"x"`, "{}"} {
		err := json.Unmarshal([]byte(badJSON), &ValueArray{})
		assert.Error(t, err)
		assert.IsType(t, &json.UnmarshalTypeError{}, err)
	}
}

func shouldBeSameSlice(t *testing.T, s0 []Value, s1 []Value) {
	old := s0[0]
	s0[0] = String("temp-value")
	assert.Equal(t, s0, s1, "ValueArrays should be sharing same slice but it was copied instead")
	s0[0] = old
}

func shouldNotBeSameSlice(t *testing.T, s0 []Value, s1 []Value) {
	old := s0[0]
	s0[0] = String("temp-value")
	assert.NotEqual(t, s0, s1, "ValueArrays should not be sharing same slice but they are")
	s0[0] = old
}
