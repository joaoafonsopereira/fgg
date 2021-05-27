package fg

import (
	"github.com/rhu1/fgg/internal/base"
	"reflect"
	"strings"
)

type Operator string

const (
	// arithmetic
	ADD = Operator("+")
	SUB = Operator("-")
	// logical
	LAND = Operator("&&")
	LOR  = Operator("||")
	// relational
	GT = Operator(">")
	LT = Operator("<")
)

var OpToRule = map[Operator]string{
	ADD:  "Add",
	SUB:  "Sub",
	LAND: "LAnd",
	LOR:  "LOr",
	GT:   "Gt",
	LT:   "Lt",
}

/* "Exported" constructors */

func NewBinaryOp(left, right FGExpr, op Operator) FGExpr {
	baseBop := BaseBinaryOperation{left, right, op}
	switch op {
	case GT, LT:
		return Comparison{baseBop}
	default:
		return BinaryOperation{baseBop}
	}
}

// "Base class" for binary operations. Provides the methods common to
// arithmetic, logical and relational operations.
type BaseBinaryOperation struct {
	left, right FGExpr
	op          Operator
}

func (b BaseBinaryOperation) IsValue() bool { return false }

func (b BaseBinaryOperation) CanEval([]base.Decl) bool { return true }

func (b BaseBinaryOperation) String() string {
	var sb strings.Builder
	sb.WriteString(b.left.String())
	sb.WriteString(" ")
	sb.WriteString(string(b.op))
	sb.WriteString(" ")
	sb.WriteString(b.right.String())
	return sb.String()
}

func (b BaseBinaryOperation) ToGoString(ds []base.Decl) string {
	var sb strings.Builder
	sb.WriteString(b.left.ToGoString(ds))
	sb.WriteString(" ")
	sb.WriteString(string(b.op))
	sb.WriteString(" ")
	sb.WriteString(b.right.ToGoString(ds))
	return sb.String()
}

func (b BaseBinaryOperation) Subs(subs map[Variable]FGExpr) FGExpr {
	return NewBinaryOp(b.left.Subs(subs), b.right.Subs(subs), b.op)
}

/**********************************************************************************/

// Operations whose output's type is equal (or appropriately related)
// to the type of the inputs
type BinaryOperation struct {
	//left, right FGExpr
	//op          Operator
	BaseBinaryOperation
}

var _ FGExpr = BinaryOperation{}

func (b BinaryOperation) Eval(ds []Decl) (FGExpr, string) {
	if !b.left.IsValue() {
		e, rule := b.left.Eval(ds)
		return NewBinaryOp(e, b.right, b.op), rule
	}
	if !b.right.IsValue() {
		e, rule := b.right.Eval(ds)
		return NewBinaryOp(b.left, e, b.op), rule
	}

	left, right := match(b.left, b.right)

	rawRes := rawBinop(left.Val(), right.Val(), b.op)

	switch left := left.(type) {
	case PrimitiveLiteral:
		return PrimitiveLiteral{rawRes, left.tag}, OpToRule[b.op]

	case NamedPrimitiveLiteral:
		primLit := PrimitiveLiteral{rawRes, left.tag}
		return NamedPrimitiveLiteral{primLit, left.typ}, OpToRule[b.op]

	case BoolVal:
		return BoolVal{rawRes.(bool)}, OpToRule[b.op]

	case Int32Val:
		return Int32Val{rawRes.(int32)}, OpToRule[b.op]

	case Int64Val:
		return Int64Val{rawRes.(int64)}, OpToRule[b.op]

	case Float32Val:
		return Float32Val{rawRes.(float32)}, OpToRule[b.op]

	case Float64Val:
		return Float64Val{rawRes.(float64)}, OpToRule[b.op]

	case StringVal:
		return StringVal{rawRes.(string)}, OpToRule[b.op]
	}
	panic("Unsupported binary operation: " +
		b.left.String() + " " + string(b.op) + " " + b.right.String())
}

func (b BinaryOperation) Typing(ds []Decl, gamma Gamma, allowStupid bool) (Type, FGExpr) {
	ltype, ltree := b.left.Typing(ds, gamma, allowStupid)
	rtype, rtree := b.right.Typing(ds, gamma, allowStupid)

	// enough to verify ltype -- if rtype is a 'wrong' type, it will not pass
	// any of the Impls tests below
	switch b.op {
	case ADD:
		if !isNumeric(ds, ltype) && !isString(ds, ltype) {
			panic("Add doesn't support type: " + ltype.String())
		}
	case SUB:
		if !isNumeric(ds, ltype) {
			panic("Sub doesn't support type: " + ltype.String())
		}
	case LAND, LOR:
		if !isBoolean(ds, ltype) {
			// TODO replace by string(b.op)
			panic("LAND/LOR doesn't support type: " + ltype.String())
		}
	}

	newTree := NewBinaryOp(ltree, rtree, b.op)

	// verify that ltype and rtype are compatible;
	// if they are, return the most general type
	if ltype.Impls(ds, rtype) {
		return rtype, newTree
	}
	if rtype.Impls(ds, ltype) {
		return ltype, newTree
	}
	panic("mismatched types " + ltype.String() + " and " + rtype.String())

}

