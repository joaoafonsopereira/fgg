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

func NewTypedPrimitiveValue(lit PrimitiveLiteral, typ Type) TypedPrimitiveValue {
	return TypedPrimitiveValue{lit, typ}
}

/* Remaining "exported" constructors */

func NewBoolLit(lit string) PrimitiveLiteral {
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

/******************************************************************************/
/* PrimtValue - base interface for primitive values */

type PrimtValue interface {
	FGExpr
	Val() interface{}
}

func (x PrimitiveLiteral) Val() interface{}      { return x.payload }
func (x TypedPrimitiveValue) Val() interface{} { return x.lit.payload }

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

func (x PrimitiveLiteral) Subs(map[Variable]FGExpr) FGExpr {
	return x
}

func (x PrimitiveLiteral) Eval([]Decl) (FGExpr, string) {
	panic("Cannot reduce: " + x.String())
}

func (x PrimitiveLiteral) Typing([]Decl, Gamma, bool) (Type, FGExpr) {
	return NewUndefTPrimitive(x.tag), x
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
		payload = "\"" + p + "\""
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
/* TypedPrimitiveValue */

// Essentially a PrimitiveLiteral whose type was already determined.
// Need this in order not to break type-safety at each (small) step of evaluation.
type TypedPrimitiveValue struct {
	lit PrimitiveLiteral
	typ Type
}

var _ FGExpr = TypedPrimitiveValue{}

func (t TypedPrimitiveValue) Subs(subs map[Variable]FGExpr) FGExpr {
	return t
}

func (t TypedPrimitiveValue) Eval(ds []Decl) (FGExpr, string) {
	panic("Cannot reduce: " + t.String())
}

func (t TypedPrimitiveValue) Typing(ds []Decl, gamma Gamma, allowStupid bool) (Type, FGExpr) {
	return t.typ, t
}

func (t TypedPrimitiveValue) IsValue() bool {
	return true
}

func (t TypedPrimitiveValue) CanEval(ds []base.Decl) bool {
	return false
}

func (t TypedPrimitiveValue) String() string {
	var b strings.Builder
	b.WriteString(t.typ.String())
	b.WriteString("Val(")
	b.WriteString(t.lit.String())
	b.WriteString(")")
	return b.String()
}

func (t TypedPrimitiveValue) ToGoString(ds []base.Decl) string {
	return t.String()
}

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


/* Predicates */

type PrimtPredicate func(PrimType) bool

var (
	isBool       = func(t_P PrimType) bool { return t_P.Tag() == BOOL }
	isString     = func(t_P PrimType) bool { return t_P.Tag() == STRING }
	isInt        = func(t_P PrimType) bool { return t_P.Tag() == INT32 || t_P.Tag() == INT64 }
	isFloat      = func(t_P PrimType) bool { return t_P.Tag() == FLOAT32 || t_P.Tag() == FLOAT64 }
	isNumeric    = Or(isInt, isFloat)
	isComparable = func(_ PrimType) bool { return true } // enough to be a TPrimitive (underlying) ??
)

func Or(pred1, pred2 PrimtPredicate) PrimtPredicate {
	return func(t_P PrimType) bool { return pred1(t_P) || pred2(t_P) }
}

func evalPrimtPredicate(ds []Decl, predicate PrimtPredicate, t Type) bool {
	under := t.Underlying(ds)
	if t_P, ok := under.(PrimType); ok {
		return predicate(t_P)
	}
	return false
}

