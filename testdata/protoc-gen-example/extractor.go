package main

import (
	"reflect"
)

// 判断val中是否含名称为name的field。（类型为 *uint32）
// 如果有，返回true, 取值。否则返回false, 0
func GetUint32(val reflect.Value, name string) (bool, uint32) {
	field := val.FieldByName(name)
	if field.IsValid() {
		v, ok := field.Interface().(*uint32)
		if ok && v != nil {
			return true, *v
		}
	}
	return false, 0
}

func GetBool(val reflect.Value, name string) (bool, bool) {
	field := val.FieldByName(name)
	if field.IsValid() {
		v, ok := field.Interface().(*bool)
		if ok && v != nil {
			return true, *v
		}
	}
	return false, false
}

func GetFieldPointer[T uint32 | uint64 | int32 | int64 | float32 | float64](val reflect.Value, name string) (bool, T) {
	var zero T
	field := val.FieldByName(name)
	if field.IsValid() {
		v, ok := field.Interface().(*T)
		if ok && v != nil {
			return true, *v
		}
	}
	return false, zero
}

func GetFieldArray[T any](val reflect.Value, name string) (bool, []T) {
	field := val.FieldByName(name)
	if field.IsValid() {
		v, ok := field.Interface().([]T)
		if ok && v != nil {
			return true, v
		}
	}
	return false, nil
}
