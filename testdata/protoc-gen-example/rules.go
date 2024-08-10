package main

import "fmt"

type RuleFunc[T any] func(T) (bool, string)

// 判断数值是否在指定范围内 [left, right]
func NumberRangeLR[T uint32 | uint64 | int32 | int64 | float32 | float64](left, right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val >= left && val <= right {
			return true, ""
		}
		message := fmt.Sprintf("数值不在指定范围[%v, %v]", left, right)
		return false, message
	}
}

// 判断数值是否在指定范围内 [left, right)
func NumberRangeL[T uint32 | uint64 | int32 | int64 | float32 | float64](left, right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val >= left && val <= right {
			return true, ""
		}
		message := fmt.Sprintf("数值不在指定范围[%v, %v)", left, right)
		return false, message
	}
}

// 判断数值是否在指定范围内 (left, right]
func NumberRangeR[T uint32 | uint64 | int32 | int64 | float32 | float64](left, right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val >= left && val <= right {
			return true, ""
		}
		message := fmt.Sprintf("数值不在指定范围(%v, %v]", left, right)
		return false, message
	}
}

// 判断数值是否在指定范围内 (left, right)
func NumberRange[T uint32 | uint64 | int32 | int64 | float32 | float64](left, right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val >= left && val <= right {
			return true, ""
		}
		message := fmt.Sprintf("数值不在指定范围(%v, %v)", left, right)
		return false, message
	}
}
