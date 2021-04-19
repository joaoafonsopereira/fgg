package fg

import (
	"github.com/rhu1/fgg/internal/base"
	"strconv"
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
)

var NamesToTags = map[string]Tag{
	"int32": INT32,
	"int64": INT64,
}

var TagsToNames = map[Tag]string{
	INT32: "int32",
	INT64: "int64",
}

func TagFromName(name string) Tag {
	//tag, ok := NamesToTags[name]
	return NamesToTags[name]
}

func NameFromTag(tag Tag) string {
	return TagsToNames[tag]
}

/* "Exported" constructors (( ? for fgg (monomorph) ? )) */

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
	if i, ok := newIntLit(lit); ok {
		return i // CHECKME(jp): maybe assign tag INT_OR_FLOAT in fgg (for type inference -- e.g. for 13.0, infer int or float?)
	} else if f, ok := newFloatLit(lit); ok {
		return f
	} else {
		panic("Float const too big??")
	}
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

func (b PrimitiveLiteral) Typing([]Decl, Gamma, bool) Type {
	return TPrimitive{tag: b.tag, undefined: true}
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
	panic("implement me PrimitiveLiteral.ToGoString")
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

var _ FGExpr = BoolVal{}
var _ FGExpr = Int32Val{}
var _ FGExpr = Int64Val{}
var _ FGExpr = Float32Val{}
var _ FGExpr = Float64Val{}

func (x BoolVal) GetValue() bool       { return x.val }
func (x Int32Val) GetValue() int32     { return x.val }
func (x Int64Val) GetValue() int64     { return x.val }
func (x Float32Val) GetValue() float32 { return x.val }
func (x Float64Val) GetValue() float64 { return x.val }

func (x BoolVal) Subs(map[Variable]FGExpr) FGExpr    { return x }
func (x Int32Val) Subs(map[Variable]FGExpr) FGExpr   { return x }
func (x Int64Val) Subs(map[Variable]FGExpr) FGExpr   { return x }
func (x Float32Val) Subs(map[Variable]FGExpr) FGExpr { return x }
func (x Float64Val) Subs(map[Variable]FGExpr) FGExpr { return x }

func (x BoolVal) Eval([]Decl) (FGExpr, string)    { panic("Cannot reduce: " + x.String()) }
func (x Int32Val) Eval([]Decl) (FGExpr, string)   { panic("Cannot reduce: " + x.String()) }
func (x Int64Val) Eval([]Decl) (FGExpr, string)   { panic("Cannot reduce: " + x.String()) }
func (x Float32Val) Eval([]Decl) (FGExpr, string) { panic("Cannot reduce: " + x.String()) }
func (x Float64Val) Eval([]Decl) (FGExpr, string) { panic("Cannot reduce: " + x.String()) }

func (x BoolVal) Typing([]Decl, Gamma, bool) Type    { return TPrimitive{tag: BOOL} }
func (x Int32Val) Typing([]Decl, Gamma, bool) Type   { return TPrimitive{tag: INT32} }
func (x Int64Val) Typing([]Decl, Gamma, bool) Type   { return TPrimitive{tag: INT64} }
func (x Float32Val) Typing([]Decl, Gamma, bool) Type { return TPrimitive{tag: FLOAT32} }
func (x Float64Val) Typing([]Decl, Gamma, bool) Type { return TPrimitive{tag: FLOAT64} }

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

func (x BoolVal) String() string    { return strconv.FormatBool(x.val) }
func (x Int32Val) String() string   { return strconv.FormatInt(int64(x.val), 10) }
func (x Int64Val) String() string   { return strconv.FormatInt(x.val, 10) }
func (x Float32Val) String() string { return strconv.FormatFloat(float64(x.val), 'E', -1, 32) }
func (x Float64Val) String() string { return strconv.FormatFloat(x.val, 'E', -1, 64) }

func (x BoolVal) ToGoString([]base.Decl) string    { return x.String() }
func (x Int32Val) ToGoString([]base.Decl) string   { return x.String() }
func (x Int64Val) ToGoString([]base.Decl) string   { return x.String() }
func (x Float32Val) ToGoString([]base.Decl) string { return x.String() }
func (x Float64Val) ToGoString([]base.Decl) string { return x.String() }

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

func maxTag(t1, t2 Tag) Tag {
	if t1 > t2 {
		return t1
	}
	return t2
}

/* Predicates */

func isNumeric(t TPrimitive) bool {
	switch t.Tag() {
	case INT32, INT64, FLOAT32, FLOAT64:
		return true
	}
	return false
}

//func isString(t TPrimitive) bool {
//	return t.Tag() == STRING
//}

/* Accessors -- return underlying value of a FGExpr */

func exprToInt32Val(expr FGExpr) Int32Val {
	switch e := expr.(type) {
	case Int32Val:
		return e
	case PrimitiveLiteral:
		return Int32Val{int32(e.payload.(int64))}
	}
	panic("Expr is not an int32")
}

func exprToInt64Val(expr FGExpr) Int64Val {
	switch e := expr.(type) {
	case Int64Val:
		return e
	case PrimitiveLiteral:
		return Int64Val{e.payload.(int64)}
	}
	panic("Expr is not an int64")
}

func exprToFloat32Val(expr FGExpr) Float32Val {
	switch e := expr.(type) {
	case Float32Val:
		return e
	case PrimitiveLiteral:
		return Float32Val{float32(e.payload.(float64))}
	}
	panic("Expr is not a float32")
}

func exprToFloat64Val(expr FGExpr) Float64Val {
	switch e := expr.(type) {
	case Float64Val:
		return e
	case PrimitiveLiteral:
		return Float64Val{e.payload.(float64)}
	}
	panic("Expr is not a float64")
}

//func exprToString(expr FGExpr) string {
//	switch e := expr.(type) {
//	case StringVal:
//		return e.val
//	case PrimitiveLiteral:
//		return e.payload.(string)
//	}
//	panic("Expr is not a string")
//}
