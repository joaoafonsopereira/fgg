package fg

import "github.com/rhu1/fgg/internal/base"

/* Aliases from base */

type Name = base.Name
type FGNode = base.AstNode
type Decl = base.Decl

/* Name, Context, Type, Coercion */

// Name: see Aliases (at top)

type Gamma map[Name]Type // Variable? though is an Expr

type Type interface {
	base.Type
	Ok(ds []Decl)
	Underlying(ds []Decl) Type

	AssignableTo(ds []Decl, t Type) (bool, Coercion)
}

// Represents the coercion a compiler would insert
// after checking assignability. They are needed to ensure
// type safety holds after every small-step of evaluation.
// Cf. e.g. TestStructLit at fg_prims_test.go
// N.B. by making this a function, one can write a closure that captures a type
// inside an assignability test, which can later (outside AssignableTo)
// be applied to the desired FGExpr (Cf. e.g. Call.Typing)
type Coercion func(FGExpr) FGExpr

func noOpCoercion(expr FGExpr) FGExpr { return expr }

/* AST base interfaces: FGNode, Decl, Spec, Expr */

// FGNode, Decl: see Aliases (at top)

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
	Typing(ds []Decl, gamma Gamma, allowStupid bool) (Type, FGExpr)
	//Typing(ds []Decl, gamma Gamma, allowStupid bool) Type

	// string is the type name of the "actually evaluated" expr (within the eval context)
	// CHECKME: resulting Exprs are not "parsed" from source, OK?
	Eval(ds []Decl) (FGExpr, string)

	//IsPanic() bool  // TODO "explicit" FG panic -- cf. underlying runtime panic
}

/* Helpers */

func isStructType(ds []Decl, t Type) bool {
	_, ok := t.Underlying(ds).(STypeLit)
	return ok
}

func isInterfaceType(ds []Decl, t Type) bool {
	_, ok := t.Underlying(ds).(ITypeLit)
	return ok
}

func getInterface(ds []Decl, u Type) ITypeLit {
	return u.Underlying(ds).(ITypeLit)
}

func isPrimitiveType(t Type) bool {
	_, ok := t.(TPrimitive)
	return ok
}
