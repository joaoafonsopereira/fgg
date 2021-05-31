package fgg

import (
	"github.com/rhu1/fgg/internal/base"
	"regexp"
	"strconv"
	"strings"
)

/* constants */

type Tag int // maybe rename to TypeTag

const (
	BOOL Tag = iota
	INT32
	INT64
	FLOAT32
	FLOAT64
	STRING
)

var NamesToTags = map[string]Tag{
	"bool":    BOOL,
	"int32":   INT32,
	"int64":   INT64,
	"float32": FLOAT32,
	"float64": FLOAT64,
	"string":  STRING,
}

var TagsToNames = map[Tag]string{
	BOOL:    "bool",
	INT32:   "int32",
	INT64:   "int64",
	FLOAT32: "float32",
	FLOAT64: "float64",
	STRING:  "string",
}

func TagFromName(name string) Tag {
	if tag, ok := NamesToTags[name]; ok {
		return tag
	}
	//return NamesToTags[name]
	panic("Unknown primitive type")
}

func NameFromTag(tag Tag) string {
	return TagsToNames[tag]
}

/* "Exported" constructors */

func NewBool(lit string) PrimitiveLiteral {
	b, _ := strconv.ParseBool(lit)
	return PrimitiveLiteral{b, BOOL}
}

func NewIntLit(lit string) PrimitiveLiteral {
	if i, ok := newIntLit(lit); ok {
		return i
	} else {
		panic("Int const too big??")
	}
}

func NewFloatLit(lit string) PrimitiveLiteral {
	// try to fit literal into an integer first
	// (int can always be 'converted' back to float)
	if hasNoFractionalPart(lit) {
		truncated := truncateFractionalPart(lit)
		if i, ok := newIntLit(truncated); ok {
			return i // CHECKME(jp): maybe assign tag INT_OR_FLOAT in fgg (for type inference -- e.g. for 13.0, infer int or float?)
		}
	}
	if f, ok := newFloatLit(lit); ok {
		return f
	} else {
		panic("Float const too big??")
	}
}

func NewStringLit(lit string) PrimitiveLiteral {
	trim := strings.ReplaceAll(lit, "\"", "")

	return PrimitiveLiteral{trim, STRING}
}

// Converts a PrimitiveLiteral Ast node in a new node
// corresponding to a value of the expected/declared type.
func ConvertLitNode(lit PrimitiveLiteral, decldType Type) PrimtValue {
	switch t := decldType.(type) {
	case TPrimitive:
		return valueFromLiteral(lit, t)
	case TNamed:
		return NamedPrimitiveLiteral{lit, t}
	}
	panic("Literal: " + lit.String() + " can't assume type: " + decldType.String())
}

/******************************************************************************/
/* PrimtValue - base interface for primitive values */

type PrimtValue interface {
	FGGExpr
	Val() interface{}
	//Typ() Type
}

func (x PrimitiveLiteral) Val() interface{}      { return x.payload }
func (x NamedPrimitiveLiteral) Val() interface{} { return x.payload }
func (x BoolVal) Val() interface{}               { return x.val }
func (x Int32Val) Val() interface{}              { return x.val }
func (x Int64Val) Val() interface{}              { return x.val }
func (x Float32Val) Val() interface{}            { return x.val }
func (x Float64Val) Val() interface{}            { return x.val }
func (x StringVal) Val() interface{}             { return x.val }

/******************************************************************************/
/* PrimitiveLiteral */

// Represents a literal whose type is still undefined
//  (e.g. 123 is 'assignable' to either int32, int64 or MyInt, but the type is
//        only determined upon 'assignment')
// An int/float payload is saved as int64/float64.
type PrimitiveLiteral struct {
	payload interface{}
	tag     Tag
}

var _ PrimtValue = PrimitiveLiteral{}

func (x PrimitiveLiteral) Payload() interface{} { return x.payload }
func (x PrimitiveLiteral) Tag() Tag             { return x.tag }

func (x PrimitiveLiteral) Subs(map[Variable]FGGExpr) FGGExpr {
	return x
}

func (x PrimitiveLiteral) TSubs(_ map[TParam]Type) FGGExpr {
	return x
}

func (x PrimitiveLiteral) Eval([]Decl) (FGGExpr, string) {
	panic("Cannot reduce: " + x.String())
}

