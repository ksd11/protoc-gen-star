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

var debug_is_required bool = false
var debug_field_type string = ""
var debug_field_name string = ""
var debug_field_value string = ""
var debug_ignore_empty bool = false
var debug_rules map[string]interface{} = make(map[string]interface{})

func debug() {
	res := ""
	if debug_is_required {
		res += "required "
	} else {
		res += "optional "
	}
	res += debug_field_type + " " + debug_field_name
	if debug_ignore_empty {
		res += " = \"\" [ignore_empty]"
		fmt.Fprintln(os.Stderr, res)
		return
	}

	res += " = " + debug_field_value + "["
	for k, v := range debug_rules {
		res += k + "=" + fmt.Sprintf("%v", v) + ","
	}
	res += "]"
	fmt.Fprintln(os.Stderr, res)
}

func debug_clear() {
	debug_is_required = false
	debug_field_type = ""
	debug_field_name = ""
	debug_field_value = ""
	debug_ignore_empty = false
	debug_rules = make(map[string]interface{})
}

/*
*

	检测字段是否复合规范
	1. 若字段是必须的，是否已经设置
	2. 字段的类型是否一致
	3. 字段是否符合校验规则
*/
func ParseField(f pgs.Field, rawData map[string]string) (isValidate bool, msg []string) {
	debug_clear() // for debug
	debug_field_name = f.Name().String()

	// 检验字段是否是必须的
	isValidate, msg = checkRequired(f, rawData)
	if !isValidate {
		return
	}

	// 检验字段类型和校验信息
	isValidate, msg = checkRule(f, rawData)

	debug() // for debug
	return
}

func checkRequired(f pgs.Field, rawData map[string]string) (isValidate bool, msg []string) {
	isValidate = true
	msg = []string{}
	if f.Required() {
		debug_is_required = true // for debug
		if _, ok := rawData[f.Name().String()]; !ok {
			isValidate = false
			msg = append(msg, fmt.Sprintf("字段 %s 是必须的", f.Name().String()))
		}
	}
	return
}

/*
*

	处理数值类型的数据
	1. 先获取Number的所有校验规则
	2. 然后验证Number对应的value是否符合校验规则
*/
func handleNumber[T Number](value_any any, rules protoreflect.ProtoMessage) (isValidate bool, msg []string) {
	parsedRules := parseNumber[T](rules)
	return validateRules[T](value_any.(T), parsedRules)
}

func checkRule(f pgs.Field, rawData map[string]string) (isValidate bool, msg []string) {
	isValidate = true
	msg = []string{}
	ruleContext, err := rulesContext(f)
	if err != nil {
		isValidate = false
		msg = append(msg, err.Error())
		return
	}
	// for debug
	debug_field_type = ruleContext.Typ

	// ignore_empty
	reflect_val := getValue(ruleContext.Rules)
	ok, ignore_empty := GetBool(reflect_val, "IgnoreEmpty")
	if ok && ignore_empty && rawData[f.Name().String()] == "" {
		ignore_empty = true
		debug_ignore_empty = true // for debug
		return
	}

	// 校验类型
	value_any, err := TypeConvertFuncMap[ruleContext.Typ](rawData[f.Name().String()])
	if err != nil {
		isValidate = false
		msg = append(msg, err.Error())
		return
	}
	debug_field_value = fmt.Sprintf("%v", rawData[f.Name().String()])

	// validate
	switch ruleContext.Typ {
	case "uint32", "fixed32":
		return handleNumber[uint32](value_any, ruleContext.Rules)
	case "uint64", "fixed64":
		return handleNumber[uint64](value_any, ruleContext.Rules)
	case "int32", "sint32", "sfixed32":
		return handleNumber[int32](value_any, ruleContext.Rules)
	case "int64", "sint64", "sfixed64":
		return handleNumber[int64](value_any, ruleContext.Rules)
	case "double":
		return handleNumber[float64](value_any, ruleContext.Rules)
	case "float":
		return handleNumber[float32](value_any, ruleContext.Rules)
	default:
		isValidate = false
		msg = append(msg, fmt.Sprintf("不支持类型 %s", ruleContext.Typ))
	}

	return
}

func getValue(numberRules protoreflect.ProtoMessage) reflect.Value {
	val := reflect.ValueOf(numberRules)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	return val
}

// 返回一堆验证函数
func parseNumber[T Number](numberRules protoreflect.ProtoMessage) []RuleFunc[T] {
	val := getValue(numberRules)

	var ok bool
	var rules []RuleFunc[T]

	ok, constVal := GetFieldPointer[T](val, "Const")
	if ok {
		rules = append(rules, NumberConst(constVal))
		debug_rules["const"] = constVal // for debug
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
			rules = append(rules, NumberRange(left_value, right_value))
		} else if left == "(" && right == "]" {
			rules = append(rules, NumberRangeR(left_value, right_value))
		} else if left == "[" && right == ")" {
			rules = append(rules, NumberRangeL(left_value, right_value))
		} else {
			rules = append(rules, NumberRangeLR(left_value, right_value))
		}
		debug_rules["range"] = fmt.Sprintf("%v%v,%v%v", left, left_value, right_value, right) // for debug
	} else {
		if lt {
			rules = append(rules, NumberLt(ltVal))
			debug_rules["lt"] = ltVal // for debug
		}
		if lte {
			rules = append(rules, NumberLte(lteVal))
			debug_rules["lte"] = lteVal // for debug
		}
		if gt {
			rules = append(rules, NumberGt(gtVal))
			debug_rules["gt"] = gtVal // for debug
		}
		if gte {
			rules = append(rules, NumberGte(gteVal))
			debug_rules["gte"] = gteVal // for debug
		}
	}

	ok, in := GetFieldArray[T](val, "In")
	if ok {
		rules = append(rules, NumberIn(in))
		debug_rules["in"] = in // for debug
	}
	ok, not_in := GetFieldArray[T](val, "NotIn")
	if ok {
		rules = append(rules, NumberNotIn(in))
		debug_rules["not_in"] = not_in // for debug
	}
	return rules
}

// 验证value是否满足规则，只要有任意一个规则不通过，则不通过
func validateRules[T any](val T, rules []RuleFunc[T]) (check bool, msg []string) {
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

func rulesContext(f pgs.Field) (out shared.RuleContext, err error) {
	out.Field = f

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
		err = fmt.Errorf("unknown rule type (%T)", rules.Type)
	}

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
