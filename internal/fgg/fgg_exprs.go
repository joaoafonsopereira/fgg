/*
 * This file contains defs for "concrete" syntax w.r.t. exprs.
 * Base ("abstract") types, interfaces, etc. are in fg.go.
 */

package fgg

import (
	"fmt"
	"reflect"
	"strings"
)

var _ = fmt.Errorf
var _ = reflect.Append
var _ = strings.Compare

/* Public constructors */

func NewVariable(id Name) Variable                            { return Variable{id} }
func NewStructLit(u_S Type, es []FGGExpr) StructLit         { return StructLit{u_S, es} }
func NewSelect(e FGGExpr, f Name) Select                      { return Select{e, f} }
func NewCall(e FGGExpr, m Name, us []Type, es []FGGExpr) Call { return Call{e, m, us, es} }
func NewAssert(e FGGExpr, t Type) Assert                      { return Assert{e, t} }
func NewSprintf(format string, args []FGGExpr) Sprintf        { return Sprintf{format, args} }

/* Variable */

type Variable struct {
	name Name
}

var _ FGGExpr = Variable{}

func (x Variable) GetName() Name { return x.name }

func (x Variable) Subs(m map[Variable]FGGExpr) FGGExpr {
	res, ok := m[x]
	if !ok {
		panic("Unknown var: " + x.String())
	}
	return res
}

func (x Variable) TSubs(subs EtaOpen) FGGExpr {
	return x
}

func (x Variable) Eval(ds []Decl) (FGGExpr, string) {
	panic("Cannot evaluate free variable: " + x.name)
}

// TODO: refactor as Typing and StupidTyping? (clearer than bool param)
func (x Variable) Typing(ds []Decl, delta Delta, gamma Gamma,
	allowStupid bool) (Type, FGGExpr) {
	res, ok := gamma[x.name]
	if !ok {
		panic("Var not in env: " + x.String())
	}
	return res, x
}

// From base.Expr
func (x Variable) IsValue() bool {
	return false
}

func (x Variable) CanEval(ds []Decl) bool {
	return false
}

func (x Variable) String() string {
	return x.name
}

func (x Variable) ToGoString(ds []Decl) string {
	return x.name
}

/* StructLit */

type StructLit struct {
	u_S   Type
	elems []FGGExpr
}

var _ FGGExpr = StructLit{}

func (s StructLit) GetNamedType() TNamed { panic("GetNamedType kinda deprecated") }
func (s StructLit) GetType() Type { return s.u_S }
func (s StructLit) GetElems() []FGGExpr  { return s.elems }

func (s StructLit) Subs(subs map[Variable]FGGExpr) FGGExpr {
	es := make([]FGGExpr, len(s.elems))
	for i := 0; i < len(s.elems); i++ {
		es[i] = s.elems[i].Subs(subs)
	}
	return StructLit{s.u_S, es}
}

func (s StructLit) TSubs(subs EtaOpen) FGGExpr {
	es := make([]FGGExpr, len(s.elems))
	for i := 0; i < len(s.elems); i++ {
		es[i] = s.elems[i].TSubs(subs)
	}
	return StructLit{s.u_S.SubsEtaOpen(subs).(TNamed), es}
}

func (s StructLit) Eval(ds []Decl) (FGGExpr, string) {
	es := make([]FGGExpr, len(s.elems))
	done := false
	var rule string
	for i := 0; i < len(s.elems); i++ {
		v := s.elems[i]
		if !done && !v.IsValue() {
			v, rule = v.Eval(ds)
			done = true
		}
		es[i] = v
	}
	if done {
		return StructLit{s.u_S, es}, rule
	} else {
		panic("Cannot reduce: " + s.String())
	}
}

func (s StructLit) Typing(ds []Decl, delta Delta, gamma Gamma,
	allowStupid bool) (Type, FGGExpr) {
	s.u_S.Ok(ds, delta)
	if !isStructType(ds, s.u_S) {
		panic("Struct literal: " + s.u_S.String() + " is not a struct type")
	}
	fs := fields(ds, s.u_S)
	if len(s.elems) != len(fs) {
		var b strings.Builder
		b.WriteString("Arity mismatch: args=[")
		writeExprs(&b, s.elems)
		b.WriteString("], fields=[")
		writeFieldDecls(&b, fs)
		b.WriteString("]\n\t")
		b.WriteString(s.String())
		panic(b.String())
	}
	elems := make([]FGGExpr, len(s.elems))
	for i := 0; i < len(s.elems); i++ {
		u, newSubtree := s.elems[i].Typing(ds, delta, gamma, allowStupid)
		u_f := fs[i].u
		ok, coercion := u.AssignableToDelta(ds, delta, u_f)
		if !ok {
			panic("Arg expr must be assignable to field type: arg=" + u.String() +
				", field=" + u_f.String() + "\n\t" + s.String())
		}
		elems[i] = coercion(newSubtree)
	}
	return s.u_S, StructLit{s.u_S, elems}
}

