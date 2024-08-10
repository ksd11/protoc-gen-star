package main

import "fmt"

// Number 是一个泛型约束，表示可以是以下任意一种数值类型
type Number interface {
	uint32 | uint64 | int32 | int64 | float32 | float64
}

type RuleFunc[T any] func(T) (bool, string)

// 判断数值是否在指定范围内 [left, right]
func NumberRangeLR[T Number](left, right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val >= left && val <= right {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 不在指定范围[%v, %v]", val, left, right)
		return false, message
	}
}

// 判断数值是否在指定范围内 [left, right)
func NumberRangeL[T Number](left, right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val >= left && val <= right {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 不在指定范围[%v, %v)", val, left, right)
		return false, message
	}
}

// 判断数值是否在指定范围内 (left, right]
func NumberRangeR[T Number](left, right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val >= left && val <= right {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 不在指定范围(%v, %v]", val, left, right)
		return false, message
	}
}

// 判断数值是否在指定范围内 (left, right)
func NumberRange[T Number](left, right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val >= left && val <= right {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 不在指定范围(%v, %v)", val, left, right)
		return false, message
	}
}

func NumberLt[T Number](right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val < right {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 必须小于%v", val, right)
		return false, message
	}
}

func NumberLte[T Number](right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val <= right {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 必须小于等于%v", val, right)
		return false, message
	}
}

func NumberGt[T Number](right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val > right {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 必须大于%v", val, right)
		return false, message
	}
}

func NumberGte[T Number](right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val >= right {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 必须大于等于%v", val, right)
		return false, message
	}
}

// Contains 检查元素是否在数组中
func Contains[T comparable](arr []T, elem T) bool {
	for _, v := range arr {
		if v == elem {
			return true
		}
	}
	return false
}

func NumberIn[T Number](right []T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if Contains(right, val) {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 应该在数组 %v", val, right)
		return false, message
	}
}

func NumberNotIn[T Number](right []T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if !Contains(right, val) {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 不应该在数组 %v", val, right)
		return false, message
	}
}

func NumberConst[T Number](right T) RuleFunc[T] {
	return func(val T) (bool, string) {
		if val == right {
			return true, ""
		}
		message := fmt.Sprintf("数值 %v 必须等于%v", val, right)
		return false, message
	}
}
