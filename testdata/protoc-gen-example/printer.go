package main

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strings"

	"bytes"

	"github.com/envoyproxy/protoc-gen-validate/templates/shared"
	"github.com/envoyproxy/protoc-gen-validate/validate"
	pgs "github.com/lyft/protoc-gen-star/v2"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type PrinterModule struct {
	*pgs.ModuleBase
}

func ASTPrinter() *PrinterModule { return &PrinterModule{ModuleBase: &pgs.ModuleBase{}} }

func (p *PrinterModule) Name() string { return "printer" }

func (p *PrinterModule) Execute(targets map[string]pgs.File, packages map[string]pgs.Package) []pgs.Artifact {
	buf := &bytes.Buffer{}

	for _, f := range targets {
		p.printFile(f, buf)
	}

	return p.Artifacts()
}

func (p *PrinterModule) printFile(f pgs.File, buf *bytes.Buffer) {
	p.Push(f.Name().String())
	defer p.Pop()

	buf.Reset()
	v := initPrintVisitor(buf, "")
	p.CheckErr(pgs.Walk(v, f), "unable to print AST tree")

	out := buf.String()

	if ok, _ := p.Parameters().Bool("log_tree"); ok {
		p.Logf("Proto Tree:\n%s", out)
	}

	p.AddGeneratorFile(
		f.InputPath().SetExt(".tree.txt").String(),
		out,
	)
}

const (
	startNodePrefix = "┳ "
	subNodePrefix   = "┃"
	leafNodePrefix  = "┣"
	leafNodeSpacer  = "━ "
)

type PrinterVisitor struct {
	pgs.Visitor
	prefix string
	w      io.Writer
}

func initPrintVisitor(w io.Writer, prefix string) pgs.Visitor {
	v := PrinterVisitor{
		prefix: prefix,
		w:      w,
	}
	v.Visitor = pgs.PassThroughVisitor(&v)
	return v
}

func (v PrinterVisitor) leafPrefix() string {
	if strings.HasSuffix(v.prefix, subNodePrefix) {
		return strings.TrimSuffix(v.prefix, subNodePrefix) + leafNodePrefix
	}
	return v.prefix
}

func (v PrinterVisitor) writeSubNode(str string) pgs.Visitor {
	fmt.Fprintf(v.w, "%s%s%s\n", v.leafPrefix(), startNodePrefix, str)
	return initPrintVisitor(v.w, fmt.Sprintf("%s%v", v.prefix, subNodePrefix))
}

func (v PrinterVisitor) writeLeaf(str string) {
	fmt.Fprintf(v.w, "%s%s%s\n", v.leafPrefix(), leafNodeSpacer, str)
}

func (v PrinterVisitor) VisitFile(f pgs.File) (pgs.Visitor, error) {
	return v.writeSubNode("File: " + f.Name().String()), nil
}

func (v PrinterVisitor) VisitMessage(m pgs.Message) (pgs.Visitor, error) {
	return v.writeSubNode("Message: " + m.Name().String()), nil
}

func (v PrinterVisitor) VisitEnum(e pgs.Enum) (pgs.Visitor, error) {
	return v.writeSubNode("Enum: " + e.Name().String()), nil
}

func (v PrinterVisitor) VisitService(s pgs.Service) (pgs.Visitor, error) {
	return v.writeSubNode("Service: " + s.Name().String()), nil
}

func (v PrinterVisitor) VisitEnumValue(ev pgs.EnumValue) (pgs.Visitor, error) {
	v.writeLeaf(ev.Name().String())
	return nil, nil
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

func parseNumber[T any](numberRules protoreflect.ProtoMessage) {
	val := reflect.ValueOf(numberRules)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	var ok bool

	ok, _ = GetFieldPointer[bool](val, "IgnoreEmpty")
	if ok {
		fmt.Fprintln(os.Stderr, "ignore_empty")
	}

	ok, constVal := GetFieldPointer[T](val, "Const")
	if ok {
		fmt.Fprintln(os.Stderr, "const: value = ", constVal)
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

		fmt.Fprintln(os.Stderr, "uint32: range ", left, left_value, right_value, right)
	} else {
		if lt {
			fmt.Fprintln(os.Stderr, "uint32: value < ", ltVal)
		}
		if lte {
			fmt.Fprintln(os.Stderr, "uint32: value <= ", lteVal)
		}
		if gt {
			fmt.Fprintln(os.Stderr, "uint32: value > ", gtVal)
		}
		if gte {
			fmt.Fprintln(os.Stderr, "uint32: value >= ", gteVal)
		}
	}

	ok, in := GetFieldArray[T](val, "In")
	if ok {
		fmt.Fprintln(os.Stderr, "uint32: value in ", in)
	}
	ok, not_in := GetFieldArray[T](val, "NotIn")
	if ok {
		fmt.Fprintln(os.Stderr, "uint32: value not in ", not_in)
	}
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

func (v PrinterVisitor) VisitField(f pgs.Field) (pgs.Visitor, error) {

	res := f.Name().String() + "\t"

	// 检验字段是否存在
	if f.Required() {
		res += "required" + "\t"
	} else {
		res += "optional" + "\t"
	}

	// 字段类型
	res += f.Type().ProtoType().String() + "\t"

	// 验证信息
	parseRule(f.Name().String(), f)

	v.writeLeaf(res)
	return nil, nil
}

func (v PrinterVisitor) VisitMethod(m pgs.Method) (pgs.Visitor, error) {
	v.writeLeaf(m.Name().String())
	return nil, nil
}