// From base.Expr
func (s StructLit) IsValue() bool {
	for _, v := range s.elems {
		if !v.IsValue() {
			return false
		}
	}
	return true
}

func (s StructLit) CanEval(ds []Decl) bool {
	for _, v := range s.elems {
		if v.CanEval(ds) {
			return true
		} else if !v.IsValue() {
			return false
		}
	}
	return false
}

func (s StructLit) String() string {
	var b strings.Builder
	b.WriteString(s.u_S.String())
	b.WriteString("{")
	writeExprs(&b, s.elems)
	b.WriteString("}")
	return b.String()
}

func (s StructLit) ToGoString(ds []Decl) string {
	var b strings.Builder
	b.WriteString(s.u_S.ToGoString(ds))
	b.WriteString("{")
	under := s.u_S.Underlying(ds).(STypeLit)
	if len(s.elems) > 0 {
		b.WriteString(under.fDecls[0].field)
		b.WriteString(":")
		b.WriteString(s.elems[0].ToGoString(ds))
		for i, v := range s.elems[1:] {
			b.WriteString(", ")
			b.WriteString(under.fDecls[i+1].field)
			b.WriteString(":")
			b.WriteString(v.ToGoString(ds))
		}
	}
	b.WriteString("}")
	return b.String()
}

/* Select */

type Select struct {
	e_S   FGGExpr
	field Name
}

var _ FGGExpr = Select{}

func (s Select) GetExpr() FGGExpr { return s.e_S }
func (s Select) GetField() Name   { return s.field }

func (s Select) Subs(subs map[Variable]FGGExpr) FGGExpr {
	return Select{s.e_S.Subs(subs), s.field}
}

func (s Select) TSubs(subs EtaOpen) FGGExpr {
	return Select{s.e_S.TSubs(subs), s.field}
}

func (s Select) Eval(ds []Decl) (FGGExpr, string) {
	if !s.e_S.IsValue() {
		e, rule := s.e_S.Eval(ds)
		return Select{e, s.field}, rule
	}
	v := s.e_S.(StructLit)
	fds := fields(ds, v.u_S)
	for i := 0; i < len(fds); i++ {
		if fds[i].field == s.field {
			return v.elems[i], "Select"
		}
	}
	panic("Field not found: " + s.field + "\n\t" + s.String())
}

func (s Select) Typing(ds []Decl, delta Delta, gamma Gamma,
	allowStupid bool) (Type, FGGExpr) {
	u, e_S := s.e_S.Typing(ds, delta, gamma, allowStupid)
	if !IsStructType(ds, u) {
		panic("Illegal select on expr of non-struct type: " + u.String() +
			"\n\t" + s.String())
	}
	fds := fields(ds, u.(TNamed))
	for _, v := range fds {
		if v.field == s.field {
			return v.u, Select{e_S, s.field}
		}
	}
	panic("Field " + s.field + " not found in type: " + u.String() +
		"\n\t" + s.String())
}

// From base.Expr
func (s Select) IsValue() bool {
	return false
}

func (s Select) CanEval(ds []Decl) bool {
	if s.e_S.CanEval(ds) {
		return true
	} else if !s.e_S.IsValue() {
		return false
	}
	for _, v := range fields(ds, s.e_S.(StructLit).u_S) { // N.B. "purely operational", no typing aspect
		if v.field == s.field {
			return true
		}
	}
	return false
}

func (s Select) String() string {
	return s.e_S.String() + "." + s.field
}

func (s Select) ToGoString(ds []Decl) string {
	return s.e_S.ToGoString(ds) + "." + s.field
}

/* Call */

type Call struct {
	e_recv FGGExpr
	meth   Name
	t_args []Type // Rename u_args?
	args   []FGGExpr
}

var _ FGGExpr = Call{}

func (c Call) GetRecv() FGGExpr   { return c.e_recv } // Called GetReceiver in fg
func (c Call) GetMethod() Name    { return c.meth }
func (c Call) GetTArgs() []Type   { return c.t_args }
func (c Call) GetArgs() []FGGExpr { return c.args }

