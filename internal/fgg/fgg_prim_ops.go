package fgg

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

func NewBinaryOp(left, right FGGExpr, op Operator) FGGExpr {
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
	left, right FGGExpr
	op          Operator
}

func (b BaseBinaryOperation) IsValue() bool { return false }

func (b BaseBinaryOperation) CanEval(ds []base.Decl) bool {
	leftOk := b.left.IsValue() || b.left.CanEval(ds)
	rightOk := b.right.IsValue() || b.right.CanEval(ds)
	return leftOk && rightOk
}

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

func (b BaseBinaryOperation) Subs(subs map[Variable]FGGExpr) FGGExpr {
	return NewBinaryOp(b.left.Subs(subs), b.right.Subs(subs), b.op)
}

func (b BaseBinaryOperation) TSubs(subs EtaOpen) FGGExpr {
	return NewBinaryOp(b.left.TSubs(subs), b.right.TSubs(subs), b.op)
}

/**********************************************************************************/

// Operations whose output's type is equal (or appropriately related)
// to the type of the inputs
type BinaryOperation struct {
	//left, right FGGExpr
	//op          Operator
	BaseBinaryOperation
}

var _ FGGExpr = BinaryOperation{}

func (b BinaryOperation) Eval(ds []Decl) (FGGExpr, string) {
	if !b.left.IsValue() {
		e, rule := b.left.Eval(ds)
		return NewBinaryOp(e, b.right, b.op), rule
	}
	if !b.right.IsValue() {
		e, rule := b.right.Eval(ds)
		return NewBinaryOp(b.left, e, b.op), rule
	}

	left := b.left.(PrimtValue)
	right := b.right.(PrimtValue)
	rawRes := rawBinop(left.Val(), right.Val(), b.op)

	switch left := left.(type) {
	case PrimitiveLiteral:
		return PrimitiveLiteral{rawRes, left.tag}, OpToRule[b.op]

	case TypedPrimitiveValue:
		primLit := PrimitiveLiteral{rawRes, left.lit.tag}
		return TypedPrimitiveValue{primLit, left.typ}, OpToRule[b.op]
	}
	panic("Unsupported binary operation: " +
		b.left.String() + " " + string(b.op) + " " + b.right.String())
}

func (b BinaryOperation) Typing(ds []Decl, delta Delta, gamma Gamma, allowStupid bool) (Type, FGGExpr) {
	ltype, ltree := b.left.Typing(ds, delta, gamma, allowStupid)
	rtype, rtree := b.right.Typing(ds, delta, gamma, allowStupid)

	var pred PrimtPredicate
	switch b.op {
	case ADD:
		pred = Or(isNumeric, isString)
	case SUB:
		pred = isNumeric
	case LAND, LOR:
		pred = isBool
	}
	if ok := evalPrimtPredicate(ds, delta, pred, ltype); !ok {
		panic("operator " + string(b.op) + " not defined for type: " + ltype.String())
	}
	// also check if op defined for rtype?
	if ok := evalPrimtPredicate(ds, delta, pred, rtype); !ok {
		panic("operator " + string(b.op) + " not defined for type: " + rtype.String())
	}

	// verify that ltype and rtype are compatible;
	// if they are, return the most general type
	if ok, coercion := ltype.AssignableToDelta(ds, delta, rtype); ok {
		return rtype, NewBinaryOp(coercion(ltree), rtree, b.op)
	}
	if ok, coercion := rtype.AssignableToDelta(ds, delta, ltype); ok {
		return ltype, NewBinaryOp(ltree, coercion(rtree), b.op)
	}
	panic("mismatched types " + ltype.String() + " and " + rtype.String())

}

// Different from "pure" BinaryOperation -- output is always boolean.
type Comparison struct {
	//left, right FGGExpr
	//op          Operator
	BaseBinaryOperation
}

var _ FGGExpr = Comparison{}

func (c Comparison) Eval(ds []Decl) (FGGExpr, string) {
	if !c.left.IsValue() {
		e, rule := c.left.Eval(ds)
		return NewBinaryOp(e, c.right, c.op), rule
	}
	if !c.right.IsValue() {
		e, rule := c.right.Eval(ds)
		return NewBinaryOp(c.left, e, c.op), rule
	}

	left := c.left.(PrimtValue)
	right := c.right.(PrimtValue)
	res := rawBinop(left.Val(), right.Val(), c.op).(bool)
	return PrimitiveLiteral{res, BOOL}, OpToRule[c.op] // according to the spec, the result of a comparison is an "untyped" boolean
}

func (c Comparison) Typing(ds []Decl, delta Delta, gamma Gamma, allowStupid bool) (Type, FGGExpr) {
	ltype, ltree := c.left.Typing(ds, delta, gamma, allowStupid)
	rtype, rtree := c.right.Typing(ds, delta, gamma, allowStupid)

	if ok := evalPrimtPredicate(ds, delta, isComparable, ltype); !ok {
		panic("operator " + string(c.op) + " not defined for type: " + ltype.String())
	}
	if ok := evalPrimtPredicate(ds, delta, isComparable, rtype); !ok {
		panic("operator " + string(c.op) + " not defined for type: " + rtype.String())
	}

	var newTree FGGExpr

	if ok, coercion := ltype.AssignableToDelta(ds, delta, rtype); ok {
		newTree = NewBinaryOp(coercion(ltree), rtree, c.op)
	} else if ok, coercion := rtype.AssignableToDelta(ds, delta, ltype); ok {
		newTree = NewBinaryOp(ltree, coercion(rtree), c.op)
	} else {
		panic("mismatched types " + ltype.String() + " and " + rtype.String())
	}

	return NewUndefTPrimitive(BOOL), newTree // according to the spec, the result of a comparison is an "untyped" boolean
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