func (x PrimitiveLiteral) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr) {
	return TPrimitive{tag: x.tag, undefined: true}, x
}

func (x PrimitiveLiteral) IsValue() bool {
	return true
}

func (x PrimitiveLiteral) CanEval([]base.Decl) bool {
	return false
}

func (x PrimitiveLiteral) String() string {
	var payload string
	switch p := x.payload.(type) {
	case bool:
		payload = strconv.FormatBool(p)
	case string:
		payload = p
	case int64:
		payload = strconv.FormatInt(p, 10)
	case float64:
		payload = strconv.FormatFloat(p, 'E', -1, 64)
	default:
		panic("PrimitiveLiteral.String() for unsupported type")
	}
	tag := NameFromTag(x.tag)
	return "PrimitiveLiteral{payload=" + payload + ", tag=" + tag + "}"
}

func (x PrimitiveLiteral) ToGoString([]base.Decl) string {
	return x.String()
}

/******************************************************************************/
/* NamedPrimitiveLiteral */

// Essentially a PrimitiveLiteral whose type was already determined as a TNamed.
// Need this in order not to break type-safety at each (small) step of evaluation.
type NamedPrimitiveLiteral struct {
	PrimitiveLiteral
	typ TNamed
}

var _ PrimtValue = NamedPrimitiveLiteral{}

func (x NamedPrimitiveLiteral) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr) {
	return x.typ, x
}

func (x NamedPrimitiveLiteral) String() string {
	return "Named" + x.PrimitiveLiteral.String()
}

func (x NamedPrimitiveLiteral) ToGoString(ds []base.Decl) string {
	return "Named" + x.PrimitiveLiteral.ToGoString(ds)
}

/******************************************************************************/
/* Values of pre-declared (primitive) types */

type (
	BoolVal    struct{ val bool }
	Int32Val   struct{ val int32 }
	Int64Val   struct{ val int64 }
	Float32Val struct{ val float32 }
	Float64Val struct{ val float64 }
	StringVal  struct{ val string }
)

var _ PrimtValue = BoolVal{}
var _ PrimtValue = Int32Val{}
var _ PrimtValue = Int64Val{}
var _ PrimtValue = Float32Val{}
var _ PrimtValue = Float64Val{}
var _ PrimtValue = StringVal{}

func (x BoolVal) GetValue() bool       { return x.val }
func (x Int32Val) GetValue() int32     { return x.val }
func (x Int64Val) GetValue() int64     { return x.val }
func (x Float32Val) GetValue() float32 { return x.val }
func (x Float64Val) GetValue() float64 { return x.val }
func (x StringVal) GetValue() string   { return x.val }

func (x BoolVal) Subs(map[Variable]FGGExpr) FGGExpr    { return x }
func (x Int32Val) Subs(map[Variable]FGGExpr) FGGExpr   { return x }
func (x Int64Val) Subs(map[Variable]FGGExpr) FGGExpr   { return x }
func (x Float32Val) Subs(map[Variable]FGGExpr) FGGExpr { return x }
func (x Float64Val) Subs(map[Variable]FGGExpr) FGGExpr { return x }
func (x StringVal) Subs(map[Variable]FGGExpr) FGGExpr  { return x }

func (x BoolVal) TSubs(map[TParam]Type) FGGExpr    { return x }
func (x Int32Val) TSubs(map[TParam]Type) FGGExpr   { return x }
func (x Int64Val) TSubs(map[TParam]Type) FGGExpr   { return x }
func (x Float32Val) TSubs(map[TParam]Type) FGGExpr { return x }
func (x Float64Val) TSubs(map[TParam]Type) FGGExpr { return x }
func (x StringVal) TSubs(map[TParam]Type) FGGExpr  { return x }

func (x BoolVal) Eval([]Decl) (FGGExpr, string)    { panic("Cannot reduce: " + x.String()) }
func (x Int32Val) Eval([]Decl) (FGGExpr, string)   { panic("Cannot reduce: " + x.String()) }
func (x Int64Val) Eval([]Decl) (FGGExpr, string)   { panic("Cannot reduce: " + x.String()) }
func (x Float32Val) Eval([]Decl) (FGGExpr, string) { panic("Cannot reduce: " + x.String()) }
func (x Float64Val) Eval([]Decl) (FGGExpr, string) { panic("Cannot reduce: " + x.String()) }
func (x StringVal) Eval([]Decl) (FGGExpr, string)  { panic("Cannot reduce: " + x.String()) }