func (c Call) Subs(subs map[Variable]FGGExpr) FGGExpr {
	e := c.e_recv.Subs(subs)
	args := make([]FGGExpr, len(c.args))
	for i := 0; i < len(c.args); i++ {
		args[i] = c.args[i].Subs(subs)
	}
	return Call{e, c.meth, c.t_args, args}
}

func (c Call) TSubs(subs EtaOpen) FGGExpr {
	targs := make([]Type, len(c.t_args))
	for i := 0; i < len(c.t_args); i++ {
		targs[i] = c.t_args[i].SubsEtaOpen(subs)
	}
	args := make([]FGGExpr, len(c.args))
	for i := 0; i < len(c.args); i++ {
		args[i] = c.args[i].TSubs(subs)
	}
	return Call{c.e_recv.TSubs(subs), c.meth, targs, args}
}

func (c Call) Eval(ds []Decl) (FGGExpr, string) {
	if !c.e_recv.IsValue() {
		e, rule := c.e_recv.Eval(ds)
		return Call{e, c.meth, c.t_args, c.args}, rule
	}
	args := make([]FGGExpr, len(c.args))
	done := false
	var rule string
	for i := 0; i < len(c.args); i++ {
		e := c.args[i]
		if !done && !e.IsValue() {
			e, rule = e.Eval(ds)
			done = true
		}
		args[i] = e
	}
	if done {
		return Call{c.e_recv, c.meth, c.t_args, args}, rule
	}
	// c.e and c.args all values
	t := concreteType(c.e_recv).(TNamed)
	x0, xs, e := body(ds, t, c.meth, c.t_args) // panics if method not found

	subs := make(map[Variable]FGGExpr)
	subs[Variable{x0.name}] = c.e_recv
	for i := 0; i < len(xs); i++ {
		subs[Variable{xs[i].name}] = c.args[i]
	}
	return e.Subs(subs), "Call" // N.B. single combined substitution map slightly different to R-Call
}

func (c Call) Typing(ds []Decl, delta Delta, gamma Gamma, allowStupid bool) (Type, FGGExpr) {
	u0, e_recv := c.e_recv.Typing(ds, delta, gamma, allowStupid)
	var g Sig
	if tmp, ok := methodsDelta(ds, delta, bounds(delta, u0))[c.meth]; !ok { // !!! submission version had "methods(m)"
		panic("Method not found: " + c.meth + " in " + u0.String())
	} else {
		g = tmp
	}
	if len(c.t_args) != len(g.Psi.tFormals) {
		var b strings.Builder
		b.WriteString("Arity mismatch: type actuals=[")
		writeTypes(&b, c.t_args)
		b.WriteString("], formals=[")
		b.WriteString(g.Psi.String())
		b.WriteString("]\n\t")
		b.WriteString(c.String())
		panic(b.String())
	}
	if len(c.args) != len(g.pDecls) {
		var b strings.Builder
		b.WriteString("Arity mismatch: args=[")
		writeExprs(&b, c.args)
		b.WriteString("], params=[")
		writeParamDecls(&b, g.pDecls)
		b.WriteString("]\n\t")
		b.WriteString(c.String())
		panic(b.String())
	}
	// duplicates MakeEtaDelta
	eta := MakeEtaOpen(g.Psi, c.t_args) // CHECKME: applying this subs vs. adding to a new delta?  // Cf. MakeEta TODO CHECK THIS
	for i := 0; i < len(c.t_args); i++ {
		u := g.Psi.tFormals[i].u_I.SubsEtaOpen(eta)
		u_I := getInterface(ds, u)
		if !ImplsDelta(ds, delta, c.t_args[i], u_I) {
			panic("Type actual must implement type formal: actual=" +
				c.t_args[i].String() + ", param=" + u.String() + "\n\t" + c.String())
		}
	}
	args := make([]FGGExpr, len(c.args))
	for i := 0; i < len(c.args); i++ {
		// CHECKME: submission version's notation, (~\tau :> ~\rho)[subs], slightly unclear
		u_a, newSubtree := c.args[i].Typing(ds, delta, gamma, allowStupid)
		//.TSubs(subs)  // !!! submission version, subs also applied to ~tau, ..
		// ..falsely captures "repeat" var occurrences in recursive calls, ..
		// ..e.g., bad monomorph (Box) example.
		// The ~beta morally do not occur in ~tau, they only bind ~rho
		u_p := g.pDecls[i].u.SubsEtaOpen(eta)
		ok, coercion := u_a.AssignableToDelta(ds, delta, u_p)
		if !ok {
			panic("Arg expr must be assignable to param type: arg=" + u_a.String() +
				", param=" + u_p.String() + "\n\t" + c.String())
		}
		args[i] = coercion(newSubtree)
	}
	return g.u_ret.SubsEtaOpen(eta), // subs necessary, c.psi info (i.e., bounds) will be "lost" after leaving this context
	Call{e_recv, c.meth, c.t_args, args}
}

