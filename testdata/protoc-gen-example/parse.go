package main

import (
	"fmt"
	"os"
	"reflect"

	"github.com/envoyproxy/protoc-gen-validate/templates/shared"
	"github.com/envoyproxy/protoc-gen-validate/validate"
	pgs "github.com/lyft/protoc-gen-star/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

/*
*

	检测字段是否复合规范
	1. 若字段是必须的，是否已经设置
	2. 字段的类型是否一致
	3. 字段是否符合校验规则
*/
func ParseField(f pgs.Field, rawData map[string]interface{}) (isValidate bool, msg []string) {
	// 检验字段是否是必须的
	isValidate, msg = checkRequired(f, rawData)
	if !isValidate {
		return
	}

	// 字段类型
	// res += f.Type().ProtoType().String() + "\t"
	isValidate, msg = checkType(f, rawData)
	if !isValidate {
		return
	}

	// 验证信息
	// parseRule(f.Name().String(), f)
	isValidate, msg = checkRule(f, rawData)
	return
}

func checkRequired(f pgs.Field, rawData map[string]interface{}) (isValidate bool, msg []string) {
	isValidate = true
	msg = []string{}
	if f.Required() {
		if _, ok := rawData[f.Name().String()]; !ok {
			isValidate = false
			msg = append(msg, fmt.Sprintf("字段 %s 是必须的", f.Name().String()))
		}
	}
	return
}

func checkType(f pgs.Field, rawData map[string]interface{}) (isValidate bool, msg []string) {
	isValidate = true
	msg = []string{}
	// try{
	// 	rawData[f.Name().String()].(f.Type().ProtoType().String()
	// }
	return
}

func checkRule(f pgs.Field, rawData map[string]interface{}) (isValidate bool, msg []string) {
	isValidate = true
	msg = []string{}
	// try{
	// 	rawData[f.Name().String()].(f.Type().ProtoType().String()
	// }
	return
}

func resolveRules(typ interface{ IsEmbed() bool }, rules *validate.FieldRules) (ruleType string, rule proto.Message, messageRule *validate.MessageRules, wrapped bool) {
	switch r := rules.GetType().(type) {
	case *validate.FieldRules_Float:
		ruleType, rule, wrapped = "float", r.Float, typ.IsEmbed()
	case *validate.FieldRules_Double:
		ruleType, rule, wrapped = "double", r.Double, typ.IsEmbed()
	case *validate.FieldRules_Int32:
		ruleType, rule, wrapped = "int32", r.Int32, typ.IsEmbed()
	case *validate.FieldRules_Int64:
		ruleType, rule, wrapped = "int64", r.Int64, typ.IsEmbed()
	case *validate.FieldRules_Uint32:
		ruleType, rule, wrapped = "uint32", r.Uint32, typ.IsEmbed()
	case *validate.FieldRules_Uint64:
		ruleType, rule, wrapped = "uint64", r.Uint64, typ.IsEmbed()
	case *validate.FieldRules_Sint32:
		ruleType, rule, wrapped = "sint32", r.Sint32, false
	case *validate.FieldRules_Sint64:
		ruleType, rule, wrapped = "sint64", r.Sint64, false
	case *validate.FieldRules_Fixed32:
		ruleType, rule, wrapped = "fixed32", r.Fixed32, false
	case *validate.FieldRules_Fixed64:
		ruleType, rule, wrapped = "fixed64", r.Fixed64, false
	case *validate.FieldRules_Sfixed32:
		ruleType, rule, wrapped = "sfixed32", r.Sfixed32, false
	case *validate.FieldRules_Sfixed64:
		ruleType, rule, wrapped = "sfixed64", r.Sfixed64, false
	case *validate.FieldRules_Bool:
		ruleType, rule, wrapped = "bool", r.Bool, typ.IsEmbed()
	case *validate.FieldRules_String_:
		ruleType, rule, wrapped = "string", r.String_, typ.IsEmbed()
	case *validate.FieldRules_Bytes:
		ruleType, rule, wrapped = "bytes", r.Bytes, typ.IsEmbed()
	case *validate.FieldRules_Enum:
		ruleType, rule, wrapped = "enum", r.Enum, false
	case *validate.FieldRules_Repeated:
		ruleType, rule, wrapped = "repeated", r.Repeated, false
	case *validate.FieldRules_Map:
		ruleType, rule, wrapped = "map", r.Map, false
	case *validate.FieldRules_Any:
		ruleType, rule, wrapped = "any", r.Any, false
	case *validate.FieldRules_Duration:
		ruleType, rule, wrapped = "duration", r.Duration, false
	case *validate.FieldRules_Timestamp:
		ruleType, rule, wrapped = "timestamp", r.Timestamp, false
	case nil:
		if ft, ok := typ.(pgs.FieldType); ok && ft.IsRepeated() {
			return "repeated", &validate.RepeatedRules{}, rules.Message, false
		} else if ok && ft.IsMap() && ft.Element().IsEmbed() {
			return "map", &validate.MapRules{}, rules.Message, false
		} else if typ.IsEmbed() {
			return "message", rules.GetMessage(), rules.GetMessage(), false
		}
		return "none", nil, nil, false
	default:
		ruleType, rule, wrapped = "error", nil, false
	}

	return ruleType, rule, rules.Message, wrapped
}

// 返回一堆验证函数
func parseNumber[T uint32 | uint64 | int32 | int64 | float32 | float64](numberRules protoreflect.ProtoMessage) []RuleFunc[T] {
	val := reflect.ValueOf(numberRules)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var ok bool
	var rules []RuleFunc[T]

	ok, _ = GetBool(val, "IgnoreEmpty")
	if ok {
		fmt.Fprintln(os.Stderr, "ignore_empty")
	}

	ok, constVal := GetFieldPointer[T](val, "Const")
	if ok {
		fmt.Fprintf(os.Stderr, "%v : const value = %v\n", val.Type(), constVal)
	}

	lt, ltVal := GetFieldPointer[T](val, "Lt")
	lte, lteVal := GetFieldPointer[T](val, "Lte")
	gt, gtVal := GetFieldPointer[T](val, "Gt")
	gte, gteVal := GetFieldPointer[T](val, "Gte")

	if (lt || lte) && (gt || gte) {
		left := "["
		left_value := gteVal
		if gt {
			left = "("
			left_value = gtVal
		}
		right := "]"
		right_value := lteVal
		if lt {
			right = ")"
			right_value = ltVal
		}
		if left == "(" && right == ")" {
			rules = append(rules, NumberRange[T](left_value, right_value))
		} else if left == "(" && right == "]" {
			rules = append(rules, NumberRangeR[T](left_value, right_value))
		} else if left == "[" && right == ")" {
			rules = append(rules, NumberRangeL[T](left_value, right_value))
		} else {
			rules = append(rules, NumberRangeLR[T](left_value, right_value))
		}
		// if left_value < right_value {
		// 	fmt.Fprintf(os.Stderr, ": range %v %v,%v %v", left, left_value, right_value, right)
		// }

		fmt.Fprintf(os.Stderr, "%v : range %v %v,%v %v\n", val.Type(), left, left_value, right_value, right)
	} else {
		if lt {
			fmt.Fprintf(os.Stderr, "%v : value < %v\n", val.Type(), ltVal)
		}
		if lte {
			fmt.Fprintf(os.Stderr, "%v: value <= %v\n", val.Type(), lteVal)
		}
		if gt {
			fmt.Fprintf(os.Stderr, "%v: value > %v\n", val.Type(), gtVal)
		}
		if gte {
			fmt.Fprintf(os.Stderr, "%v: value >= %v\n", val.Type(), gteVal)
		}
	}

	ok, in := GetFieldArray[T](val, "In")
	if ok {
		fmt.Fprintf(os.Stderr, "%v: value in %v\n", val.Type(), in)
	}
	ok, not_in := GetFieldArray[T](val, "NotIn")
	if ok {
		fmt.Fprintf(os.Stderr, "%v: value not in %v\n", val.Type(), not_in)
	}
	return rules
}

// 验证value是否满足规则，只要有任意一个规则不通过，则不通过
func checkRules[T any](val T, rules []RuleFunc[T]) (check bool, msg []string) {
	check = true
	msg = []string{}
	for _, rule := range rules {
		if ok, m := rule(val); !ok {
			msg = append(msg, m)
			check = false
		}
	}
	return
}

func parseRule(name string, f pgs.Field) (out shared.RuleContext, err error) {

	var rules validate.FieldRules
	if _, err = f.Extension(validate.E_Rules, &rules); err != nil {
		return
	}

	var wrapped bool
	if out.Typ, out.Rules, out.MessageRules, wrapped = resolveRules(f.Type(), &rules); wrapped {
		out.WrapperTyp = out.Typ
		out.Typ = "wrapper"
	}

	if out.Typ == "error" {
		err = fmt.Errorf("unknown rule type (%T)", rules)
	}

	if out.Rules == nil {
		return
	}
	// 只有复合类型或者具有验证条件的字段才会下来

	fmt.Fprintln(os.Stderr, "----------------")
	fmt.Fprintln(os.Stderr, "name:", name)

	switch out.Typ {
	// case "string":
	// 	fmt.Fprintln(os.Stderr, "string:", out.Rules.GetConst())
	case "uint32", "fixed32":
		parseNumber[uint32](out.Rules)
	case "uint64", "fixed64":
		parseNumber[uint64](out.Rules)
	case "int32", "sint32", "sfixed32":
		parseNumber[int32](out.Rules)
	case "int64", "sint64", "sfixed64":
		parseNumber[int64](out.Rules)
	case "double":
		parseNumber[float64](out.Rules)
	case "float":
		parseNumber[float32](out.Rules)
	default:
		fmt.Fprintln(os.Stderr, "unknown type")
	}
	return

}
