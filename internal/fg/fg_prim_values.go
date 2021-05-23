package fg

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

/* "Exported" constructors for fgg (monomorph) */

func NewPrimitiveLiteral(pload interface{}, tag Tag) PrimitiveLiteral {
	return PrimitiveLiteral{pload, tag}
}
func NewBoolVal(val bool) BoolVal          { return BoolVal{val} }
func NewInt32Val(val int32) Int32Val       { return Int32Val{val} }
func NewInt64Val(val int64) Int64Val       { return Int64Val{val} }
func NewFloat32Val(val float32) Float32Val { return Float32Val{val} }
func NewFloat64Val(val float64) Float64Val { return Float64Val{val} }
func NewStringVal(val string) StringVal    { return StringVal{val} }

/* Remaining "exported" constructors */

func NewBool(lit string) BoolVal {
	b, _ := strconv.ParseBool(lit)
	return BoolVal{b}
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

func NewStringLit(lit string) StringVal {
	trim := strings.ReplaceAll(lit, "\"", "")
	return StringVal{trim}
}

		return makeInt64Val(n)
	case FLOAT32:
		return makeFloat32Val(n)
	case FLOAT64:
		return makeFloat64Val(n)
	}
	panic("Unexpected declared type for " + n.String() + ": " + decldType.String())
}

/******************************************************************************/
/* PrimitiveLiteral */

// Represents a literal whose type is still undefined (e.g. 123 can either be int32 or int64)
// Either way, an int/float payload is always saved as an int64/float64;
type PrimitiveLiteral struct {
	payload interface{}
	tag     Tag
}

var _ FGExpr = PrimitiveLiteral{}

func (b PrimitiveLiteral) Payload() interface{} { return b.payload }
func (b PrimitiveLiteral) Tag() Tag             { return b.tag }

func (b PrimitiveLiteral) Subs(map[Variable]FGExpr) FGExpr {
	return b
}

func (b PrimitiveLiteral) Eval([]Decl) (FGExpr, string) {
	panic("Cannot reduce: " + b.String())
}

func (b PrimitiveLiteral) Typing([]Decl, Gamma, bool) (Type, FGExpr) {
	return TPrimitive{tag: b.tag, undefined: true}, b
}

func (b PrimitiveLiteral) IsValue() bool {
	return true
}

func (b PrimitiveLiteral) CanEval([]base.Decl) bool {
	return false
}

func (b PrimitiveLiteral) String() string {
	var payload string
	switch p := b.payload.(type) {
	case int64:
		payload = strconv.FormatInt(p, 10)
	case float64:
		payload = strconv.FormatFloat(p, 'E', -1, 64)
	default:
		panic("PrimitiveLiteral.String() for unsupported type")
	}
	tag := NameFromTag(b.tag)
	return "PrimitiveLiteral{payload=" + payload + ", tag=" + tag + "}"
}

func (b PrimitiveLiteral) ToGoString(ds []base.Decl) string {
	return b.String()
}

/* Values of pre-declared (primitive) types */

type (
	// structs ou simplesmente underlying types?
	BoolVal    struct{ val bool }
	Int32Val   struct{ val int32 }
	Int64Val   struct{ val int64 }
	Float32Val struct{ val float32 }
	Float64Val struct{ val float64 }
	StringVal  struct{ val string }
)

var _ FGExpr = BoolVal{}
var _ FGExpr = Int32Val{}
var _ FGExpr = Int64Val{}
var _ FGExpr = Float32Val{}
var _ FGExpr = Float64Val{}
var _ FGExpr = StringVal{}

func (x BoolVal) GetValue() bool       { return x.val }
func (x Int32Val) GetValue() int32     { return x.val }
func (x Int64Val) GetValue() int64     { return x.val }
func (x Float32Val) GetValue() float32 { return x.val }
func (x Float64Val) GetValue() float64 { return x.val }
func (x StringVal) GetValue() string   { return x.val }

