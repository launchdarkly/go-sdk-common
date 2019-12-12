package ldvalue

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNullValue(t *testing.T) {
	v := Null()

	assert.Equal(t, NullType, v.Type())
	assert.True(t, v.IsNull())
	assert.False(t, v.IsNumber())
	assert.False(t, v.IsInt())

	assert.Equal(t, Null(), v)
	assert.Equal(t, Value{}, v)

	// treating a null as a non-null produces empty values
	assert.False(t, v.AsBool())
	assert.Equal(t, float64(0), v.AsFloat64())
	assert.Equal(t, "", v.AsString())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestBoolValue(t *testing.T) {
	tv := Bool(true)

	assert.Equal(t, BoolType, tv.Type())
	assert.True(t, tv.AsBool())
	assert.False(t, tv.IsNull())
	assert.False(t, tv.IsNumber())
	assert.False(t, tv.IsInt())

	assert.Equal(t, Bool(true), tv)
	assert.NotEqual(t, Bool(false), tv)

	// treating a bool as a non-bool produces empty values
	assert.Equal(t, float64(0), tv.AsFloat64())
	assert.Equal(t, "", tv.AsString())
	assert.Equal(t, 0, tv.Count())
	assert.Equal(t, Null(), tv.GetByIndex(0))
	assert.Equal(t, Null(), tv.GetByKey("x"))

	fv := Bool(false)

	assert.Equal(t, BoolType, fv.Type())
	assert.False(t, fv.AsBool())
	assert.False(t, fv.IsNull())
	assert.False(t, fv.IsNumber())
	assert.False(t, fv.IsInt())

	assert.Equal(t, Bool(false), fv)
	assert.NotEqual(t, Bool(true), fv)
}

func TestIntValue(t *testing.T) {
	v := Int(2)

	assert.Equal(t, NumberType, v.Type())
	assert.Equal(t, 2, v.AsInt())
	assert.Equal(t, float64(2), v.AsFloat64())
	assert.False(t, v.IsNull())
	assert.True(t, v.IsNumber())
	assert.True(t, v.IsInt())

	assert.Equal(t, Int(2), v)
	assert.Equal(t, Float64(2), v)
	assert.NotEqual(t, Float64(2.5), v)

	// treating a number as a non-number produces empty values
	assert.False(t, v.AsBool())
	assert.Equal(t, "", v.AsString())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestFloat64Value(t *testing.T) {
	v := Float64(2.75)

	assert.Equal(t, NumberType, v.Type())
	assert.Equal(t, 2, v.AsInt())
	assert.Equal(t, 2.75, v.AsFloat64())
	assert.False(t, v.IsNull())
	assert.True(t, v.IsNumber())
	assert.False(t, v.IsInt())

	floatButReallyInt := Float64(2.0)
	assert.Equal(t, NumberType, floatButReallyInt.Type())
	assert.Equal(t, 2, floatButReallyInt.AsInt())
	assert.Equal(t, 2.0, floatButReallyInt.AsFloat64())
	assert.False(t, floatButReallyInt.IsNull())
	assert.True(t, floatButReallyInt.IsNumber())
	assert.True(t, floatButReallyInt.IsInt())

	assert.Equal(t, Float64(2.75), v)
	assert.NotEqual(t, Float64(2.5), v)

	// treating a number as a non-number produces empty values
	assert.False(t, v.AsBool())
	assert.Equal(t, "", v.AsString())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestStringValue(t *testing.T) {
	v := String("abc")

	assert.Equal(t, StringType, v.Type())
	assert.Equal(t, "abc", v.AsString())
	assert.False(t, v.IsNull())
	assert.False(t, v.IsNumber())
	assert.False(t, v.IsInt())
	assert.Equal(t, v, String("abc"))

	assert.Equal(t, String("abc"), v)
	assert.NotEqual(t, String("def"), v)

	// treating a string as a non-string produces empty values
	assert.False(t, v.AsBool())
	assert.Equal(t, float64(0), v.AsFloat64())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestRawValue(t *testing.T) {
	rawJson := json.RawMessage([]byte("[1]"))
	v := Raw(rawJson)

	assert.Equal(t, RawType, v.Type())
	assert.Equal(t, rawJson, v.AsRaw())
	assert.False(t, v.IsNull())
	assert.False(t, v.IsNumber())
	assert.False(t, v.IsInt())

	// conversion of other types to Raw is covered in value_serialization_test

	// treating a Raw as a non-Raw produces empty values
	assert.False(t, v.AsBool())
	assert.Equal(t, float64(0), v.AsFloat64())
	assert.Equal(t, 0, v.Count())
	assert.Equal(t, Null(), v.GetByIndex(0))
	assert.Equal(t, Null(), v.GetByKey("x"))
}

func TestConvertPrimitivesFromArbitraryValue(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		v := CopyArbitraryValue(nil)
		assert.Equal(t, Null(), v)
	})
	t.Run("Value", func(t *testing.T) {
		originalValue := Int(1)
		assert.Equal(t, originalValue, CopyArbitraryValue(originalValue))
	})
	t.Run("bool", func(t *testing.T) {
		tv := CopyArbitraryValue(true)
		assert.Equal(t, Bool(true), tv)

		fv := CopyArbitraryValue(false)
		assert.Equal(t, Bool(false), fv)
	})
	t.Run("int8", func(t *testing.T) {
		v := CopyArbitraryValue(int8(1))
		assert.Equal(t, Int(1), v)
		assert.Equal(t, Float64(1), v)
	})
	t.Run("uint8", func(t *testing.T) {
		v := CopyArbitraryValue(uint8(1))
		assert.Equal(t, Int(1), v)
		assert.Equal(t, Float64(1), v)
	})
	t.Run("int16", func(t *testing.T) {
		v := CopyArbitraryValue(int16(1))
		assert.Equal(t, Int(1), v)
		assert.Equal(t, Float64(1), v)
	})
	t.Run("uint16", func(t *testing.T) {
		v := CopyArbitraryValue(uint16(1))
		assert.Equal(t, Int(1), v)
		assert.Equal(t, Float64(1), v)
	})
	t.Run("int", func(t *testing.T) {
		v := CopyArbitraryValue(int(1))
		assert.Equal(t, Int(1), v)
		assert.Equal(t, Float64(1), v)
	})
	t.Run("uint", func(t *testing.T) {
		v := CopyArbitraryValue(uint(1))
		assert.Equal(t, Int(1), v)
		assert.Equal(t, Float64(1), v)
	})
	t.Run("int32", func(t *testing.T) {
		v := CopyArbitraryValue(int32(1))
		assert.Equal(t, Int(1), v)
		assert.Equal(t, Float64(1), v)
	})
	t.Run("uint32", func(t *testing.T) {
		v := CopyArbitraryValue(uint32(1))
		assert.Equal(t, Int(1), v)
		assert.Equal(t, Float64(1), v)
	})
	t.Run("float32", func(t *testing.T) {
		v := CopyArbitraryValue(float32(2.5))
		assert.Equal(t, Float64(2.5), v)
	})
	t.Run("float64", func(t *testing.T) {
		v := CopyArbitraryValue(float64(2.5))
		assert.Equal(t, Float64(2.5), v)
	})
	t.Run("string", func(t *testing.T) {
		v := CopyArbitraryValue("x")
		assert.Equal(t, String("x"), v)
	})
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
		}{X: 2}
		v := CopyArbitraryValue(s)
		assert.Equal(t, ObjectBuild().Set("x", Int(2)).Build(), v)
	})
}

func TestConvertPrimitivesToArbitraryValue(t *testing.T) {
	assert.Nil(t, Null().AsArbitraryValue())
	assert.Equal(t, true, Bool(true).AsArbitraryValue())
	assert.Equal(t, false, Bool(false).AsArbitraryValue())
	assert.Equal(t, float64(2), Int(2).AsArbitraryValue())
	assert.Equal(t, "x", String("x").AsArbitraryValue())
}