func (x BoolVal) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr)  { return TPrimitive{tag: BOOL}, x }
func (x Int32Val) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr) { return TPrimitive{tag: INT32}, x }
func (x Int64Val) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr) { return TPrimitive{tag: INT64}, x }
func (x Float32Val) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr) {
	return TPrimitive{tag: FLOAT32}, x
}
func (x Float64Val) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr) {
	return TPrimitive{tag: FLOAT64}, x
}
func (x StringVal) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr) {
	return TPrimitive{tag: STRING}, x
}

func (x BoolVal) IsValue() bool    { return true }
func (x Int32Val) IsValue() bool   { return true }
func (x Int64Val) IsValue() bool   { return true }
func (x Float32Val) IsValue() bool { return true }
func (x Float64Val) IsValue() bool { return true }
func (x StringVal) IsValue() bool  { return true }

func (x BoolVal) CanEval([]base.Decl) bool    { return false }
func (x Int32Val) CanEval([]base.Decl) bool   { return false }
func (x Int64Val) CanEval([]base.Decl) bool   { return false }
func (x Float32Val) CanEval([]base.Decl) bool { return false }
func (x Float64Val) CanEval([]base.Decl) bool { return false }
func (x StringVal) CanEval([]base.Decl) bool  { return false }

func (x BoolVal) String() string {
	return chavetize("BoolVal", strconv.FormatBool(x.val))
}

func (x Int32Val) String() string {
	return chavetize("Int32Val", strconv.FormatInt(int64(x.val), 10))
}

func (x Int64Val) String() string {
	return chavetize("Int64Val", strconv.FormatInt(x.val, 10))

}
func (x Float32Val) String() string {
	return chavetize("Float32Val", strconv.FormatFloat(float64(x.val), 'E', -1, 32))
}

func (x Float64Val) String() string {
	return chavetize("Float64Val", strconv.FormatFloat(x.val, 'E', -1, 64))
}

func (x StringVal) String() string {
	return chavetize("StringVal", "\""+x.val+"\"")
}

func (x BoolVal) ToGoString([]base.Decl) string    { return x.String() }
func (x Int32Val) ToGoString([]base.Decl) string   { return x.String() }
func (x Int64Val) ToGoString([]base.Decl) string   { return x.String() }
func (x Float32Val) ToGoString([]base.Decl) string { return x.String() }
func (x Float64Val) ToGoString([]base.Decl) string { return x.String() }
func (x StringVal) ToGoString([]base.Decl) string  { return x.String() }

/******************************************************************************/
/* Helpers */

func newIntLit(lit string) (PrimitiveLiteral, bool) {
	if i, err := strconv.ParseInt(lit, 10, 32); err == nil {
		return PrimitiveLiteral{i, INT32}, true
	}
	if i, err := strconv.ParseInt(lit, 10, 64); err == nil {
		return PrimitiveLiteral{i, INT64}, true
	}
	return PrimitiveLiteral{}, false
}

func newFloatLit(lit string) (PrimitiveLiteral, bool) {
	if f, err := strconv.ParseFloat(lit, 32); err == nil {
		return PrimitiveLiteral{f, FLOAT32}, true
	}
	if f, err := strconv.ParseFloat(lit, 64); err == nil {
		return PrimitiveLiteral{f, FLOAT64}, true
	}
	return PrimitiveLiteral{}, false
}

// checks if the fractional part of the argument only contains zeros
func hasNoFractionalPart(x string) bool {
	var zerosFractional = regexp.MustCompile(`^[-+]?\d*\.0*$`)
	return zerosFractional.FindString(x) != ""
}

// extracts the leading integer from the fractional number represented by
// the string x. E.g. truncateFractionalPart("42.0000") == "42"
// Pre: hasNoFractionalPart(x)
func truncateFractionalPart(x string) string {
	var leadingInt = regexp.MustCompile(`^[-+]?\d*`)
	return leadingInt.FindString(x)
}

func maxTag(t1, t2 Tag) Tag {
	if t1 > t2 {
		return t1
	}
	return t2
}

/* valueFromLiteral */