func (x BoolVal) Subs(map[Variable]FGExpr) FGExpr    { return x }
func (x Int32Val) Subs(map[Variable]FGExpr) FGExpr   { return x }
func (x Int64Val) Subs(map[Variable]FGExpr) FGExpr   { return x }
func (x Float32Val) Subs(map[Variable]FGExpr) FGExpr { return x }
func (x Float64Val) Subs(map[Variable]FGExpr) FGExpr { return x }
func (x StringVal) Subs(map[Variable]FGExpr) FGExpr  { return x }

func (x BoolVal) Eval([]Decl) (FGExpr, string)    { panic("Cannot reduce: " + x.String()) }
func (x Int32Val) Eval([]Decl) (FGExpr, string)   { panic("Cannot reduce: " + x.String()) }
func (x Int64Val) Eval([]Decl) (FGExpr, string)   { panic("Cannot reduce: " + x.String()) }
func (x Float32Val) Eval([]Decl) (FGExpr, string) { panic("Cannot reduce: " + x.String()) }
func (x Float64Val) Eval([]Decl) (FGExpr, string) { panic("Cannot reduce: " + x.String()) }
func (x StringVal) Eval([]Decl) (FGExpr, string)  { panic("Cannot reduce: " + x.String()) }

func (x BoolVal) Typing([]Decl, Gamma, bool) (Type, FGExpr)    { return TPrimitive{tag: BOOL}, x }
func (x Int32Val) Typing([]Decl, Gamma, bool) (Type, FGExpr)   { return TPrimitive{tag: INT32}, x }
func (x Int64Val) Typing([]Decl, Gamma, bool) (Type, FGExpr)   { return TPrimitive{tag: INT64}, x }
func (x Float32Val) Typing([]Decl, Gamma, bool) (Type, FGExpr) { return TPrimitive{tag: FLOAT32}, x }
func (x Float64Val) Typing([]Decl, Gamma, bool) (Type, FGExpr) { return TPrimitive{tag: FLOAT64}, x }
func (x StringVal) Typing([]Decl, Gamma, bool) (Type, FGExpr)  { return TPrimitive{tag: STRING}, x }

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

/* Accessors -- return underlying value of a FGExpr */

func makeInt32Val(expr FGExpr) Int32Val {
	switch e := expr.(type) {
	case Int32Val:
		return e
	case PrimitiveLiteral:
		return Int32Val{int32(e.payload.(int64))}
	}
	panic("Expr " + expr.String() + " is not an int32")
}

func makeInt64Val(expr FGExpr) Int64Val {
	switch e := expr.(type) {
	case Int64Val:
		return e
	case PrimitiveLiteral:
		return Int64Val{e.payload.(int64)}
	}
	panic("Expr " + expr.String() + " is not an int64")
}

func makeFloat32Val(expr FGExpr) Float32Val {
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

func makeFloat64Val(expr FGExpr) Float64Val {
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

func makeStringVal(expr FGExpr) StringVal {
	switch e := expr.(type) {
	case StringVal:
		return e
	case PrimitiveLiteral:
		return StringVal{e.payload.(string)}
	}
	panic("Expr " + expr.String() + " is not a string")
}

/* Predicates */

func isBoolean(ds []Decl, t Type) bool {
	under := t.Underlying(ds)
	if t_P, ok := under.(TPrimitive); ok {
		return t_P.Tag() == BOOL
	}
	return false
}

// TODO maybe the predicates could be directly associated with each type, instead
//  of enumerating all the matching types here
func isNumeric(ds []Decl, t Type) bool {
	under := t.Underlying(ds)
	if t_P, ok := under.(TPrimitive); ok {
		switch t_P.Tag() {
		case INT32, INT64, FLOAT32, FLOAT64:
			return true
		}
	}
	return false
}

func isString(ds []Decl, t Type) bool {
	under := t.Underlying(ds)
	if t_P, ok := under.(TPrimitive); ok {
		return t_P.Tag() == STRING
	}
	return false
}

func isComparable(ds []Decl, t Type) bool {
	// TODO
	under := t.Underlying(ds)
	_, ok := under.(TPrimitive)
	return ok
}

/* Strings */

// similar to parenthesize but inserts chavetas
func chavetize(tname, body string) string {
	var sb strings.Builder
	sb.WriteString(tname)
	sb.WriteString("{ ")
	sb.WriteString(body)
	sb.WriteString(" }")
	return sb.String()
}