// From base.Expr
func (c Call) IsValue() bool {
	return false
}

func (c Call) CanEval(ds []Decl) bool {
	if c.e_recv.CanEval(ds) {
		return true
	} else if !c.e_recv.IsValue() {
		return false
	}
	for _, v := range c.args {
		if v.CanEval(ds) {
			return true
		} else if !v.IsValue() {
			return false
		}
	}
	u_S := concreteType(c.e_recv).(TNamed)
	md := getMethDecl(ds, u_S.t_name, c.meth)
	return len(md.Psi_recv.tFormals) == len(u_S.u_args) && // Needed, or also disregard?
		len(md.Psi_meth.tFormals) == len(c.t_args) &&
		len(md.pDecls) == len(c.args)
}

func (c Call) String() string {
	var b strings.Builder
	b.WriteString(c.e_recv.String())
	b.WriteString(".")
	b.WriteString(c.meth)
	b.WriteString("(")
	writeTypes(&b, c.t_args)
	b.WriteString(")(")
	writeExprs(&b, c.args)
	b.WriteString(")")
	return b.String()
}

func (c Call) ToGoString(ds []Decl) string {
	var b strings.Builder
	b.WriteString(c.e_recv.ToGoString(ds))
	b.WriteString(".")
	b.WriteString(c.meth)
	b.WriteString("(")
	writeToGoTypes(ds, &b, c.t_args)
	b.WriteString(")(")
	writeToGoExprs(ds, &b, c.args)
	b.WriteString(")")
	return b.String()
}

/* Assert */

type Assert struct {
	e_I    FGGExpr
	u_cast Type
}

var _ FGGExpr = Assert{}

func (a Assert) GetExpr() FGGExpr { return a.e_I }
func (a Assert) GetType() Type    { return a.u_cast }

func (a Assert) Subs(subs map[Variable]FGGExpr) FGGExpr {
	return Assert{a.e_I.Subs(subs), a.u_cast}
}

func (a Assert) TSubs(subs EtaOpen) FGGExpr {
	return Assert{a.e_I.TSubs(subs), a.u_cast.SubsEtaOpen(subs)}
}

func (a Assert) Eval(ds []Decl) (FGGExpr, string) {
	if !a.e_I.IsValue() {
		e, rule := a.e_I.Eval(ds)
		return Assert{e, a.u_cast}, rule
	}
	ok, _ := concreteType(a.e_I).AssignableToDelta(ds, Delta{}, a.u_cast) // Empty Delta -- not super clear in submission version
	if ok {
		return a.e_I, "Assert"
	}
	panic("Cannot reduce: " + a.String())
}

func (a Assert) Typing(ds []Decl, delta Delta, gamma Gamma, allowStupid bool) (Type, FGGExpr) {
	a.u_cast.Ok(ds, delta)
	u, e_I := a.e_I.Typing(ds, delta, gamma, allowStupid)
	newAst := Assert{e_I, a.u_cast}
	if !IsIfaceLikeType(ds, u) {
		if allowStupid {
			return a.u_cast, newAst
		} else {
			panic("Expr must be an interface-like type (in a non-stupid context): found " +
				u.String() + " for\n\t" + a.String())
		}
	}
	// u is a TParam or an interface type TName
	if IsIfaceLikeType(ds, a.u_cast) { // T-ASSERT_I
		return a.u_cast, newAst // No further checks -- N.B., Robert said they are looking to refine this
	} else { // T-ASSERT_S
		u_bound := bounds(delta, u)
		if ImplsDelta(ds, delta, a.u_cast, getInterface(ds, u_bound)) {
			return a.u_cast, newAst
		}
		panic("Struct type assertion must implement expr type: asserted=" +
			a.u_cast.String() + ", expr=" + u.String())
	}
}

// CHECKME: make isStuck alternative? (i.e., bad cast)
// From base.fgg
func (a Assert) IsValue() bool {
	return false
}