func valueFromLiteral(lit PrimitiveLiteral, decldType TPrimitive) PrimtValue {
	switch decldType.Tag() {
	case INT32:
		return makeInt32Val(lit)
	case INT64:
		return makeInt64Val(lit)
	case FLOAT32:
		return makeFloat32Val(lit)
	case FLOAT64:
		return makeFloat64Val(lit)
	case STRING:
		return makeStringVal(lit)
	}
	panic("Unexpected declared type for " + lit.String() + ": " + decldType.String())
}

/* Accessors -- return underlying value of a FGGExpr */

func makeBoolVal(expr FGGExpr) BoolVal {
	switch e := expr.(type) {
	case BoolVal:
		return e
	case PrimitiveLiteral:
		return BoolVal{e.payload.(bool)}
	}
	panic("Expr " + expr.String() + " is not a bool")
}

func makeInt32Val(expr FGGExpr) Int32Val {
	switch e := expr.(type) {
	case Int32Val:
		return e
	case PrimitiveLiteral:
		return Int32Val{int32(e.payload.(int64))}
	}
	panic("Expr " + expr.String() + " is not an int32")
}

func makeInt64Val(expr FGGExpr) Int64Val {
	switch e := expr.(type) {
	case Int64Val:
		return e
	case PrimitiveLiteral:
		return Int64Val{e.payload.(int64)}
	}
	panic("Expr " + expr.String() + " is not an int64")
}

func makeFloat32Val(expr FGGExpr) Float32Val {
	switch e := expr.(type) {
	case Float32Val:
		return e
	case PrimitiveLiteral:
		switch p := e.payload.(type) {
		case int64:
			return Float32Val{float32(p)}
		case float64:
			return Float32Val{float32(p)}
		}
	}
	panic("Expr " + expr.String() + " is not a float32")
}

func makeFloat64Val(expr FGGExpr) Float64Val {
	switch e := expr.(type) {
	case Float64Val:
		return e
	case PrimitiveLiteral:
		switch p := e.payload.(type) {
		case int64:
			return Float64Val{float64(p)}
		case float64:
			return Float64Val{p}
		}
	}
	panic("Expr " + expr.String() + " is not a float64")
}

func makeStringVal(expr FGGExpr) StringVal {
	switch e := expr.(type) {
	case StringVal:
		return e
	case PrimitiveLiteral:
		return StringVal{e.payload.(string)}
	}
	panic("Expr " + expr.String() + " is not a string")
}

func makeNamedPrimtLiteral(expr FGGExpr, typ TNamed) NamedPrimitiveLiteral {
	switch e := expr.(type) {
	case NamedPrimitiveLiteral:
		return e
	case PrimitiveLiteral:
		return NamedPrimitiveLiteral{e, typ}
	}
	panic("Expr " + expr.String() + " is not a NamedPrimitiveLiteral")
}

/* Predicates */

type PrimtPredicate func(TPrimitive) bool

func evalPrimtPredicate(ds []Decl, t Type, predicate PrimtPredicate) bool {
	under := t.Underlying(ds)
	if t_P, ok := under.(TPrimitive); ok {
		return predicate(t_P)
	}
	//else if t_I, ok := under.(ITypeLit); ok {
	// test each type in type list
	//}
	return false
}

func isBoolean(ds []Decl, t Type) bool {
	pred := func(t_P TPrimitive) bool { return t_P.Tag() == BOOL }
	return evalPrimtPredicate(ds, t, pred)
}

// TODO maybe the predicates could be directly associated with each type, instead
//  of enumerating all the matching types here
func isNumeric(ds []Decl, t Type) bool {
	pred := func(t_P TPrimitive) bool {
		tag := t_P.Tag()
		return tag == INT32 || tag == INT64 || tag == FLOAT32 || tag == FLOAT64
	}
	return evalPrimtPredicate(ds, t, pred)
}

func isString(ds []Decl, t Type) bool {
	pred := func(t_P TPrimitive) bool { return t_P.Tag() == STRING }
	return evalPrimtPredicate(ds, t, pred)
}

func isComparable(ds []Decl, t Type) bool {
	// TODO just tests that t's underlying is a primitive
	pred := func(_ TPrimitive) bool { return true }
	return evalPrimtPredicate(ds, t, pred)
}

/* Strings */

// similar to parenthesize but inserts a chaveta
func chavetize(tname, body string) string {
	var sb strings.Builder
	sb.WriteString(tname)
	sb.WriteString("{ ")
	sb.WriteString(body)
	sb.WriteString(" }")
	return sb.String()
}
