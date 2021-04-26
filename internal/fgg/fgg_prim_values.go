package fgg

import (
	"github.com/rhu1/fgg/internal/base"
	"regexp"
	"strconv"
	"strings"
)

// constants
// CHECKME(jp): maybe these "representations" will change

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
}

var TagsToNames = map[Tag]string{
	BOOL:    "bool",
	INT32:   "int32",
	INT64:   "int64",
	FLOAT32: "float32",
	FLOAT64: "float64",
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

/* "Exported" constructors (( ? for fgg (monomorph) ? )) */

func NewBool(lit string) BoolVal {
	b, _ := strconv.ParseBool(lit)
	return BoolVal{b}
}

func NewIntLit(lit string) NumericLiteral {
	if i, ok := newIntLit(lit); ok {
		return i
	} else {
		panic("Int const too big??")
	}
}

func NewFloatLit(lit string) NumericLiteral {
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

func ValueFromLiteral(n NumericLiteral, decldType TPrimitive) FGGExpr {
	switch decldType.Tag() {
	case INT32:
		return makeInt32Val(n)
	case INT64:
		return makeInt64Val(n)
	case FLOAT32:
		return makeFloat32Val(n)
	case FLOAT64:
		return makeFloat64Val(n)
	}
	panic("Unexpected declared type for " + n.String() + ": " + decldType.String())
}

/******************************************************************************/
/* NumericLiteral */

// Represents a literal whose type is still undefined (e.g. 123 can either be int32 or int64)
// Either way, an int/float payload is always saved as an int64/float64;
type NumericLiteral struct {
	payload interface{}
	tag     Tag
}

var _ FGGExpr = NumericLiteral{}

func (b NumericLiteral) Payload() interface{} { return b.payload }
func (b NumericLiteral) Tag() Tag             { return b.tag }

func (b NumericLiteral) Subs(map[Variable]FGGExpr) FGGExpr {
	return b
}

func (b NumericLiteral) TSubs(subs map[TParam]Type) FGGExpr {
	return b
}

func (b NumericLiteral) Eval([]Decl) (FGGExpr, string) {
	panic("Cannot reduce: " + b.String())
}

func (b NumericLiteral) Typing(ds []Decl, delta Delta, gamma Gamma, allowStupid bool) Type {
	return TPrimitive{tag: b.tag, undefined: true}
}

func (b NumericLiteral) IsValue() bool {
	return true
}

func (b NumericLiteral) CanEval([]base.Decl) bool {
	return false
}

func (b NumericLiteral) String() string {
	var payload string
	switch p := b.payload.(type) {
	case int64:
		payload = strconv.FormatInt(p, 10)
	case float64:
		payload = strconv.FormatFloat(p, 'E', -1, 64)
	default:
		panic("NumericLiteral.String() for unsupported type")
	}
	tag := NameFromTag(b.tag)
	return "NumericLiteral{payload=" + payload + ", tag=" + tag + "}"
}

func (b NumericLiteral) ToGoString(ds []base.Decl) string {
	panic("implement me NumericLiteral.ToGoString")
}

/* Values of pre-declared (primitive) types */

type (
	// structs ou simplesmente underlying types?
	BoolVal    struct{ val bool }
	Int32Val   struct{ val int32 }
	Int64Val   struct{ val int64 }
	Float32Val struct{ val float32 }
	Float64Val struct{ val float64 }
	// ...
)

var _ FGGExpr = BoolVal{}
var _ FGGExpr = Int32Val{}
var _ FGGExpr = Int64Val{}
var _ FGGExpr = Float32Val{}
var _ FGGExpr = Float64Val{}

func (x BoolVal) GetValue() bool       { return x.val }
func (x Int32Val) GetValue() int32     { return x.val }
func (x Int64Val) GetValue() int64     { return x.val }
func (x Float32Val) GetValue() float32 { return x.val }
func (x Float64Val) GetValue() float64 { return x.val }

func (x BoolVal) Subs(map[Variable]FGGExpr) FGGExpr    { return x }
func (x Int32Val) Subs(map[Variable]FGGExpr) FGGExpr   { return x }
func (x Int64Val) Subs(map[Variable]FGGExpr) FGGExpr   { return x }
func (x Float32Val) Subs(map[Variable]FGGExpr) FGGExpr { return x }
func (x Float64Val) Subs(map[Variable]FGGExpr) FGGExpr { return x }

func (x BoolVal) TSubs(subs map[TParam]Type) FGGExpr    { return x }
func (x Int32Val) TSubs(subs map[TParam]Type) FGGExpr   { return x }
func (x Int64Val) TSubs(subs map[TParam]Type) FGGExpr   { return x }
func (x Float32Val) TSubs(subs map[TParam]Type) FGGExpr { return x }
func (x Float64Val) TSubs(subs map[TParam]Type) FGGExpr { return x }

func (x BoolVal) Eval([]Decl) (FGGExpr, string)    { panic("Cannot reduce: " + x.String()) }
func (x Int32Val) Eval([]Decl) (FGGExpr, string)   { panic("Cannot reduce: " + x.String()) }
func (x Int64Val) Eval([]Decl) (FGGExpr, string)   { panic("Cannot reduce: " + x.String()) }
func (x Float32Val) Eval([]Decl) (FGGExpr, string) { panic("Cannot reduce: " + x.String()) }
func (x Float64Val) Eval([]Decl) (FGGExpr, string) { panic("Cannot reduce: " + x.String()) }

func (x BoolVal) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr)    { return TPrimitive{tag: BOOL}, x }
func (x Int32Val) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr)   { return TPrimitive{tag: INT32}, x }
func (x Int64Val) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr)   { return TPrimitive{tag: INT64}, x }
func (x Float32Val) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr) { return TPrimitive{tag: FLOAT32}, x }
func (x Float64Val) Typing([]Decl, Delta, Gamma, bool) (Type, FGGExpr) { return TPrimitive{tag: FLOAT64}, x }

