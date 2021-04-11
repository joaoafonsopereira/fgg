package fg

import (
	"reflect"
	"strconv"
)

import "github.com/rhu1/fgg/internal/base"

/* Export */

func NewTPrimitive(t Tag, undefined bool) TPrimitive       { return TPrimitive{t, undefined} }

/* Aliases from base */

type Name = base.Name
type FGNode = base.AstNode
type Decl = base.Decl

type Type base.Type //

/* Constants */

var STRING_TYPE = TNamed("string")
var PRIMITIVE_TYPES = make(map[Type]Type)

func init() {
	PRIMITIVE_TYPES[STRING_TYPE] = STRING_TYPE
}

/* Name, Context, Type */

// Name: see Aliases (at top)

type Gamma map[Name]Type // Variable? though is an Expr

//type Type Name // Type definition (cf. alias)
type TNamed Name // Represents types declared by the user -- structs, interfaces

var _ Type = TNamed("")
var _ Spec = TNamed("")

// Pre: t0, t are known types
// t0 <: t
func (t0 TNamed) Impls(ds []Decl, t base.Type) bool {
	if _, ok := t.(Type); !ok {
		panic("Expected FGR type, not " + reflect.TypeOf(t).String() +
			":\n\t" + t.String())
	}

	if _, ok := t.(TNamed); !ok {
		return false // test that t is not a TNamed or that t is a TPrimitive ?
	}

	t_fg := t.(TNamed)
	if isStructType(ds, t_fg) {
		return isStructType(ds, t0) && t0 == t_fg
	}

	gs := methods(ds, t_fg) // t is a t_I
	gs0 := methods(ds, t0)  // t0 may be any
	for k, g := range gs {
		g0, ok := gs0[k]
		if !ok || !g.EqExceptVars(g0) {
			return false
		}
	}
	return true
}

// t_I is a Spec, but not t_S -- this aspect is currently "dynamically typed"
// From Spec
func (t TNamed) GetSigs(ds []Decl) []Sig {
	if !isInterfaceType(ds, t) { // isStructType would be more efficient
		panic("Cannot use non-interface type as a Spec: " + t.String())
	}
	td := getTDecl(ds, t).(ITypeLit)
	var res []Sig
	for _, s := range td.specs {
		res = append(res, s.GetSigs(ds)...)
	}
	return res
}

func (t0 TNamed) Equals(t base.Type) bool {
	if _, ok := t.(Type); !ok {
		panic("Expected FGR type, not " + reflect.TypeOf(t).String() +
			":\n\t" + t.String())
	}
	if _, ok := t.(TNamed); !ok {
		return false
	}
	return t0 == t.(TNamed)
}

func (t0 TNamed) String() string {
	return string(t0)
}

/* Primitive types */

type TPrimitive struct {
	//name      string
	tag       Tag
	undefined bool
}

var _ Type = TPrimitive{}

func (t TPrimitive) isUndefined() bool { return t.undefined }

// Pre: t0.isUndefined() && !t.isUndefined()
func (t0 TPrimitive) fitsIn(t TPrimitive) bool {
	switch t0.tag {
	case INT8, INT32, INT64:
		return t0.tag <= t.tag && INT8 <= t.tag && t.tag <= INT64 // kind of hardcoded
	default:
		panic("TODO")
	}
}

func (t0 TPrimitive) Impls(ds []base.Decl, t base.Type) bool {
	if _, ok := t.(Type); !ok {
		panic("Expected FGR type, not " + reflect.TypeOf(t).String() +
			":\n\t" + t.String())
	}

	if t_P, ok := t.(TPrimitive); ok {
		//if t_P.isUndefined() {
		//	panic("This should not be possible")
		//}

		if t0.isUndefined() {
			return t0.fitsIn(t_P)
		} else {
			return t0 == t_P // or .Equals?
		}
	} else {
		// only true if t == Any
		return isInterfaceType(ds, t) && len(methods(ds, t)) == 0
	}
}

func (t0 TPrimitive) Equals(t base.Type) bool {
	if _, ok := t.(Type); !ok {
		panic("Expected FGR type, not " + reflect.TypeOf(t).String() +
			":\n\t" + t.String())
	}
	if _, ok := t.(TPrimitive); !ok {
		return false
	}
	return t0 == t.(TPrimitive)
}

func (t TPrimitive) String() string {
	return "TPrimitive{tag=" + strconv.Itoa(int(t.tag)) + ", undefined=" + strconv.FormatBool(t.undefined) + "}"
}


/* AST base intefaces: FGNode, Decl, TDecl, Spec, Expr */

// FGNode, Decl: see Aliases (at top)

type TDecl interface { // Rename TypeDecl
	Decl
	GetType() Type // In FG, GetType() == Type(GetName())
}

// A Sig or a Type (specifically a t_I -- bad t_S usage raises a run-time error, cf. Type.GetSigs)
type Spec interface {
	FGNode
	GetSigs(ds []Decl) []Sig
}

type FGExpr interface {
	base.Expr
	Subs(subs map[Variable]FGExpr) FGExpr

	// N.B. gamma should be treated immutably (and ds, of course)
	// (No typing rule modifies gamma, except the T-Func bootstrap)
	Typing(ds []Decl, gamma Gamma, allowStupid bool) Type

	// string is the type name of the "actually evaluated" expr (within the eval context)
	// CHECKME: resulting Exprs are not "parsed" from source, OK?
	Eval(ds []Decl) (FGExpr, string)

	//IsPanic() bool  // TODO "explicit" FG panic -- cf. underlying runtime panic
}

/* Helpers */

func isStructType(ds []Decl, t Type) bool {
	for _, v := range ds {
		if d, ok := v.(STypeLit); ok && d.t_S == t {
			return true
		}
	}
	return false
}

func isInterfaceType(ds []Decl, t Type) bool {
	for _, v := range ds {
		d, ok := v.(ITypeLit)
		if ok && d.t_I == t {
			return true
		}
	}
	return false
}

func isPrimitiveType(t Type) bool {
	_, ok := t.(TPrimitive)
	return ok
}
