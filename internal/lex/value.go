package lex

import (
	"strconv"
)

type Value struct {
	data any
}

func stringValue(v string) Value {
	return Value{
		data: v,
	}
}

func intValue(v int32) Value {
	return Value{
		data: v,
	}
}

func floatValue(v float64) Value {
	return Value{
		data: v,
	}
}

func parseIntValue(text []byte) (Value, error) {
	v, err := strconv.ParseInt(string(text), 10, 32)
	if err != nil {
		return Value{}, err
	}
	return intValue(int32(v)), nil
}

func parseFloatValue(text []byte) (Value, error) {
	v, err := strconv.ParseFloat(string(text), 64)
	if err != nil {
		return Value{}, err
	}
	return floatValue(v), nil
}

func (v Value) String() string {
	switch value := v.data.(type) {
	case string:
		return value
	case int32:
		return strconv.FormatInt(int64(value), 10)
	case float64:
		return strconv.FormatFloat(value, 'g', -1, 64)
	default:
		return ""
	}
}

func (v *Value) AsInt32() (int32, bool) {
	result, ok := v.data.(int32)
	return result, ok
}

func (v *Value) AsFloat64() (float64, bool) {
	result, ok := v.data.(float64)
	return result, ok
}

func (v *Value) AsString() (string, bool) {
	result, ok := v.data.(string)
	return result, ok
}