func (a Assert) CanEval(ds []Decl) bool {
	if a.e_I.CanEval(ds) {
		return true
	} else if !a.e_I.IsValue() {
		return false
	}
	ok, _ := concreteType(a.e_I).AssignableToDelta(ds, Delta{}, a.u_cast)
	return ok
}

func (a Assert) String() string {
	var b strings.Builder
	b.WriteString(a.e_I.String())
	b.WriteString(".(")
	b.WriteString(a.u_cast.String())
	b.WriteString(")")
	return b.String()
}

func (a Assert) ToGoString(ds []Decl) string {
	var b strings.Builder
	b.WriteString(a.e_I.ToGoString(ds))
	b.WriteString(".(")
	b.WriteString(a.u_cast.ToGoString(ds))
	b.WriteString(")")
	return b.String()
}

/* Type conversions */

// Simplified version of Go's type conversions.
// The main goal is to allow conversions such as:
//   - int32(1), MyInt(1), MyInt(int32(1))
//   - S(struct{}{}), struct{}(S{})
// Essentially, it only supports conversions between types with similar
// underlying types (Cf. validConversion).
// In particular, conversions such as float32(int32(1)) are not supported.
type Convert struct {
	typ  Type
	expr FGGExpr
}

var _ FGGExpr = Convert{}

func (c Convert) Subs(subs map[Variable]FGGExpr) FGGExpr {
	return Convert{c.typ, c.expr.Subs(subs)}
}

func (c Convert) TSubs(eta EtaOpen) FGGExpr {
	return Convert{c.typ.SubsEtaOpen(eta), c.expr.TSubs(eta)}
}

func (c Convert) Eval(ds []Decl) (FGGExpr, string) {
	if !c.expr.IsValue() {
		e, rule := c.expr.Eval(ds)
		return Convert{c.typ, e}, rule
	}

	var converted FGGExpr
	switch e := c.expr.(type) {
	case PrimitiveLiteral:
		ptype := c.typ.Underlying(ds).(PrimType)
		convdLit := rawConversion(e, ptype.Tag())
		if _, ok := c.typ.(UndefTPrimitive); ok {
			converted = convdLit
		} else {
			converted = TypedPrimitiveValue{convdLit, c.typ}
		}
	case TypedPrimitiveValue:
		converted = TypedPrimitiveValue{e.lit, c.typ}
	case StructLit:
		converted = StructLit{c.typ, e.elems}
	default:
		panic("Unsupported conversion: " + c.String())
	}
	return converted, "Convert"
}

// Performs the conversion falling back on actual Go conversions.
// N.B. For each Tag case, switches *only* on the types that
// lit.payload might have (in a correct conversion).
// Todo factor this with the exactly equal FG code
func rawConversion(lit PrimitiveLiteral, tag Tag) PrimitiveLiteral {
	var convdPayload interface{}
	switch tag {
	case BOOL, STRING:
		convdPayload = lit.payload // payload can only be bool/string

	case INT32:
		pload := lit.payload.(int32) // lit.payload can only be int32
		convdPayload = int32(pload)  //  redundant
	case INT64:
		switch pload := lit.payload.(type) {
		case int32:
			convdPayload = int64(pload)
		case int64:
			convdPayload = int64(pload) // redundant
		}
	case FLOAT32:
		switch pload := lit.payload.(type) {
		case int32:
			convdPayload = float32(pload)
		case int64:
			convdPayload = float32(pload)
		case float32:
			convdPayload = float32(pload) // redundant
		}
	case FLOAT64:
		switch pload := lit.payload.(type) {
		case int32:
			convdPayload = float64(pload)
		case int64:
			convdPayload = float64(pload)
		case float32:
			convdPayload = float64(pload)
		case float64:
			convdPayload = float64(pload) // redundant
		}
	}
	return PrimitiveLiteral{convdPayload, tag}
}

func (c Convert) Typing(ds []Decl, delta Delta, gamma Gamma, allowStupid bool) (Type, FGGExpr) {
	c.typ.Ok(ds, delta)
	u_expr, expr := c.expr.Typing(ds, delta, gamma, allowStupid)
	if validConversion(ds, delta, u_expr, c.typ) {
		return c.typ, Convert{c.typ, expr}
	}
	panic("Invalid type conversion: " + c.String())
}

