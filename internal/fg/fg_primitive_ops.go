package fg

import "github.com/rhu1/fgg/internal/base"

const (
	// arithmetic
	PLUS  = "+"
	MINUS = "-"
	// logical
	AND = "&&"
	OR  = "||"
	// relational
	GT = ">"
	LT = "<"
)

/* "Exported" constructors */

func NewBinaryOp(left, right FGExpr, op string) FGExpr {
	switch op {
	case PLUS:
		return Add{left, right}
	case MINUS:
		//a.push(fg.NewSub(l,r))
	case AND:
	case OR:
	case GT:
	case LT:
	}
	panic("Unknown binary operation")
}

/* Primitive operations - arithmetic */

type Add struct {
	left, right FGExpr
}

var _ FGExpr = Add{}

func (a Add) Subs(subs map[Variable]FGExpr) FGExpr {
	return Add{a.left.Subs(subs), a.right.Subs(subs)}
}

func (a Add) Eval(ds []Decl) (FGExpr, string) {
	if !a.left.IsValue() {
		e, rule := a.left.Eval(ds)
		return Add{e, a.right}, rule
	}
	if !a.right.IsValue() {
		e, rule := a.right.Eval(ds)
		return Add{a.left, e}, rule
	}

	left, right := match(a.left, a.right)

	var result FGExpr
	switch l := left.(type) {
	case PrimitiveLiteral: // both are PrimitiveLiterals
		switch l.tag {
		case INT32, INT64:
			ll := l.payload.(int64)
			rr := right.(PrimitiveLiteral).payload.(int64)
			result = PrimitiveLiteral{ll + rr, l.tag}
		case FLOAT32, FLOAT64:
			ll := l.payload.(float64)
			rr := right.(PrimitiveLiteral).payload.(float64)
			result = PrimitiveLiteral{ll + rr, l.tag}
		}
	case Int32Val:
		result = Int32Val{l.val + right.(Int32Val).val}
	case Int64Val:
		result = Int64Val{l.val + right.(Int64Val).val}
	case Float32Val:
		result = Float32Val{l.val + right.(Float32Val).val}
	case Float64Val:
		result = Float64Val{l.val + right.(Float64Val).val}
		//case StringVal:
		//	result = Float64Val{l.val + a.right.(Float64Val).val}
		//default:
		//	panic()
	}
	return result, "Add"
}

// Returns the matching representation for x and y.
// If either x or y already has a concrete representation(type), returns both
// as having that representation -- e.g. match(Int32Val, PrimitiveLiteral) = (Int32Val, Int32Val)
// If both x and y have undefined representation (both PrimitiveLiteral's),
// returns the one with the 'highest' tag.
//
// Pre: x and y can be matched (match can only be invoked during Eval, i.e. after Typing)
func match(x, y FGExpr) (interface{}, interface{}) {
	switch xx := x.(type) {
	case PrimitiveLiteral:
		if yy, ok := y.(PrimitiveLiteral); ok {
			t := maxTag(xx.tag, yy.tag)

			if t == INT32 || t == INT64 {
				return PrimitiveLiteral{xx.payload, t},
					PrimitiveLiteral{yy.payload, t}
			}
			if t == FLOAT32 || t == FLOAT64 {
				return PrimitiveLiteral{anyToFloat64(xx.payload), t},
					PrimitiveLiteral{anyToFloat64(yy.payload), t}
			}
		} else {
			// invert -- will fall into one of the cases below
			y, x := match(yy, xx)
			return x, y
		}
	case Int32Val:
		return xx, exprToInt32Val(y)
	case Int64Val:
		return xx, exprToInt64Val(y)
	case Float32Val:
		return xx, exprToFloat32Val(y)
	case Float64Val:
		return xx, exprToFloat64Val(y)
		//case StringVal:
		//	result = Float64Val{xx.val + a.right.(Float64Val).val}
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
		panic("anyToFloat64: unsupported type masked as Any")
	}
}

func (a Add) Typing(ds []Decl, gamma Gamma, allowStupid bool) Type {
	lt := a.left.Typing(ds, gamma, allowStupid)
	rt := a.left.Typing(ds, gamma, allowStupid)
	if !isPrimitiveType(lt) || !isPrimitiveType(rt) {
		panic("+: operands are not primitive types")
	}
	ltype := lt.(TPrimitive)
	rtype := rt.(TPrimitive)

	// TODO verificar q left é ou um tipo numérico ou string
	//if !isNumeric(ltype) && !isString(ltype) {
	//	panic()
	//}

	if ltype.Impls(ds, rtype) {
		return rtype
	} else if rtype.Impls(ds, ltype) {
		return ltype
	} else {
		panic("mismatched types " + ltype.String() + " and " + rtype.String())
	}
}

func (a Add) IsValue() bool {
	return false
}

func (a Add) CanEval([]base.Decl) bool {
	return true
}

func (a Add) String() string {
	return a.left.String() + " + " + a.right.String()
}

func (a Add) ToGoString(ds []base.Decl) string {
	return a.left.ToGoString(ds) + " + " + a.right.ToGoString(ds)
}
