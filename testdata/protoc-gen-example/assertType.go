package main

import "strconv"

type ConvertFunc[T any] func(string) (T, error)

var TypeConvertFuncMap map[string]ConvertFunc[any] = make(map[string]ConvertFunc[any])

func init() {
	TypeConvertFuncMap["uint32"] = StringToUint32
	TypeConvertFuncMap["fixed32"] = StringToUint32
	TypeConvertFuncMap["uint64"] = StringToUint64
	TypeConvertFuncMap["fixed64"] = StringToUint64
	TypeConvertFuncMap["int32"] = StringToInt32
	TypeConvertFuncMap["sint32"] = StringToInt32
	TypeConvertFuncMap["sfixed32"] = StringToInt32
	TypeConvertFuncMap["int64"] = StringToInt64
	TypeConvertFuncMap["sint64"] = StringToInt64
	TypeConvertFuncMap["sfixed64"] = StringToInt64
	TypeConvertFuncMap["double"] = StringToFloat64
	TypeConvertFuncMap["float"] = StringToFloat32
}

// Convert string to int32
func StringToInt32(s string) (any, error) {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return int32(i), nil
}

// Convert string to uint32
func StringToUint32(s string) (any, error) {
	i, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, err
	}
	return uint32(i), nil
}

// Convert string to int64
func StringToInt64(s string) (any, error) {
	return strconv.ParseInt(s, 10, 64)
}

// Convert string to uint32
func StringToUint64(s string) (any, error) {
	i, err := strconv.ParseUint(s, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint64(i), nil
}

// Convert string to float32
func StringToFloat32(s string) (any, error) {
	f, err := strconv.ParseFloat(s, 32)
	if err != nil {
		return 0, err
	}
	return float32(f), nil
}

// Convert string to float64
func StringToFloat64(s string) (any, error) {
	return strconv.ParseFloat(s, 64)
}

func StringToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}