func (x BoolVal) IsValue() bool    { return true }
func (x Int32Val) IsValue() bool   { return true }
func (x Int64Val) IsValue() bool   { return true }
func (x Float32Val) IsValue() bool { return true }
func (x Float64Val) IsValue() bool { return true }

func (x BoolVal) CanEval([]base.Decl) bool    { return false }
func (x Int32Val) CanEval([]base.Decl) bool   { return false }
func (x Int64Val) CanEval([]base.Decl) bool   { return false }
func (x Float32Val) CanEval([]base.Decl) bool { return false }
func (x Float64Val) CanEval([]base.Decl) bool { return false }

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

func (x BoolVal) ToGoString([]base.Decl) string    { return x.String() }
func (x Int32Val) ToGoString([]base.Decl) string   { return x.String() }
func (x Int64Val) ToGoString([]base.Decl) string   { return x.String() }
func (x Float32Val) ToGoString([]base.Decl) string { return x.String() }
func (x Float64Val) ToGoString([]base.Decl) string { return x.String() }

/* Helpers */

func newIntLit(lit string) (NumericLiteral, bool) {
	if i, err := strconv.ParseInt(lit, 10, 32); err == nil {
		return NumericLiteral{i, INT32}, true
	}
	if i, err := strconv.ParseInt(lit, 10, 64); err == nil {
		return NumericLiteral{i, INT64}, true
	}
	return NumericLiteral{}, false
}

func newFloatLit(lit string) (NumericLiteral, bool) {
	if f, err := strconv.ParseFloat(lit, 32); err == nil {
		return NumericLiteral{f, FLOAT32}, true
	}
	if f, err := strconv.ParseFloat(lit, 64); err == nil {
		return NumericLiteral{f, FLOAT64}, true
	}
	return NumericLiteral{}, false
}

// checks if the fractional part of the argument only contains zeros
// TODO should this be moved to the parser/adaptor??
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

/* Accessors -- return underlying value of a FGGExpr */

func makeInt32Val(expr FGGExpr) Int32Val {
	switch e := expr.(type) {
	case Int32Val:
		return e
	case NumericLiteral:
		return Int32Val{int32(e.payload.(int64))}
	}
	panic("Expr is not an int32")
}

func makeInt64Val(expr FGGExpr) Int64Val {
	switch e := expr.(type) {
	case Int64Val:
		return e
	case NumericLiteral:
		return Int64Val{e.payload.(int64)}
	}
	panic("Expr is not an int64")
}

func makeFloat32Val(expr FGGExpr) Float32Val {
	switch e := expr.(type) {
	case Float32Val:
		return e
	case NumericLiteral:
		switch p := e.payload.(type) {
		case int64:
			return Float32Val{float32(p)}
		case float64:
			return Float32Val{float32(p)}
		}
	}
	panic("Expr is not a float32")
}

func makeFloat64Val(expr FGGExpr) Float64Val {
	switch e := expr.(type) {
	case Float64Val:
		return e
	case NumericLiteral:
		switch p := e.payload.(type) {
		case int64:
			return Float64Val{float64(p)}
		case float64:
			return Float64Val{p}
		}
	}
	panic("Expr is not a float64")
}

//func exprToString(expr FGGExpr) string {
//	switch e := expr.(type) {
//	case StringVal:
//		return e.val
//	case NumericLiteral:
//		return e.payload.(string)
//	}
//	panic("Expr is not a string")
//}

/* Predicates */

func isBoolean(t Type) bool {
	if t_P, ok := t.(TPrimitive); ok {
		return t_P.Tag() == BOOL
	}
	return false
}

// TODO maybe the predicates could be directly associated with each type, instead
//  of enumerating all the matching types here
func isNumeric(t Type) bool {
	if t_P, ok := t.(TPrimitive); ok {
		switch t_P.Tag() {
		case INT32, INT64, FLOAT32, FLOAT64:
			return true
		}
	}
	return false
}

func isString(t Type) bool {
	if t_P, ok := t.(TPrimitive); ok {
		return t_P.Tag() == STRING
	}
	return false
}

func isComparable(t Type) bool {
	// TODO
	_, ok := t.(TPrimitive)
	return ok
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