func validConversion(ds []Decl, delta Delta, u1, u2 Type) bool {
	if u1.Underlying(ds).Equals(u2.Underlying(ds)) {
		return true
	}
	if u1, ok := u1.(UndefTPrimitive); ok {
		return u1.RepresentableBy(ds, delta, u2)
	}
	return false
}

func (c Convert) IsValue() bool {
	return false
}

func (c Convert) CanEval(ds []Decl) bool {
	t_expr := concreteType(c.expr)
	return validConversion(ds, Delta{}, t_expr, c.typ)
}

func (c Convert) String() string {
	var b strings.Builder
	b.WriteString(c.typ.String())
	b.WriteString("(")
	b.WriteString(c.expr.String())
	b.WriteString(")")
	return b.String()
}

func (c Convert) ToGoString(ds []Decl) string {
	var b strings.Builder
	b.WriteString("main.")
	b.WriteString(c.typ.String())
	b.WriteString("(")
	b.WriteString(c.expr.String())
	b.WriteString(")")
	return b.String()
}

/* fmt.Sprintf */

type Sprintf struct {
	format string // Includes surrounding quotes
	args   []FGGExpr
}

var _ FGGExpr = Sprintf{}

func (s Sprintf) GetFormat() string  { return s.format }
func (s Sprintf) GetArgs() []FGGExpr { return s.args }

func (s Sprintf) Subs(subs map[Variable]FGGExpr) FGGExpr {
	args := make([]FGGExpr, len(s.args))
	for i := 0; i < len(args); i++ {
		args[i] = s.args[i].Subs(subs)
	}
	return Sprintf{s.format, args}
}

func (s Sprintf) TSubs(subs EtaOpen) FGGExpr {
	return s
}

func (s Sprintf) Eval(ds []Decl) (FGGExpr, string) {
	args := make([]FGGExpr, len(s.args))
	done := false
	var rule string
	for i := 0; i < len(s.args); i++ {
		v := s.args[i]
		if !done && !v.IsValue() {
			v, rule = v.Eval(ds)
			done = true
		}
		args[i] = v
	}
	if done {
		return Sprintf{s.format, args}, rule
	} else {
		cast := make([]interface{}, len(args))
		for i := range args {
			cast[i] = args[i] // N.B. inside fgg this is, e.g., a StructLit (not the struct itself, as in native Go)
		}
		template := s.format[1 : len(s.format)-1] // Remove surrounding quote chars
		str := fmt.Sprintf(template, cast...)
		str = strings.ReplaceAll(str, "\"", "") // HACK because StringLit.String() includes quotes
		// FIXME: currently user remplates cannot include xplicit quote chars
		return NewStringLit(str), "Sprintf"
	}
}

// TODO: [Warning] not "fully" type checked, cf. MISSING/EXTRA
func (s Sprintf) Typing(ds []Decl, delta Delta, gamma Gamma, allowStupid bool) (Type, FGGExpr) {
	args := make([]FGGExpr, len(s.args))
	for i := 0; i < len(s.args); i++ {
		_, args[i] = s.args[i].Typing(ds, delta, gamma, allowStupid)
	}
	return NewTPrimitive(STRING), Sprintf{s.format, args}
}

// From base.Expr
func (s Sprintf) IsValue() bool {
	return false
}

func (s Sprintf) CanEval(ds []Decl) bool {
	return true
}

func (s Sprintf) String() string {
	var b strings.Builder
	b.WriteString("fmt.Sprintf(")
	b.WriteString(s.format)
	if len(s.args) > 0 {
		b.WriteString(", ")
		writeExprs(&b, s.args)
	}
	b.WriteString(")")
	return b.String()
}

func (s Sprintf) ToGoString(ds []Decl) string {
	var b strings.Builder
	b.WriteString("fmt.Sprintf(")
	b.WriteString(s.format)
	if len(s.args) > 0 {
		b.WriteString(", ")
		writeToGoExprs(ds, &b, s.args)
	}
	b.WriteString(")")
	return b.String()
}

/* Aux, helpers */

func writeExprs(b *strings.Builder, es []FGGExpr) {
	if len(es) > 0 {
		b.WriteString(es[0].String())
		for _, v := range es[1:] {
			b.WriteString(", " + v.String())
		}
	}
}

func writeToGoExprs(ds []Decl, b *strings.Builder, es []FGGExpr) {
	if len(es) > 0 {
		b.WriteString(es[0].ToGoString(ds))
		for _, v := range es[1:] {
			b.WriteString(", " + v.ToGoString(ds))
		}
	}
}