// Different from "pure" BinaryOperation -- output is always boolean.
type Comparison struct {
	//left, right FGExpr
	//op          Operator
	BaseBinaryOperation
}

var _ FGExpr = Comparison{}

func (c Comparison) Eval(ds []Decl) (FGExpr, string) {
	if !c.left.IsValue() {
		e, rule := c.left.Eval(ds)
		return NewBinaryOp(e, c.right, c.op), rule
	}
	if !c.right.IsValue() {
		e, rule := c.right.Eval(ds)
		return NewBinaryOp(c.left, e, c.op), rule
	}

	left, right := match(c.left, c.right) // panics if not able to match
	res := rawBinop(left.Val(), right.Val(), c.op).(bool)
	return BoolVal{res}, OpToRule[c.op]
}

func (c Comparison) Typing(ds []Decl, gamma Gamma, allowStupid bool) (Type, FGExpr) {
	ltype, ltree := c.left.Typing(ds, gamma, allowStupid)
	rtype, rtree := c.right.Typing(ds, gamma, allowStupid)

	// enough to verify ltype -- if rtype is a 'wrong' type, it will not pass
	// the Impls tests below
	if !isComparable(ds, ltype) {
		panic("GT/LT doesn't support type: " + ltype.String())
	}
	if !ltype.Impls(ds, rtype) && !rtype.Impls(ds, ltype) {
		panic("mismatched types " + ltype.String() + " and " + rtype.String())
	}

	return TPrimitive{tag: BOOL}, NewBinaryOp(ltree, rtree, c.op)
}

/* Helpers */

func rawBinop(left, right interface{}, op Operator) interface{} {
	switch lval := left.(type) {
	case bool:
		rval := right.(bool)
		switch op {
		case LAND:
			return lval && rval
		case LOR:
			return lval || rval
		}
	case int32:
		rval := right.(int32)
		switch op {
		case ADD:
			return lval + rval
		case SUB:
			return lval - rval
		case GT:
			return lval > rval
		case LT:
			return lval < rval
		}
	case int64:
		rval := right.(int64)
		switch op {
		case ADD:
			return lval + rval
		case SUB:
			return lval - rval
		case GT:
			return lval > rval
		case LT:
			return lval < rval
		}
	case float32:
		rval := right.(float32)
		switch op {
		case ADD:
			return lval + rval
		case SUB:
			return lval - rval
		case GT:
			return lval > rval
		case LT:
			return lval < rval
		}
	case float64:
		rval := right.(float64)
		switch op {
		case ADD:
			return lval + rval
		case SUB:
			return lval - rval
		case GT:
			return lval > rval
		case LT:
			return lval < rval
		}
	case string:
		rval := right.(string)
		switch op {
		case ADD:
			return lval + rval
		case GT:
			return lval > rval
		case LT:
			return lval < rval
		}
	}
	panic("Unsupported raw binOp: " + string(op) +
		" for type: " + reflect.TypeOf(left).String())
}

// Returns the matching representation for x and y.
// If either x or y already has a concrete representation(type), returns both
// as having that representation -- e.g. match(Int32Val, PrimitiveLiteral) = (Int32Val, Int32Val)
// If both x and y have undefined representation (both PrimitiveLiteral's),
// returns the one with the 'highest' tag.
//
// Pre: x and y are compatible (match can only be invoked during Eval, i.e. after Typing)
func match(x, y FGExpr) (PrimtValue, PrimtValue) {
	switch xx := x.(type) {
	case PrimitiveLiteral:
		if yy, ok := y.(PrimitiveLiteral); ok {
			t := maxTag(xx.tag, yy.tag)
			if t == FLOAT32 || t == FLOAT64 {
				return PrimitiveLiteral{anyToFloat64(xx.payload), t},
					PrimitiveLiteral{anyToFloat64(yy.payload), t}
			} else {
				return PrimitiveLiteral{xx.payload, t},
					PrimitiveLiteral{yy.payload, t}
			}
		} else {
			// invert -- will fall into one of the cases below
			y, x := match(yy, xx)
			return x, y
		}

	case BoolVal:
		return xx, makeBoolVal(y)
	case Int32Val:
		return xx, makeInt32Val(y)
	case Int64Val:
		return xx, makeInt64Val(y)
	case Float32Val:
		return xx, makeFloat32Val(y)
	case Float64Val:
		return xx, makeFloat64Val(y)
	case StringVal:
		return xx, makeStringVal(y)
	case NamedPrimitiveLiteral:
		return xx, makeNamedPrimtLiteral(y, xx.typ)
	}

	panic("Can't match " + x.String() + " with " + y.String())
}

func anyToFloat64(x interface{}) float64 {
	switch xx := x.(type) {
	case int64:
		return float64(xx)
	case float64:
		return xx
	default:
		panic("anyToFloat64: unsupported type " + reflect.TypeOf(x).String())
	}
}
