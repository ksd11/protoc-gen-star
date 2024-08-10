package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// 返回数据的接口
func ReadJsonData() map[string]string {
	return simpleJsonData
}

// 返回一条field的验证结果
func OutputOneFieldValidateResult(name string, isValidate bool, msg []string) {
	if !isValidate {
		fmt.Fprintln(os.Stderr, "name :", name)
		for _, m := range msg {
			fmt.Fprintln(os.Stderr, m)
		}
	}
	fmt.Fprintln(os.Stderr, "-------------")
}

// 对simple.proto的简单测试集
var simpleJsonData map[string]string

func init() {
	jsonStr := `{"float_val": "0.3"
				, "double_val": "0.05"
				, "int32_val": "3"
				, "int64_val": ""
				, "uint32_val": "5"
				, "uint64_val": "3"
				, "sint32_val": "3"
				, "sint64_val": "3"
				, "fixed32_val": "3"
				, "fixed64_val": "3"
				, "sfixed32_val": "3"
				, "sfixed64_val": "3"}`
	err := json.Unmarshal([]byte(jsonStr), &simpleJsonData)
	if err != nil {
		fmt.Println("Error:", err)
	}
}
